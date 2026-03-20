// Package internal contains the core logic for all Forge commands.
// Functions here are language-agnostic helpers called by cmd/ subcommands.
package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Clone clones a git repository and optionally sets up the development environment.
// lang can be "py" for Python, "go" for Go, or empty to skip environment setup.
// username is used for Go module paths, falls back to git config if empty.
func Clone(repo string, lang string, username string) error {

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
	switch lang {
	case "py":
		return SetupPython()
	case "go":
		return SetupGo(username)
	}

	// no lang provided: clone only, no environment setup
	return nil
}

// run executes a shell command and pipes stdout/stderr directly to the terminal.
// It is used by all internal functions to run external tools like git, go, and python.
func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
