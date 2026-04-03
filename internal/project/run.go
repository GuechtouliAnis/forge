package project

import (
	"os"
	"os/exec"
)

// run executes a shell command and pipes stdout/stderr directly to the terminal.
// It is used by all internal functions to run external tools like git, go, and python.
func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
