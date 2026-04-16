package commands

import "fmt"

type vaultLinksCmd struct{}

func (c *vaultLinksCmd) Execute(args map[string]any, send SendFunc) (any, error) {
	if vaultService == nil {
		return nil, fmt.Errorf("vault service not initialized")
	}

	id, _ := args["id"].(string)
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}

	return vaultService.Links(id)
}

func (c *vaultLinksCmd) Description() string {
	return "Show all connections for a note (relates_to, links_to, backlinks)"
}

func init() {
	Register("vault.links", &vaultLinksCmd{}, Info{})
}
