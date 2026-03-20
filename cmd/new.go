package cmd

import (
	"github.com/GuechtouliAnis/forge/internal"
	"github.com/spf13/cobra"
)

var (
	newPy bool
	newGo bool
)

var newCmd = &cobra.Command{
	Use:   "new [name]",
	Short: "Scaffold a new project locally",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		// resolve language from flags, empty string means no language setup
		lang := ""
		if newPy {
			lang = "py"
		} else if newGo {
			lang = "go"
		}

		// true tells CreateProject to run git init and initial commit
		return internal.CreateProject(args[0], lang, false)
	},
}

// newCmd scaffolds a new project directory locally without initializing git.
// Use --py for Python projects or --go for Go projects.
// Example: forge new --py myproject

// init registers the new command and its flags with the root command.
func init() {
	newCmd.Flags().BoolVar(&newPy, "py", false, "Python Project")
	newCmd.Flags().BoolVar(&newGo, "go", false, "Go Project")
	rootCmd.AddCommand(newCmd)
}
