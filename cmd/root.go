package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var configPath string

var rootCmd = &cobra.Command{
	Use:   "pictl",
	Short: "pictl — manage a fleet of Raspberry Pis over SSH",
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Path to cluster configuration file")
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
