package commands

import "fmt"

type vaultTagsCmd struct{}

func (c *vaultTagsCmd) Execute(args map[string]any, send SendFunc) (any, error) {
	if vaultService == nil {
		return nil, fmt.Errorf("vault service not initialized")
	}

	return vaultService.Tags()
}

func (c *vaultTagsCmd) Description() string {
	return "List all tags with note counts"
}

func init() {
	Register("vault.tags", &vaultTagsCmd{}, Info{})
}
