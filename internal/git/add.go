package git

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/GuechtouliAnis/forge/internal/config"
)

// fileState is the terminal classification assigned to a single file after
// it has passed through (or been rejected by) the AddGit guardrail
// pipeline. It exists purely to drive summary counts and status text.
type fileState int

const (
	stateStaged           fileState = iota // approved and successfully staged
	stateBlockedHardcoded                  // .env / env.default_file — unconditional
	stateBlockedPattern                    // matched a blocklist_patterns glob
	stateSkippedLarge                      // exceeded max_file_size_mb, not confirmed
	stateFailedLarge                       // "fail" mode, dry-run only — would abort in a real run
	stateFailed                            // approved, but `git add` itself failed
)

// fileResult records the outcome of a single file as it moves through the
// pipeline. Reason holds the matched blocklist pattern (for blocked files)
// or the git error text (for failed stages) — used only for display.
type fileResult struct {
	Path   string
	State  fileState
	Reason string
	SizeMB float64
	Code   string
}

// AddGit wraps `git add` with a validation layer applied to every file
// before it reaches the index:
//
//  1. Environment files (.env and the project's configured env.default_file)
//     are unconditionally rejected — no configuration option re-enables
//     staging them through this command. `git add` directly bypasses the
//     guard, by design.
//  2. Files whose base name matches a configured blocklist glob pattern
//     (e.g. "*.pem", "id_rsa") are rejected next, case-insensitively.
//  3. Every remaining file is size-checked; anything over max_file_size_mb
//     prompts for confirmation before staging, unless the threshold is
//     disabled via -1.
//
// dryRun previews every decision above without touching the index and
// without prompting on stdin — oversized files are simply reported as
// "would-prompt" rather than blocking execution on user input.
//
// Design note: blocking or skipping a file is treated as expected,
// non-erroring behavior — the guardrails did their job. AddGit only
// returns a non-nil error for fatal preconditions (not a repo, no path
// given, a path that doesn't exist) or if `git add` itself fails for an
// approved file, mirroring how CleanGit treats "nothing to do" as success
// rather than failure.
func AddGit(cfg config.GitAdd, envDefaultFile string, rawPaths []string, dryRun bool) error {

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

	// At least one path is required
	if len(rawPaths) == 0 {
		return fmt.Errorf("[git add]: at least one path is required")
	}

	// Expand directories into individual files
	files, err := resolvePaths(ctx, rawPaths)
	if err != nil {
		return err
	}

	// Classify every file through the guardrail pipeline

	// blocked / skipped entries — terminal, not staged
	var results []fileResult

	// approved entries — pending the staging step
	var toStage []fileResult

	for _, f := range files {
		// Cooperative cancellation: check for Ctrl+C/SIGTERM between
		// files rather than only at the start, so a large batch can be
		// interrupted mid-loop instead of running to completion.
		if ctx.Err() != nil {
			fmt.Println("\n[git add]: Interrupted — remaining files were not evaluated.")
			break
		}

		result, shouldStage, err := classifyFile(f, cfg, envDefaultFile, dryRun)
		if err != nil {
			// Only a stat failure (file vanished between resolution and
			// evaluation) reaches here
			return err
		}

		// Sort into files that pass the guardrail pipeline (to be staged)
		// vs. files that were flagged/skipped (reported but not staged).
		if shouldStage {
			toStage = append(toStage, result)
		} else {
			results = append(results, result)
		}
	}

	// dry-run before any index mutation
	if dryRun {
		printDryRunSummary(results, toStage)
		return nil
	}

	// stage every approved file
	staged, failed := stageFiles(ctx, toStage)

	// final summary
	blocked := countState(results, stateBlockedHardcoded) + countState(results, stateBlockedPattern)
	skipped := countState(results, stateSkippedLarge)

	fmt.Printf("\n[git add]: %d file(s) staged and ready to commit, %d blocked, %d skipped.\n",
		staged, blocked, skipped)

	// A partial `git add` failure surfaces as a non-zero exit so CI
	// pipelines and calling scripts can detect it, rather than reporting
	// success while silently having left a file unstaged.
	if failed > 0 {
		return fmt.Errorf("[git add]: %d file(s) failed to stage", failed)
	}

	return nil
}
