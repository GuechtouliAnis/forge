package git

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// BranchInfo holds metadata about a branch collected during the clean pass.
type BranchInfo struct {
	Name      string
	DaysOld   int
	Behind    int
	Merged    bool
	Current   bool
	Protected bool
}

// CleanGit scans local branches and suggests or removes stale ones.
// Dry-run is the default — --remove triggers deletion with confirmation, --force skips it.
func CleanGit(days int, behind int, remove bool, force bool) error {
	// pre-flight: confirm we're in a git repo
	if err := exec.Command("git", "rev-parse", "--is-inside-work-tree").Run(); err != nil {
		return fmt.Errorf("not a git repository")
	}

	// pre-flight: check git version >= 2.0
	verOut, err := exec.Command("git", "--version").Output()
	if err != nil {
		return fmt.Errorf("could not determine git version")
	}
	// Clean the "git version X.Y.Z" output to extract just the version number string
	verStr := strings.TrimPrefix(strings.TrimSpace(string(verOut)), "git version ")

	// Isolate the major version (e.g., "2" from "2.40.1") and ensure it meets the minimum requirement
	major, err := strconv.Atoi(strings.Split(verStr, ".")[0])
	if err != nil || major < 2 {
		return fmt.Errorf("git 2.0+ required, found: %s", verStr)
	}

	// Synchronize with the remote and remove local references to branches that no longer exist on the server
	fetch := exec.Command("git", "fetch", "--prune")

	// Redirect git's output and errors directly to the terminal so the user sees real-time progress
	fetch.Stdout = os.Stdout
	fetch.Stderr = os.Stderr
	if err := fetch.Run(); err != nil {
		return fmt.Errorf("git fetch --prune failed: %w", err)
	}

	// identify default branch
	defaultBranch := defaultBranch()

	// identify current branch — protect from self-deletion
	currentOut, err := exec.Command("git", "branch", "--show-current").Output()
	if err != nil {
		return fmt.Errorf("could not determine current branch: %w", err)
	}
	currentBranch := strings.TrimSpace(string(currentOut))

	// Retrieve a clean list of local branch names, excluding prefixes like 'refs/heads/'
	branchOut, err := exec.Command("git", "branch", "--format=%(refname:short)").Output()
	if err != nil {
		return fmt.Errorf("could not list branches: %w", err)
	}

	// Parse the newline-separated output into a slice, trimming whitespace and skipping empty strings
	branches := []string{}
	for _, b := range strings.Split(strings.TrimSpace(string(branchOut)), "\n") {
		b = strings.TrimSpace(b)
		if b != "" {
			branches = append(branches, b)
		}
	}

	// If only one branch exists (default branch), nothing to evaluate for deletion
	if len(branches) <= 1 {
		fmt.Println("No branches to clean.")
		return nil
	}

	// evaluate each branch
	var stale []BranchInfo
	for _, name := range branches {
		info := BranchInfo{Name: name}

		// main, master and the default branch are protected branches that are excluded from forge git clean.
		if name == "main" || name == "master" || name == defaultBranch {
			info.Protected = true
			stale = append(stale, info)
			continue
		}

		// Identify if the branch is currently checked out; we skip analysis for the
		// active branch to prevent accidental deletion and keep the workspace stable.
		if name == currentBranch {
			info.Current = true
			stale = append(stale, info)
			continue
		}

		// days since last commit on branch
		dateOut, err := exec.Command("git", "log", "-1", "--format=%ct", name).Output()
		if err == nil {
			ts, err := strconv.ParseInt(strings.TrimSpace(string(dateOut)), 10, 64)
			if err == nil {
				info.DaysOld = int(time.Since(time.Unix(ts, 0)).Hours() / 24)
			}
		}

		// commits behind default branch
		behindOut, err := exec.Command("git", "rev-list", "--count", name+".."+defaultBranch).Output()
		if err == nil {
			info.Behind, _ = strconv.Atoi(strings.TrimSpace(string(behindOut)))
		}

		// merged check
		mergeErr := exec.Command("git", "merge-base", "--is-ancestor", name, defaultBranch).Run()
		info.Merged = mergeErr == nil

		stale = append(stale, info)
	}

	// print table
	fmt.Printf("\n%-30s %-10s %-10s %-10s %s\n", "BRANCH", "DAYS OLD", "BEHIND", "MERGED", "STATUS")
	fmt.Println(strings.Repeat("-", 75))

	var toDelete []BranchInfo
	for _, b := range stale {
		switch {
		case b.Protected:
			fmt.Printf("%-30s %-10s %-10s %-10s %s\n", b.Name, "-", "-", "-", "[PROTECTED]")
		case b.Current:
			fmt.Printf("%-30s %-10s %-10s %-10s %s\n", b.Name, "-", "-", "-", "[CURRENT — skipped]")
		case b.DaysOld >= days || b.Behind >= behind:
			status := "[TO BE DELETED]"
			if !remove {
				status = "[STALE]"
			}
			fmt.Printf("%-30s %-10d %-10d %-10v %s\n", b.Name, b.DaysOld, b.Behind, b.Merged, status)
			toDelete = append(toDelete, b)
		default:
			fmt.Printf("%-30s %-10d %-10d %-10v %s\n", b.Name, b.DaysOld, b.Behind, b.Merged, "[OK]")
		}
	}

	fmt.Println()

	if len(toDelete) == 0 {
		fmt.Println("No stale branches found.")
		return nil
	}

	// dry-run: just inform
	if !remove {
		fmt.Printf("%d stale branch(es) found. Run with --remove to delete.\n", len(toDelete))
		return nil
	}

	// --remove: confirm unless --force
	if !force {
		fmt.Printf("Delete %d branch(es)? [y/N]: ", len(toDelete))
		var input string
		fmt.Scanln(&input)
		if strings.ToLower(input) != "y" && strings.ToLower(input) != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	// delete
	for _, b := range toDelete {
		flag := "-d"
		if !b.Merged {
			flag = "-D" // force delete unmerged branches
		}
		out, err := exec.Command("git", "branch", flag, b.Name).CombinedOutput()
		if err != nil {
			fmt.Printf("[FAILED]  %s — %s\n", b.Name, strings.TrimSpace(string(out)))
		} else {
			fmt.Printf("[DELETED] %s\n", b.Name)
		}
	}

	return nil
}

// defaultBranch attempts to detect the repo default branch from remote HEAD.
func defaultBranch() string {
	out, err := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD").Output()
	if err == nil {
		parts := strings.Split(strings.TrimSpace(string(out)), "/")
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
	}
	// fallback
	return "main"
}
