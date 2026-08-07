package git

import (
	"fmt"
	"strings"
)

// printCommitDryRunSummary renders the commit dry-run preview: the message
// that would be used, and a table of staged files with their status codes.
// Nothing here mutates git state.
func printCommitDryRunSummary(staged []StagedFile, message string, amend bool) {
	if amend {
		fmt.Printf("[git commit]: [dry-run] would amend previous commit with message: %s\n", message)
	} else {
		fmt.Printf("[git commit]: [dry-run] would commit with message: %s\n", message)
	}

	if len(staged) == 0 {
		if amend {
			fmt.Println("[git commit]: [dry-run] no staged changes — only the message would change")
		}
		return
	}

	fmt.Printf("\n%-50s %s\n", "FILE", "CODE")
	fmt.Println(strings.Repeat("-", 60))
	for _, f := range staged {
		fmt.Printf("%-50s %s\n", f.Path, f.Code)
	}
	fmt.Printf("\n[git commit]: Dry-run complete. %d file(s) would be included.\n", len(staged))
}
