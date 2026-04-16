package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove pigo integrations",
	Long: `Removes the pi extension and Claude Code skill that pigo install placed.
Does NOT remove ~/.pigo/ or your vault data — that's yours.`,
	RunE: runUninstall,
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}

func runUninstall(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("find home directory: %w", err)
	}

	var removed []string

	// Remove pi extensions.
	piExt := filepath.Join(home, ".pi", "extensions", "pigo.ts")
	if fileExists(piExt) {
		if err := os.Remove(piExt); err != nil {
			return fmt.Errorf("remove pi extension: %w", err)
		}
		removed = append(removed, piExt)
	}

	ollamaExt := filepath.Join(home, ".pi", "extensions", "ollama.js")
	if fileExists(ollamaExt) {
		if err := os.Remove(ollamaExt); err != nil {
			return fmt.Errorf("remove ollama extension: %w", err)
		}
		removed = append(removed, ollamaExt)
	}

	// Remove Claude Code skill.
	claudeSkill := filepath.Join(home, ".claude", "commands", "pigo.md")
	if fileExists(claudeSkill) {
		if err := os.Remove(claudeSkill); err != nil {
			return fmt.Errorf("remove claude skill: %w", err)
		}
		removed = append(removed, claudeSkill)
	}

	if len(removed) == 0 {
		fmt.Println("nothing to remove — no pigo integrations found")
	} else {
		fmt.Println("pigo uninstall complete:")
		for _, path := range removed {
			fmt.Printf("  removed %s\n", path)
		}
	}

	fmt.Println("\nnote: ~/.pigo/ and your vault data were not touched")
	return nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
