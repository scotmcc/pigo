package commands

import (
	"fmt"

	"github.com/scotmcc/pigo/internal/vault"
)

type vaultEditCmd struct{}

func (c *vaultEditCmd) Execute(args map[string]any, send SendFunc) (any, error) {
	if vaultService == nil {
		return nil, fmt.Errorf("vault service not initialized")
	}

	id, _ := args["id"].(string)
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}

	input := vault.EditInput{ID: id}

	if body, ok := args["body"].(string); ok {
		input.Body = &body
	}

	if rawTags, ok := args["tags"].([]any); ok {
		var tags []string
		for _, t := range rawTags {
			if s, ok := t.(string); ok {
				tags = append(tags, s)
			}
		}
		input.Tags = tags
	}

	if err := vaultService.Edit(input); err != nil {
		return nil, err
	}

	return map[string]string{"status": "updated", "id": id}, nil
}

func (c *vaultEditCmd) Description() string {
	return "Edit an existing note"
}

func init() {
	Register("vault.edit", &vaultEditCmd{}, Info{})
}
