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
)

// CleanGit scans local branches and suggests or removes stale ones.
// Dry-run is the default — --remove triggers deletion with confirmation, --force skips it.
func CleanGit(cfg config.GitClean, remove bool, force bool, offlineFlag bool) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	offline := offlineFlag || !cfg.FetchRemote

	days := cfg.StaleDays
	behind := cfg.CommitsBehind

	// pre-flight: confirm we're in a git repo
	if err := gitCheck(ctx); err != nil {
		return err
	}

	// Synchronize with the remote and remove local references to branches that no longer exist on the server.
	// This is skipped entirely when fetch_remote=false or --offline is passed. If a fetch is attempted and
	// fails, it is treated as non-fatal: we fall back to local-only diagnosis using existing refs.
	if !offline {
		if fallback := gitFetch(ctx, cfg); fallback {
			offline = true
		}
	} else {
		fmt.Println("fetch_remote is disabled — doing local diagnosis only.")
	}

	if ctx.Err() != nil {
		return reportInterrupted(0, 0)
	}

	defaultBranch := defaultBranch(ctx)

	if offline {
		if info, err := os.Stat(".git/FETCH_HEAD"); err == nil {
			fmt.Printf("Last successful fetch: %s ago.\n\n", time.Since(info.ModTime()).Round(time.Minute))
		} else {
			fmt.Println("No record of a prior successful fetch — remote-tracking data may be unreliable.")
			fmt.Println()
		}
	}

	currentOut, err := exec.CommandContext(ctx, "git", "branch", "--show-current").Output()
	if err != nil {
		return fmt.Errorf("[git clean]: could not determine current branch: %w", err)
	}
	currentBranch := strings.TrimSpace(string(currentOut))

	// Batch-query all branches and their absolute ages.
	// Fields are NUL-separated (%00) rather than a printable delimiter like "|",
	// since NUL is one of the few bytes git-check-ref-format guarantees cannot
	// appear inside a ref name — this is the same convention git's own plumbing
	// commands use (git diff -z, git ls-files -z) to make output unambiguous
	// regardless of what characters a name legally contains.
	refOut, err := exec.CommandContext(ctx, "git", "for-each-ref",
		"--format=%(refname:short)%00%(committerdate:unix)", "refs/heads/").Output()
	if err != nil {
		return fmt.Errorf("[git clean]: could not list branches: %w", err)
	}

	branchLines := strings.Split(strings.TrimSpace(string(refOut)), "\n")
	if len(branchLines) == 0 || (len(branchLines) == 1 && branchLines[0] == "") {
		fmt.Println("No branches to clean.")
		return nil
	}

	staleMap := make(map[string]*BranchInfo)
	var branchNames []string
	now := time.Now()

	for _, line := range branchLines {
		parts := strings.SplitN(line, "\x00", 2)
		if len(parts) != 2 {
			continue
		}
		name := parts[0]
		tsStr := parts[1]

		info := &BranchInfo{Name: name}
		if ts, err := strconv.ParseInt(tsStr, 10, 64); err == nil {
			info.DaysOld = int(now.Sub(time.Unix(ts, 0)).Hours() / 24)
		}
		staleMap[name] = info
		branchNames = append(branchNames, name)
	}

	if len(branchNames) <= 1 {
		fmt.Println("No branches to clean.")
		return nil
	}

	// Batch-query merged status
	mergedOut, err := exec.CommandContext(ctx, "git", "branch", "--format=%(refname:short)", "--merged", defaultBranch).Output()
	if err == nil {
		for _, name := range strings.Split(strings.TrimSpace(string(mergedOut)), "\n") {
			name = strings.TrimSpace(name)
			if info, exists := staleMap[name]; exists {
				info.Merged = true
			}
		}
	}

	// Classify and queue branches for the "Behind" check.
	var toEvaluate []string

	for _, name := range branchNames {
		info := staleMap[name]
		switch name {
		case "main", "master", defaultBranch:
			info.Protected = true
		case currentBranch:
			info.Current = true
		default:
			toEvaluate = append(toEvaluate, name)
		}
	}

	// Evaluate "Commits Behind" concurrently.
	// Workers only ever write to the `results` channel; staleMap is mutated
	// exclusively by this goroutine as it drains `results`, so there is no
	// concurrent write to shared state — provably race-free regardless of
	// worker count.
	if len(toEvaluate) > 0 {
		numWorkers := computeWorkerCount(len(toEvaluate), cfg.MaxWorkers)
		jobs := make(chan string)
		results := make(chan behindResult, len(toEvaluate))
		var wg sync.WaitGroup
		wg.Add(numWorkers)

		for w := 0; w < numWorkers; w++ {
			go func() {
				defer wg.Done()
				for name := range jobs {
					select {
					case <-ctx.Done():
						results <- behindResult{name: name, interrupted: true}
					default:
						behindCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
						behindOut, err := exec.CommandContext(behindCtx, "git", "rev-list", "--count", name+".."+defaultBranch).Output()
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

		go func() {
			wg.Wait()
			close(results)
		}()

		for r := range results {
			info := staleMap[r.name]
			info.Behind = r.behind
			info.Interrupted = r.interrupted
		}
	}

	// Reassemble slice to preserve iteration stability and original structure
	var stale []BranchInfo
	for _, name := range branchNames {
		stale = append(stale, *staleMap[name])
	}

	if ctx.Err() != nil {
		evaluated := 0
		for _, b := range stale {
			if !b.Interrupted {
				evaluated++
			}
		}
		return reportInterrupted(evaluated, len(stale))
	}

	fmt.Printf("\n%-30s %-10s %-10s %-10s %s\n", "BRANCH", "DAYS OLD", "BEHIND", "MERGED", "STATUS")
	fmt.Println(strings.Repeat("-", 75))

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

	if len(toDelete) == 0 {
		fmt.Println("No stale branches found.")
		return nil
	}

	if !remove {
		fmt.Printf("%d stale branch(es) found. Run with --remove to delete.\n", len(toDelete))
		return nil
	}

	if !force {
		fmt.Printf("Delete %d branch(es)? [y/N]: ", len(toDelete))
		var input string
		fmt.Scanln(&input)
		if strings.ToLower(input) != "y" && strings.ToLower(input) != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	for _, b := range toDelete {
		flag := "-d"
		if !b.Merged {
			flag = "-D"
		}
		out, err := exec.CommandContext(ctx, "git", "branch", flag, b.Name).CombinedOutput()
		if err != nil {
			fmt.Printf("[FAILED]  %s — %s\n", b.Name, strings.TrimSpace(string(out)))
		} else {
			fmt.Printf("[DELETED] %s\n", b.Name)
		}
	}

	return nil
}
