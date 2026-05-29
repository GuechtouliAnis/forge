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

// InvalidKeysMode controls how ParseEnv handles keys that fail validation.
// Matches the invalid_keys config field: "warn" | "skip" | "include".
type InvalidKeysMode string

const (
	InvalidKeysWarn    InvalidKeysMode = "warn"    // print warning, include in output
	InvalidKeysSkip    InvalidKeysMode = "skip"    // silently exclude from output
	InvalidKeysInclude InvalidKeysMode = "include" // include silently, no warning
)

// ParseEnv reads a .env file and returns a sanitized string suitable for .env.example.
// Values are stripped from key=value pairs, inline comments are preserved.
// Warns if duplicate keys are detected.
// Invalid key handling is governed by cfg.InvalidKeys ("warn" | "skip" | "include").
func ParseEnv(path, invalidKeys string) (string, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return "", err
	}

	mode := InvalidKeysMode(invalidKeys)
	if mode == "" {
		// Defaults to [warn]
		mode = InvalidKeysWarn
	}

	lines := strings.Split(string(data), "\n")

	seen := make(map[string]bool)
	var result []string

	for _, line := range lines {
		// remove "export " if .env line starts with it
		line = strings.TrimPrefix(line, "export ")

		// resolve the key before transforming, to apply invalid_keys logic
		ln := strings.TrimSpace(line)
		if ln != "" && !strings.HasPrefix(ln, "#") {
			eqIdx := strings.Index(line, "=")
			if eqIdx > 0 {
				key := strings.TrimSpace(line[:eqIdx])

				// ? [WARN] - duplicate key
				if seen[key] {
					fmt.Printf("[warn] duplicate key: %q\n", key)
				}
				seen[key] = true

				keyCode := ValidateKey(key)
				if keyCode != KeyValid && keyCode != KeyIsLowercase {
					// Hard invalid keys (starts with digit, invalid chars)
					switch mode {
					case InvalidKeysSkip:
						continue
					case InvalidKeysWarn:
						fmt.Printf("[warn] %s — included in .env.example\n", invalidKeyReason(keyCode, key))
						// fall through: include despite warning
					case InvalidKeysInclude:
						// fall through silently
					}
				} else if keyCode == KeyIsLowercase {
					// Soft invalid (lowercase)
					switch mode {
					case InvalidKeysSkip:
						continue
					case InvalidKeysWarn:
						fmt.Printf("[warn] key %q contains lowercase characters\n", key)
						// include despite warning (same as legacy behavior)
					case InvalidKeysInclude:
						// fall through silently
					}
				}
			}
		}

		transformed := transformLine(line)
		result = append(result, transformed)
	}
	return strings.Join(result, "\n"), nil
}

// ValidateKey checks if a .env key conforms to standard naming rules.
// Valid keys must match [A-Z][A-Z0-9_]* — uppercase letters, digits, and underscores only.
// Returns KeyValid if the key passes, otherwise returns a code indicating the first violation found:
//   - KeyStartsWithDigit — key begins with a digit
//   - KeyIsLowercase     — key contains a lowercase letter
//   - KeyInvalidChars    — key contains a character outside [A-Z0-9_]
func ValidateKey(key string) int {

	if unicode.IsDigit(rune(key[0])) {
		return KeyStartsWithDigit
	}
	for _, c := range key {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != '_' {
			return KeyInvalidChars
		}
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
			rest := stripped[eqIdx+1:]

			// reuse quoted value logic
			trimmed := strings.TrimSpace(rest)
			if len(trimmed) > 0 && (trimmed[0] == '"' || trimmed[0] == '\'') {
				quote := trimmed[0]
				closeIdx := strings.IndexByte(trimmed[1:], quote)
				if closeIdx >= 0 {
					after := strings.TrimSpace(trimmed[closeIdx+2:])
					if strings.HasPrefix(after, "#") {
						return "# " + key + "=  " + after
					}
					return "# " + key + "="
				}
			}

			// unquoted: first # is comment
			hashIdx := strings.Index(rest, "#")
			if hashIdx >= 0 {
				comment := strings.TrimSpace(rest[hashIdx:])
				return "# " + key + "=  " + comment
			}
			return "# " + key + "="
		}
		// plain comment, return as-is
		return line
	}

	// if "=" not found in line
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

// invalidKeyReason returns a human-readable reason string for a failed key validation code.
func invalidKeyReason(code int, key string) string {
	switch code {
	case KeyStartsWithDigit:
		return fmt.Sprintf("key %q starts with a digit", key)
	case KeyInvalidChars:
		return fmt.Sprintf("key %q contains invalid characters", key)
	default:
		return fmt.Sprintf("key %q is invalid", key)
	}
}
