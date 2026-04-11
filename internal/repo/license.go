package repo

import (
	_ "embed" // required for go:embed directives
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

//go:embed templates/licenses/agpl.license
var licenseAGPL string

//go:embed templates/licenses/apache.license
var licenseApache string

//go:embed templates/licenses/bsd.license
var licenseBSD string

//go:embed templates/licenses/gpl.license
var licenseGPL string

//go:embed templates/licenses/mit.license
var licenseMIT string

// CreateLicense writes a LICENSE file to the given path.
// If path is empty, it defaults to the current directory.
// Author is read from git config user.name, falls back to a prompt if not set.
// Year is inferred from the system clock.
func CreateLicense(license string, path string) error {
	if path == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		path = cwd
	}

	licensePath := filepath.Join(path, "LICENSE")
	licenseMDPath := filepath.Join(path, "LICENSE.md")

	// check if a LICENSE already exists at the target path
	_, err1 := os.Stat(licensePath)
	_, err2 := os.Stat(licenseMDPath)
	if err1 == nil || err2 == nil {
		fmt.Print("A LICENSE already exists. Overwrite? [y/N]: ")
		var input string
		fmt.Scanln(&input)
		if input != "y" && input != "Y" && strings.ToLower(input) != "yes" {
			fmt.Println("Aborted")
			return nil
		}
	}

	// select the appropriate template based on license type
	var content string
	switch strings.ToLower(license) {
	case "mit", "":
		content = licenseMIT
	case "apache":
		content = licenseApache
	case "gpl":
		content = licenseGPL
	case "agpl":
		content = licenseAGPL
	case "bsd":
		content = licenseBSD
	default:
		return fmt.Errorf("unsupported license %q — supported: mit, apache, gpl, agpl, bsd", license)
	}

	// prefer git config user.name over prompting
	var author string
	out, err := exec.Command("git", "config", "user.name").Output()
	if err != nil || strings.TrimSpace(string(out)) == "" {
		fmt.Print("Author name: ")
		fmt.Scanln(&author)
	} else {
		author = strings.TrimSpace(string(out))
	}

	year := strconv.Itoa(time.Now().Year())
	content = strings.ReplaceAll(content, "{author}", author)
	content = strings.ReplaceAll(content, "{year}", year)

	// 0644 = owner read/write, group and others read only
	if err := os.WriteFile(licensePath, []byte(content), 0644); err != nil {
		return err
	}

	displayName := license
	if displayName == "" {
		displayName = "mit"
	}
	fmt.Printf("Generated LICENSE (%s)\n", strings.ToUpper(displayName))

	return nil
}
