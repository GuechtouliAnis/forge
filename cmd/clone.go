package cmd

import (
	"github.com/GuechtouliAnis/forge/internal"
	"github.com/spf13/cobra"
)

var cloneLang string

var cloneCmd = &cobra.Command{
	Use:   "clone [repo]",
	Short: "Clone a repo and setup environment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo := args[0]
		return internal.Clone(repo, cloneLang, githubUsername)
	},
}

func init() {
	cloneCmd.Flags().StringVarP(&cloneLang, "lang", "l", "", "Language setup: py or go")

	rootCmd.AddCommand(cloneCmd)
}
