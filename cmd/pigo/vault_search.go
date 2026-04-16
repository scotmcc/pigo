package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var vaultSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search the vault",
	Long:  "Search notes by title, tags, and content. Uses fuzzy matching and semantic similarity.",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runVaultSearch,
}

var searchLimit int

func init() {
	vaultSearchCmd.Flags().IntVar(&searchLimit, "limit", 10, "max results")
	vaultCmd.AddCommand(vaultSearchCmd)
}

func runVaultSearch(cmd *cobra.Command, args []string) error {
	svc, cleanup, err := setupVault()
	if err != nil {
		return err
	}
	defer cleanup()

	query := strings.Join(args, " ")
	result, err := svc.Search(query, searchLimit)
	if err != nil {
		return err
	}

	if jsonFlag {
		return printJSON(result)
	}

	for _, w := range result.Warnings {
		fmt.Fprintf(os.Stderr, "warning: %s\n", w)
	}

	if len(result.Results) == 0 {
		fmt.Println("no results")
		return nil
	}

	for _, r := range result.Results {
		heading := ""
		if r.Heading != "" {
			heading = fmt.Sprintf(" > %s", r.Heading)
		}
		fmt.Printf("%.2f  %s%s\n", r.Score, r.Title, heading)
	}

	return nil
}
