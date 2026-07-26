package cmdgit

import (
	"fmt"
	"os"

	"github.com/GuechtouliAnis/forge/internal/config"
	"github.com/GuechtouliAnis/forge/internal/git"
	"github.com/spf13/cobra"
)

var (
	commitMessageFlag string
	commitAmendFlag   bool
	commitDryRunFlag  bool
)

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Commit staged changes with an enforced message format",
	Long: `Validates a commit message against a configurable format before committing.
The format supports {domain} and {message} placeholders - {domain} is
constrained to a configured list of valid values, {message} is capped at a
configurable maximum length. If both format and message_max_length are unset,
validation is skipped entirely. --amend reuses the previous commit's message
when no new message is supplied, or validates and applies a new one if -m is given.
--dry-run runs all the same validation and checks but stops short of actually
committing, printing the message and staged files that would be included instead.`,
	// No positional arguments — the message is supplied via -m, per spec.
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("could not determine working directory: %w", err)
		}

		cfg, err := config.Load(cwd)
		if err != nil {
			return fmt.Errorf("could not load .forge.toml: %w", err)
		}

		// Distinguishes "-m not passed" from "-m passed as empty string" —
		// only the former should trigger CommitGit's amend-reuse path.
		messageProvided := cmd.Flags().Changed("message")

		if err := git.CommitGit(&cfg.Git.Commit, commitMessageFlag, messageProvided, commitAmendFlag, commitDryRunFlag); err != nil {

			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	commitCmd.Flags().StringVarP(&commitMessageFlag, "message", "m", "",
		"The commit message to validate and commit")
	commitCmd.Flags().BoolVar(&commitAmendFlag, "amend", false,
		"Amend the previous commit")
	commitCmd.Flags().BoolVar(&commitDryRunFlag, "dry-run", false,
		"Run all validation and checks without actually committing")
	gitCmd.AddCommand(commitCmd)
}
