package cmd

import (
	"github.com/GuechtouliAnis/forge/internal/repo"
	"github.com/spf13/cobra"
)

// gitignoreCmd generates a .gitignore file in the current directory.
// Accepts an optional language arg: py/python or go/golang.
// Defaults to a generic gitignore if no language is provided.
var gitignoreCmd = &cobra.Command{
	Use:   "gitignore [language]",
	Short: "Generate a .gitignore for the current project",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		lang := ""
		if len(args) > 0 {
			lang = args[0]
		}
		return repo.CreateGitignore(lang)
	},
}

// init registers the gitignore command with the repo command.
func init() {
	repoCmd.AddCommand(gitignoreCmd)
}
