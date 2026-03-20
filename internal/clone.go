package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func Clone(repo string, lang string, username string) error {

	fmt.Println("Cloning", repo)
	if err := run("git", "clone", repo); err != nil {
		return err
	}

	// extract repo name from the file path
	dir := filepath.Base(repo)
	if ext := filepath.Ext(dir); ext == ".git" {
		dir = dir[:len(dir)-len(ext)]
	}

	// cd into cloned repo
	if err := os.Chdir(dir); err != nil {
		return err
	}

	// route to the correct -l flag
	switch lang {
	case "py":
		return SetupPython()
	case "go":
		return SetupGo(username) // pass username down
	}

	return nil
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
