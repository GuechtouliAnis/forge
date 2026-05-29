package env

import (
	"fmt"
	"os"
	"strings"
)

// WriteEnvExample writes content to path as a .env.example file.
// If the file already exists, the user is prompted for confirmation before overwriting.
func WriteEnvExample(path string, content string) error {
	_, err := os.Stat(path)
	if err == nil {
		var input string
		fmt.Print(".env.example already exists, overwrite? (y/n): ")
		fmt.Scan(&input)
		if strings.ToLower(input) != "y" && strings.ToLower(input) != "yes" {
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
