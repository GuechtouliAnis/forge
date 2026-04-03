package cmd

import (
	"github.com/GuechtouliAnis/forge/internal/project"
	"github.com/spf13/cobra"
)

var (
	clonePy bool
	cloneGo bool
)

var cloneCmd = &cobra.Command{
	Use:   "clone [repo]",
	Short: "Clone a repo and setup environment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo := args[0]

		// resolve language from flags, empty string means no language setup
		lang := ""
		if clonePy {
			lang = "py"
		} else if cloneGo {
			lang = "go"
		}

		// true tells CreateProject to run git init and initial commit
		return project.Clone(repo, lang, githubUsername)
	},
}

// cloneCmd clones a git repository and optionally sets up the development environment.
// Use --py for Python projects or --go for Go projects.
// Use -u to provide a GitHub username for Go module paths.

// init registers the clone command and its flags with the root command.
func init() {
	cloneCmd.Flags().BoolVar(&clonePy, "py", false, "Set up Python environment")
	cloneCmd.Flags().BoolVar(&cloneGo, "go", false, "Set up Go environment")
	rootCmd.AddCommand(cloneCmd)
}
