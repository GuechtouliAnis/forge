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
