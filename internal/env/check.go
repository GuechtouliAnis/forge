package env

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
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
func CheckEnv(path string, level int) ([]CheckIssue, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("[env check]: %w", err)
	}

	// kick off .env.example read concurrently
	type exampleResult struct {
		keys map[string]ExampleKey
		err  error
	}

	exampleCh := make(chan exampleResult, 1)
	go func() {
		examplePath := filepath.Join(filepath.Dir(path), ".env.example")
		keys, err := parseKeysFromExample(examplePath)
		exampleCh <- exampleResult{keys, err}
	}()

	seen := make(map[string]bool)
	var issues []CheckIssue

	add := func(lineNum, severity int, msg string) {
		if severity >= level {
			issues = append(issues, CheckIssue{Line: lineNum, Severity: severity, Message: msg})
		}
	}

	var prevLine string
	lineNum := 0
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		raw := scanner.Text()
		lineNum++

		if raw != strings.TrimRight(raw, " \t") {
			add(lineNum, LevelWarn, "trailing whitespace")
		}

		// warn: two consecutive blank lines
		if lineNum > 1 && strings.TrimSpace(raw) == "" && strings.TrimSpace(prevLine) == "" {
			add(lineNum, LevelWarn, "consecutive blank lines")
		}
		prevLine = raw

		line := strings.TrimPrefix(raw, "export ")

		// skip comment lines, but warn if commented key has a value
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			// strip line from comment sign
			stripped := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "#"))
			if key, value, found := strings.Cut(stripped, "="); found {
				if key == "" {
					continue
				}
				key = strings.TrimSpace(key)
				value = strings.TrimSpace(value)
				if hashIdx := strings.Index(value, "#"); hashIdx >= 0 {
					value = strings.TrimSpace(value[:hashIdx])
				}
				if value != "" {
					add(lineNum, LevelWarn, fmt.Sprintf("commented key %q has a value", key))
				}
			}
			continue
		}

		// skip blank lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			add(lineNum, LevelError, fmt.Sprintf("malformed line, no '=' found: %q", strings.TrimSpace(line)))
			continue
		}

		// error: key contains space (API = KEY)
		if strings.ContainsAny(key, " \t") {
			add(lineNum, LevelError, fmt.Sprintf("key contains spaces: %q", key))
			continue
		}

		trimmedKey := strings.TrimSpace(key)

		// guard: key is empty (line starts with '=')
		if trimmedKey == "" {
			add(lineNum, LevelError, "malformed line, empty key")
			continue
		}

		switch ValidateKey(trimmedKey) {
		case KeyStartsWithDigit:
			add(lineNum, LevelError, fmt.Sprintf("key starts with digit: %q", trimmedKey))
			// don't add to seen
			continue
		case KeyInvalidChars:
			add(lineNum, LevelError, fmt.Sprintf("key contains invalid characters: %q", trimmedKey))
			// don't add to seen
			continue
		case KeyIsLowercase:
			add(lineNum, LevelWarn, fmt.Sprintf("key contains lowercase: %q", trimmedKey))
		}

		// error: duplicate key
		if seen[trimmedKey] {
			add(lineNum, LevelError, fmt.Sprintf("duplicate key: %q", trimmedKey))
		}
		seen[trimmedKey] = true

		// warn: empty value (KEY=)
		checkEmpty := strings.TrimSpace(value)
		if hashIdx := strings.Index(checkEmpty, "#"); hashIdx >= 0 {
			checkEmpty = strings.TrimSpace(checkEmpty[:hashIdx])
		}
		if checkEmpty == "" {
			add(lineNum, LevelWarn, fmt.Sprintf("empty value for key: %q", trimmedKey))
			continue
		}

		// error: unclosed quotation
		trimmedVal := strings.TrimSpace(value)
		if len(trimmedVal) > 0 && (trimmedVal[0] == '"' || trimmedVal[0] == '\'') {
			quote := trimmedVal[0]
			if strings.IndexByte(trimmedVal[1:], quote) < 0 {
				add(lineNum, LevelError, fmt.Sprintf("unclosed quote for key: %q", trimmedKey))
				continue
			}
		} else {
			checkVal := trimmedVal
			if hashIdx := strings.Index(trimmedVal, "#"); hashIdx >= 0 {
				checkVal = strings.TrimSpace(trimmedVal[:hashIdx])
			}
			if strings.ContainsAny(checkVal, " \t") {
				add(lineNum, LevelError, fmt.Sprintf("unquoted value contains spaces for key: %q", trimmedKey))
			}
		}
	}

	// trailing blank line — check last two scanned lines
	trimmed := strings.TrimRight(string(data), "\n")
	if string(data) != trimmed && strings.TrimSpace(prevLine) == "" {
		add(lineNum, LevelWarn, "file ends with blank line")
	}

	// conformity check — drain goroutine result
	if result := <-exampleCh; result.err == nil {
		examplePath := filepath.Join(filepath.Dir(path), ".env.example")
		for k, meta := range result.keys {
			if meta.HasValue {
				issues = append(issues, CheckIssue{
					Line:     0,
					Severity: LevelWarn,
					File:     examplePath,
					Message:  fmt.Sprintf("key %q has a value set — example files should use empty or placeholder values", k),
				})
			}
			if !seen[k] {
				add(0, LevelWarn, fmt.Sprintf("key %q exists in .env.example but not in .env", k))
			}
		}
		for k := range seen {
			if _, exists := result.keys[k]; !exists {
				add(0, LevelWarn, fmt.Sprintf("key %q exists in .env but not in .env.example", k))
			}
		}
	}

	sort.Slice(issues, func(i, j int) bool {
		if issues[i].Line == 0 {
			return false
		}
		if issues[j].Line == 0 {
			return true
		}
		return issues[i].Line < issues[j].Line
	})

	return issues, nil
}

type ExampleKey struct {
	HasValue bool
}

// parseKeysFromExample reads a .env.example file and returns a map of keys.
// HasValue is true if a key has an actual value set — which it shouldn't in an example file.
func parseKeysFromExample(path string) (map[string]ExampleKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("[env check]: %w", err)
	}
	keys := make(map[string]ExampleKey)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimPrefix(scanner.Text(), "export ")
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		if key, value, found := strings.Cut(line, "="); found {
			key = strings.TrimSpace(key)
			if h := strings.Index(value, "#"); h >= 0 {
				value = strings.TrimSpace(value[:h])
			}
			keys[key] = ExampleKey{HasValue: strings.TrimSpace(value) != ""}
		}
	}
	return keys, nil
}

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
