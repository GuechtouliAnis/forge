package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func SetupGo(username string) error {
	fmt.Println("Setting up Go environment...")

	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		// get current dir name for the module path
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		dir := filepath.Base(wd)

		// only fetch from git config if username wasn't passed in
		if username == "" {
			out, err := exec.Command("git", "config", "--global", "user.name").Output()
			if err != nil {
				return err
			}
			username = strings.TrimSpace(string(out))
		}

		if err := run("go", "mod", "init", "github.com/"+username+"/"+dir); err != nil {
			return err
		}
	}

	return run("go", "mod", "tidy")
}
