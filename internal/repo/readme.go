package repo

import (
	_ "embed" // required for go:embed directives
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed templates/readme/template.readme
var readmeFile string

// CreateReadme writes a README.md to the given path.
// If path is empty, it defaults to the current directory.
// Project name is inferred from the target directory name.
// Author is read from git config user.name, falls back to a prompt if not set.
func CreateReadme(path string) error {
	// default to current directory if no path provided
	if path == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		path = cwd
	}

	readmePath := filepath.Join(path, "README.md")
	readmePlain := filepath.Join(path, "README")

	// check if a README already exists at the target path
	_, err1 := os.Stat(readmePath)
	_, err2 := os.Stat(readmePlain)
	if err1 == nil || err2 == nil {
		fmt.Print("A README already exists. Overwrite? [y/N]: ")
		var input string
		fmt.Scanln(&input)
		if input != "y" && input != "Y" && strings.ToLower(input) != "yes" {
			fmt.Println("Aborted")
			return nil
		}
	}

	// infer project name from the target directory name
	dirName := filepath.Base(path)
	content := strings.ReplaceAll(readmeFile, "{project-name}", dirName)

	// prefer git config user.name over prompting — cleaner UX
	// falls back to prompt if git is not configured or not available
	var author string
	out, err := exec.Command("git", "config", "user.name").Output()
	if err != nil || strings.TrimSpace(string(out)) == "" {
		fmt.Print("Author name: ")
		fmt.Scanln(&author)
	} else {
		author = strings.TrimSpace(string(out))
	}

	authorLink := fmt.Sprintf("[%s](https://github.com/%s)", author, author)
	content = strings.ReplaceAll(content, "{author}", authorLink)

	// 0644 = owner read/write, group and others read only
	if err := os.WriteFile(readmePath, []byte(content), 0644); err != nil {
		return err
	}

	fmt.Println("Generated README.md")
	return nil
}
