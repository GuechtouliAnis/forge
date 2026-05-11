// Package cmd contains all CLI subcommands for Forge.
// Each file in this package defines one subcommand and registers it with the root command.
package cmd

import (
	"fmt"
	"os"

	cmdconfig "github.com/GuechtouliAnis/forge/cmd/config"
	cmdenv "github.com/GuechtouliAnis/forge/cmd/env"
	cmdgit "github.com/GuechtouliAnis/forge/cmd/git"
	cmdrepo "github.com/GuechtouliAnis/forge/cmd/repo"
	"github.com/spf13/cobra"
)

// Package cmd contains all CLI subcommands for Forge.
// Each subdirectory defines a command group and registers it via Register(root).
var rootCmd = &cobra.Command{
	Use:           "forge",
	Short:         "Developer CLI for scaffolding repos and managing env files",
	Long:          "Forge — scaffold repositories, generate licenses, READMEs, and manage environment files without the boilerplate.",
	Version:       "1.5.0",
	SilenceErrors: true,
	SilenceUsage:  true,
}

// username is a persistent flag available to all subcommands
// used for Go module paths in the format github.com/username/project
// falls back to git config user.name if not provided
func init() {
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	// register all command groups
	cmdconfig.Register(rootCmd)
	cmdenv.Register(rootCmd)
	cmdgit.Register(rootCmd)
	cmdrepo.Register(rootCmd)
}

// Execute is called by main.go to start the CLI.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
