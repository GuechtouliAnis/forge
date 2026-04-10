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

	// check for existing keys in .env file
	existing := make(map[string]bool)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimPrefix(line, "export ")
		if strings.TrimSpace(line) == "" {
			continue
		}
		eqIdx := strings.Index(line, "=")
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			if eqIdx > 0 {
				key := strings.TrimSpace(line[1:eqIdx]) // skip the '#'
				for _, preset := range selected {
					for _, presetKey := range presets[preset] {
						// if the key we are trying to insert is the same key that is commented
						if key == presetKey {
							fmt.Printf("warning: %s exists but is commented out\n", key)
						}
					}
				}
			}
			continue
		}
		if eqIdx > 0 {
			existing[strings.TrimSpace(line[:eqIdx])] = true
		}
	}

	// build lines to append
	skipped := 0
	total := 0

	// loop through the selected lists of predefined values (eg. : values of --db)
	for _, preset := range selected {
		for _, key := range presets[preset] {
			total++
			if existing[key] {
				fmt.Printf("warning: %s already exists, skipping\n", key)
				skipped++
				continue
			}
		}
	}

	if skipped == total {
		return fmt.Errorf("ERROR: all predefined vars already exist in %s", path)
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
		fmt.Fprintf(f, "# %s - added by forge env add\n", preset)
		for _, key := range presets[preset] {
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
