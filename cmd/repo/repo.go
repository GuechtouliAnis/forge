package cmdrepo

import "github.com/spf13/cobra"

// repoCmd is the parent command for all repo-related subcommands.
// Running `forge repo` without a subcommand prints the help message.
var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Manage repository lifecycle",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func Register(root *cobra.Command) {
	root.AddCommand(repoCmd)
}
