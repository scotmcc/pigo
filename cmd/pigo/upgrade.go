package main

import (
	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade pigo integrations to the latest version",
	Long:  "Re-installs the pi extension and Claude Code skill from the current binary. Same as 'pigo install'.",
	RunE:  runInstall, // upgrade is just install — it's idempotent
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
