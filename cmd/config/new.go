package cmdconfig

import (
	"github.com/GuechtouliAnis/forge/internal/config"
	"github.com/spf13/cobra"
)

// configNewCmd generates a .forge.toml in the current or specified directory.
// Accepts an optional path arg — defaults to current directory if omitted.
var configNewCmd = &cobra.Command{
	Use:   "new [path]",
	Short: "Generate a .forge.toml for the current project",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := ""
		if len(args) > 0 {
			path = args[0]
		}
		return config.CreateForgeToml(path)
	},
}

func init() {
	configCmd.AddCommand(configNewCmd)
}
