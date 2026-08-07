package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/GuechtouliAnis/forge/internal/config"
)

// defaultBranch attempts to detect the repository's default branch from remote HEAD.
func defaultBranch(ctx context.Context) string {
	out, err := exec.CommandContext(ctx, "git", "symbolic-ref", "refs/remotes/origin/HEAD").Output()
	if err == nil {
		parts := strings.Split(strings.TrimSpace(string(out)), "/")
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
	}
	return "main"
}

// gitFetch attempts `git fetch --prune` against origin within the configured
// timeout. It returns true if the caller should fall back to offline mode
// (i.e. the fetch failed or timed out).
func gitFetch(ctx context.Context, cfg config.GitClean) bool {
	fmt.Println("Fetching from origin...")

	fetchTimeout := time.Duration(cfg.FetchTimeoutSeconds) * time.Second
	if fetchTimeout <= 0 {
		fetchTimeout = 30 * time.Second
	}

	fetchCtx, cancel := context.WithTimeout(ctx, fetchTimeout)
	defer cancel()

	fetch := exec.CommandContext(fetchCtx, "git", "fetch", "--prune")
	fetch.Stdout = os.Stdout
	fetch.Stderr = os.Stderr
	fetch.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	if err := fetch.Run(); err != nil {
		if fetchCtx.Err() == context.DeadlineExceeded {
			fmt.Printf("Warning: git fetch --prune timed out after %s.\n", fetchTimeout)
		} else {
			fmt.Printf("Warning: git fetch --prune failed (%v).\n", err)
		}
		fmt.Println("This failed, doing local diagnosis only — results reflect the last successful fetch and may be stale.")
		return true
	}
	return false
}
