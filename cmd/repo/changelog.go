package cmdrepo

import (
	"github.com/GuechtouliAnis/forge/internal/repo"
	"github.com/spf13/cobra"
)

// changelogCmd generates a CHANGELOG.md scaffold in the current or specified directory.
// Accepts an optional path arg — defaults to current directory if omitted.
var changelogCmd = &cobra.Command{
	Use:   "changelog [path]",
	Short: "Generate a CHANGELOG.md scaffold for the current project",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := ""
		if len(args) > 0 {
			path = args[0]
		}
		return repo.CreateChangelog(path)
	},
}

// init registers the changelog command with the repo command.
func init() {
	repoCmd.AddCommand(changelogCmd)
}
