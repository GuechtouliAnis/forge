// Package internal contains the core logic for all Forge commands.
package repo

import (
	_ "embed" // required for go:embed directives
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// gitignore templates are embedded into the binary at compile time
// so the binary is self-contained and requires no external template files
//
//go:embed templates/gitignore/python.gitignore
var pyGitignore string

//go:embed templates/gitignore/go.gitignore
var goGitignore string

//go:embed templates/gitignore/generic.gitignore
var genericGitignore string

// CreateGitignore writes a .gitignore file to the given path.
// If path is empty, it defaults to the current directory.
// lang can be "py" or "python" for Python, "go" or "golang" for Go, or empty for a generic gitignore.
// Templates are embedded at compile time from internal/repo/templates/gitignore/.
func CreateGitignore(lang string, path string) error {
	if path == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		path = cwd
	}

	gitignorePath := filepath.Join(path, ".gitignore")

	// check if .gitignore exists at the target path
	if _, err := os.Stat(gitignorePath); err == nil {
		fmt.Print(".gitignore already exists. Overwrite? [y/N]: ")
		var input string
		fmt.Scanln(&input)
		if input != "y" && input != "Y" && strings.ToLower(input) != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	var content string

	// select the appropriate template based on language
	switch lang {
	case "py", "python":
		content = pyGitignore
	case "go", "golang":
		content = goGitignore
	case "":
		content = genericGitignore
	default:
		return fmt.Errorf("unsupported language %q — supported: py, python, go, golang", lang)
	}

	// 0644 = owner read/write, group and others read only
	if err := os.WriteFile(gitignorePath, []byte(content), 0644); err != nil {
		return err
	}

	fmt.Println("Generated .gitignore")
	return nil
}
