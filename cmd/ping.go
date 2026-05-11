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
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig("cluster.yaml")
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}
		ssh.PingAll(cfg.Nodes)
	},
}

func init() {
	rootCmd.AddCommand(pingCMD)
}
