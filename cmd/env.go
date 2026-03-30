package cmd

import (
	"github.com/GuechtouliAnis/forge/internal"
	"github.com/spf13/cobra"
)

var (
	envAppend   bool
	envRecreate bool
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Generate a .env.example from .env",
	RunE: func(cmd *cobra.Command, args []string) error {
		if envAppend {
			return internal.AppendMissing(".env", ".env.example")
		}
		content, err := internal.ParseEnv(".env")
		if err != nil {
			return err
		}
		if envRecreate {
			return internal.WriteEnvExampleForce(".env.example", content)
		}
		return internal.WriteEnvExample(".env.example", content)
	},
}

func init() {
	envCmd.Flags().BoolVarP(&envAppend, "append", "a", false, "Append missing keys to .env.example")
	envCmd.Flags().BoolVarP(&envRecreate, "recreate", "c", false, "Create .env.example from scratch")
	rootCmd.AddCommand(envCmd)
}
