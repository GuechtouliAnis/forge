// Package cmd contains all CLI subcommands for Forge.
// Each file in this package defines one subcommand and registers it with the root command.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// Base "forge" command
var rootCmd = &cobra.Command{
	Use:           "forge",
	Short:         "Developer CLI for scaffolding repos and managing env files",
	Long:          "Forge — scaffold repositories, generate licenses, READMEs, and manage environment files without the boilerplate.",
	Version:       "1.3.0",
	SilenceErrors: true,
	SilenceUsage:  true,
}

var githubUsername string

// username is a persistent flag available to all subcommands
// used for Go module paths in the format github.com/username/project
// falls back to git config user.name if not provided
func init() {
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.PersistentFlags().StringVarP(&githubUsername, "username", "u", "", "Your GitHub username")
}

// Execute is called by main.go to start the CLI.
// Cobra takes over and routes to the right command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
