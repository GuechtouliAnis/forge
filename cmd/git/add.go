package cmdgit

import (
	"fmt"
	"os"

	"github.com/GuechtouliAnis/forge/internal/config"
	"github.com/GuechtouliAnis/forge/internal/git"
	"github.com/spf13/cobra"
)

// dryRunFlag backs the --dry-run/-d flag. Declared at package scope so it
// can be read inside RunE, matching cobra's standard flag-binding pattern.
var dryRunFlag bool

var addCmd = &cobra.Command{
	Use:   "add <path> [path...]",
	Short: "Stage files with guardrails for secrets, size thresholds, and blocklist patterns",
	Long: `Wraps 'git add' with a configurable validation layer applied before any file
reaches the index. Blocklist pattern matching intercepts sensitive files such as
private keys. File size thresholds prompt for confirmation before large assets
are staged. Environment files (.env and env.default_file) are unconditionally
rejected regardless of configuration — use git add directly to bypass.
All checks are dry-run safe.`,
	// At least one path is required; unlike commitCmd's ExactArgs(1), add
	// accepts a variadic path list per the spec's `variadic = true`.
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		paths := args

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("could not determine working directory: %w", err)
		}

		cfg, err := config.Load(cwd)
		if err != nil {
			return fmt.Errorf("could not load .forge.toml: %w", err)
		}

		// NOTE: assumes the resolved env default filename lives at
		// cfg.Env.DefaultFile — adjust this field reference to match the
		// actual config.Env struct if the name differs.
		err = git.AddGit(cfg.Git.Add, cfg.Env.DefaultFile, paths, dryRunFlag)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	addCmd.Flags().BoolVarP(&dryRunFlag, "dry-run", "d", false,
		"Preview which files would be staged, blocked, or prompted without touching the index")
	gitCmd.AddCommand(addCmd)
}
