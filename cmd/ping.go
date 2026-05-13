package cmd

import (
	"fmt"

	"github.com/kingnathanal/pictl/internal/config"
	"github.com/kingnathanal/pictl/internal/ssh"
	"github.com/spf13/cobra"
)

var pingCMD = &cobra.Command{
	Use:   "ping",
	Short: "Ping all nodes in the cluster to check connectivity",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig(getConfigPath())
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		ssh.PingAll(cfg.Nodes)
		
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pingCMD)
}
