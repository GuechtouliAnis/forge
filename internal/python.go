package internal

import (
	"fmt"
	"os"
)

func SetupPython() error {
	fmt.Println("Setting up Python environment...")

	if err := run("python", "-m", "venv", "venv"); err != nil {
		return err
	}

	if err := run("venv/bin/pip", "install", "--upgrade", "pip"); err != nil {
		return err
	}

	// check if pyproject exists
	if _, err := os.Stat("pyproject.toml"); err == nil {
		run("bash", "-c", "source venv/bin/activate && pip install -e .")
	} else if _, err := os.Stat("requirements.txt"); err == nil {
		run("bash", "-c", "source venv/bin/activate && pip install -r requirements.txt")
	} else {
		fmt.Println("No dependency file found, venv ready.")
	}

	fmt.Println("Environment ready - activate with: source venv/bin/activate")
	return nil
}
