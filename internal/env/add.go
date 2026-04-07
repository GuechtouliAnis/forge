package env

import (
	"fmt"
	"os"
	"strings"
)

// AddEnv appends predefined variable sets to a .env file.
// Skips keys that already exist, warns per skipped key.
// Returns an error if all keys in the selected presets already exist.
func AddEnv(path string, selected []string) error {
	// read existing keys
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	existing := make(map[string]bool)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimPrefix(line, "export ")
		if strings.HasPrefix(strings.TrimSpace(line), "#") || strings.TrimSpace(line) == "" {
			continue
		}
		eqIdx := strings.Index(line, "=")
		if eqIdx > 0 {
			existing[strings.TrimSpace(line[:eqIdx])] = true
		}
	}

	// build lines to append
	skipped := 0
	total := 0

	for _, preset := range selected {
		keys, ok := presets[preset]
		if !ok {
			continue
		}
		for _, key := range keys {
			total++
			if existing[key] {
				fmt.Printf("warning: %s already exists, skipping\n", key)
				skipped++
				continue
			}
		}
	}

	if skipped == total {
		return fmt.Errorf("all predefined vars already exist in %s", path)
	}

	// append to file
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// ensure we start on a new line
	if len(data) > 0 && data[len(data)-1] != '\n' {
		fmt.Fprintln(f)
	}

	for _, preset := range selected {
		keys, ok := presets[preset]
		if !ok {
			continue
		}
		fmt.Fprintf(f, "# %s - added by forge env add\n", preset)
		for _, key := range keys {
			if existing[key] {
				continue
			}
			value := "\"\""
			if v, ok := hostVars[key]; ok {
				value = v
			}
			fmt.Fprintln(f, key+"="+value)
		}
	}

	return nil
}
