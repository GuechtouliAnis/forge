package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/GuechtouliAnis/forge/internal/git"
)

// RestoreFile recovers a file from git history using fuzzy path matching.
// Collision detection prompts before overwriting dirty/staged files.
// --latest skips the version menu and restores from the most recent commit where the file existed.
// --commit allows pinning to a specific commit hash.
func RestoreFile(search string, latest bool, dryRun bool, commitHash string) error {
	// Set up a cancellable context tied to OS interrupt/termination signals.
	// This allows long-running operations below to be aborted cleanly
	// mid-flight rather than leaving the terminal in an inconsistent state
	// on Ctrl+C.
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM)
	defer stop()

	if err := git.GitCheck(ctx); err != nil {
		return err
	}

	resolvedPath, err := resolveTargetPath(ctx, search)
	if err != nil || resolvedPath == "" {
		return err
	}

	proceed, err := confirmCollisions(ctx, resolvedPath)
	if err != nil {
		return err
	}
	if !proceed {
		return nil
	}

	targetHash, err := resolveTargetCommit(ctx, resolvedPath, latest, commitHash)
	if err != nil || targetHash == "" {
		return err
	}

	if dryRun {
		printDryRun(resolvedPath, targetHash)
		return nil
	}

	return checkoutAndReport(ctx, resolvedPath, targetHash)
}

// resolveTargetPath fuzzy-matches search against historical paths and
// resolves it to a single path, prompting the user if there's more than
// one match. Returns "" with a nil error if the user cancels.
func resolveTargetPath(ctx context.Context, search string) (string, error) {
	logOutput, err := fetchHistoricalPaths(ctx)
	if err != nil {
		return "", err
	}

	matches := findMatchingPaths(logOutput, search)
	if len(matches) == 0 {
		return "", fmt.Errorf("[git restore]: no historical paths found matching %q", search)
	}

	return promptPathSelection(search, matches)
}

// confirmCollisions checks path for uncommitted changes and gitignore
// status, prompting the user before a restore would overwrite either.
// proceed is false if the user declined either prompt.
func confirmCollisions(ctx context.Context, path string) (proceed bool, err error) {

	dirty, err := hasUncommittedChanges(ctx, path)
	if err != nil {
		return false, err
	}
	if dirty && !confirmOverwrite(path) {
		fmt.Println("[git restore]: Aborted.")
		return false, nil
	}

	ignored, err := isGitIgnored(ctx, path)
	if err != nil {
		return false, err
	}

	if ignored && !confirmIgnoredRestore(path) {
		fmt.Println("[git restore]: Aborted.")
		return false, nil
	}

	return true, nil
}

// resolveTargetCommit determines which commit to restore path from: a
// pinned --commit hash, the most recent restorable commit for --latest,
// or an interactive menu over the last 10 commits that touched the file.
// Returns "" with a nil error if the user cancels the menu.
func resolveTargetCommit(ctx context.Context, path string, latest bool, commitHash string) (string, error) {
	if commitHash != "" {
		return commitHash, nil
	}
	if latest {
		return findLatestRestorableCommit(ctx, path)
	}

	raw, err := fetchCommitHistory(ctx, path, 10)
	if err != nil {
		return "", err
	}

	commits := parseCommitHistory(raw)
	if len(commits) == 0 {
		return "", fmt.Errorf("[git restore]: no commit history found for %s", path)
	}

	restorable := restorableCommits(commits)
	if len(restorable) == 0 {
		return "", fmt.Errorf("[git restore]: no restorable commits found for %s", path)
	}

	return promptCommitSelection(path, restorable)
}

// checkoutAndReport performs the actual git checkout + unstage of path at
// hash, then prints a summary of what was restored.
func checkoutAndReport(ctx context.Context, path, hash string) error {

	if err := checkoutFile(ctx, hash, path); err != nil {
		return err
	}

	if err := unstageFile(ctx, path); err != nil {
		return err
	}

	printRestored(path, hash, diffStats(ctx, hash, path))

	return nil
}

// checkoutFile restores path's content from hash via `git checkout`.
func checkoutFile(ctx context.Context, hash, path string) error {

	out, err := exec.CommandContext(ctx, "git", "checkout", hash, "--", path).CombinedOutput()
	if err != nil {
		return fmt.Errorf("[git restore]: git checkout failed: %s — %w",
			strings.TrimSpace(string(out)), err)
	}

	return nil
}

// unstageFile removes path from the index after checkout, since
// `git checkout <commit> -- <path>` stages the restored content.
func unstageFile(ctx context.Context, path string) error {

	out, err := exec.CommandContext(ctx, "git", "restore", "--staged", path).CombinedOutput()
	if err != nil {
		return fmt.Errorf("[git restore]: restored file but failed to unstage it: %s — %w",
			strings.TrimSpace(string(out)), err)
	}

	return nil
}
