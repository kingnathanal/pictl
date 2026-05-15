package cmd

import (
	"fmt"

	"github.com/kingnathanal/pictl/internal/config"
	internalssh "github.com/kingnathanal/pictl/internal/ssh"
	"github.com/spf13/cobra"
)

var infoCMD = &cobra.Command{
	Use:   "info",
	Short: "Collect and display stats about all nodes in the cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig(getConfigPath())
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		fmt.Println("Collecting node stats...")
		nodes := config.FilterNodes(cfg.Nodes, targetNode, targetRole)
		results := internalssh.CollectAll(nodes, cfg.SSHKeyPath)
		internalssh.PrintStatsTable(results)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(infoCMD)
}