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

func init() {
	envCmd.Flags().BoolVarP(&envYes, "yes", "y", false, "Create .env.example from scratch")
	rootCmd.AddCommand(envCmd)
}
