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
		lang := ""
		if giPy {
			lang = "py"
		} else if giGo {
			lang = "go"
		}
		return internal.CreateGitignore(lang)
	},
}

func init() {
	gitignoreCmd.Flags().BoolVar(&giPy, "py", false, "Python gitignore")
	gitignoreCmd.Flags().BoolVar(&giGo, "go", false, "Go gitignore")
	rootCmd.AddCommand(gitignoreCmd)
}
