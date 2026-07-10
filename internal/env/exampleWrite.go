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
			fmt.Println("[env example]: Aborted.")
			return nil
		}
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return err
	}
	fmt.Printf("[env example]: .env.example created at %s.\n", path)
	return nil
}

// WriteEnvExampleForce writes content to path as a .env.example file without prompting.
// Used when the -y flag is passed to forge env.
func WriteEnvExampleForce(path string, content string) error {
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return err
	}
	fmt.Printf("[env example]: .env.example created at %s.\n", path)
	return nil
}
