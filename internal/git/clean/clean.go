package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/GuechtouliAnis/forge/internal/config"
	"github.com/GuechtouliAnis/forge/internal/git"
)

// CleanGit scans all local branches in the current repository and identifies
// which ones are "stale" — i.e. eligible for deletion — based on two
// independent, configurable criteria:
//
//  1. Age: the branch has not been committed to in cfg.StaleDays days.
//  2. Divergence: the branch is cfg.CommitsBehind (or more) commits behind
//     the repository's default branch.
//
// A branch is flagged as stale if EITHER threshold is enabled and met (see
// the `ageEnabled`/`behindEnabled` logic below). Protected branches (main,
// master, or the detected default branch) and the currently checked-out
// branch are always exempt and reported separately.
//
// Execution modes:
//   - Dry-run (default): stale branches are listed but nothing is deleted.
//   - remove=true: after listing, the user is prompted to confirm deletion.
//   - remove=true, force=true: deletion proceeds without confirmation.
//
// Network behavior:
//   - By default (offline=false), the function attempts `git fetch` before
//     diagnosis, so remote-tracking data (merged status, ahead/behind counts)
//     reflects the server's current state.
//   - If offlineFlag is set, or cfg.FetchRemote is false, the fetch step is
//     skipped entirely and diagnosis relies solely on local refs.
//   - If a fetch is attempted but fails (e.g. no network), it degrades
//     gracefully to local-only diagnosis.
func CleanGit(cfg config.GitClean, remove bool, force bool, offlineFlag bool) error {

	// Set up a cancellable context tied to OS interrupt/termination signals.
	// This allows long-running operations below (particularly the concurrent
	// "commits behind" evaluation) to be aborted cleanly mid-flight rather
	// than leaving the terminal in an inconsistent state on Ctrl+C.
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM)

	defer stop()

	// Verify we are actually inside a git repository.
	// Fails fast with a clear error otherwise.
	if err := git.GitCheck(ctx); err != nil {
		return err
	}

	// Resolve the effective offline mode: an explicit --offline flag always
	// wins, but the config's fetch_remote=false setting has the same effect
	// if no flag was passed.
	offline := offlineFlag || !cfg.FetchRemote

	days := cfg.StaleDays
	behind := cfg.CommitsBehind

	// Synchronize with the remote and remove local references to branches
	// that no longer exist on the server. This step is skipped entirely
	// when fetch_remote=false or --offline is passed. If a fetch is
	// attempted and fails, it is treated as non-fatal: we fall back to
	// local-only diagnosis using existing refs rather than erroring out.
	if !offline {
		if fallback := gitFetch(ctx, cfg); fallback {
			offline = true
		}
	} else {
		fmt.Println("[git clean]: fetch_remote is disabled — doing local diagnosis only.")
	}

	// Bail out early if the user interrupted during the fetch step.
	if ctx.Err() != nil {
		return reportInterrupted(0, 0)
	}

	// Determine the repository's default branch (e.g. main/master), used
	// both as a protected branch and as the comparison point for the
	// "commits behind" calculation.
	defaultBranch := defaultBranch(ctx)

	// If we ended up in offline mode (whether by choice or by fetch
	// failure), inform the user how stale the local remote-tracking data
	// might be, so they can judge how much to trust "merged"/"behind"
	// results below.
	if offline {
		if info, err := os.Stat(".git/FETCH_HEAD"); err == nil {
			lastFetchTime := time.Since(info.ModTime()).Round(time.Minute)

			if lastFetchTime < time.Hour {
				fmt.Printf("[git clean]: Last successful fetch: %.0fm ago.\n\n",
					lastFetchTime.Minutes())
			} else {
				fmt.Printf("[git clean]: Last successful fetch: %dh%dm ago.\n\n",
					int(lastFetchTime.Hours()),
					int(lastFetchTime.Minutes())%60)
			}
		} else {
			fmt.Println("[git clean]: No record of a prior successful fetch — remote-tracking data may be unreliable.")
			fmt.Println()
		}
	}

	// Identify the current branch and exempt it from deletion
	currentOut, err := exec.CommandContext(ctx, "git", "branch", "--show-current").Output()
	if err != nil {
		return fmt.Errorf("[git clean]: could not determine current branch: %w", err)
	}
	currentBranch := strings.TrimSpace(string(currentOut))

	// Batch-fetch all branches with their commit timestamps
	// Rather than shelling out once per branch (which would be O(n) process
	// spawns), we issue a single `git for-each-ref` call and parse its
	// output. Fields are NUL-separated (%00) since NUL is one of the few bytes
	// git-check-ref-format guarantees cannot appear inside a ref name.
	refOut, err := exec.CommandContext(ctx, "git", "for-each-ref",
		"--format=%(refname:short)%00%(committerdate:unix)", "refs/heads/").Output()

	if err != nil {
		return fmt.Errorf("[git clean]: could not list branches: %w", err)
	}

	branchLines := strings.Split(strings.TrimSpace(string(refOut)), "\n")
	if len(branchLines) == 0 || (len(branchLines) == 1 && branchLines[0] == "") {
		fmt.Println("[git clean]: No branches to clean.")
		return nil
	}

	// staleMap indexes BranchInfo by branch name for O(1) lookups during the
	// subsequent merged-status and behind-count passes. branchNames
	// preserves the original iteration order so final output is stable and
	// deterministic across runs.
	staleMap := make(map[string]*BranchInfo)
	var branchNames []string
	now := time.Now()

	for _, line := range branchLines {
		parts := strings.SplitN(line, "\x00", 2)
		if len(parts) != 2 {
			// Malformed line
			// Skip defensively
			continue
		}

		branchName := parts[0]
		timestampString := parts[1]

		info := &BranchInfo{Name: branchName}
		if lastCommitTimestamp, err := strconv.ParseInt(timestampString, 10, 64); err == nil {
			info.DaysOld = int(now.Sub(time.Unix(lastCommitTimestamp, 0)).Hours() / 24)
		}
		staleMap[branchName] = info
		branchNames = append(branchNames, branchName)
	}

	// Only the current branch exists — nothing to evaluate.
	if len(branchNames) <= 1 {
		fmt.Println("[git clean]: No branches to clean.")
		return nil
	}

	// Batch-fetch merged status: which branches are already fully
	// merged into the default branch. Non-fatal on failure — if this call
	// errors, all branches simply keep their zero-value (unmerged) status,
	// and evaluation continues on age/behind criteria alone.
	mergedOut, err := exec.CommandContext(ctx, "git", "branch",
		"--format=%(refname:short)", "--merged", defaultBranch).Output()

	if err == nil {
		for _, name := range strings.Split(strings.TrimSpace(string(mergedOut)), "\n") {
			name = strings.TrimSpace(name)
			if info, exists := staleMap[name]; exists {
				info.Merged = true
			}
		}
	}

	// Classify branches, queue the rest for "behind" analysis
	// Protected and current branches are settled immediately and excluded
	// from the (more expensive) concurrent "commits behind" evaluation.
	// Everything else goes into toEvaluate.
	var toEvaluate []string

	// Create a quick lookup map for user-protected branches
	isConfigProtected := make(map[string]bool)
	for _, b := range cfg.ProtectedBranches {
		isConfigProtected[b] = true
	}

	for _, name := range branchNames {
		info := staleMap[name]

		// Changing 'switch name' to an expressionless switch lets us evaluate conditions
		switch {
		case name == "main" || name == "master" || name == defaultBranch || isConfigProtected[name]:
			info.Protected = true
		case name == currentBranch:
			info.Current = true
		default:
			toEvaluate = append(toEvaluate, name)
		}
	}

	// Concurrently compute "commits behind" for each candidate
	// This is the most expensive part of the scan (one `git rev-list` call
	// per branch), so it is parallelized across a worker pool sized by
	// computeWorkerCount. Concurrency-safety note: workers only ever write
	// to the `results` channel; staleMap is mutated exclusively by this
	// goroutine as it drains `results`, so there is no concurrent write to
	// shared state — provably race-free regardless of worker count.
	if len(toEvaluate) > 0 {
		numWorkers := computeWorkerCount(len(toEvaluate), cfg.MaxWorkers)
		jobs := make(chan string)
		results := make(chan behindResult, len(toEvaluate))
		var wg sync.WaitGroup
		wg.Add(numWorkers)

		// Worker goroutines: each pulls branch names off `jobs` until the
		// channel is closed, computes how many commits the default branch
		// is ahead of that branch, and reports the result.
		for w := 0; w < numWorkers; w++ {
			go func() {
				defer wg.Done()
				for name := range jobs {
					select {
					case <-ctx.Done():
						// Respect cancellation: report as interrupted rather
						// than spawning another subprocess.
						results <- behindResult{name: name, interrupted: true}
					default:
						// Each individual `git rev-list` call gets its own
						// 5-second timeout so a single hung subprocess can't
						// stall the entire worker pool indefinitely.
						behindCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
						behindOut, err := exec.CommandContext(behindCtx, "git", "rev-list", "--count",
							name+".."+defaultBranch).Output()
						cancel()

						n := 0
						if err == nil {
							n, _ = strconv.Atoi(strings.TrimSpace(string(behindOut)))
						}
						results <- behindResult{name: name, behind: n}
					}
				}
			}()
		}

		// Producer: feeds branch names into `jobs`, aborting early if the
		// context is cancelled mid-dispatch.
		go func() {
			defer close(jobs)
			for _, name := range toEvaluate {
				select {
				case jobs <- name:
				case <-ctx.Done():
					return
				}
			}
		}()

		// Closer: waits for all workers to finish, then closes `results` so
		// the consumer loop below terminates.
		go func() {
			wg.Wait()
			close(results)
		}()

		// Consumer: the only goroutine that writes into staleMap, draining
		// results as they arrive and merging them into the corresponding
		// BranchInfo entries.
		for r := range results {
			info := staleMap[r.name]
			info.Behind = r.behind
			info.Interrupted = r.interrupted
		}
	}

	// Rebuilt in original branchNames order so output is stable and
	// deterministic, rather than relying on Go's unordered map iteration.
	var stale []BranchInfo
	for _, name := range branchNames {
		stale = append(stale, *staleMap[name])
	}

	// If interrupted during the concurrent phase, report partial progress
	// instead of silently returning incomplete/misleading results.
	if ctx.Err() != nil {
		evaluated := 0
		for _, b := range stale {
			if !b.Interrupted {
				evaluated++
			}
		}
		return reportInterrupted(evaluated, len(stale))
	}

	// Render the report table
	fmt.Printf("\n%-30s %-10s %-10s %-10s %s\n", "BRANCH", "DAYS OLD", "BEHIND", "MERGED", "STATUS")
	fmt.Println(strings.Repeat("-", 75))

	// Each threshold is independently toggleable: a zero/negative config
	// value disables that criterion entirely (ageEnabled/behindEnabled),
	// allowing users to clean by age only, by divergence only, or both.
	ageEnabled := days > 0
	behindEnabled := behind > 0
	var toDelete []BranchInfo

	for _, b := range stale {
		behindStr := strconv.Itoa(b.Behind)

		switch {
		case b.Protected:
			fmt.Printf("%-30s %-10s %-10s %-10s %s\n", b.Name, "-", "-", "-", "[PROTECTED]")
		case b.Current:
			fmt.Printf("%-30s %-10s %-10s %-10s %s\n", b.Name, "-", "-", "-", "[CURRENT — skipped]")
		case (ageEnabled && b.DaysOld >= days) || (behindEnabled && b.Behind >= behind):
			// Flagged stale if EITHER enabled criterion is met — this is a
			// deliberate OR, not AND: a branch that's old OR far behind is
			// still worth flagging, since either condition independently
			// signals abandonment.
			status := "[TO BE DELETED]"
			if !remove {
				status = "[STALE]"
			}
			fmt.Printf("%-30s %-10d %-10s %-10v %s\n", b.Name, b.DaysOld, behindStr, b.Merged, status)
			toDelete = append(toDelete, b)
		default:
			fmt.Printf("%-30s %-10d %-10s %-10v %s\n", b.Name, b.DaysOld, behindStr, b.Merged, "[OK]")
		}
	}

	fmt.Println()

	// exit early if nothing to do
	if len(toDelete) == 0 {
		fmt.Println("[git clean]: No stale branches found.")
		return nil
	}

	// Report findings only
	if !remove {
		fmt.Printf("[git clean]: %d stale branch(es) found. Run with --remove to delete.\n", len(toDelete))
		return nil
	}

	// confirmation gate is skipped when --force is passed
	if !force {
		fmt.Printf("[git clean]: Delete %d branch(es)? [y/N]: ", len(toDelete))
		var input string
		fmt.Scanln(&input)
		if strings.ToLower(input) != "y" && strings.ToLower(input) != "yes" {
			fmt.Println("[git clean]: Aborted.")
			return nil
		}
	}

	// Merged branches use the safe delete flag (-d), which git itself
	// refuses if the branch has unmerged commits. Unmerged branches use the
	// force flag (-D), since we already know via cfg thresholds that the
	// branch is intentionally being discarded despite being unmerged.
	for _, b := range toDelete {
		flag := "-d"

		if !b.Merged {
			if !force {
				// Safe mode: Warn the user and skip it, or let git -d fail gracefully.
				fmt.Printf("[git clean]: [SKIPPED]  %s contains unmerged commits (run with --force to delete)\n", b.Name)
				continue
			}
			// Explicitly authorized mode
			flag = "-D"
		}

		out, err := exec.CommandContext(ctx, "git", "branch", flag, b.Name).CombinedOutput()

		if err != nil {
			fmt.Printf("[git clean]: [FAILED]   %s — %s\n", b.Name, strings.TrimSpace(string(out)))
		} else {
			fmt.Printf("[git clean]: [DELETED]  %s\n", b.Name)
		}
	}

	return nil
}
