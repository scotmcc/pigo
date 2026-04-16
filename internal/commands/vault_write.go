package commands

import (
	"fmt"

	"github.com/scotmcc/pigo/internal/vault"
)

type vaultWriteCmd struct{}

func (c *vaultWriteCmd) Execute(args map[string]any, send SendFunc) (any, error) {
	if vaultService == nil {
		return nil, fmt.Errorf("vault service not initialized")
	}

	title, _ := args["title"].(string)
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}

	body, _ := args["body"].(string)

	var tags []string
	if rawTags, ok := args["tags"].([]any); ok {
		for _, t := range rawTags {
			if s, ok := t.(string); ok {
				tags = append(tags, s)
			}
		}
	}

	return vaultService.Write(vault.WriteInput{
		Title: title,
		Tags:  tags,
		Body:  body,
	})
}

func (c *vaultWriteCmd) Description() string {
	return "Create a new note"
}

func init() {
	Register("vault.write", &vaultWriteCmd{}, Info{})
}
