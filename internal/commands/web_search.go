package commands

import "fmt"

type webSearchCmd struct{}

func (c *webSearchCmd) Execute(args map[string]any, send SendFunc) (any, error) {
	if searchClient == nil {
		return nil, fmt.Errorf("search not configured — set [search] url in ~/.pigo/config.toml")
	}

	q, _ := args["q"].(string)
	if q == "" {
		return nil, fmt.Errorf("q (query) is required")
	}

	limit := 10
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	return searchClient.Search(q, limit)
}

func (c *webSearchCmd) Description() string {
	return "Search the web via SearXNG"
}

func init() {
	Register("web.search", &webSearchCmd{}, Info{})
}
