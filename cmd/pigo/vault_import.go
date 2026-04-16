package main

import (
	"fmt"
	"strings"

	"github.com/scotmcc/pigo/internal/vault"
	"github.com/spf13/cobra"
)

var vaultImportCmd = &cobra.Command{
	Use:   "import [url]",
	Short: "Import a web page as a vault note",
	Long:  "Fetches a URL, converts HTML to markdown, and saves it as a note with the page title and source URL.",
	Args:  cobra.ExactArgs(1),
	RunE:  runVaultImport,
}

var importTags string

func init() {
	vaultImportCmd.Flags().StringVar(&importTags, "tags", "", "additional comma-separated tags")
	vaultCmd.AddCommand(vaultImportCmd)
}

func runVaultImport(cmd *cobra.Command, args []string) error {
	svc, cleanup, err := setupVault()
	if err != nil {
		return err
	}
	defer cleanup()

	var tags []string
	if importTags != "" {
		for _, t := range strings.Split(importTags, ",") {
			tags = append(tags, strings.TrimSpace(t))
		}
	}

	result, err := svc.Import(vault.ImportInput{
		URL:  args[0],
		Tags: tags,
	})
	if err != nil {
		return err
	}

	if jsonFlag {
		return printJSON(result)
	}

	fmt.Printf("imported: %s (%s)\n", result.ID, result.FilePath)
	return nil
}
