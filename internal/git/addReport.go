package git

import (
	"fmt"
	"strings"
)

// countState returns how many results in the slice carry the given state.
func countState(results []fileResult, state fileState) int {
	n := 0
	for _, r := range results {
		if r.State == state {
			n++
		}
	}
	return n
}

// printDryRunSummary renders the structured preview required by the
// dry_run flow: every file's path, its would-be status, and its size,
// followed by aggregate counts. Nothing in either slice has been touched —
// this is purely informational output.
func printDryRunSummary(settled []fileResult, wouldStage []fileResult) {
	// Nothing to evaluate at all — skip the table entirely rather than
	// printing an empty header followed immediately by a zero-count
	// summary. This happens when the working tree is fully clean within
	// the given paths (e.g. everything already staged with no further
	// changes, after the porcelain Y-status filtering in changedFilesUnder).
	if len(settled) == 0 && len(wouldStage) == 0 {
		fmt.Println("[git add]: Nothing to stage — working tree clean within the given path(s).")
		return
	}

	fmt.Printf("\n%-50s %-6s %-15s %s\n", "FILE", "CODE", "STATUS", "SIZE")
	fmt.Println(strings.Repeat("-", 80))

	for _, r := range settled {
		status := "blocked"
		switch r.State {
		case stateSkippedLarge:
			status = "would-skip"
		case stateFailedLarge:
			status = "would-fail"
		}
		fmt.Printf("%-50s %-6s %-15s %.1f MB\n", r.Path, r.Code, status, r.SizeMB)
	}
	for _, r := range wouldStage {
		fmt.Printf("%-50s %-6s %-15s %.1f MB\n", r.Path, r.Code, "would-stage", r.SizeMB)
	}

	blocked := countState(settled, stateBlockedHardcoded) + countState(settled, stateBlockedPattern)
	skipped := countState(settled, stateSkippedLarge)

	fmt.Printf("\n[git add]: Dry-run complete. No files staged.\n%d would-stage, %d blocked, %d skipped.\n",
		len(wouldStage), blocked, skipped)
}
