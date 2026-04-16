// Package config handles loading and accessing pigo configuration.
// Config is a plain struct — no global state, no singletons.
// The caller (main.go) loads it and passes it where needed.
package config

// Config is the top-level configuration for pigo.
type Config struct {
	Vault  VaultConfig  `toml:"vault"`
	DB     DBConfig     `toml:"db"`
	Ollama OllamaConfig `toml:"ollama"`
	Server ServerConfig `toml:"server"`
	Git    GitConfig    `toml:"git"`
}

// VaultConfig controls where markdown notes are stored.
type VaultConfig struct {
	Path string `toml:"path"`
}

// DBConfig controls the SQLite database location.
type DBConfig struct {
	Path string `toml:"path"`
}

// OllamaConfig controls the connection to the Ollama embedding service.
type OllamaConfig struct {
	Endpoint string `toml:"endpoint"`
	Model    string `toml:"model"`
}

// ServerConfig controls the pigo daemon listeners.
type ServerConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`      // HTTP API port
	PipePort int    `toml:"pipe_port"` // TCP pipe port for persistent connections
}

// GitConfig controls automatic git operations on the vault.
type GitConfig struct {
	AutoCommit bool   `toml:"auto_commit"`
	Remote     string `toml:"remote"`
}
