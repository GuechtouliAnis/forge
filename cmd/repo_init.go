package cmd

import (
	"github.com/GuechtouliAnis/forge/internal/repo"
	"github.com/spf13/cobra"
)

// initCmd initializes a new git repository with forge scaffolding.
// Accepts an optional path arg — defaults to current directory if omitted.
// Lang and license default to generic/mit if not provided.
var repoInitCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize a new git repository with forge scaffolding",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := ""
		if len(args) > 0 {
			path = args[0]
		}
		lang, _ := cmd.Flags().GetString("lang")
		license, _ := cmd.Flags().GetString("license")
		return repo.CreateRepo(path, lang, license)
	},
}

// init registers the repo init command with the repo command.
func init() {
	repoInitCmd.Flags().String("lang", "", "language for .gitignore — py/python or go/golang")
	repoInitCmd.Flags().String("license", "", "license type — mit, apache, gpl, agpl, bsd (default: mit)")
	repoCmd.AddCommand(repoInitCmd)
}
