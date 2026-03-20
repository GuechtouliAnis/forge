package cmd

import (
	"github.com/GuechtouliAnis/forge/internal"
	"github.com/spf13/cobra"
)

var (
	initPy bool
	initGo bool
)

var initCmd = &cobra.Command{
	Use:   "init [name]",
	Short: "Scaffold a new project with git initialized",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		// resolve language from flags, empty string means no language setup
		lang := ""
		if initPy {
			lang = "py"
		} else if initGo {
			lang = "go"
		}

		// true tells CreateProject to run git init and initial commit
		return internal.CreateProject(args[0], lang, true)
	},
}

// initCmd scaffolds a new project directory and initializes a git repository.
// It is identical to new but also runs git init and creates an initial commit.
// Example: forge init --py myproject

// init registers the init command and its flags with the root command.
func init() {
	initCmd.Flags().BoolVar(&initPy, "py", false, "Python project")
	initCmd.Flags().BoolVar(&initGo, "go", false, "Go project")
	rootCmd.AddCommand(initCmd)
}
