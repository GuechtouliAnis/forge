package cmdgit

import (
	"fmt"
	"os"

	"github.com/GuechtouliAnis/forge/internal/config"
	gitClean "github.com/GuechtouliAnis/forge/internal/git/clean"
	"github.com/spf13/cobra"
)

var gitCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Detect and remove stale local branches",
	Long: `Scans local branches and flags ones that are stale by age or commits behind.
Dry-run is the default — use --remove to delete, --force to skip confirmation.

Unmerged branches are safely guarded and skipped during deletion unless --force is passed.
main, master, and the default branch are always protected.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("could not determine working directory: %w", err)
		}

		cfg, err := config.Load(cwd)
		if err != nil {
			return fmt.Errorf("could not load .forge.toml: %w", err)
		}

		// flags override toml, toml overrides defaults
		days, _ := cmd.Flags().GetInt("days")
		behind, _ := cmd.Flags().GetInt("behind")
		remove, _ := cmd.Flags().GetBool("remove")
		force, _ := cmd.Flags().GetBool("force")
		offline, _ := cmd.Flags().GetBool("offline")

		// 0 is a meaningful, intentional value (disables that detection axis),
		// so once we know the flag wasn't explicitly set, trust the config value as-is.
		if !cmd.Flags().Changed("days") {
			days = cfg.Git.Clean.StaleDays
		}
		if !cmd.Flags().Changed("behind") {
			behind = cfg.Git.Clean.CommitsBehind
		}

		cleanCfg := config.GitClean{
			StaleDays:           days,
			CommitsBehind:       behind,
			FetchRemote:         cfg.Git.Clean.FetchRemote,
			MaxWorkers:          cfg.Git.Clean.MaxWorkers,
			ProtectedBranches:   cfg.Git.Clean.ProtectedBranches,
			FetchTimeoutSeconds: cfg.Git.Clean.FetchTimeoutSeconds,
		}

		return gitClean.CleanGit(cleanCfg, remove, force, offline)
	},
}

func init() {
	gitCleanCmd.Flags().Int("days", 30, "days since last commit before branch is considered stale (0 disables)")
	gitCleanCmd.Flags().Int("behind", 10, "commits behind base before branch is considered stale (0 disables)")
	gitCleanCmd.Flags().Bool("remove", false, "show branches to delete and prompt for confirmation")
	gitCleanCmd.Flags().Bool("force", false, "skip confirmation and force-delete unmerged branches (-D)")
	gitCleanCmd.Flags().Bool("offline", false, "skip git fetch and diagnose using local refs only, overriding fetch_remote")
	gitCmd.AddCommand(gitCleanCmd)
}
