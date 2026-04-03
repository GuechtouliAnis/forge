// Package internal contains the core logic for all Forge commands.
package project

import (
	_ "embed" // required for go:embed directives
	"fmt"
	"os"
)

// gitignore templates are embedded into the binary at compile time
// so the binary is self-contained and requires no external template files
//
//go:embed templates/python.gitignore
var pyGitignore string

//go:embed templates/go.gitignore
var goGitignore string

//go:embed templates/generic.gitignore
var genericGitignore string

// CreateGitignore writes a .gitignore file to the current directory.
// lang can be "py" for Python, "go" for Go, or empty for a generic gitignore.
// Templates are embedded at compile time from internal/templates/.
func CreateGitignore(lang string) error {
	var content string

	// select the appropriate template based on language
	switch lang {
	case "py":
		content = pyGitignore
	case "go":
		content = goGitignore
	default:
		// no lang provided: use the generic template
		content = genericGitignore
	}

	// 0644 = owner read/write, group and others read only
	if err := os.WriteFile(".gitignore", []byte(content), 0644); err != nil {
		return err
	}

	fmt.Println("Generated .gitignore")
	return nil
}
