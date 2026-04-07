package env

import (
	"fmt"
	"os"
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
}

// CheckEnv validates a .env file and returns a list of issues.
// It checks for key naming rules, duplicate keys, malformed lines, and more.
func CheckEnv(path string, level int) ([]CheckIssue, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	seen := make(map[string]bool)
	var issues []CheckIssue

	add := func(lineNum, severity int, msg string) {
		if severity >= level {
			issues = append(issues, CheckIssue{Line: lineNum, Severity: severity, Message: msg})
		}
	}

	for i, raw := range lines {
		lineNum := i + 1

		// warn: trailing whitespace
		if raw != strings.TrimRight(raw, " \t") {
			add(lineNum, LevelWarn, "trailing whitespace")
		}

		// warn: two consecutive blank lines
		if i > 0 && strings.TrimSpace(raw) == "" && strings.TrimSpace(lines[i-1]) == "" {
			add(lineNum, LevelWarn, "consecutive blank lines")
		}

		line := strings.TrimPrefix(raw, "export ")

		// skip comment lines, but warn if commented key has a value
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			stripped := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "#"))
			eqIdx := strings.Index(stripped, "=")
			if eqIdx > 0 {
				key := strings.TrimSpace(stripped[:eqIdx])
				value := strings.TrimSpace(stripped[eqIdx+1:])
				// strip inline comment from value
				hashIdx := strings.Index(value, "#")
				if hashIdx >= 0 {
					value = strings.TrimSpace(value[:hashIdx])
				}
				if value != "" {
					add(lineNum, LevelWarn, fmt.Sprintf("commented key has value: %q — intentional?", key))
				}
			}
			continue
		}

		// skip blank lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		eqIdx := strings.Index(line, "=")

		// error: malformed line, no = found (KEY alone)
		if eqIdx < 0 {
			add(lineNum, LevelError, fmt.Sprintf("malformed line, no '=' found: %q", strings.TrimSpace(line)))
			continue
		}

		key := line[:eqIdx]
		value := line[eqIdx+1:]

		// error: key contains space (API = KEY)
		if strings.ContainsAny(key, " \t") {
			add(lineNum, LevelError, fmt.Sprintf("key contains spaces: %q", key))
			continue
		}

		trimmedKey := strings.TrimSpace(key)

		switch ValidateKey(trimmedKey) {
		case KeyStartsWithDigit:
			add(lineNum, LevelError, fmt.Sprintf("key starts with digit: %q", trimmedKey))
		case KeyInvalidChars:
			add(lineNum, LevelError, fmt.Sprintf("key contains invalid characters: %q", trimmedKey))
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
		hashIdx := strings.Index(checkEmpty, "#")
		if hashIdx >= 0 {
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
			closeIdx := strings.IndexByte(trimmedVal[1:], quote)
			if closeIdx < 0 {
				add(lineNum, LevelError, fmt.Sprintf("unclosed quote for key: %q", trimmedKey))
				continue
			}
		}

		// error: unquoted value with spaces (KEY=hello world)
		if !strings.HasPrefix(trimmedVal, "\"") && !strings.HasPrefix(trimmedVal, "'") {
			// strip inline comment first
			hashIdx := strings.Index(trimmedVal, "#")
			checkVal := trimmedVal
			if hashIdx >= 0 {
				checkVal = strings.TrimSpace(trimmedVal[:hashIdx])
			}
			if strings.ContainsAny(checkVal, " \t") {
				add(lineNum, LevelError, fmt.Sprintf("unquoted value contains spaces for key: %q", trimmedKey))
			}
		}
	}

	// warn: ending blank line
	if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		add(len(lines), LevelWarn, "file ends with blank line")
	}

	// warn: conformity with .env.example (skip silently if not found)
	exampleKeys, err := parseKeysFromExample(".env.example")
	if err == nil {
		for k := range seen {
			if !exampleKeys[k] {
				add(0, LevelWarn, fmt.Sprintf("key %q exists in .env but not in .env.example", k))
			}
		}
		for k := range exampleKeys {
			if !seen[k] {
				add(0, LevelWarn, fmt.Sprintf("key %q exists in .env.example but not in .env", k))
			}
		}
	}

	return issues, nil
}

// parseKeysFromExample reads a .env.example file and returns a set of keys.
func parseKeysFromExample(path string) (map[string]bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	keys := make(map[string]bool)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimPrefix(line, "export ")
		if strings.HasPrefix(strings.TrimSpace(line), "#") || strings.TrimSpace(line) == "" {
			continue
		}
		eqIdx := strings.Index(line, "=")
		if eqIdx > 0 {
			keys[strings.TrimSpace(line[:eqIdx])] = true
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
	if issue.Line == 0 {
		return fmt.Sprintf("%s - %s - %s", prefix, path, issue.Message)
	}
	return fmt.Sprintf("%s - %s:%d - %s", prefix, path, issue.Line, issue.Message)
}
