package config

import (
	"errors"
	"os"

	"github.com/BurntSushi/toml"
)

// Load reads config from path. If the file doesn't exist, returns defaults.
// If the file exists but is malformed, returns an error.
//
// The pattern: start with defaults, then overlay whatever the file provides.
// Missing fields in the TOML keep their default values automatically,
// because we decode into an already-populated struct.
func Load(path string) (Config, error) {
	cfg := Default()

	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}

	_, err = toml.DecodeFile(path, &cfg)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}
