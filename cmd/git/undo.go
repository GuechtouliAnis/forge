package cmdgit

import (
	"github.com/GuechtouliAnis/forge/internal/git"
	"github.com/spf13/cobra"
)

// undoCmd reverts the last commit, buffering the message to .git/forge/undo_msg.txt.
// Soft reset by default — use --hard to wipe uncommitted changes too.
var undoCmd = &cobra.Command{
	Use:   "undo",
	Short: "Revert the last commit, buffering the message for reuse",
	RunE: func(cmd *cobra.Command, args []string) error {
		hard, _ := cmd.Flags().GetBool("hard")
		return git.UndoCommit(hard)
	},
}

func init() {
	undoCmd.Flags().Bool("hard", false, "destructive reset — wipes uncommitted changes")
	gitCmd.AddCommand(undoCmd)
}
