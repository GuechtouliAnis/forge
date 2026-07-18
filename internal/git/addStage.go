package git

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// stageFiles runs `git add` on every approved file, printing a per-file
// status line as it goes and returning aggregate counts for the final
// summary. A single file's failure is non-fatal — one bad file (e.g. a
// permissions issue) should never prevent the rest of the batch from being
// staged, matching CleanGit's per-branch-failure tolerance during deletion.
func stageFiles(ctx context.Context, files []fileResult) (staged, failed int) {
	for _, f := range files {
		if ctx.Err() != nil {
			fmt.Println("\n[git add]: Interrupted — remaining files were not staged.")
			break
		}

		if err := gitAddFile(ctx, f.Path); err != nil {
			failed++
			fmt.Printf("[git add]: git add failed for '%s': %s\n", f.Path, err.Error())
			continue
		}

		staged++
		fmt.Printf("[git add]: %s staged.\n", f.Path)
	}
	return staged, failed
}

// gitAddFile stages a single file via `git add`, returning the command's
// combined output as the error text on failure so the caller can surface a
// precise, per-file reason rather than a generic failure message.
func gitAddFile(ctx context.Context, path string) error {
	out, err := exec.CommandContext(ctx, "git", "add", path).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}
