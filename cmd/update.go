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
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig("cluster.yaml")
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}
		fmt.Println("Running apt update on all nodes...")
		results := ssh.UpdateAll(cfg.Nodes, cfg.SSHKeyPath)
		ssh.PrintUpdateResults(results)
	},
}

func init() {
	rootCmd.AddCommand(updateCMD)
}