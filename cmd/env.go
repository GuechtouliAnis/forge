package cmd

import "github.com/spf13/cobra"

// envCmd is the parent command for all env-related subcommands.
// Running `forge env` without a subcommand prints the help message.
var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage .env files — generate, validate, populate, and sync",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(envCmd)
}
