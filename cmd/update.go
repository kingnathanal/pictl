package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/kingnathanal/pictl/internal/config"
	"github.com/kingnathanal/pictl/internal/ssh"
)

var updateCMD = &cobra.Command{
	Use:   "update",
	Short: "Run apt update && apt upgrade on all cluster nodes",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig(getConfigPath())
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		fmt.Println("Running apt update on all nodes...")
		results := ssh.UpdateAll(cfg.Nodes, cfg.SSHKeyPath)
		ssh.PrintUpdateResults(results)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCMD)
}