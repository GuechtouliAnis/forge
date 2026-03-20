package cmd

import (
	"github.com/GuechtouliAnis/forge/internal"
	"github.com/spf13/cobra"
)

var (
	giPy bool
	giGo bool
)

var gitignoreCmd = &cobra.Command{
	Use:   "gitignore",
	Short: "Generate a .gitignore for the current project",
	RunE: func(cmd *cobra.Command, args []string) error {

		// resolve language from flags, empty string means no language setup
		lang := ""
		if giPy {
			lang = "py"
		} else if giGo {
			lang = "go"
		}

		// true tells CreateProject to run git init and initial commit
		return internal.CreateGitignore(lang)
	},
}

// gitignoreCmd generates a .gitignore file in the current directory.
// Use --py for Python, --go for Go, or no flag for a generic gitignore.

// init registers the gitignore command and its flags with the root command.
func init() {
	gitignoreCmd.Flags().BoolVar(&giPy, "py", false, "Python gitignore")
	gitignoreCmd.Flags().BoolVar(&giGo, "go", false, "Go gitignore")
	rootCmd.AddCommand(gitignoreCmd)
}
