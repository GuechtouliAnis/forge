package internal

import (
	"fmt"
	"os"
)

// SetupPython creates a virtual environment and installs dependencies.
// It checks for pyproject.toml first, then requirements.txt.
// If neither exists, the venv is created and left empty, ready to use.
func SetupPython() error {
	fmt.Println("Setting up Python environment...")

	// create the virtual environment in ./venv
	if err := run("python", "-m", "venv", "venv"); err != nil {
		return err
	}

	// upgrade pip inside the venv before installing any dependencies
	if err := run("venv/bin/pip", "install", "--upgrade", "pip"); err != nil {
		return err
	}

	// pyproject.toml takes priority: modern standard, installs as editable package
	if _, err := os.Stat("pyproject.toml"); err == nil {
		run("venv/bin/pip", "install", "-e", ".")
	} else if _, err := os.Stat("requirements.txt"); err == nil {
		// fall back to requirements.txt: classic dependency file
		run("venv/bin/pip", "install", "-r", "requirements.txt")
	} else {
		// no dependency file found: venv is ready but empty
		fmt.Println("No dependency file found, venv ready.")
	}

	fmt.Println("Environment ready - activate with: source venv/bin/activate")
	return nil
}
