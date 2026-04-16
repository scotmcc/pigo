package main

import (
	"fmt"

	"github.com/scotmcc/pigo/internal/keys"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the pigo version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("pigo %s (%s, %s)\n", keys.Version, keys.Commit, keys.Date)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
