package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/GuechtouliAnis/forge/internal/repo" // Adjust import as needed
	"github.com/pelletier/go-toml/v2"
)

// Load reads .forge.toml from path and merges it over defaults.
// Missing file is not an error — defaults are returned as-is.
func Load(path string) (*Config, error) {
	cfg := defaults()
	targetName := ".forge.toml"

	// 1. FAST PATH: Attempt direct read first
	tomlPath := filepath.Join(path, targetName)
	data, err := os.ReadFile(tomlPath)

	// 2. SLOW PATH: If not found, resolve the actual filename
	if errors.Is(err, os.ErrNotExist) {
		actualName, resolveErr := repo.ResolveCaseInsensitive(path, targetName)
		if resolveErr != nil {
			return nil, resolveErr
		}

		// If still empty, it truly doesn't exist -> return defaults safely
		if actualName == "" {
			return cfg, nil
		}

		// Read using the discovered casing (e.g., ".Forge.toml")
		data, err = os.ReadFile(filepath.Join(path, actualName))
	}

	// Catch any other errors (like Permission Denied)
	if err != nil {
		return nil, err
	}

	// decode into pre-populated defaults — absent keys keep their default value
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
