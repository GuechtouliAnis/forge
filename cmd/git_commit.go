package cmd

import (
	"fmt"
	"os"

	"github.com/GuechtouliAnis/forge/internal/config"
	"github.com/GuechtouliAnis/forge/internal/git"
	"github.com/spf13/cobra"
)

var commitCmd = &cobra.Command{
	Use:   "commit <message>",
	Short: "Opinionated git helpers (not a git replacement)",
	Long: `A thin layer on top of git for enforcing commit conventions and reducing friction.
Forge does not replicate git — it handles the parts git leaves to you.
For branching, history, rebasing, and anything else: use git directly.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		commitMsg := args[0]

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("could not determine working directory: %w", err)
		}

		cfg, err := config.Load(cwd)
		if err != nil {
			return fmt.Errorf("could not load .forge.toml: %w", err)
		}

		valid, err := git.ValidateCommit(commitMsg, &cfg.Git.Commit)
		if err != nil {
			return fmt.Errorf("could not validate commit: %w", err)
		}
		if !valid {
			fmt.Fprintln(os.Stderr, "✗ commit message does not match the required format")
			os.Exit(1)
		}
		fmt.Println("✓ valid")
		return nil
	},
}

func init() {
	gitCmd.AddCommand(commitCmd)
}
