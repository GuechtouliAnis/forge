package git

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// stagedFiles returns the list of files currently staged for commit, or
// nil if none are staged. The default (non-amend) flow treats an empty
// result as NOTHING_STAGED; amend does not require staged changes, since
// it can rewrite the previous commit's message or tree alone — but the
// list is fetched regardless so dry-run output can show what would be
// included alongside the amend, if anything.
func stagedFiles(ctx context.Context) ([]string, error) {
	out, err := exec.CommandContext(ctx, "git", "diff", "--cached", "--name-only").Output()
	if err != nil {
		return nil, fmt.Errorf("[git commit]: could not check staged changes: %w", err)
	}
	trimmed := strings.TrimSpace(string(out))
	if trimmed == "" {
		return nil, nil
	}
	return strings.Split(trimmed, "\n"), nil
}

// stagedIgnoredFiles returns every file that is currently staged despite
// matching a .gitignore (or other standard exclude) rule. This can only
// happen via an explicit force-add (`git add -f`), since forge git add's
// own guardrails never bypass ignore rules on their own — so surfacing
// it here is a safety net for whatever staged the index, not just files
// forge itself staged.
func stagedIgnoredFiles(ctx context.Context) ([]string, error) {
	out, err := exec.CommandContext(ctx, "git", "ls-files", "-c", "-i", "--exclude-standard").Output()
	if err != nil {
		return nil, fmt.Errorf("[git commit]: could not check staged files against .gitignore: %w", err)
	}
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
