package repo

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//go:embed templates/changelog/template.changelog
var changelogFile string

// CreateChangelog writes a CHANGELOG.md scaffold to the given path.
// If path is empty, it defaults to the current directory.
// Prompts before overwriting an existing changelog.
func CreateChangelog(path string) error {
	if path == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		path = cwd
	}

	for _, name := range []string{"CHANGELOG.md", "CHANGELOG"} {
		exists, err := CheckFileExists(path, name)
		if err != nil {
			return err
		}
		if exists {
			fmt.Print("A CHANGELOG already exists. Overwrite? [y/N]: ")
			var input string
			fmt.Scanln(&input)
			if input != "y" && input != "Y" && strings.ToLower(input) != "yes" {
				fmt.Println("Aborted")
				return nil
			}
			// remove existing file regardless of casing before writing
			RemoveFileInsensitive(path, "CHANGELOG.md")
			RemoveFileInsensitive(path, "CHANGELOG")
			break
		}
	}

	content := strings.ReplaceAll(changelogFile, "{date}", time.Now().Format("2006-01-02"))

	if err := os.WriteFile(filepath.Join(path, "CHANGELOG.md"), []byte(content), 0644); err != nil {
		return err
	}

	fmt.Println("Generated CHANGELOG.md")
	return nil
}
