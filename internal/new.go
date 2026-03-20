package internal

import (
	"fmt"
	"os"
)

func CreateProject(name string, lang string, withGit bool) error {

	if err := os.Mkdir(name, 0755); err != nil {
		return err
	}

	if err := os.Chdir(name); err != nil {
		return err
	}

	if err := CreateGitignore(lang); err != nil {
		return err
	}

	switch lang {
	case "py":
		if err := SetupPython(); err != nil {
			return err
		}
	case "go":
		if err := SetupGo(""); err != nil {
			return err
		}
	}

	if withGit {
		if err := run("git", "init"); err != nil {
			return err
		}
		if err := run("git", "add", "."); err != nil {
			return err
		}
		if err := run("git", "commit", "-m", "init: "+name); err != nil {
			return err
		}
	}

	fmt.Println("Project", name, "ready.")
	return nil
}
