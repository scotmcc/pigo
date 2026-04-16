package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var vaultLinksCmd = &cobra.Command{
	Use:   "links [id]",
	Short: "Show all connections for a note",
	Long:  "Shows relates_to (auto-discovered), links_to ([[wiki-links]]), and backlinks (who links here).",
	Args:  cobra.ExactArgs(1),
	RunE:  runVaultLinks,
}

func init() {
	vaultCmd.AddCommand(vaultLinksCmd)
}

func runVaultLinks(cmd *cobra.Command, args []string) error {
	svc, cleanup, err := setupVault()
	if err != nil {
		return err
	}
	defer cleanup()

	info, err := svc.Links(args[0])
	if err != nil {
		return err
	}

	if jsonFlag {
		return printJSON(info)
	}

	fmt.Printf("Links for: %s\n\n", info.NoteID)

	if len(info.RelatesTo) > 0 {
		fmt.Printf("  Relates to:  %s\n", strings.Join(info.RelatesTo, ", "))
	} else {
		fmt.Println("  Relates to:  (none)")
	}

	if len(info.LinksTo) > 0 {
		fmt.Printf("  Links to:    %s\n", strings.Join(info.LinksTo, ", "))
	} else {
		fmt.Println("  Links to:    (none)")
	}

	if len(info.Backlinks) > 0 {
		fmt.Printf("  Backlinks:   %s\n", strings.Join(info.Backlinks, ", "))
	} else {
		fmt.Println("  Backlinks:   (none)")
	}

	return nil
}
