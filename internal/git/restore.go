package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
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

	// Verify we are actually inside a git repository.
	// Fails fast with a clear error otherwise.
	if err := gitCheck(ctx); err != nil {
		return err
	}

	// gather unique historical paths via git log
	logOut, err := exec.CommandContext(ctx, "git", "log", "--all", "--name-only", "--pretty=format:").Output()
	if err != nil {
		return fmt.Errorf("could not read git history: %w", err)
	}

	// deduplicate paths using a map
	seen := make(map[string]bool)
	var matches []string
	for _, line := range strings.Split(string(logOut), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || seen[line] {
			continue
		}
		seen[line] = true
		// fuzzy match — path contains the search term
		if strings.Contains(strings.ToLower(line), strings.ToLower(search)) {
			matches = append(matches, line)
		}
	}

	if len(matches) == 0 {
		return fmt.Errorf("no historical paths found matching %q", search)
	}

	// resolve target path — prompt if multiple matches
	var resolvedPath string
	if len(matches) == 1 {
		resolvedPath = matches[0]
		fmt.Printf("[FOUND] %s\n", resolvedPath)
	} else {
		fmt.Printf("[FOUND] Multiple matches for %q:\n", search)
		for i, m := range matches {
			fmt.Printf("  [%d] %s\n", i+1, m)
		}
		fmt.Print("Select (0 to cancel): ")
		var choice int
		fmt.Scan(&choice)
		if choice == 0 {
			fmt.Println("Cancelled.")
			return nil
		}
		if choice < 1 || choice > len(matches) {
			return fmt.Errorf("invalid selection")
		}
		resolvedPath = matches[choice-1]
	}

	// collision detection — prompt before overwriting a dirty/staged file
	statusOut, err := exec.CommandContext(ctx, "git", "status", "--porcelain", resolvedPath).Output()
	if err != nil {
		return fmt.Errorf("could not check file status: %w", err)
	}
	if strings.TrimSpace(string(statusOut)) != "" {
		fmt.Printf("%s has uncommitted changes. Overwrite? [y/N]: ", resolvedPath)
		var input string
		fmt.Scanln(&input)
		if strings.ToLower(input) != "y" && strings.ToLower(input) != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	// check if file is gitignored — warn but don't block
	ignoredOut, ignoreErr := exec.CommandContext(ctx, "git", "check-ignore", "-q", resolvedPath).Output()
	// check-ignore exits 1 (non-zero) when the path is simply not ignored —
	// that's expected and not a real error. Only treat it as a failure when
	// it's neither "matched" (exit 0) nor a plain "not ignored" (exit 1).
	if ignoreErr != nil {
		if exitErr, ok := ignoreErr.(*exec.ExitError); !ok || exitErr.ExitCode() > 1 {
			return fmt.Errorf("could not check .gitignore status: %w", ignoreErr)
		}
	}
	if len(ignoredOut) > 0 {
		fmt.Printf("WARNING: %s is gitignored. Restore anyway? [y/N]: ", resolvedPath)
		var input string
		fmt.Scanln(&input)
		if strings.ToLower(input) != "y" && strings.ToLower(input) != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	// resolve commit hash to restore from
	var targetHash string
	if commitHash != "" {
		// pinned to specific commit via --commit flag
		targetHash = commitHash
	} else if latest {
		// find the most recent commit where the file actually existed (added > 0)
		hashesOut, err := exec.CommandContext(ctx, "git", "log", "--all", "--numstat",
			"--pretty=format:%H", "--", resolvedPath).Output()
		if err != nil || strings.TrimSpace(string(hashesOut)) == "" {
			return fmt.Errorf("no commit history found for %s", resolvedPath)
		}
		lines := strings.Split(strings.TrimSpace(string(hashesOut)), "\n")
		for i := 0; i < len(lines); i++ {
			line := strings.TrimSpace(lines[i])
			// full commit hash line
			if len(line) == 40 && !strings.Contains(line, "\t") {
				// look ahead for numstat line
				for j := i + 1; j < len(lines); j++ {
					s := strings.TrimSpace(lines[j])
					if s == "" {
						continue
					}
					fields := strings.Fields(s)
					if len(fields) >= 2 && fields[0] != "0" {
						// file had additions in this commit — not a deletion commit
						targetHash = line
						break
					}
					break
				}
			}
			if targetHash != "" {
				break
			}
		}
		if targetHash == "" {
			return fmt.Errorf("no restorable commit found for %s (all commits deleted it)", resolvedPath)
		}
	} else {
		// fetch last 10 commits that touched this file with stats
		logOut, err := exec.CommandContext(ctx, "git", "log", "-n", "10", "--numstat",
			"--pretty=format:%h|%cr|%s", "--", resolvedPath).Output()
		if err != nil || strings.TrimSpace(string(logOut)) == "" {
			return fmt.Errorf("no commit history found for %s", resolvedPath)
		}

		// parse and display commit history for selection
		type commitEntry struct {
			hash    string
			display string
			deleted bool
		}
		var commits []commitEntry
		lines := strings.Split(strings.TrimSpace(string(logOut)), "\n")
		for i := 0; i < len(lines); i++ {
			line := strings.TrimSpace(lines[i])
			if !strings.Contains(line, "|") {
				continue
			}
			parts := strings.SplitN(line, "|", 3)
			if len(parts) < 3 {
				continue
			}
			hash, age, subject := parts[0], parts[1], parts[2]
			// next non-empty line should be the numstat
			stats := ""
			deleted := false
			for j := i + 1; j < len(lines); j++ {
				s := strings.TrimSpace(lines[j])
				if s != "" {
					fields := strings.Fields(s)
					if len(fields) >= 2 {
						stats = fmt.Sprintf("+%s / -%s", fields[0], fields[1])
						if fields[0] == "0" && fields[1] != "0" {
							deleted = true
						}
					}
					break
				}
			}
			label := ""
			if deleted {
				label = " ⚠ deleted in this commit"
			}
			commits = append(commits, commitEntry{
				hash:    hash,
				display: fmt.Sprintf("%s — %s — %s %s%s", hash, age, subject, stats, label),
				deleted: deleted,
			})
		}

		if len(commits) == 0 {
			return fmt.Errorf("no commit history found for %s", resolvedPath)
		}
		// filter out deletion commits
		var restorable []commitEntry
		for _, c := range commits {
			if !c.deleted {
				restorable = append(restorable, c)
			}
		}
		if len(restorable) == 0 {
			return fmt.Errorf("no restorable commits found for %s", resolvedPath)
		}

		fmt.Printf("History for %s:\n", resolvedPath)
		for i, c := range restorable {
			fmt.Printf("  [%d] %s\n", i+1, c.display)
		}
		fmt.Print("Select (0 to cancel): ")
		var choice int
		fmt.Scan(&choice)
		if choice == 0 {
			fmt.Println("Cancelled.")
			return nil
		}
		if choice < 1 || choice > len(restorable) {
			return fmt.Errorf("invalid selection")
		}

		targetHash = restorable[choice-1].hash
	}

	// dry-run — just show what was found, no restoration
	if dryRun {
		fmt.Printf("[DRY-RUN] Would restore: %s from commit %s\n", resolvedPath, targetHash)
		return nil
	}

	// restore the file from the resolved commit
	out, err := exec.CommandContext(ctx, "git", "checkout", targetHash, "--", resolvedPath).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git checkout failed: %s — %w", strings.TrimSpace(string(out)), err)
	}

	if unstageOut, err := exec.CommandContext(ctx, "git", "restore", "--staged", resolvedPath).CombinedOutput(); err != nil {
		return fmt.Errorf("restored file but failed to unstage it: %s — %w", strings.TrimSpace(string(unstageOut)), err)
	}

	// get diff stats from the selected commit directly
	statsOut, _ := exec.CommandContext(ctx, "git", "show", "--numstat", "--pretty=format:",
		targetHash, "--", resolvedPath).Output()
	stats := ""
	if fields := strings.Fields(strings.TrimSpace(string(statsOut))); len(fields) >= 2 {
		stats = fmt.Sprintf(" (+%s, -%s)", fields[0], fields[1])
	}

	fmt.Printf("[RESTORED] %s%s from commit %s\n", resolvedPath, stats, targetHash)
	return nil
}
