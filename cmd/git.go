package cmd

import "github.com/spf13/cobra"

// gitCmd is the parent command for git-related subcommands.
// Forge is not a git replacement — it is a thin opinionated layer on top of git
// that enforces commit structure and automates repetitive tasks.
// For anything beyond that, use git directly.
var gitCmd = &cobra.Command{
	Use:   "git",
	Short: "Opinionated git helpers (not a git replacement)",
	Long: `A thin layer on top of git for enforcing commit conventions and reducing friction.
Forge does not replicate git — it handles the parts git leaves to you.
For branching, history, rebasing, and anything else: use git directly.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(gitCmd)
}
