package commands

import "fmt"

type vaultSearchCmd struct{}

func (c *vaultSearchCmd) Execute(args map[string]any, send SendFunc) (any, error) {
	if vaultService == nil {
		return nil, fmt.Errorf("vault service not initialized")
	}

	q, _ := args["q"].(string)
	if q == "" {
		return nil, fmt.Errorf("q (query) is required")
	}

	limit := 10
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	return vaultService.Search(q, limit)
}

func (c *vaultSearchCmd) Description() string {
	return "Search notes by title, tags, and content"
}

func init() {
	Register("vault.search", &vaultSearchCmd{}, Info{})
}
