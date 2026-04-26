package cmdgit

import (
	"github.com/GuechtouliAnis/forge/internal/git"
	"github.com/spf13/cobra"
)

// restoreCmd recovers a file from git history using fuzzy path matching.
// Collision detection blocks overwrites of dirty or staged files unless --force is passed.
var restoreCmd = &cobra.Command{
	Use:   "restore <search>",
	Short: "Recover a file from git history using fuzzy path matching",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		latest, _ := cmd.Flags().GetBool("latest")
		force, _ := cmd.Flags().GetBool("force")
		commit, _ := cmd.Flags().GetString("commit")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		return git.RestoreFile(args[0], latest, force, dryRun, commit)
	},
}

func init() {
	restoreCmd.Flags().BoolP("latest", "l", false, "skip version menu, restore from most recent commit")
	restoreCmd.Flags().Bool("force", false, "overwrite dirty or ignored files without prompt")
	restoreCmd.Flags().String("commit", "", "restore from a specific commit hash")
	restoreCmd.Flags().BoolP("dry-run", "d", false, "search and preview match without restoring")
	gitCmd.AddCommand(restoreCmd)
}
