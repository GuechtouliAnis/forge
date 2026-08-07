package git

import (
	"fmt"
	"strings"
)

// commitEntry represents a single commit in a file's history, annotated
// with whether it deleted the file (so restore can skip it).
type commitEntry struct {
	hash    string
	display string
	deleted bool
}

// findMatchingPaths deduplicates the paths in logOutput and returns those
// whose path contains search (case-insensitive fuzzy match).
func findMatchingPaths(logOutput string, search string) []string {

	seen := make(map[string]bool)

	var matches []string

	needle := strings.ToLower(search)

	for _, line := range strings.Split(logOutput, "\n") {

		line = strings.TrimSpace(line)
		if line == "" || seen[line] {
			continue
		}

		seen[line] = true
		if strings.Contains(strings.ToLower(line), needle) {
			matches = append(matches, line)
		}
	}
	return matches
}

// parseLatestRestorableCommit scans `--numstat --pretty=format:%H` output
// and returns the first commit hash whose numstat line shows additions
// (i.e. is not a deletion commit).
func parseLatestRestorableCommit(raw string) string {

	lines := strings.Split(strings.TrimSpace(raw), "\n")

	for i := 0; i < len(lines); i++ {

		line := strings.TrimSpace(lines[i])
		if len(line) != 40 || strings.Contains(line, "\t") {
			continue
		}

		for j := i + 1; j < len(lines); j++ {
			s := strings.TrimSpace(lines[j])
			if s == "" {
				continue
			}

			fields := strings.Fields(s)
			if len(fields) >= 2 && fields[0] != "0" {
				return line
			}
			break
		}
	}
	return ""
}

// parseCommitHistory turns raw `%h|%cr|%s` + numstat output (from
// `git log --numstat`) into commitEntry values, flagging any commit that
// deleted the file.
func parseCommitHistory(raw string) []commitEntry {

	var commits []commitEntry

	lines := strings.Split(strings.TrimSpace(raw), "\n")

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
			label = " !! deleted in this commit"
		}

		commits = append(commits, commitEntry{
			hash:    hash,
			display: fmt.Sprintf("%s — %s — %s %s%s", hash, age, subject, stats, label),
			deleted: deleted,
		})
	}

	return commits
}

// restorableCommits filters out commits that deleted the file.
func restorableCommits(commits []commitEntry) []commitEntry {

	var restorable []commitEntry

	for _, c := range commits {
		if !c.deleted {
			restorable = append(restorable, c)
		}
	}

	return restorable
}
