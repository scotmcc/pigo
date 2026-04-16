package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/scotmcc/pigo/internal/config"
	"github.com/scotmcc/pigo/internal/db"
	"github.com/scotmcc/pigo/internal/git"
	"github.com/scotmcc/pigo/internal/ollama"
	"github.com/scotmcc/pigo/internal/vault"
)

// setupVault creates a vault.Service by loading config and wiring dependencies.
// Auto-creates ~/.pigo/ and vault directory on first use — no manual setup required.
func setupVault() (*vault.Service, func(), error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, nil, err
	}

	// Auto-create vault directory if it doesn't exist.
	firstRun := false
	if _, err := os.Stat(cfg.Vault.Path); os.IsNotExist(err) {
		firstRun = true
	}
	if err := os.MkdirAll(cfg.Vault.Path, 0755); err != nil {
		return nil, nil, fmt.Errorf("create vault directory: %w", err)
	}

	database, err := db.Open(cfg.DB.Path)
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}

	if err := database.Migrate(); err != nil {
		database.Close()
		return nil, nil, fmt.Errorf("migrate database: %w", err)
	}

	repo, err := git.Open(cfg.Vault.Path)
	if err != nil {
		database.Close()
		return nil, nil, fmt.Errorf("open git repo: %w", err)
	}

	client := ollama.NewClient(cfg.Ollama.Endpoint, cfg.Ollama.Model)

	svc := vault.NewService(database, repo, client, cfg.Vault.Path)

	cleanup := func() {
		database.Close()
	}

	if firstRun {
		fmt.Fprintf(os.Stderr, "initialized pigo vault at %s\n", cfg.Vault.Path)
		fmt.Fprintf(os.Stderr, "run 'pigo install' to set up harness integrations\n\n")
	}

	return svc, cleanup, nil
}

// loadConfig reads config from the --config flag or default location.
func loadConfig() (config.Config, error) {
	path := cfgPath
	if path == "" {
		path = filepath.Join(config.Default().Vault.Path, "..", "config.toml")
	}
	return config.Load(path)
}
