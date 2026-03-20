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
		lang := ""
		if newPy {
			lang = "py"
		} else if newGo {
			lang = "go"
		}
		return internal.CreateProject(args[0], lang, false)
	},
}

func init() {
	newCmd.Flags().BoolVar(&newPy, "py", false, "Python Project")
	newCmd.Flags().BoolVar(&newGo, "go", false, "Go Project")
	rootCmd.AddCommand(newCmd)
}
