package main

import (
	"github.com/spf13/cobra"
)

// rootCmd is the base command. Running "pigo" with no subcommand prints help.
var rootCmd = &cobra.Command{
	Use:   "pigo",
	Short: "A resident AI knowledge server",
	Long:  "pigo is an always-on knowledge system. Markdown vault, semantic search, git versioning, fact extraction.",
}

// Global flags available to all subcommands.
var (
	cfgPath  string
	jsonFlag bool
)

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgPath, "config", "", "config file (default ~/.pigo/config.toml)")
	rootCmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "output as JSON (for scripting and extensions)")
}
