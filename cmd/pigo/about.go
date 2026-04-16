package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/scotmcc/pigo/internal/config"
	"github.com/scotmcc/pigo/internal/keys"
	"github.com/spf13/cobra"
)

var aboutCmd = &cobra.Command{
	Use:   "about",
	Short: "Show pigo build info, config, and system status",
	Run:   runAbout,
}

func init() {
	rootCmd.AddCommand(aboutCmd)
}

func runAbout(cmd *cobra.Command, args []string) {
	home, _ := os.UserHomeDir()
	cfg := config.Default()

	fmt.Println("pigo — a resident AI knowledge server")
	fmt.Println()

	// Build info.
	fmt.Println("Build:")
	fmt.Printf("  version    %s\n", keys.Version)
	fmt.Printf("  commit     %s\n", keys.Commit)
	fmt.Printf("  built      %s\n", keys.Date)
	fmt.Printf("  go         %s\n", runtime.Version())
	fmt.Printf("  os/arch    %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Println()

	// Paths.
	fmt.Println("Paths:")
	fmt.Printf("  vault      %s  %s\n", cfg.Vault.Path, pathStatus(cfg.Vault.Path))
	fmt.Printf("  database   %s  %s\n", cfg.DB.Path, pathStatus(cfg.DB.Path))
	configPath := filepath.Join(home, ".pigo", "config.toml")
	fmt.Printf("  config     %s  %s\n", configPath, pathStatus(configPath))
	fmt.Println()

	// Integrations.
	fmt.Println("Integrations:")
	piExt := filepath.Join(home, ".pi", "extensions", "pigo.ts")
	fmt.Printf("  pi ext     %s\n", integrationStatus(piExt, filepath.Join(home, ".pi")))
	claudeSkill := filepath.Join(home, ".claude", "commands", "pigo.md")
	fmt.Printf("  claude     %s\n", integrationStatus(claudeSkill, filepath.Join(home, ".claude")))
	fmt.Println()

	// Services.
	fmt.Println("Services:")
	fmt.Printf("  ollama     %s  %s\n", cfg.Ollama.Endpoint, serviceStatus(cfg.Ollama.Endpoint+"/api/tags"))
	fmt.Printf("  server     %s:%d  %s\n", cfg.Server.Host, cfg.Server.Port, serviceStatus(fmt.Sprintf("http://%s:%d/health", cfg.Server.Host, cfg.Server.Port)))
	fmt.Printf("  pipe       %s:%d\n", cfg.Server.Host, cfg.Server.PipePort)

	// Vault stats.
	dbPath := cfg.DB.Path
	if fileExistsCheck(dbPath) {
		printVaultStats(dbPath)
	}
}

func pathStatus(path string) string {
	if info, err := os.Stat(path); err == nil {
		if info.IsDir() {
			return "(exists)"
		}
		return "(exists)"
	}
	return "(not created)"
}

func integrationStatus(filePath, parentDir string) string {
	if fileExistsCheck(filePath) {
		return "installed"
	}
	if dirExistsCheck(parentDir) {
		return "not installed (run: pigo install)"
	}
	return "not detected"
}

func serviceStatus(url string) string {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "(offline)"
	}
	resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return "(running)"
	}
	return fmt.Sprintf("(status %d)", resp.StatusCode)
}

func printVaultStats(dbPath string) {
	// Import would create a circular dependency, so we use a lightweight query.
	// Only print if we can open the DB.
	db, err := openDBForStats(dbPath)
	if err != nil {
		return
	}
	defer db.Close()

	var noteCount, chunkCount, factCount int
	db.QueryRow("SELECT COUNT(*) FROM notes").Scan(&noteCount)
	db.QueryRow("SELECT COUNT(*) FROM chunks").Scan(&chunkCount)
	db.QueryRow("SELECT COUNT(*) FROM facts").Scan(&factCount)

	fmt.Println()
	fmt.Println("Vault:")
	fmt.Printf("  notes      %d\n", noteCount)
	fmt.Printf("  chunks     %d\n", chunkCount)
	fmt.Printf("  facts      %d\n", factCount)
}

func fileExistsCheck(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExistsCheck(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
