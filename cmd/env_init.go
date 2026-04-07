package cmd

import (
	"github.com/GuechtouliAnis/forge/internal/env"
	"github.com/spf13/cobra"
)

var envInitNoGitignoreFile bool

// envInitCmd initializes a .env file from .env.example and registers it in .gitignore.
var envInitCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize a .env file from .env.example",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true
		path := ".env"
		if len(args) > 0 {
			path = args[0]
		}
		return env.InitEnv(path, !envInitNoGitignoreFile)
	},
}

func init() {
	envInitCmd.Flags().BoolVarP(&envInitNoGitignoreFile, "no-gitignore", "n", false, "skip adding to .gitignore")
	envCmd.AddCommand(envInitCmd)
}
