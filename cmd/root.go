package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Base "forge" command
var rootCmd = &cobra.Command{
	Use:   "forge",
	Short: "Personal dev CLI",
	Long:  "Forge - build, clone, and scaffold projects your way",
}

// Execute is called by main.go to start the CLI.
// Cobra takes over and routes to the right command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
