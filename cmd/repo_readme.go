package cmd

import (
	"github.com/GuechtouliAnis/forge/internal/repo"
	"github.com/spf13/cobra"
)

// readmeCmd generates a README.md in the current or specified directory.
// Accepts an optional path arg — defaults to current directory if omitted.
var readmeCmd = &cobra.Command{
	Use:   "readme [path]",
	Short: "Generate a README.md for the current project",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := ""
		if len(args) > 0 {
			path = args[0]
		}
		return repo.CreateReadme(path)
	},
}

// init registers the readme command with the repo command.
func init() {
	repoCmd.AddCommand(readmeCmd)
}
