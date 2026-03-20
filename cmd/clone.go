package cmd

import (
	"github.com/GuechtouliAnis/forge/internal"
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

		lang := ""
		if clonePy {
			lang = "py"
		} else if cloneGo {
			lang = "go"
		}

		return internal.Clone(repo, lang, githubUsername)
	},
}

func init() {
	cloneCmd.Flags().BoolVar(&clonePy, "py", false, "Set up Python environment")
	cloneCmd.Flags().BoolVar(&cloneGo, "go", false, "Set up Go environment")

	rootCmd.AddCommand(cloneCmd)
}
