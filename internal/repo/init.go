package repo

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CreateRepo initializes a new git repository at the given path with forge scaffolding.
// If path is empty, it initializes the current directory.
// If path is given, it fails if the directory already exists to prevent accidental overwrites.
func CreateRepo(path string, lang string, license string) error {
	// default to current directory if no path provided
	if path == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		path = cwd
	} else {
		// refuse to init into an existing directory — prevents accidental overwrites
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("directory already exists: %s", path)
		}
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
	}

	// git init — sets up the .git directory at the target path
	cmd := exec.Command("git", "init", path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git init failed: %w", err)
	}

	// scaffold .gitignore first — must exist before git add to prevent secret leaks
	if err := CreateGitignore(lang, path); err != nil {
		return err
	}
	if err := CreateReadme(path); err != nil {
		return err
	}
	if err := CreateLicense(license, path); err != nil {
		return err
	}
	if err := CreateChangelog(path); err != nil {
		return err
	}

	// stage all scaffolded files
	gc := exec.Command("git", "-C", path, "add", ".")
	gc.Stdout = os.Stdout
	gc.Stderr = os.Stderr
	if err := gc.Run(); err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}

	// safety check — abort if .env is staged, meaning gitignore didn't catch it
	out, err := exec.Command("git", "-C", path, "ls-files", "--others", "--exclude-standard", ".env").Output()
	if err == nil && strings.TrimSpace(string(out)) != "" {
		return fmt.Errorf("aborting commit: .env is not ignored — check your .gitignore")
	}

	// initial commit — marks the scaffold baseline
	gc = exec.Command("git", "-C", path, "commit", "-m", "[INIT] forge repo init")
	gc.Stdout = os.Stdout
	gc.Stderr = os.Stderr
	if err := gc.Run(); err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}

	fmt.Printf("Initialized repo in %s\n", path)
	return nil
}
