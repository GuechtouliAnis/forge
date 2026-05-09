package env

import (
	"bufio"
	"bytes"
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
		return fmt.Errorf("[env add]: %w", err)
	}

	// build preset key set for O(1) lookup
	presetKeys := make(map[string]bool)
	for _, preset := range selected {
		for _, key := range presets[preset] {
			presetKeys[key] = true
		}
	}

	// check for existing keys in .env file
	existing := make(map[string]bool)
	// process .env line by line
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimPrefix(scanner.Text(), "export ")
		// if is empty line, skip
		if strings.TrimSpace(line) == "" {
			continue
		}
		// if line is comment
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			// if line is comment BUT has equal sign and a key
			// >> # KEY=    (flagged as a warning)
			if key, _, found := strings.Cut(line[1:], "="); found {
				if presetKeys[strings.TrimSpace(key)] {
					fmt.Printf("warning: %s exists but is commented out\n", strings.TrimSpace(key))
				}
			}
			continue
		}
		// keeping track of existing keys to avoid duplicates
		if key, _, found := strings.Cut(line, "="); found {
			existing[strings.TrimSpace(key)] = true
		}
	}

	type entry struct {
		preset string
		key    string
		value  string
	}

	var toWrite []entry
	skipped, total := 0, 0

	// loop through the selected lists of predefined values (eg. : values of --db)
	for _, preset := range selected {
		for _, key := range presets[preset] {
			total++
			if existing[key] {
				fmt.Printf("warning: %s already exists, skipping\n", key)
				skipped++
				continue
			}
			value := `""`
			if v, ok := hostVars[key]; ok {
				value = v
			}
			toWrite = append(toWrite, entry{preset, key, value})
		}
	}

	if skipped == total {
		return fmt.Errorf("[env add]: all predefined vars already exist in %s", path)
	}

	// append to file
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("[env add]: %w", err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	// ensure starting on a new line
	if len(data) > 0 && data[len(data)-1] != '\n' {
		fmt.Fprintln(w)
	}

	// add a comment above keys added by forge
	currentPreset := ""
	for _, e := range toWrite {
		if e.preset != currentPreset {
			fmt.Fprintf(w, "# %s - added by forge env add\n", e.preset)
			currentPreset = e.preset
		}
		fmt.Fprintln(w, e.key+"="+e.value)
	}

	return nil
}
