package cmdenv

import (
	"fmt"
	"os"

	"github.com/GuechtouliAnis/forge/internal/env"
	"github.com/spf13/cobra"
)

var envCheckError bool

// envCheckCmd validates a .env file against key naming rules.
// By default, prints both warnings and errors with line numbers.
// Use -e to show errors only.
var envCheckCmd = &cobra.Command{
	Use:   "check [path]",
	Short: "Validate a .env file against key naming rules",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := ".env"
		if len(args) > 0 {
			path = args[0]
		}

		level := env.LevelWarn
		if envCheckError {
			level = env.LevelError
		}

		issues, err := env.CheckEnv(path, level)
		if err != nil {
			return fmt.Errorf("[env check]: %w", err)
		}

		for _, issue := range issues {
			fmt.Fprintln(os.Stderr, env.FormatIssue(path, issue))
		}

		if len(issues) == 0 {
			fmt.Println("no issues found")
		}

		return nil
	},
}

func init() {
	envCheckCmd.Flags().BoolVarP(&envCheckError, "error", "e", false, "show errors only")
	envCmd.AddCommand(envCheckCmd)
}
