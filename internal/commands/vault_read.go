package commands

import "fmt"

type vaultReadCmd struct{}

func (c *vaultReadCmd) Execute(args map[string]any, send SendFunc) (any, error) {
	if vaultService == nil {
		return nil, fmt.Errorf("vault service not initialized")
	}

	id, _ := args["id"].(string)
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}

	return vaultService.Read(id)
}

func (c *vaultReadCmd) Description() string {
	return "Read a note by ID"
}

func init() {
	Register("vault.read", &vaultReadCmd{}, Info{})
}
