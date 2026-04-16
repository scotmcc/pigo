package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/scotmcc/pigo/internal/config"
	"github.com/scotmcc/pigo/internal/db"
	"github.com/scotmcc/pigo/internal/detect"
	"github.com/spf13/cobra"
)

// doctorCheck is one row in the doctor report: what was checked, whether it
// passed, a short detail line, and an actionable fix (only shown on failure).
type doctorCheck struct {
	Name   string `json:"name"`
	OK     bool   `json:"ok"`
	Detail string `json:"detail,omitempty"`
	Fix    string `json:"fix,omitempty"`
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check pigo's dependencies and report their health",
	Long: `Runs all pigo dependency checks:
  - sqlite-vec extension loaded
  - Ollama daemon reachable
  - embedding model pulled
  - pi extension installed
  - Claude Code skill installed

Exits 0 if all checks pass, 1 otherwise — useful in scripts and CI.`,
	RunE: runDoctor,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

func runDoctor(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("find home directory: %w", err)
	}

	configPath := filepath.Join(home, ".pigo", "config.toml")
	cfg, err := config.Load(configPath)
	if err != nil {
		cfg = config.Default()
	}

	checks := collectDoctorChecks(home, cfg)

	if jsonFlag {
		return emitDoctorJSON(checks)
	}
	emitDoctorText(checks)

	if anyFailed(checks) {
		os.Exit(1)
	}
	return nil
}

// collectDoctorChecks runs every check and returns them in display order.
// Downstream checks (embedding model) are skipped with a note when their
// prerequisite fails — no point retrying a call that can't possibly succeed.
func collectDoctorChecks(home string, cfg config.Config) []doctorCheck {
	checks := []doctorCheck{checkSQLiteVec()}

	ollamaCheck, ollamaUp := checkOllama(cfg)
	checks = append(checks, ollamaCheck)

	if ollamaUp {
		checks = append(checks, checkEmbeddingModel(cfg))
	} else {
		checks = append(checks, doctorCheck{
			Name:   fmt.Sprintf("embedding model %q", cfg.Ollama.Model),
			OK:     false,
			Detail: "skipped — ollama not reachable",
		})
	}

	checks = append(checks, checkPiExtension(home))
	checks = append(checks, checkClaudeSkill(home))
	return checks
}

func checkSQLiteVec() doctorCheck {
	tmpPath := filepath.Join(os.TempDir(), fmt.Sprintf("pigo-doctor-%d.db", os.Getpid()))
	defer os.Remove(tmpPath)

	d, err := db.Open(tmpPath)
	if err != nil {
		return doctorCheck{
			Name:   "sqlite-vec extension",
			OK:     false,
			Detail: fmt.Sprintf("database open failed: %v", err),
		}
	}
	defer d.Close()

	var version string
	if err := d.Conn().QueryRow("SELECT vec_version()").Scan(&version); err != nil {
		return doctorCheck{
			Name:   "sqlite-vec extension",
			OK:     false,
			Detail: "vec_version() failed — extension not loaded",
			Fix:    "rebuild pigo or download a release binary",
		}
	}
	return doctorCheck{
		Name:   "sqlite-vec extension",
		OK:     true,
		Detail: fmt.Sprintf("loaded (%s)", version),
	}
}

// checkOllama reports whether the daemon is usable. The second return value
// tells the caller whether to bother running model checks.
func checkOllama(cfg config.Config) (doctorCheck, bool) {
	c := doctorCheck{Name: "ollama daemon"}

	if !detect.OllamaOnPath() {
		c.Detail = "ollama not installed"
		c.Fix = "run 'pigo install' to install ollama, or install manually from https://ollama.com/download"
		return c, false
	}
	if !detect.OllamaReachable(cfg.Ollama.Endpoint) {
		c.Detail = fmt.Sprintf("installed but not reachable at %s", cfg.Ollama.Endpoint)
		c.Fix = "start ollama with 'ollama serve' (or your service manager)"
		return c, false
	}

	c.OK = true
	c.Detail = fmt.Sprintf("reachable at %s", cfg.Ollama.Endpoint)
	return c, true
}

func checkEmbeddingModel(cfg config.Config) doctorCheck {
	c := doctorCheck{Name: fmt.Sprintf("embedding model %q", cfg.Ollama.Model)}

	has, err := detect.OllamaHasModel(cfg.Ollama.Endpoint, cfg.Ollama.Model)
	if err != nil {
		c.Detail = fmt.Sprintf("model list query failed: %v", err)
		return c
	}
	if !has {
		c.Detail = "not pulled"
		c.Fix = fmt.Sprintf("run: ollama pull %s", cfg.Ollama.Model)
		return c
	}

	c.OK = true
	c.Detail = "pulled"
	return c
}

func checkPiExtension(home string) doctorCheck {
	path := filepath.Join(home, ".pi", "extensions", "pigo.ts")
	c := doctorCheck{Name: "pi extension"}

	if _, err := os.Stat(path); err != nil {
		c.Detail = "not installed"
		c.Fix = "run 'pigo install'"
		return c
	}

	c.OK = true
	c.Detail = fmt.Sprintf("installed (%s)", path)
	return c
}

func checkClaudeSkill(home string) doctorCheck {
	path := filepath.Join(home, ".claude", "commands", "pigo.md")
	c := doctorCheck{Name: "claude code skill"}

	if _, err := os.Stat(path); err != nil {
		c.Detail = "not installed"
		c.Fix = "run 'pigo install'"
		return c
	}

	c.OK = true
	c.Detail = fmt.Sprintf("installed (%s)", path)
	return c
}

func emitDoctorText(checks []doctorCheck) {
	fmt.Println("pigo doctor:")
	fmt.Println()

	passed := 0
	for _, c := range checks {
		symbol := "✓"
		if !c.OK {
			symbol = "✗"
		} else {
			passed++
		}
		fmt.Printf("  [%s] %s — %s\n", symbol, c.Name, c.Detail)
		if !c.OK && c.Fix != "" {
			fmt.Printf("      → %s\n", c.Fix)
		}
	}

	fmt.Println()
	if passed == len(checks) {
		fmt.Printf("all %d checks passed.\n", len(checks))
		return
	}
	fmt.Printf("%d of %d checks passed — see fixes above.\n", passed, len(checks))
}

func emitDoctorJSON(checks []doctorCheck) error {
	result := struct {
		Checks []doctorCheck `json:"checks"`
		Passed int           `json:"passed"`
		Failed int           `json:"failed"`
	}{Checks: checks}
	for _, c := range checks {
		if c.OK {
			result.Passed++
		} else {
			result.Failed++
		}
	}
	if err := printJSON(result); err != nil {
		return err
	}
	if result.Failed > 0 {
		os.Exit(1)
	}
	return nil
}

func anyFailed(checks []doctorCheck) bool {
	for _, c := range checks {
		if !c.OK {
			return true
		}
	}
	return false
}
