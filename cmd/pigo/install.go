package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/scotmcc/pigo/internal/assets"
	"github.com/scotmcc/pigo/internal/config"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Set up pigo and install integrations",
	Long: `Detects available AI harnesses and installs the appropriate integration.

Creates ~/.pigo/ (vault + config) if it doesn't exist.
If pi harness is found (~/.pi/extensions/), installs the pi extension.
If Claude Code is found (~/.claude/), installs the Claude Code skill.

Safe to re-run — overwrites integrations with the latest version.`,
	RunE: runInstall,
}

func init() {
	rootCmd.AddCommand(installCmd)
}

func runInstall(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("find home directory: %w", err)
	}

	var installed []string

	// 1. Create ~/.pigo/ structure.
	pigoDir := filepath.Join(home, ".pigo")
	vaultDir := config.Default().Vault.Path
	if err := os.MkdirAll(vaultDir, 0755); err != nil {
		return fmt.Errorf("create vault directory: %w", err)
	}
	installed = append(installed, fmt.Sprintf("created %s", pigoDir))

	// Write default config if it doesn't exist.
	configPath := filepath.Join(pigoDir, "config.toml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := writeDefaultConfig(configPath); err != nil {
			return err
		}
		installed = append(installed, "wrote default config.toml")
	}

	// 2. Detect pi harness.
	piExtDir := filepath.Join(home, ".pi", "extensions")
	if dirExists(piExtDir) {
		dest := filepath.Join(piExtDir, "pigo.ts")
		if err := os.WriteFile(dest, []byte(assets.PiExtension), 0644); err != nil {
			return fmt.Errorf("write pi extension: %w", err)
		}
		installed = append(installed, fmt.Sprintf("installed pi extension → %s", dest))

		ollamaDest := filepath.Join(piExtDir, "ollama.js")
		if err := os.WriteFile(ollamaDest, []byte(assets.OllamaExtension), 0644); err != nil {
			return fmt.Errorf("write ollama extension: %w", err)
		}
		installed = append(installed, fmt.Sprintf("installed ollama provider → %s", ollamaDest))
	}

	// 3. Detect Claude Code.
	claudeDir := filepath.Join(home, ".claude")
	if dirExists(claudeDir) {
		skillDir := filepath.Join(claudeDir, "commands")
		if err := os.MkdirAll(skillDir, 0755); err != nil {
			return fmt.Errorf("create claude commands dir: %w", err)
		}
		dest := filepath.Join(skillDir, "pigo.md")
		if err := os.WriteFile(dest, []byte(assets.ClaudeSkill), 0644); err != nil {
			return fmt.Errorf("write claude skill: %w", err)
		}
		installed = append(installed, fmt.Sprintf("installed claude skill → %s", dest))
	}

	// Report.
	fmt.Println("pigo install complete:")
	for _, line := range installed {
		fmt.Printf("  %s\n", line)
	}

	if !dirExists(piExtDir) && !dirExists(filepath.Join(home, ".claude")) {
		fmt.Println("\n  no pi harness or claude code detected — vault is ready, integrations skipped")
		fmt.Println("  re-run 'pigo install' after installing a harness")
	}

	return nil
}

func writeDefaultConfig(path string) error {
	content := `# pigo configuration
# See docs/ARCHITECTURE.md for details.

[vault]
# path = "~/.pigo/vault"

[db]
# path = "~/.pigo/pigo.db"

[ollama]
# endpoint = "http://localhost:11434"
# model = "nomic-embed-text"

[server]
# host = "127.0.0.1"
# port = 9876
# pipe_port = 9877

[git]
# auto_commit = true
# remote = ""

[search]
# url = "https://searxng.example.com"  # SearXNG instance for web search
`
	return os.WriteFile(path, []byte(content), 0644)
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
