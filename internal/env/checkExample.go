package env

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
)

// kick off .env.example read concurrently
type ExampleResult struct {
	keys map[string]ExampleKey
	err  error
}

type ExampleKey struct {
	HasValue bool
}

// parseKeysFromExample reads a .env.example file and returns a map of keys.
// HasValue is true if a key has an actual value set — which it shouldn't in an example file.
func ParseKeysFromExample(path string) (map[string]ExampleKey, error) {
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
			if key == "" {
				continue
			}
			if ValidateKey(key) == KeyInvalidChars || ValidateKey(key) == KeyStartsWithDigit {
				continue
			}
			if h := strings.Index(value, "#"); h >= 0 {
				value = strings.TrimSpace(value[:h])
			}
			keys[key] = ExampleKey{HasValue: strings.TrimSpace(value) != ""}
		}
	}
	return keys, nil
}
