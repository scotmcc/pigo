package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var vaultListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all notes in the vault",
	RunE:  runVaultList,
}

func init() {
	vaultCmd.AddCommand(vaultListCmd)
}

func runVaultList(cmd *cobra.Command, args []string) error {
	svc, cleanup, err := setupVault()
	if err != nil {
		return err
	}
	defer cleanup()

	items, err := svc.List()
	if err != nil {
		return err
	}

	if jsonFlag {
		return printJSON(items)
	}

	if len(items) == 0 {
		fmt.Println("vault is empty")
		return nil
	}

	for _, item := range items {
		tags := ""
		if len(item.Tags) > 0 {
			tags = fmt.Sprintf("  [%s]", strings.Join(item.Tags, ", "))
		}
		fmt.Printf("%s  %s%s\n", item.UpdatedAt, item.Title, tags)
	}

	return nil
}
