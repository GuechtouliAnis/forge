// Package internal contains the core logic for all Forge commands.
// Functions here are language-agnostic helpers called by cmd/ subcommands.
package project

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/GuechtouliAnis/forge/internal/lang"
)

// Clone clones a git repository and optionally sets up the development environment.
// lang can be "py" for Python, "go" for Go, or empty to skip environment setup.
// username is used for Go module paths, falls back to git config if empty.
func Clone(repo string, language string, username string) error {

	fmt.Println("Cloning", repo)
	if err := run("git", "clone", repo); err != nil {
		return err
	}

	// extract the repo name from the URL
	// e.g. git@github.com:user/myproject.git → myproject
	dir := filepath.Base(repo)
	if ext := filepath.Ext(dir); ext == ".git" {
		dir = dir[:len(dir)-len(ext)]
	}

	// move into the cloned directory for subsequent setup steps
	if err := os.Chdir(dir); err != nil {
		return err
	}

	// route to the correct environment setup based on lang flag
	switch language {
	case "py":
		return lang.SetupPython()
	case "go":
		return lang.SetupGo(username)
	}

	// no lang provided: clone only, no environment setup
	return nil
}
