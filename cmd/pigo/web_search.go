package main

import (
	"fmt"
	"strings"

	"github.com/scotmcc/pigo/internal/search"
	"github.com/spf13/cobra"
)

var webSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search the web via SearXNG",
	Long:  "Queries your configured SearXNG instance and returns results. Configure with [search] url in config.toml.",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runWebSearch,
}

var webSearchLimit int

func init() {
	webSearchCmd.Flags().IntVar(&webSearchLimit, "limit", 10, "max results")
	webCmd.AddCommand(webSearchCmd)
}

func runWebSearch(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	client := search.NewClient(cfg.Search.URL)
	query := strings.Join(args, " ")

	results, err := client.Search(query, webSearchLimit)
	if err != nil {
		return err
	}

	if jsonFlag {
		return printJSON(results)
	}

	if len(results) == 0 {
		fmt.Println("no results")
		return nil
	}

	for i, r := range results {
		fmt.Printf("%d. %s\n", i+1, r.Title)
		fmt.Printf("   %s\n", r.URL)
		if r.Content != "" {
			// Trim snippet to one line.
			snippet := strings.ReplaceAll(r.Content, "\n", " ")
			if len(snippet) > 120 {
				snippet = snippet[:120] + "..."
			}
			fmt.Printf("   %s\n", snippet)
		}
		fmt.Println()
	}

	return nil
}
