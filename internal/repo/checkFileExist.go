package repo

import (
	"os"
	"path/filepath"
	"strings"
)

// CheckFileExists checks if any file in the directory matches the given name case-insensitively.
func CheckFileExists(dir string, name string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}
	for _, entry := range entries {
		if strings.EqualFold(entry.Name(), name) {
			return true, nil
		}
	}
	return false, nil
}

// RemoveFileInsensitive removes a file from dir matching name case-insensitively.
// No-ops if the file doesn't exist.
func RemoveFileInsensitive(dir string, name string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if strings.EqualFold(entry.Name(), name) {
			return os.Remove(filepath.Join(dir, entry.Name()))
		}
	}
	return nil
}
