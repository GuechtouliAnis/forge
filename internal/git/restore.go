package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// RestoreFile recovers a file from git history using fuzzy path matching.
// Collision detection blocks overwrites of dirty/staged files unless --force is passed.
// --latest skips the version menu and restores from the most recent commit where the file existed.
// --commit allows pinning to a specific commit hash.
func RestoreFile(search string, latest bool, force bool, dryRun bool, commitHash string) error {
	// confirm we're in a git repo
	if err := exec.Command("git", "rev-parse", "--is-inside-work-tree").Run(); err != nil {
		return fmt.Errorf("not a git repository")
	}

	// gather unique historical paths via git log
	logOut, err := exec.Command("git", "log", "--all", "--name-only", "--pretty=format:").Output()
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
		fmt.Print("Select: ")
		var choice int
		fmt.Scan(&choice)
		if choice < 1 || choice > len(matches) {
			return fmt.Errorf("invalid selection")
		}
		resolvedPath = matches[choice-1]
	}

	// collision detection — block if file is dirty or staged unless --force
	statusOut, err := exec.Command("git", "status", "--porcelain", resolvedPath).Output()
	if err != nil {
		return fmt.Errorf("could not check file status: %w", err)
	}
	if strings.TrimSpace(string(statusOut)) != "" && !force {
		return fmt.Errorf("file has uncommitted changes — use --force to overwrite")
	}

	// check if file is gitignored — warn but don't block unless --force skips prompt
	ignoredOut, _ := exec.Command("git", "check-ignore", "-q", resolvedPath).Output()
	if len(ignoredOut) > 0 && !force {
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
		hashesOut, err := exec.Command("git", "log", "--all", "--numstat",
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
		logOut, err := exec.Command("git", "log", "-n", "10", "--numstat",
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
		fmt.Print("Select: ")
		var choice int
		fmt.Scan(&choice)
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
	out, err := exec.Command("git", "checkout", targetHash, "--", resolvedPath).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git checkout failed: %s — %w", strings.TrimSpace(string(out)), err)
	}

	exec.Command("git", "restore", "--staged", resolvedPath).Run()

	// get diff stats from the selected commit directly
	statsOut, _ := exec.Command("git", "show", "--numstat", "--pretty=format:", targetHash, "--", resolvedPath).Output()
	stats := ""
	if fields := strings.Fields(strings.TrimSpace(string(statsOut))); len(fields) >= 2 {
		stats = fmt.Sprintf(" (+%s, -%s)", fields[0], fields[1])
	}

	fmt.Printf("[RESTORED] %s%s from commit %s\n", resolvedPath, stats, targetHash)
	return nil
}
