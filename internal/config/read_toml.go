package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// Load reads .forge.toml from path and merges it over defaults.
// Missing file is not an error — defaults are returned as-is.
func Load(path string) (*Config, error) {
	cfg := defaults()

	tomlPath := filepath.Join(path, ".forge.toml")
	data, err := os.ReadFile(tomlPath)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return nil, err
	}

	// decode into pre-populated defaults — absent keys keep their default value
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
