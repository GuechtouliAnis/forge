package env

import (
	"fmt"
	"os"
	"strings"
)

// InitEnv creates a .env file from .env.example and registers it in .gitignore.
// If .env.example does not exist, an empty file is created.
// If .gitignore does not exist, it is created with the target path.
func InitEnv(path string, updateGitignoreFile bool) error {
	// error if target already exists
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("ERROR: %s already exists — forge does not overwrite .env files for security reasons", path)
	}

	// copy from .env.example if it exists, otherwise create empty
	example, err := os.ReadFile(".env.example")
	if err == nil {
		if err := os.WriteFile(path, example, 0644); err != nil {
			return fmt.Errorf("ERROR: could not write %s: %w", path, err)
		}
		fmt.Printf("created %s from .env.example\n", path)
	} else {
		if err := os.WriteFile(path, []byte{}, 0644); err != nil {
			return fmt.Errorf("ERROR: could not write %s: %w", path, err)
		}
		fmt.Printf("created empty %s (.env.example not found)\n", path)
	}

	// update .gitignore
	if updateGitignoreFile {
		if err := updateGitignore(path); err != nil {
			return fmt.Errorf("ERROR: could not update .gitignore: %w", err)
		}
	} else {
		fmt.Printf("WARNING: skipping .gitignore update — this risks committing %s and leaking secrets\n", path)
	}

	return nil
}

// updateGitignore appends path to .gitignore if not already present.
// Creates .gitignore if it does not exist.
func updateGitignore(path string) error {
	data, err := os.ReadFile(".gitignore")
	if err == nil {
		// check if already present
		for _, line := range strings.Split(string(data), "\n") {
			if strings.TrimSpace(line) == path {
				fmt.Printf(".gitignore already contains %s\n", path)
				return nil
			}
		}

		// append
		f, err := os.OpenFile(".gitignore", os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}

		defer f.Close()

		if len(data) > 0 && data[len(data)-1] != '\n' {
			fmt.Fprintln(f)
		}

		fmt.Fprintln(f, path)
		fmt.Printf("added %s to .gitignore\n", path)

		return nil
	}

	// create .gitignore
	if err := os.WriteFile(".gitignore", []byte(path+"\n"), 0644); err != nil {
		return err
	}

	fmt.Printf("created .gitignore with %s\n", path)
	return nil
}
