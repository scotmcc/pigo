package commands

import (
	"fmt"

	"github.com/scotmcc/pigo/internal/assets"
)

// soulGetCmd returns the soul content, or the welcome prompt if no soul exists.
type soulGetCmd struct{}

func (c *soulGetCmd) Execute(args map[string]any, send SendFunc) (any, error) {
	if vaultService == nil {
		return nil, fmt.Errorf("vault service not initialized")
	}

	result, err := vaultService.Read("soul")
	if err != nil {
		// No soul — return the welcome prompt so the AI can start the flow.
		return map[string]any{
			"exists":  false,
			"welcome": assets.WelcomePrompt,
		}, nil
	}

	return map[string]any{
		"exists":  true,
		"content": result.Body,
		"raw":     result.RawContent,
	}, nil
}

func (c *soulGetCmd) Description() string {
	return "Get the soul file, or the welcome prompt if none exists"
}

func init() {
	Register("soul.get", &soulGetCmd{}, Info{})
}
