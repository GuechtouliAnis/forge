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
		lang := ""
		if initPy {
			lang = "py"
		} else if initGo {
			lang = "go"
		}
		return internal.CreateProject(args[0], lang, true)
	},
}

func init() {
	initCmd.Flags().BoolVar(&initPy, "py", false, "Python project")
	initCmd.Flags().BoolVar(&initGo, "go", false, "Go project")
	rootCmd.AddCommand(initCmd)
}
