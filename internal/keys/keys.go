// Package keys defines constants used across pigo.
// No magic strings — everything references this package.
package keys

// Build metadata — injected at compile time via ldflags.
// Example: go build -ldflags "-X github.com/scotmcc/pigo/internal/keys.Version=1.0.0"
//
// These are var (not const) so ldflags can override them.
// If not set, they fall back to "dev" values.
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

// Command names used in the command registry.
const (
	VaultRead   = "vault.read"
	VaultWrite  = "vault.write"
	VaultEdit   = "vault.edit"
	VaultSearch = "vault.search"
	VaultList   = "vault.list"

	FactsConsolidate = "facts.consolidate"
	FactsSearch      = "facts.search"
	FactsTopics      = "facts.topics"

	SystemMethods = "system.methods"
)

// Config keys matching TOML sections and fields.
const (
	ConfigVaultPath     = "vault.path"
	ConfigDBPath        = "db.path"
	ConfigOllamaURL     = "ollama.endpoint"
	ConfigOllamaModel   = "ollama.model"
	ConfigServerHost    = "server.host"
	ConfigServerPort    = "server.port"
	ConfigGitAutoCommit = "git.auto_commit"
	ConfigGitRemote     = "git.remote"
)
