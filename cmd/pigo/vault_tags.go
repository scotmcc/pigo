package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var vaultTagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "List all tags with note counts",
	Long:  "Shows every tag used in the vault and how many notes have it. Sorted by count.",
	RunE:  runVaultTags,
}

func init() {
	vaultCmd.AddCommand(vaultTagsCmd)
}

func runVaultTags(cmd *cobra.Command, args []string) error {
	svc, cleanup, err := setupVault()
	if err != nil {
		return err
	}
	defer cleanup()

	tags, err := svc.Tags()
	if err != nil {
		return err
	}

	if jsonFlag {
		return printJSON(tags)
	}

	if len(tags) == 0 {
		fmt.Println("no tags in vault")
		return nil
	}

	for _, t := range tags {
		fmt.Printf("  %3d  %s\n", t.Count, t.Tag)
	}

	return nil
}
