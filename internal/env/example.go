// Package internal contains the core logic for all Forge commands.
// Functions here are language-agnostic helpers called by cmd/ subcommands.
package env

import (
	"fmt"
	"os"
	"strings"
	"unicode"
)

// validateKey return codes.
// iota assigns incrementing integers starting from 0 (KeyValid = 0, KeyStartsWithDigit = 1, ...).
// Used in ParseEnv's switch to distinguish warning-only cases from hard invalid ones.
const (
	KeyValid = iota
	KeyStartsWithDigit
	KeyInvalidChars
	KeyIsLowercase
)

// ParseEnv reads a .env file and returns a sanitized string suitable for .env.example.
// Values are stripped from key=value pairs, inline comments are preserved.
// Warns if duplicate keys are detected.
func ParseEnv(path string) (string, error) {

	data, err := os.ReadFile(path)

	if err != nil {
		return "", err
	}

	lines := strings.Split(string(data), "\n")

	seen := make(map[string]bool)
	var result []string

	for _, line := range lines {
		// remove "export " if .env line starts with it
		line = strings.TrimPrefix(line, "export ")
		ln := transformLine(line)

		// look for the index in which the first value ends
		eqIdx := strings.Index(ln, "=")
		if eqIdx > 0 {
			key := strings.TrimSpace(ln[:eqIdx])

			// remove comment on "continue" to exclude invalid keys from .env.example
			switch ValidateKey(key) {
			case KeyStartsWithDigit:
				fmt.Printf("warning: key %q starts with digit\n", key)
				// continue
			case KeyInvalidChars:
				fmt.Printf("warning: key %q contains invalid characters\n", key)
				// continue
			case KeyIsLowercase:
				fmt.Printf("warning: key %q contains lowercase characters\n", key)
				// continue
			}
			if seen[key] {
				fmt.Printf("warning: duplicate key %s\n", key)
			}
			seen[key] = true
		}
		result = append(result, ln)
	}
	return strings.Join(result, "\n"), nil
}

// ValidateKey checks if a .env key is valid and returns a code indicating the result.
// Returns KeyValid if the key is valid, otherwise returns a code indicating the issue.
// Callers are responsible for handling warnings and deciding whether to skip the key.
func ValidateKey(key string) int {
	const invalidKeyChars = "$!@{} \t"

	if unicode.IsDigit(rune(key[0])) {
		return KeyStartsWithDigit
	}

	if strings.ContainsAny(key, invalidKeyChars) {
		return KeyInvalidChars
	}

	for _, c := range key {
		if unicode.IsLower(c) {
			return KeyIsLowercase
		}
	}

	return KeyValid
}

// transformLine processes a single line from a .env file.
// Comment lines are kept as-is, key=value lines have their value stripped,
// inline comments are preserved. Malformed lines return an empty string.
func transformLine(line string) string {
	// return empty line as is
	if strings.TrimSpace(line) == "" {
		return ""
	}

	// handle commented lines
	if strings.HasPrefix(strings.TrimSpace(line), "#") {
		// check if it's a commented key=value (e.g. # KEY=value) — strip the value
		stripped := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "#"))
		eqIdx := strings.Index(stripped, "=")
		if eqIdx > 0 {
			key := strings.TrimSpace(stripped[:eqIdx])
			return "# " + key + "="
		}
		// plain comment, return as-is
		return line
	}

	// if = not found in line
	eqIdx := strings.Index(line, "=")
	if eqIdx < 0 {
		return ""
	}

	key := line[:eqIdx]
	rest := line[eqIdx+1:]

	// check if value is quoted
	trimmed := strings.TrimSpace(rest)
	if len(trimmed) > 0 && (trimmed[0] == '"' || trimmed[0] == '\'') {
		quote := trimmed[0]
		// find closing quote
		closeIdx := strings.IndexByte(trimmed[1:], quote)
		if closeIdx >= 0 {
			// everything after closing quote is potential comment
			// closeIdx is relative to trimmed[1:], so +2 skips both the offset and the closing quote
			after := strings.TrimSpace(trimmed[closeIdx+2:])
			if strings.HasPrefix(after, "#") {
				return key + "=  " + after
			}
			return key + "="
		}
	}

	// unquoted: first # is comment
	hashIdx := strings.Index(rest, "#")
	if hashIdx >= 0 {
		comment := strings.TrimSpace(rest[hashIdx:])
		return key + "=  " + comment
	}

	return key + "="
}

// WriteEnvExample writes content to path as a .env.example file.
// If the file already exists, the user is prompted for confirmation before overwriting.
func WriteEnvExample(path string, content string) error {
	_, err := os.Stat(path)
	if err == nil {
		var input string
		fmt.Print(".env.example already exists, overwrite? (y/n): ")
		fmt.Scan(&input)
		if input != "y" {
			return nil // abort
		}
	}

	return os.WriteFile(path, []byte(content), 0644)
}

// WriteEnvExampleForce writes content to path as a .env.example file without prompting.
// Used when the -y flag is passed to forge env.
func WriteEnvExampleForce(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}
