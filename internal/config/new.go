package config

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/GuechtouliAnis/forge/internal/repo"
)

//go:embed templates/.forge.toml.template
var tomlFile string

func CreateForgeToml(path string) error {
	if path == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		path = cwd
	}

	exists, err := repo.CheckFileExists(path, ".forge.toml")
	if err != nil {
		return err
	}

	if exists {
		fmt.Print("A .forge.toml already exists. Overwrite? [y/N]: ")
		var input string
		fmt.Scanln(&input)
		if strings.ToLower(strings.TrimSpace(input)) != "y" {
			fmt.Println("Aborted.")
			return nil // early return
		}
		if err := repo.RemoveFileInsensitive(path, ".forge.toml"); err != nil {
			return err
		}
	}

	dest := filepath.Join(path, ".forge.toml")
	if err := os.WriteFile(dest, []byte(tomlFile), 0644); err != nil {
		return err
	}

	fmt.Println(".forge.toml created.")
	return nil
}
