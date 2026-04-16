// Package commands provides the command registry.
// Commands register themselves via init() — no manual wiring needed.
// The server and CLI both dispatch through this registry.
package commands

import (
	"fmt"
	"sort"
	"sync"
)

// SendFunc is a callback that commands use to send progress updates.
// Commands call this to emit status without knowing about transport.
//
//   send("update", "Processing chunk 3/7...", nil)
//   send("done", "Finished", resultData)
//
// For sync commands, the caller passes a no-op SendFunc.
// For async commands, it routes updates through the persistent pipe.
type SendFunc func(status, message string, data any) error

// Command is the interface every dispatchable command implements.
type Command interface {
	// Execute runs the command. The send callback is for progress updates.
	// Sync commands can ignore send — just return the result.
	// Async commands should call send("update", ...) during long work.
	Execute(args map[string]any, send SendFunc) (any, error)

	// Description returns a short human-readable description for system.methods.
	Description() string
}

// Info holds metadata about a registered command.
type Info struct {
	Async bool // if true, the server returns an immediate ack and streams updates via pipe
}

// entry is a command plus its metadata.
type entry struct {
	cmd  Command
	info Info
}

var (
	registry = make(map[string]entry)
	mu       sync.RWMutex
)

// Register adds a command to the registry. Called from init() in each command file.
// Pass Info{} for a default sync command.
func Register(name string, cmd Command, info Info) {
	mu.Lock()
	defer mu.Unlock()
	registry[name] = entry{cmd: cmd, info: info}
}

// Get retrieves a command by name. Returns nil if not found.
func Get(name string) Command {
	mu.RLock()
	defer mu.RUnlock()
	e, ok := registry[name]
	if !ok {
		return nil
	}
	return e.cmd
}

// GetInfo retrieves a command's metadata by name.
func GetInfo(name string) (Info, bool) {
	mu.RLock()
	defer mu.RUnlock()
	e, ok := registry[name]
	return e.info, ok
}

// IsAsync returns true if the named command is registered as async.
func IsAsync(name string) bool {
	info, ok := GetInfo(name)
	return ok && info.Async
}

// Dispatch looks up a command by name and executes it with the given SendFunc.
func Dispatch(name string, args map[string]any, send SendFunc) (any, error) {
	cmd := Get(name)
	if cmd == nil {
		return nil, fmt.Errorf("unknown command: %s", name)
	}
	return cmd.Execute(args, send)
}

// NoOpSend is a SendFunc that does nothing. Used for sync/CLI calls.
func NoOpSend(status, message string, data any) error { return nil }

// List returns all registered command names, sorted.
func List() []string {
	mu.RLock()
	defer mu.RUnlock()
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Methods returns a map of command name → description for system.methods.
func Methods() map[string]string {
	mu.RLock()
	defer mu.RUnlock()
	methods := make(map[string]string, len(registry))
	for name, e := range registry {
		methods[name] = e.cmd.Description()
	}
	return methods
}
