package main

import (
	"github.com/spf13/cobra"
)

// webCmd is the parent for web-related subcommands.
var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Web search and fetch",
	Long:  "Search the web via SearXNG and fetch pages.",
}

func init() {
	rootCmd.AddCommand(webCmd)
}
