package env

import "fmt"

// FormatIssue formats a CheckIssue into a human-readable string.
func FormatIssue(path string, issue CheckIssue) string {
	prefix := "[warn] "
	if issue.Severity == LevelError {
		prefix = "[error]"
	}
	file := path
	if issue.File != "" {
		file = issue.File
	}
	if issue.Line == 0 {
		return fmt.Sprintf("%s - %s - %s", prefix, file, issue.Message)
	}
	return fmt.Sprintf("%s - %s:%d - %s", prefix, file, issue.Line, issue.Message)
}
