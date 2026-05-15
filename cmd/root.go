package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var configPath string
var targetNode string
var targetRole string

var rootCmd = &cobra.Command{
	Use:   "pictl",
	Short: "pictl — manage a fleet of Raspberry Pis over SSH",
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Path to cluster configuration file")
	rootCmd.PersistentFlags().StringVar(&targetNode, "node", "", "Target a specific node by name")
	rootCmd.PersistentFlags().StringVar(&targetRole, "role", "", "Target nodes by role (e.g. 'web', 'db')")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func getConfigPath() string {
	if configPath != "" {
		return configPath
	}

	if envPath := os.Getenv("PICTL_CONFIG"); envPath != "" {
		return envPath
	}

	return "cluster.yaml"
}
