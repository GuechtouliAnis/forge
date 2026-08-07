package git

import (
	"fmt"
	"strings"
)

// promptPathSelection prints the fuzzy-match candidates and asks the user
// to pick one. Returns "" (with a nil error) if the user cancels.
func promptPathSelection(search string, matches []string) (string, error) {

	if len(matches) == 1 {
		fmt.Printf("[FOUND] %s\n", matches[0])
		return matches[0], nil
	}
	fmt.Printf("[FOUND] Multiple matches for %q:\n", search)
	for i, m := range matches {
		fmt.Printf("  [%d] %s\n", i+1, m)
	}

	fmt.Print("[git restore]: Select (0 to cancel): ")
	var choice int
	fmt.Scan(&choice)
	if choice == 0 {
		fmt.Println("[git restore]: Cancelled.")
		return "", nil
	}

	if choice < 1 || choice > len(matches) {
		return "", fmt.Errorf("[git restore]: invalid selection")
	}
	return matches[choice-1], nil
}

// confirmOverwrite warns about uncommitted changes on path and asks for
// confirmation. Returns false if the user declines.
func confirmOverwrite(path string) bool {
	fmt.Printf("[git restore]: %s has uncommitted changes. Overwrite? [y/N]: ", path)
	var input string
	fmt.Scanln(&input)
	return isYes(input)
}

// confirmIgnoredRestore warns that path is gitignored and asks for
// confirmation. Returns false if the user declines.
func confirmIgnoredRestore(path string) bool {

	fmt.Printf("[git restore]: WARNING! %s is gitignored. Restore anyway? [y/N]: ", path)
	var input string
	fmt.Scanln(&input)

	return isYes(input)
}

func isYes(input string) bool {
	return strings.ToLower(input) == "y" || strings.ToLower(input) == "yes"
}

// promptCommitSelection prints the restorable commit history for path and
// asks the user to pick one. Returns "" (with a nil error) if the user
// cancels.
func promptCommitSelection(path string, commits []commitEntry) (string, error) {

	fmt.Printf("History for %s:\n", path)
	for i, c := range commits {
		fmt.Printf("  [%d] %s\n", i+1, c.display)
	}

	fmt.Print("[git restore]: Select (0 to cancel): ")
	var choice int
	fmt.Scan(&choice)
	if choice == 0 {
		fmt.Println("[git restore]: Cancelled.")
		return "", nil
	}

	if choice < 1 || choice > len(commits) {
		return "", fmt.Errorf("[git restore]: invalid selection.")
	}

	return commits[choice-1].hash, nil
}

func printDryRun(path, hash string) {
	fmt.Printf("[git restore]: [DRY-RUN] Would restore: %s from commit %s\n", path, hash)
}

func printRestored(path, hash, stats string) {
	fmt.Printf("[git restore]: [RESTORED] %s%s from commit %s\n", path, stats, hash)
}
