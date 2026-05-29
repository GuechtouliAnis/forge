package cmdenv

import (
	"fmt"
	"os"

	"github.com/GuechtouliAnis/forge/internal/config"
	"github.com/GuechtouliAnis/forge/internal/env"
	"github.com/spf13/cobra"
)

var (
	envYes bool
)

var envExampleCmd = &cobra.Command{
	Use:   "example",
	Short: "Generate a .env.example from .env",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("[env example]: could not determine working directory: %w", err)
		}

		cfg, err := config.Load(cwd)
		if err != nil {
			return fmt.Errorf("[env example]: could not load .forge.toml: %w", err)
		}

		path := cfg.Env.DefaultFile
		examplePath := cfg.Env.ExampleFile
		invalidKeys := cfg.Env.Example.InvalidKeys

		fmt.Printf("\n[Note] forge env example: review your .env.example before committing — edge cases may apply.\n\n")
		content, err := env.ParseEnv(path, invalidKeys)
		if err != nil {
			return fmt.Errorf("[env example]: could not parse %v: %w", path, err)
		}
		if envYes {
			return env.WriteEnvExampleForce(examplePath, content)
		}
		return env.WriteEnvExample(examplePath, content)
	},
}

// envCmd generates a .env.example file from the current .env file.
// Values are stripped, comments are preserved, duplicate keys are flagged.
// Use -y to overwrite an existing .env.example without being prompted.
// init registers the env command and its flags with the root command.
func init() {
	envExampleCmd.Flags().BoolVarP(&envYes, "yes", "y", false, "Overwrite existing .env.example without prompt")
	envCmd.AddCommand(envExampleCmd)
}
