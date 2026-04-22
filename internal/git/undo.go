package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// UndoCommit reverts the last commit, buffering the commit message to .git/forge/undo_msg.txt.
// Default is soft reset — staged files are preserved.
// --hard wipes uncommitted changes and requires confirmation if the worktree is dirty.
func UndoCommit(hard bool) error {
	// confirm we're in a git repo and get the .git dir
	gitDirOut, err := exec.Command("git", "rev-parse", "--git-dir").Output()
	if err != nil {
		return fmt.Errorf("not a git repository")
	}
	gitDir := strings.TrimSpace(string(gitDirOut))

	// verify HEAD~1 exists
	// a previous branch is required to reset to
	if err := exec.Command("git", "rev-parse", "--verify", "HEAD~1").Run(); err != nil {
		return fmt.Errorf("nothing to undo — this is the initial commit")
	}

	// buffer the last commit message before resetting
	msgOut, err := exec.Command("git", "log", "-1", "--pretty=%B").Output()
	if err != nil {
		return fmt.Errorf("could not read last commit message: %w", err)
	}

	// ensure .git/forge/ exists
	forgeDir := filepath.Join(gitDir, "forge")
	if err := os.MkdirAll(forgeDir, 0755); err != nil {
		return fmt.Errorf("could not create .git/forge/: %w", err)
	}

	undoPath := filepath.Join(forgeDir, "undo_msg.txt")

	// if a buffered message already exists, prompt before overwriting
	writeMsg := true
	if _, err := os.Stat(undoPath); err == nil {
		existing, _ := os.ReadFile(undoPath)
		fmt.Printf("Buffered message exists:\n  %s\nOverwrite with:\n  %s[y/N]: ",
			strings.TrimSpace(string(existing)),
			strings.TrimSpace(string(msgOut)))
		var input string
		fmt.Scanln(&input)
		if strings.ToLower(input) != "y" && strings.ToLower(input) != "yes" {
			fmt.Println("Keeping existing buffer.")
			writeMsg = false
		}
	}

	if writeMsg {
		if err := os.WriteFile(undoPath, msgOut, 0644); err != nil {
			return fmt.Errorf("could not buffer commit message: %w", err)
		}
	}

	if hard {
		// warn if worktree is dirty — hard reset will wipe uncommitted changes permanently
		statusOut, err := exec.Command("git", "status", "--porcelain").Output()
		if err != nil {
			return fmt.Errorf("could not check worktree status: %w", err)
		}
		if strings.TrimSpace(string(statusOut)) != "" {
			fmt.Print("WARNING: uncommitted changes will be lost. Continue? [y/N]: ")
			var input string
			fmt.Scanln(&input)
			if strings.ToLower(input) != "y" && strings.ToLower(input) != "yes" {
				fmt.Println("Aborted.")
				return nil
			}
		}
		// wipe staged and unstaged changes, move HEAD back one commit
		if err := exec.Command("git", "reset", "--hard", "HEAD~1").Run(); err != nil {
			return fmt.Errorf("git reset --hard failed: %w", err)
		}
	} else {
		// soft reset — moves HEAD back but keeps files staged, nothing is lost
		if err := exec.Command("git", "reset", "--soft", "HEAD~1").Run(); err != nil {
			return fmt.Errorf("git reset --soft failed: %w", err)
		}
	}

	fmt.Printf("[UNDONE] Last commit reverted. Message buffered to %s\n", undoPath)
	return nil
}
