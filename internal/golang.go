package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// SetupGo initializes a Go module in the current directory and runs go mod tidy.
// If go.mod already exists, it skips initialization and only runs tidy.
// username is used for the module path, if empty, falls back to git config user.name.
func SetupGo(username string) error {
	fmt.Println("Setting up Go environment...")

	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		// get the current directory name to use as the module name
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		dir := filepath.Base(wd)

		// fall back to git config if username was not provided via flag
		if username == "" {
			out, err := exec.Command("git", "config", "--global", "user.name").Output()
			if err != nil {
				return err
			}
			username = strings.TrimSpace(string(out))
		}

		// initialize the module with the standard github.com path convention
		if err := run("go", "mod", "init", "github.com/"+username+"/"+dir); err != nil {
			return err
		}
	}

	// pull all dependencies declared in imports
	return run("go", "mod", "tidy")
}
