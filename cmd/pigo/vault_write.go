package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/scotmcc/pigo/internal/vault"
	"github.com/spf13/cobra"
)

var vaultWriteCmd = &cobra.Command{
	Use:   "write",
	Short: "Create a new note",
	Long:  "Create a new note in the vault. Provide title, tags, and body.",
	RunE:  runVaultWrite,
}

var (
	writeTitle string
	writeTags  string
	writeBody  string
	writeStdin bool
)

func init() {
	vaultWriteCmd.Flags().StringVar(&writeTitle, "title", "", "note title (required)")
	vaultWriteCmd.Flags().StringVar(&writeTags, "tags", "", "comma-separated tags")
	vaultWriteCmd.Flags().StringVar(&writeBody, "body", "", "note body (or use --stdin)")
	vaultWriteCmd.Flags().BoolVar(&writeStdin, "stdin", false, "read body from stdin")
	vaultWriteCmd.MarkFlagRequired("title")
	vaultCmd.AddCommand(vaultWriteCmd)
}

func runVaultWrite(cmd *cobra.Command, args []string) error {
	svc, cleanup, err := setupVault()
	if err != nil {
		return err
	}
	defer cleanup()

	body := writeBody
	if writeStdin {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}
		body = string(data)
	}

	var tags []string
	if writeTags != "" {
		for _, t := range strings.Split(writeTags, ",") {
			tags = append(tags, strings.TrimSpace(t))
		}
	}

	result, err := svc.Write(vault.WriteInput{
		Title: writeTitle,
		Tags:  tags,
		Body:  body,
	})
	if err != nil {
		return err
	}

	if jsonFlag {
		return printJSON(result)
	}

	fmt.Printf("created: %s (%s)\n", result.ID, result.FilePath)
	return nil
}
