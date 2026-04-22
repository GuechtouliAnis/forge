package cmdenv

import (
	"fmt"

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
		fmt.Printf("\n[beta] forge env example: review your .env.example before committing — edge cases may apply.\n\n")
		content, err := env.ParseEnv(".env")
		if err != nil {
			return err
		}
		if envYes {
			return env.WriteEnvExampleForce(".env.example", content)
		}
		return env.WriteEnvExample(".env.example", content)
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
