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
	for _, line := range strings.Split(string(data), "\n") {
		// trim line from prefix "export " to get the key value couple only
		line = strings.TrimPrefix(line, "export ")
		// if is empty line, skip
		if strings.TrimSpace(line) == "" {
			continue
		}
		// look for the first appearance of equal sign, which should be the splitter of key=value
		eqIdx := strings.Index(line, "=")
		// if line is comment
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			// if line is comment BUT has equal sign and a key
			// >> # KEY=    (flagged as a warning)
			if eqIdx > 0 {
				key := strings.TrimSpace(line[1:eqIdx])
				if presetKeys[key] {
					fmt.Printf("warning: %s exists but is commented out\n", key)
				}
			}
			continue
		}
		// keeping track of existing keys to avoid duplicates
		if eqIdx > 0 {
			existing[strings.TrimSpace(line[:eqIdx])] = true
		}
	}

	// count skipped vs total
	skipped := 0
	total := 0

	// loop through the selected lists of predefined values (eg. : values of --db)
	for _, preset := range selected {
		for _, key := range presets[preset] {
			total++
			if existing[key] {
				fmt.Printf("warning: %s already exists, skipping\n", key)
				skipped++
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

	// ensure starting on a new line
	if len(data) > 0 && data[len(data)-1] != '\n' {
		fmt.Fprintln(f)
	}

	// add a comment above keys added by forge
	for _, preset := range selected {
		fmt.Fprintf(f, "# %s - added by forge env add\n", preset)
		for _, key := range presets[preset] {
			// skip existing keys
			if existing[key] {
				continue
			}
			// default value of added keys to be ""
			value := "\"\""
			if v, ok := hostVars[key]; ok {
				value = v
			}
			// actually print the key into the file
			fmt.Fprintln(f, key+"="+value)
		}
	}

	return nil
}
