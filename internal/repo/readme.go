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

	// check if a README already exists at the target path
	for _, name := range []string{"README", "README.md"} {
		exists, err := CheckFileExists(path, name)
		if err != nil {
			return err
		}
		if exists {
			fmt.Print("A README already exists. Overwrite? [y/N]: ")
			var input string
			fmt.Scanln(&input)
			if input != "y" && input != "Y" && strings.ToLower(input) != "yes" {
				fmt.Println("Aborted")
				return nil
			}
			RemoveFileInsensitive(path, "README")
			RemoveFileInsensitive(path, "README.md")
			break
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
	if err := os.WriteFile(filepath.Join(path, "README.md"), []byte(content), 0644); err != nil {
		return err
	}

	fmt.Println("Generated README.md")
	return nil
}
