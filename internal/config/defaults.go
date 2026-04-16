package config

import (
	"os"
	"path/filepath"
)

// pigoDir returns the default pigo data directory: ~/.pigo
func pigoDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".pigo")
	}
	return filepath.Join(home, ".pigo")
}

// Default returns a Config with sensible defaults.
// Every field has a value — no nil checks needed downstream.
func Default() Config {
	dir := pigoDir()
	return Config{
		Vault: VaultConfig{
			Path: filepath.Join(dir, "vault"),
		},
		DB: DBConfig{
			Path: filepath.Join(dir, "pigo.db"),
		},
		Ollama: OllamaConfig{
			Endpoint: "http://localhost:11434",
			Model:    "nomic-embed-text",
		},
		Server: ServerConfig{
			Host:     "127.0.0.1",
			Port:     14159,
			PipePort: 14160,
		},
		Git: GitConfig{
			AutoCommit: true,
		},
	}
}
