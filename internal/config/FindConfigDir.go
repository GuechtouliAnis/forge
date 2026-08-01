package config

import (
	"errors"
	"os"
	"path/filepath"
)

// findConfigDir walks upward from startDir looking for .forge.toml,
// stopping at the first directory where it's found, the first directory
// containing .git (repo root — no config found), or the filesystem root
// (nothing found at all). Returns the directory to search and whether
// .forge.toml was actually found there, for Load's existing fast/slow-path
// read logic to use.
func findConfigDir(startDir string) (string, bool, error) {
	dir := startDir
	for {
		if _, err := os.Stat(filepath.Join(dir, ".forge.toml")); err == nil {
			return dir, true, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", false, err
		}

		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, false, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", false, err
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root with nothing found.
			return dir, false, nil
		}
		dir = parent
	}
}
