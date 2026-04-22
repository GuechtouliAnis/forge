package cmdrepo

import (
	"github.com/GuechtouliAnis/forge/internal/repo"
	"github.com/spf13/cobra"
)

// licenseCmd generates a LICENSE file in the current or specified directory.
// Accepts license type and optional path — defaults to MIT and current directory if omitted.
var licenseCmd = &cobra.Command{
	Use:   "license [license] [path]",
	Short: "Generate a LICENSE file for the current project",
	Args:  cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		license := ""
		path := ""
		if len(args) > 0 {
			license = args[0]
		}
		if len(args) > 1 {
			path = args[1]
		}
		return repo.CreateLicense(license, path)
	},
}

// init registers the license command with the repo command.
func init() {
	repoCmd.AddCommand(licenseCmd)
}
