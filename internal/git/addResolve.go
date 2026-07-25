package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// resolvedFile carries a path alongside whether it represents a deletion,
// i.e. a file git tracks that no longer exists on disk. Deletions are a
// legitimate `git add` target (staging the removal) and must bypass the
// size-check pipeline entirely, since there is nothing left to stat.
type resolvedFile struct {
	Path    string
	Deleted bool
	Code    string // raw two-char porcelain status, e.g. "M ", "??", "R ", "AM"
}

// resolvePaths determines the final set of files to run through the
// guardrail pipeline: explicitly-named arguments are always included
// as-is, directory arguments are expanded via git status scoped to
// real changes only.
func resolvePaths(ctx context.Context, rawPaths []string) ([]resolvedFile, error) {

	// files/paths named directly by the user, passed through untouched,
	// no expansion needed.
	// Preallocated to len(rawPaths): in the common case (no directory args)
	// nearly every rawPath ends up here, so this avoids repeated slice
	// growth/reallocation as the loop appends.
	explicit := make([]resolvedFile, 0, len(rawPaths))

	// directory arguments — need expansion via git status to discover
	// the actual changed files inside them
	var dirArgs []string

	// Deduplicate paths across explicit args and expanded directory results,
	// so a file listed twice (or caught by both) isn't processed more than once
	seen := make(map[string]bool)

	for _, p := range rawPaths {
		clean := filepath.Clean(p)

		// Path exists on disk — either a directory to expand later,
		// or a regular file to add directly.
		if info, err := os.Stat(clean); err == nil {
			if info.IsDir() {
				dirArgs = append(dirArgs, p)
				continue
			}
			if !seen[clean] {
				seen[clean] = true
				explicit = append(explicit, resolvedFile{Path: clean})
			}
			continue
		}

		// Missing from disk — check whether git recognizes this as a
		// tracked deletion before concluding the path is simply invalid.
		deleted, gerr := explicitDeletedFile(ctx, clean)
		if gerr != nil {
			return nil, gerr
		}
		if !deleted {
			return nil, fmt.Errorf("[git add]: path does not exist: %s", p)
		}

		// Confirmed tracked deletion — valid `git add` target despite
		// not existing on disk.
		if !seen[clean] {
			seen[clean] = true
			explicit = append(explicit, resolvedFile{Path: clean, Deleted: true})
		}
	}

	// Start from explicitly-named files; only expand directories if any
	// were actually given, avoiding an unnecessary git call otherwise.
	resolved := explicit

	if len(dirArgs) > 0 {
		// Expand each directory argument into the real changed files
		// git sees inside it (git status scoped to those dirs) —
		// directories themselves are never valid `git add` targets
		// on their own, only the tracked/untracked changes within them.
		changed, err := changedFilesUnder(ctx, dirArgs)
		if err != nil {
			return nil, err
		}

		// Dedupe against files already collected from explicit args,
		// in case a directory expansion re-surfaces a path the user
		// also named directly.
		for _, c := range changed {
			if !seen[c.Path] {
				seen[c.Path] = true
				resolved = append(resolved, c)
			}
		}
	}

	return resolved, nil
}

// explicitDeletedFile asks git whether a path that no longer exists on
// disk is nonetheless a tracked deletion — i.e. the file is missing from
// the working tree but git still has a record of it, either in the index
// or the last commit. A blank status line means git has no record of the
// path at all, meaning it genuinely does not exist and INVALID_PATH is
// the correct outcome.
func explicitDeletedFile(ctx context.Context, path string) (bool, error) {

	// Ask git directly rather than inferring from the filesystem
	// to distinguish "deleted but tracked" from "never existed"/"typo"
	out, err := exec.CommandContext(ctx, "git", "status", "--porcelain=v1", "--", path).Output()
	if err != nil {
		return false, fmt.Errorf("[git add]: could not check status for %s: %w", path, err)
	}

	line := strings.TrimSpace(string(out))
	if line == "" {
		// No status output, git doesn't recognize this path at all.
		return false, nil
	}

	// Deletion is marked 'D' in either the index column (X) or the
	// working-tree column (Y), depending on whether it's already staged.
	return len(line) >= 2 && (line[0] == 'D' || line[1] == 'D'), nil
}

// changedFilesUnder returns every file git considers a staging candidate
// within the given directory paths, now carrying deletion status alongside
// each path so downstream classification can skip size checks correctly.
func changedFilesUnder(ctx context.Context, dirArgs []string) ([]resolvedFile, error) {

	// Query git status scoped to the given directories only, including
	// untracked files (--untracked-files=all), so newly-created files
	// inside a directory are discovered and not silently skipped.
	args := append([]string{"status", "--porcelain=v1", "--untracked-files=all", "--"}, dirArgs...)
	out, err := exec.CommandContext(ctx, "git", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("[git add]: could not determine changed files: %w", err)
	}

	trimmed := strings.TrimRight(string(out), "\n")
	if trimmed == "" {
		// No output — no changes found under the given directories.
		return nil, nil
	}

	var files []resolvedFile

	for _, line := range strings.Split(trimmed, "\n") {
		// Porcelain v1 lines are at minimum "XY <path>" — 2 status
		// chars + space + at least 1 path char.
		if len(line) < 4 {
			continue
		}

		// index (staged) and working-tree status codes
		x, y := line[0], line[1]
		// skip "XY " prefix to get the raw path
		path := line[3:]

		// y == ' ' means the working tree has no further change beyond
		// what's already staged (X); x != '?' excludes untracked files
		// (which always need adding). So this skips entries that are
		// already fully staged — nothing left for `git add` to do.
		if y == ' ' && x != '?' {
			continue
		}

		// Renames/copies are reported as "old -> new"; only the
		// destination path is a valid add target.
		if idx := strings.Index(path, " -> "); idx != -1 {
			path = path[idx+4:]
		}

		// Porcelain quotes paths containing special characters.
		path = strings.Trim(path, "\"")
		if path == "" {
			continue
		}

		files = append(files, resolvedFile{
			Path: path,
			// Deletion can be reported in either the index (x) or
			// working-tree (y) column depending on staging state.
			Deleted: x == 'D' || y == 'D',
			Code:    string([]byte{x, y}),
		})
	}

	return files, nil
}
