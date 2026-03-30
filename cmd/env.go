package cmd

import (
	"github.com/GuechtouliAnis/forge/internal"
	"github.com/spf13/cobra"
)

var (
	envYes bool
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Generate a .env.example from .env",
	RunE: func(cmd *cobra.Command, args []string) error {
		content, err := internal.ParseEnv(".env")
		if err != nil {
			return err
		}
		if envYes {
			return internal.WriteEnvExampleForce(".env.example", content)
		}
		return internal.WriteEnvExample(".env.example", content)
	},
}

// envCmd generates a .env.example file from the current .env file.
// Values are stripped, comments are preserved, duplicate keys are flagged.
// Use -y to overwrite an existing .env.example without being prompted.
// init registers the env command and its flags with the root command.
func init() {
	envCmd.Flags().BoolVarP(&envYes, "yes", "y", false, "Overwrite existing .env.example without prompt")
	rootCmd.AddCommand(envCmd)
}
