package main

import (
	"fmt"
	"strings"

	"github.com/scotmcc/pigo/internal/vault"
	"github.com/spf13/cobra"
)

var vaultEditCmd = &cobra.Command{
	Use:   "edit [id]",
	Short: "Edit an existing note",
	Args:  cobra.ExactArgs(1),
	RunE:  runVaultEdit,
}

var (
	editBody string
	editTags string
)

func init() {
	vaultEditCmd.Flags().StringVar(&editBody, "body", "", "new body content")
	vaultEditCmd.Flags().StringVar(&editTags, "tags", "", "new comma-separated tags")
	vaultCmd.AddCommand(vaultEditCmd)
}

func runVaultEdit(cmd *cobra.Command, args []string) error {
	svc, cleanup, err := setupVault()
	if err != nil {
		return err
	}
	defer cleanup()

	input := vault.EditInput{ID: args[0]}

	if cmd.Flags().Changed("body") {
		input.Body = &editBody
	}

	if cmd.Flags().Changed("tags") {
		var tags []string
		for _, t := range strings.Split(editTags, ",") {
			tags = append(tags, strings.TrimSpace(t))
		}
		input.Tags = tags
	}

	if err := svc.Edit(input); err != nil {
		return err
	}

	if jsonFlag {
		return printJSON(map[string]string{"status": "updated", "id": args[0]})
	}

	fmt.Printf("updated: %s\n", args[0])
	return nil
}
