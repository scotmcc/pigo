package commands

import "time"

// systemPingCmd is an async test command.
// It sends a few progress updates, then completes.
// Useful for verifying the pipe + GUID routing works.
type systemPingCmd struct{}

func (c *systemPingCmd) Execute(args map[string]any, send SendFunc) (any, error) {
	send("update", "pong 1/3", nil)
	time.Sleep(500 * time.Millisecond)

	send("update", "pong 2/3", nil)
	time.Sleep(500 * time.Millisecond)

	send("update", "pong 3/3", nil)
	time.Sleep(500 * time.Millisecond)

	return map[string]string{"status": "alive"}, nil
}

func (c *systemPingCmd) Description() string {
	return "Async test — sends three pongs with delays"
}

func init() {
	Register("system.ping", &systemPingCmd{}, Info{Async: true})
}
