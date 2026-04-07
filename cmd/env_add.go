package cmd

import (
	"fmt"

	"github.com/GuechtouliAnis/forge/internal/env"
	"github.com/spf13/cobra"
)

var (
	envAddDB         bool
	envAddAI         bool
	envAddWeb        bool
	envAddRedis      bool
	envAddMonitoring bool
	envAddNeo4j      bool
)

// envAddCmd appends predefined variable sets to a .env file.
// Skips keys that already exist and warns per skipped key.
// Use --db, --ai, --web to select which variable sets to append.
var envAddCmd = &cobra.Command{
	Use:   "add [path]",
	Short: "Append predefined variable sets to .env",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		path := ".env"
		if len(args) > 0 {
			path = args[0]
		}

		var selected []string
		if envAddDB {
			selected = append(selected, "db")
		}
		if envAddAI {
			selected = append(selected, "ai")
		}
		if envAddWeb {
			selected = append(selected, "web")
		}
		if envAddRedis {
			selected = append(selected, "redis")
		}
		if envAddMonitoring {
			selected = append(selected, "monitoring")
		}
		if envAddNeo4j {
			selected = append(selected, "neo4j")
		}

		if len(selected) == 0 {
			return fmt.Errorf("ERROR: no flag provided — use --db, --ai, --web, --redis, --monitoring, or --neo4j")
		}

		return env.AddEnv(path, selected)
	},
}

func init() {
	envAddCmd.Flags().BoolVar(&envAddDB, "db", false, "append database variables")
	envAddCmd.Flags().BoolVar(&envAddAI, "ai", false, "append AI/LLM variables")
	envAddCmd.Flags().BoolVar(&envAddWeb, "web", false, "append web server variables")
	envAddCmd.Flags().BoolVar(&envAddRedis, "redis", false, "append Redis variables")
	envAddCmd.Flags().BoolVar(&envAddMonitoring, "monitoring", false, "append Grafana/Prometheus variables")
	envAddCmd.Flags().BoolVar(&envAddNeo4j, "neo4j", false, "append Neo4j variables")
	envCmd.AddCommand(envAddCmd)
}
