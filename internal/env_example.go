package internal

import (
	"fmt"
	"os"
	"strings"
)

func ParseEnv(path string) (string, error) {

	data, err := os.ReadFile(path)

	if err != nil {
		return "", err
	}

	lines := strings.Split(string(data), "\n")

	var result []string
	for _, line := range lines {
		ln := transformLine(line)
		if ln != "" {
			result = append(result, ln)
		}
	}
	return strings.Join(result, "\n"), nil
}

func transformLine(line string) string {
	equal_index := strings.Index(line, "=")
	hasht_index := strings.Index(line, "#")

	if hasht_index == 0 {
		return line
	}
	if equal_index >= 0 {

		if hasht_index > equal_index {
			return line[:equal_index+1] + "  " + line[hasht_index:]
		} else {
			return line[:equal_index+1]
		}
	}
	return ""

}

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

func WriteEnvExampleForce(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

func extractKeys(content string) map[string]bool {
	keys := make(map[string]bool)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		eq := strings.Index(line, "=")
		if eq > 0 {
			keys[line[:eq]] = true
		}
	}
	return keys
}

func AppendMissing(envPath string, examplePath string) error {
	envContent, err := os.ReadFile(envPath)
	if err != nil {
		return err
	}
	exampleContent, err := os.ReadFile(examplePath)
	if err != nil {
		return err
	}
	existingKeys := extractKeys(string(exampleContent))
	lines := strings.Split(string(envContent), "\n")
	var toAppend []string
	var pendingComment string
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			pendingComment = line
			continue
		}
		ln := transformLine(line)
		if ln == "" {
			pendingComment = ""
			continue
		}
		eq := strings.Index(ln, "=")
		if eq > 0 {
			key := ln[:eq]
			if !existingKeys[key] {
				if pendingComment != "" {
					toAppend = append(toAppend, pendingComment)
				}
				toAppend = append(toAppend, ln)
			}
		}
		pendingComment = ""
	}
	if len(toAppend) == 0 {
		fmt.Println("nothing to append")
		return nil
	}
	appended := "\n" + strings.Join(toAppend, "\n")
	f, err := os.OpenFile(examplePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(appended)
	return err
}
