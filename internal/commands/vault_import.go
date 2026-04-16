package commands

import (
	"fmt"

	"github.com/scotmcc/pigo/internal/vault"
)

type vaultImportCmd struct{}

func (c *vaultImportCmd) Execute(args map[string]any, send SendFunc) (any, error) {
	if vaultService == nil {
		return nil, fmt.Errorf("vault service not initialized")
	}

	url, _ := args["url"].(string)
	if url == "" {
		return nil, fmt.Errorf("url is required")
	}

	var tags []string
	if rawTags, ok := args["tags"].([]any); ok {
		for _, t := range rawTags {
			if s, ok := t.(string); ok {
				tags = append(tags, s)
			}
		}
	}

	send("update", "Fetching "+url+"...", nil)

	result, err := vaultService.Import(vault.ImportInput{
		URL:  url,
		Tags: tags,
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *vaultImportCmd) Description() string {
	return "Import a web page as a vault note"
}

func init() {
	Register("vault.import", &vaultImportCmd{}, Info{Async: true})
}
