package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// resolvePaths determines the final set of files to run through the
// guardrail pipeline. Two argument kinds are handled differently:
//
//   - A directly-named file is always included, exactly as given —
//     regardless of whether git considers it changed, untracked, or
//     ignored. This matters specifically for files like .env: they are
//     almost always gitignored, so relying on `git status` to surface
//     them would let the primary target of this command's guardrails
//     silently bypass evaluation entirely.
//   - A directory argument is expanded via `git status --porcelain`,
//     scoped to that directory, so we only surface files with actual
//     changes rather than walking (and reporting on) the entire tree.
func resolvePaths(ctx context.Context, rawPaths []string) ([]string, error) {
	var explicitFiles []string
	var dirArgs []string

	for _, p := range rawPaths {
		info, err := os.Stat(filepath.Clean(p))
		if err != nil {
			return nil, fmt.Errorf("[git add]: path does not exist: %s", p)
		}
		if info.IsDir() {
			dirArgs = append(dirArgs, p)
		} else {
			explicitFiles = append(explicitFiles, filepath.Clean(p))
		}
	}

	resolved := append([]string{}, explicitFiles...)
	seen := make(map[string]bool)
	for _, f := range resolved {
		seen[f] = true
	}

	if len(dirArgs) > 0 {
		changed, err := changedFilesUnder(ctx, dirArgs)
		if err != nil {
			return nil, err
		}
		for _, f := range changed {
			if !seen[f] {
				seen[f] = true
				resolved = append(resolved, f)
			}
		}
	}

	return resolved, nil
}

// changedFilesUnder returns every file git considers a *staging candidate*
// within the given directory paths — i.e. files with unstaged changes
// (modified, deleted) or untracked files. It intentionally excludes files
// that are already staged with no further working-tree changes: git's
// porcelain status reports those too (since it diffs both index-vs-HEAD
// and worktree-vs-index), but re-adding an already-fully-staged file is a
// no-op, and listing it as "would-stage" is misleading — nothing would
// actually happen if the command proceeded.
//
// git status --porcelain=v1 lines are formatted as "XY PATH":
//
//	X = status relative to the index (staged state)
//	Y = status relative to the working tree (unstaged state)
//
// A file with Y == ' ' has no working-tree changes beyond what's already
// staged — that's the case we filter out. Untracked files report as "??",
// where X == Y == '?', and are always included since nothing about them
// is staged yet.
func changedFilesUnder(ctx context.Context, dirArgs []string) ([]string, error) {
	args := append([]string{"status", "--porcelain=v1", "--untracked-files=all", "--"}, dirArgs...)
	out, err := exec.CommandContext(ctx, "git", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("[git add]: could not determine changed files: %w", err)
	}

	trimmed := strings.TrimRight(string(out), "\n")
	if trimmed == "" {
		return nil, nil
	}

	var files []string
	for _, line := range strings.Split(trimmed, "\n") {
		if len(line) < 4 {
			continue
		}

		x, y := line[0], line[1]
		path := line[3:]

		// Skip files with no working-tree-level change to stage. Untracked
		// files ("??") are the one case where X itself signals "nothing
		// staged yet" rather than "already staged" — always include those.
		if y == ' ' && x != '?' {
			continue
		}

		if idx := strings.Index(path, " -> "); idx != -1 {
			path = path[idx+4:]
		}
		path = strings.Trim(path, "\"")
		if path != "" {
			files = append(files, path)
		}
	}
	return files, nil
}
