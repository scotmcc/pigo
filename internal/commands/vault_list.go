package commands

import "fmt"

type vaultListCmd struct{}

func (c *vaultListCmd) Execute(args map[string]any, send SendFunc) (any, error) {
	if vaultService == nil {
		return nil, fmt.Errorf("vault service not initialized")
	}

	return vaultService.List()
}

func (c *vaultListCmd) Description() string {
	return "List all notes in the vault"
}

func init() {
	Register("vault.list", &vaultListCmd{}, Info{})
}
