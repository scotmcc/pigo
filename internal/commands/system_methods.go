package commands

// systemMethodsCmd implements the system.methods command.
// It returns a list of all registered commands with descriptions.
// This makes the gateway AI-discoverable — no out-of-band docs needed.
type systemMethodsCmd struct{}

func (c *systemMethodsCmd) Execute(args map[string]any, send SendFunc) (any, error) {
	return Methods(), nil
}

func (c *systemMethodsCmd) Description() string {
	return "List all available commands with descriptions"
}

func init() {
	Register("system.methods", &systemMethodsCmd{}, Info{})
}
