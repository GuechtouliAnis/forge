package internal

import (
	_ "embed"
	"fmt"
	"os"
)

//go:embed templates/python.gitignore
var pyGitignore string

//go:embed templates/go.gitignore
var goGitignore string

//go:embed templates/generic.gitignore
var genericGitignore string

func CreateGitignore(lang string) error {
	var content string

	switch lang {
	case "py":
		content = pyGitignore
	case "go":
		content = goGitignore
	default:
		content = genericGitignore
	}

	if err := os.WriteFile(".gitignore", []byte(content), 0644); err != nil {
		return err
	}

	fmt.Println("Generated .gitignore")
	return nil
}
