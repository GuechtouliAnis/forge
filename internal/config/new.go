package config

import (
	"bufio"
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
	// if no given filepath, use current path
	if path == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("[config new]: %w", err)
		}
		path = cwd
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("[config new]: path does not exist: %s", path)
		}
		return fmt.Errorf("[config new]: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("[config new]: path is not a directory: %s", path)
	}

	exists, err := repo.CheckFileExists(path, ".forge.toml")
	if err != nil {
		return fmt.Errorf("[config new]: %w", err)
	}

	if exists {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("[config new]: A .forge.toml already exists. Overwrite? [y/N]: ")
		raw, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("[config new]: failed to read input: %w", err)
		}
		confirmed := strings.ToLower(strings.TrimSpace(raw))
		if confirmed != "y" && confirmed != "yes" {
			fmt.Println("[config new]: Aborted.")
			return nil
		}
	}

	dest := filepath.Join(path, ".forge.toml")
	if err := os.WriteFile(dest, []byte(tomlFile), 0644); err != nil {
		return fmt.Errorf("[config new]: %w", err)
	}

	fmt.Println("[config new]: .forge.toml created.")

	gitPath := filepath.Join(path, ".git")
	if info, err := os.Stat(gitPath); err != nil || !info.IsDir() {
		fmt.Println("Tip: .forge.toml works best at your repo root, alongside .git/")
	}

	return nil
}
