package cmdconfig

import "github.com/spf13/cobra"

// configCmd is the parent command for all config-related subcommands.
// Running `forge config` without a subcommand prints the help message.
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage .forge.toml configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func Register(root *cobra.Command) {
	root.AddCommand(configCmd)
}
