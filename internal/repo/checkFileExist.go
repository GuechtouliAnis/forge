package repo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CheckFileExists checks if a file exists in the given directory using a
// case-insensitive match. It attempts a direct O(1) stat first for speed,
// falling back to an O(N) directory scan on case-sensitive systems (Linux).
func CheckFileExists(dir string, name string) (bool, error) {
	path := filepath.Join(dir, name)
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return false, fmt.Errorf("[repo]: failed to read directory %s: %w", dir, err)
		}
		for _, entry := range entries {
			if strings.EqualFold(entry.Name(), name) {
				return true, nil
			}
		}
		return false, nil
	}
	return false, fmt.Errorf("[repo]: failed to stat %s: %w", path, err)
}

// RemoveFileInsensitive attempts to delete a file matching name case-insensitively.
// It prioritizes a direct removal (O(1)) and only scans the directory if the
// exact match is missing. No-ops if no version of the file exists.
func RemoveFileInsensitive(dir string, name string) error {
	path := filepath.Join(dir, name)
	err := os.Remove(path)
	if err == nil {
		return nil
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("[repo]: failed to remove %s: %w", path, err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("[repo]: failed to read directory %s: %w", dir, err)
	}
	for _, entry := range entries {
		if strings.EqualFold(entry.Name(), name) {
			if err := os.Remove(filepath.Join(dir, entry.Name())); err != nil {
				return fmt.Errorf("[repo]: failed to remove %s: %w", entry.Name(), err)
			}
			return nil
		}
	}
	return nil
}

// ResolveCaseInsensitive searches for a file in the given directory that matches name
// case-insensitively. It prioritizes a direct match for performance (O(1)) and falls
// back to a directory scan (O(N)) if necessary.
func ResolveCaseInsensitive(dir string, name string) (string, error) {
	if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
		return name, nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("[repo]: failed to read directory %s: %w", dir, err)
	}
	for _, entry := range entries {
		if strings.EqualFold(entry.Name(), name) {
			return entry.Name(), nil
		}
	}
	return "", nil
}
