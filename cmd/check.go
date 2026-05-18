package cmd

import (
	"fmt"

	"github.com/kingnathanal/pictl/internal/config"
	"github.com/kingnathanal/pictl/internal/ssh"
	"github.com/spf13/cobra"
)

var checkCMD = &cobra.Command{ 
	Use: "check",
	Short: "Check the status of all nodes in the cluster",
	RunE: func(cmd *cobra.Command, args []string) error { 
		cfg, err := config.LoadConfig(getConfigPath())
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		fmt.Println("Checking node status...")
		nodes := config.FilterNodes(cfg.Nodes, targetNode, targetRole)
		results := ssh.CheckAll(nodes, cfg.SSHKeyPath)
		ssh.PrintCheckResults(results)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkCMD)
}