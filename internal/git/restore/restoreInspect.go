package git

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// fetchHistoricalPaths returns the raw `git log --all --name-only` output
// listing every path that has ever appeared in the repository's history.
func fetchHistoricalPaths(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx, "git", "log", "--all", "--name-only", "--pretty=format:").Output()
	if err != nil {
		return "", fmt.Errorf("[git restore]: could not read git history: %w", err)
	}
	return string(out), nil
}

// hasUncommittedChanges reports whether path has pending changes according
// to `git status --porcelain`.
func hasUncommittedChanges(ctx context.Context, path string) (bool, error) {
	out, err := exec.CommandContext(ctx, "git", "status", "--porcelain", path).Output()
	if err != nil {
		return false, fmt.Errorf("[git restore]: could not check file status: %w", err)
	}
	return strings.TrimSpace(string(out)) != "", nil
}

// isGitIgnored reports whether path is currently matched by .gitignore.
// `git check-ignore` exits 1 (non-zero) when the path is simply not
// ignored — that's expected and not a real error. Only a higher exit
// code, or a non-ExitError, is treated as failure.
func isGitIgnored(ctx context.Context, path string) (bool, error) {
	out, err := exec.CommandContext(ctx, "git", "check-ignore", "-q", path).Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); !ok || exitErr.ExitCode() > 1 {
			return false, fmt.Errorf("[git restore]: could not check .gitignore status: %w", err)
		}
	}
	return len(out) > 0, nil
}

// findLatestRestorableCommit walks the commit history of path (most
// recent first) and returns the hash of the most recent commit where the
// file was added/modified rather than deleted.
func findLatestRestorableCommit(ctx context.Context, path string) (string, error) {

	out, err := exec.CommandContext(ctx, "git", "log", "--all", "--numstat",
		"--pretty=format:%H", "--", path).Output()
	if err != nil || strings.TrimSpace(string(out)) == "" {
		return "", fmt.Errorf("[git restore]: no commit history found for %s", path)
	}

	hash := parseLatestRestorableCommit(string(out))
	if hash == "" {
		return "", fmt.Errorf("[git restore]: no restorable commit found for %s (all commits deleted it)", path)
	}

	return hash, nil
}

// fetchCommitHistory returns the raw `git log -n <n> --numstat` output for
// the last n commits that touched path.
func fetchCommitHistory(ctx context.Context, path string, n int) (string, error) {

	out, err := exec.CommandContext(ctx, "git", "log", "-n", fmt.Sprint(n), "--numstat",
		"--pretty=format:%h|%cr|%s", "--", path).Output()

	if err != nil || strings.TrimSpace(string(out)) == "" {
		return "", fmt.Errorf("[git restore]: no commit history found for %s", path)
	}

	return string(out), nil
}

// diffStats returns a "(+A, -B)" summary for path as of commit hash, or ""
// if it can't be determined.
func diffStats(ctx context.Context, hash, path string) string {
	out, _ := exec.CommandContext(ctx, "git", "show", "--numstat", "--pretty=format:",
		hash, "--", path).Output()
	fields := strings.Fields(strings.TrimSpace(string(out)))
	if len(fields) < 2 {
		return ""
	}
	return fmt.Sprintf(" (+%s, -%s)", fields[0], fields[1])
}
