// Package internal contains the core logic for all Forge commands.
// Functions here are language-agnostic helpers called by cmd/ subcommands.
package env

import (
	"fmt"
	"os"
	"strings"
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
		ln := transformLine(line)
		if ln == "" {
			result = append(result, ln)
			continue
		}
		eq := strings.Index(ln, "=")
		if eq > 0 {
			key := strings.TrimSpace(ln[:eq])
			if seen[key] {
				fmt.Printf("warning: duplicate key %s\n", key)
			}
			seen[key] = true
		}
		result = append(result, ln)
	}
	return strings.Join(result, "\n"), nil
}

// transformLine processes a single line from a .env file.
// Comment lines are kept as-is, key=value lines have their value stripped,
// inline comments are preserved. Malformed lines return an empty string.
func transformLine(line string) string {
	if strings.TrimSpace(line) == "" {
		return line
	}
	if strings.HasPrefix(strings.TrimSpace(line), "#") {
		return line
	}

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
