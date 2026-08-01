package git

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// StagedFile pairs a staged file's path with its git status code
// (e.g. "M", "A", "D"), as reported by `git diff --cached --name-status`.
type StagedFile struct {
	Code string
	Path string
}

// stagedFiles returns every currently staged file along with its git
// status code (e.g. "M", "A", "D"), as reported by
// `git diff --cached --name-status`.
func stagedFiles(ctx context.Context) ([]StagedFile, error) {

	// Run git diff --cached --name-status to get the staged files with their code
	out, err := exec.CommandContext(ctx, "git", "diff", "--cached", "--name-status").Output()
	if err != nil {
		return nil, fmt.Errorf("[git commit]: could not check staged changes: %w", err)
	}

	// Trim trailing newline from git's output
	trimmed := strings.TrimSpace(string(out))
	if trimmed == "" {
		return nil, nil
	}

	lines := strings.Split(trimmed, "\n")
	// Pre-size at len(lines): the common case is one file per line, so
	// this avoids reallocation without over-allocating meaningfully.
	files := make([]StagedFile, 0, len(lines))
	for _, line := range lines {
		// --name-status output is tab-separated: "<code>\t<path>".
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			// Malformed line (shouldn't happen with well-formed git
			// output) — skip rather than fail the whole call.
			continue
		}
		files = append(files, StagedFile{Code: parts[0], Path: parts[1]})
	}
	return files, nil
}

// stagedIgnoredFiles returns every file that is currently staged despite
// matching a .gitignore (or other standard exclude) rule. This can only
// happen via an explicit force-add (`git add -f`), since forge git add's
// own guardrails never bypass ignore rules on their own.
func stagedIgnoredFiles(ctx context.Context) ([]string, error) {
	out, err := exec.CommandContext(ctx, "git", "ls-files", "-c", "-i", "--exclude-standard").Output()
	if err != nil {
		return nil, fmt.Errorf("[git commit]: could not check staged files against .gitignore: %w", err)
	}

	// Trim trailing newline from git's output
	trimmed := strings.TrimSpace(string(out))
	if trimmed == "" {
		return nil, nil
	}
	return strings.Split(trimmed, "\n"), nil
}

// previousCommitMessage returns the full message of the current HEAD
// commit, for display in dry-run output when amending without a new
// message — the previous message is what --no-edit would reuse verbatim.
func previousCommitMessage(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx, "git", "log", "-1", "--pretty=%B").Output()
	if err != nil {
		return "", fmt.Errorf("[git commit]: could not read previous commit message: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
