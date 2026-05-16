package env

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/GuechtouliAnis/forge/internal/config"
)

// Check severity levels.
const (
	LevelWarn  = 1
	LevelError = 2
)

// CheckIssue represents a single validation issue found in a .env file.
type CheckIssue struct {
	Line     int
	Severity int // LevelWarn or LevelError
	Message  string
	File     string
}

// CheckEnv validates a .env file and returns a list of issues.
// It checks for key naming rules, duplicate keys, malformed lines, and more.
func CheckEnv(path string, examplePath string, level int, cfg config.EnvCheck) ([]CheckIssue, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("[env check]: %w", err)
	}

	exampleCh := make(chan ExampleResult, 1)
	go func() {
		keys, err := ParseKeysFromExample(examplePath)
		exampleCh <- ExampleResult{keys, err}
	}()

	seen := make(map[string]bool)
	var issues []CheckIssue
	var prevLine string
	var consLines uint8
	var consStart int

	lineNum := 0
	scanner := bufio.NewScanner(bytes.NewReader(data))

	for scanner.Scan() {
		raw := scanner.Text()
		lineNum++

		// ? [WARN] - trailing whitespace
		if issue := CheckTrailingWhitespace(raw, lineNum); ShouldAdd(issue, level, cfg, "trailing_whitespace") {
			issues = append(issues, *issue)
		}

		// ? [WARN] - consecutive blank lines
		if strings.TrimSpace(raw) == "" {
			consLines++
			if consLines == 1 {
				consStart = lineNum
			}
			prevLine = raw
			continue
		}

		// non-blank line — flush consecutive blank run if over threshold
		if cfg.MaxConsBlanks > 0 && consLines > uint8(cfg.MaxConsBlanks) {
			issue := &CheckIssue{
				Line:     consStart,
				Severity: LevelWarn,
				Message:  fmt.Sprintf("%d consecutive blank lines (lines %d–%d)", consLines, consStart, lineNum-1),
			}
			if ShouldAdd(issue, level, cfg, "consecutive_blank_lines") {
				issues = append(issues, *issue)
			}
		}
		consLines = 0

		line := strings.TrimPrefix(raw, "export ")

		// skip comment lines, but warn if commented key has a value
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			// strip line from comment sign
			stripped := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "#"))
			// ? [WARN] - commented_key_has_value
			if issue := CommentedHasValue(stripped, lineNum); ShouldAdd(issue, level, cfg, "commented_key_has_value") {
				issues = append(issues, *issue)
			}
			continue
		}

		// skip blank lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		key, value, found := strings.Cut(line, "=")

		// ! [ERROR] - malformed line (no equal sign found)
		if !found {
			issue := &CheckIssue{
				Line:     lineNum,
				Severity: LevelError,
				Message:  fmt.Sprintf("malformed line, no '=' found: %q", strings.TrimSpace(line))}
			issues = append(issues, *issue)
			continue
		}

		trimmedKey := strings.TrimSpace(key)

		// ! [ERROR] - key contains space (API = KEY) or value has leading whitespace (KEY= value)
		if strings.ContainsAny(key, " \t") {
			issue := &CheckIssue{
				Line:     lineNum,
				Severity: LevelError,
				Message:  fmt.Sprintf("key contains spaces: %q", strings.TrimSpace(key))}
			issues = append(issues, *issue)
			continue
		}
		if value != strings.TrimLeft(value, " \t") {
			issue := &CheckIssue{
				Line:     lineNum,
				Severity: LevelError,
				Message:  fmt.Sprintf("value has leading whitespace for key: %q", trimmedKey)}
			issues = append(issues, *issue)
			continue
		}

		// ! [ERROR] - empty key (line starts with '=')
		// Errors appended directly without a level check since errors should not be ignored
		if issue := EmptyKey(trimmedKey, lineNum); issue != nil {
			issues = append(issues, *issue)
			continue
		}

		// ! [ERROR] - validate key
		switch ValidateKey(trimmedKey) {
		case KeyStartsWithDigit:
			issues = append(issues, CheckIssue{
				Line:     lineNum,
				Severity: LevelError,
				Message:  fmt.Sprintf("key starts with digit: %q", trimmedKey)})
			continue
		case KeyInvalidChars:
			issues = append(issues, CheckIssue{
				Line:     lineNum,
				Severity: LevelError,
				Message:  fmt.Sprintf("key contains invalid characters: %q", trimmedKey)})
			continue
		case KeyIsLowercase:
			// ? [WARN] - lowercase_key
			issue := LowercaseKey(trimmedKey, lineNum, cfg.AllowedLowercase)
			if ShouldAdd(issue, level, cfg, "lowercase_key") {
				issues = append(issues, *issue)
			}
		}

		// ! [ERROR] - duplicate key
		if seen[trimmedKey] {
			issue := &CheckIssue{
				Line:     lineNum,
				Severity: LevelError,
				Message:  fmt.Sprintf("duplicate key: %q", trimmedKey)}
			issues = append(issues, *issue)
		}
		seen[trimmedKey] = true

		// ? [WARN] - empty_value (KEY=)
		if issue := EmptyValue(key, value, lineNum); ShouldAdd(issue, level, cfg, "empty_value") {
			issues = append(issues, *issue)
		}

		// ! [ERROR] - unclosed quotation
		// Errors appended directly without a level check since errors should not be ignored
		trimmedVal := strings.TrimSpace(value)
		if issue := ValidateValue(trimmedKey, trimmedVal, lineNum); issue != nil {
			issues = append(issues, *issue)
		}
	}

	// ?  [WARN] - trailing blank line, check last two scanned lines
	trimmed := strings.TrimRight(string(data), "\n")
	if string(data) != trimmed && strings.TrimSpace(prevLine) == "" {
		issue := &CheckIssue{
			Line:     lineNum,
			Severity: LevelWarn,
			Message:  "file ends with blank line"}
		if ShouldAdd(issue, level, cfg, "file_ends_with_blank") {
			issues = append(issues, *issue)
		}
	}

	// ? [WARN] - conformity check, drain goroutine result
	if result := <-exampleCh; result.err == nil {
		examplePath := filepath.Join(filepath.Dir(path), examplePath)
		for k, meta := range result.keys {
			// ? [WARN] - example has a key that has a value
			if meta.HasValue {
				issue := &CheckIssue{
					Line:     0,
					Severity: LevelWarn,
					File:     examplePath,
					Message:  fmt.Sprintf("key %q has a value set — example files should use empty or placeholder values", k)}
				if ShouldAdd(issue, level, cfg, "example_has_value") {
					issues = append(issues, *issue)
				}
			}
			// ? [WARN] - key exists in .env.example but not in .env
			if !seen[k] {
				issue := &CheckIssue{
					Line:     0,
					Severity: LevelWarn,
					Message:  fmt.Sprintf("key %q exists in %s but not in %s", k, examplePath, path)}
				if ShouldAdd(issue, level, cfg, "example_conformity") {
					issues = append(issues, *issue)
				}
			}
		}
		for k := range seen {
			// ? [WARN] - key exists in .env but not in .env.example
			issue := &CheckIssue{
				Line:     0,
				Severity: LevelWarn,
				Message:  fmt.Sprintf("key %q exists in %s but not in %s", k, path, examplePath)}
			if _, exists := result.keys[k]; !exists && ShouldAdd(issue, level, cfg, "example_conformity") {
				issues = append(issues, *issue)
			}
		}
	}

	// Sort issues by line number
	sort.Slice(issues, IssuesByLine(issues))

	return issues, nil
}
