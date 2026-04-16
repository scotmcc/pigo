package main

import (
	"github.com/spf13/cobra"
)

// vaultCmd is the parent for all vault subcommands.
var vaultCmd = &cobra.Command{
	Use:   "vault",
	Short: "Manage the knowledge vault",
	Long:  "Read, write, edit, search, and list notes in the pigo vault.",
}

func init() {
	rootCmd.AddCommand(vaultCmd)
}
