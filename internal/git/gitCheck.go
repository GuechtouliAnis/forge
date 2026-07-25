package git

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// gitCheck performs pre-flight validation: confirms we're inside a git
// repository and that the installed git version meets the minimum requirement.
func gitCheck(ctx context.Context) error {

	// Confirm git is installed
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("[git clean]: git is not installed or not in PATH")
	}

	// Confirm we are in a git reposityory
	if err := exec.CommandContext(ctx, "git", "rev-parse", "--is-inside-work-tree").Run(); err != nil {
		return fmt.Errorf("[git clean]: not a git repository")
	}

	// Extract git version
	verOut, err := exec.CommandContext(ctx, "git", "--version").Output()
	if err != nil {
		return fmt.Errorf("[git clean]: could not determine git version")
	}

	// Trim the git version from the output
	verStr := strings.TrimPrefix(strings.TrimSpace(string(verOut)), "git version ")

	// Confirm git version is 2.0+
	major, err := strconv.Atoi(strings.Split(verStr, ".")[0])
	if err != nil || major < 2 {
		return fmt.Errorf("[git clean]: git 2.0+ required, found: %s", verStr)
	}
	return nil
}
