package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/GuechtouliAnis/forge/internal/config"
)

// commitHash returns the short hash of HEAD immediately after a commit,
// for inclusion in the success message.
func commitHash(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx, "git", "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return "", fmt.Errorf("[git commit]: could not read commit hash: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// CommitGit validates and commits staged changes, or amends the previous
// commit. messageProvided must reflect whether -m was actually passed on
// the command line (via cmd.Flags().Changed("message") at the call site)
// rather than merely whether message is an empty string — this matters
// because an explicit `-m ""` is a different situation from omitting -m
// entirely, and only the latter should trigger the amend-reuse path.
//
// dryRun runs every validation and warning/block check exactly as normal,
// but stops short of the two points that mutate git state (the --no-edit
// amend and the real git commit), printing what would happen instead.
//
// Mirrors AddGit/CleanGit's error convention: only the outcomes in the
// spec's error table (NO_MESSAGE, MESSAGE_REJECTED, NOTHING_STAGED,
// GIT_COMMIT_FAILED, and CreatePattern/ValidateCommit's own errors)
// surface as a non-nil error. Everything else prints its outcome message
// directly and returns nil.
func CommitGit(cfg *config.GitCommit, message string, messageProvided bool, amend bool, dryRun bool) error {

	// Set up a cancellable context tied to OS interrupt/termination signals.
	// This allows long-running operations below to be aborted cleanly
	// mid-flight rather than leaving the terminal in an inconsistent state
	// on Ctrl+C.
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM)

	defer stop()

	// Verify we are actually inside a git repository.
	// Fails fast with a clear error otherwise.
	if err := gitCheck(ctx); err != nil {
		return err
	}

	// --- NO_MESSAGE: required unless amending without a new message -------
	if !messageProvided && !amend {
		return fmt.Errorf("[git commit]: a commit message is required")
	}

	// Get a list of staged files (without their status)
	staged, err := stagedFiles(ctx)
	if err != nil {
		return err
	}

	// If nothing staged and we are not amending, error
	if !amend && len(staged) == 0 {
		return fmt.Errorf("[git commit]: nothing staged to commit")
	}

	// Non-fatal safety net: warn or block, per staged_ignored_files_mode, if
	// anything staged (however it got staged) matches a .gitignore rule.
	if ignored, err := stagedIgnoredFiles(ctx); err != nil {
		return err
	} else if len(ignored) > 0 {
		msg := fmt.Sprintf("staged but found in .gitignore: %s", strings.Join(ignored, ", "))
		if cfg.StagedIgnoredFilesMode == "block" {
			return fmt.Errorf("[git commit]: %s", msg)
		}
		fmt.Printf("[git commit]: warning - %s\n", msg)
	}

	// --- Amend without a new message: reuse the previous message verbatim.
	if amend && !messageProvided {
		if dryRun {
			prevMsg, err := previousCommitMessage(ctx)
			if err != nil {
				return err
			}
			fmt.Printf("[git commit]: [dry-run] would amend previous commit (message unchanged): %s\n", prevMsg)
			if len(staged) > 0 {
				fmt.Printf("[git commit]: [dry-run] staged files to include: %s\n", strings.Join(staged, ", "))
			}
			return nil
		}
		out, err := exec.CommandContext(ctx, "git", "commit", "--amend", "--no-edit").CombinedOutput()
		if err != nil {
			return fmt.Errorf("[git commit]: git commit failed: %s", strings.TrimSpace(string(out)))
		}
		hash, err := commitHash(ctx)
		if err != nil {
			return err
		}
		fmt.Printf("[git commit]: amended previous commit %s\n", hash)
		return nil
	}

	// --- Validate the supplied message (new commit, or amend -m) ----------
	result, err := ValidateCommit(message, cfg)
	if err != nil {
		return fmt.Errorf("[git commit]: %w", err)
	}
	if !result.Valid {
		return fmt.Errorf("[git commit]: commit message does not match required format")
	}
	if result.CaseWarning {
		fmt.Printf("[git commit]: warning - domain case mismatch (expected %s, got %s)\n",
			result.ExpectedDomain, result.GotDomain)
	}

	// Dry run
	if dryRun {
		// Add pretty print for the dry run
		if amend {
			fmt.Printf("[git commit]: [dry-run] would amend previous commit with message: %s\n", message)
		} else {
			fmt.Printf("[git commit]: [dry-run] would commit with message: %s\n", message)
		}
		if len(staged) > 0 {
			fmt.Printf("[git commit]: [dry-run] staged files to include: %s\n", strings.Join(staged, ", "))
		} else if amend {
			fmt.Println("[git commit]: [dry-run] no staged changes — only the message would change")
		}
		return nil
	}

	// Run the actual git commit
	gitArgs := []string{"commit", "-m", message}
	if amend {
		gitArgs = []string{"commit", "--amend", "-m", message}
	}
	out, err := exec.CommandContext(ctx, "git", gitArgs...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("[git commit]: git commit failed: %s", strings.TrimSpace(string(out)))
	}

	hash, err := commitHash(ctx)
	if err != nil {
		return err
	}

	if amend {
		fmt.Printf("[git commit]: amended previous commit %s\n", hash)
	} else {
		fmt.Printf("[git commit]: committed as %s\n", hash)
	}

	return nil
}
