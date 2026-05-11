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
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig("cluster.yaml")
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}
		fmt.Println("Collecting node stats...")
		results := internalssh.CollectAll(cfg.Nodes, cfg.SSHKeyPath)
		internalssh.PrintStatsTable(results)
	},
}

func init() {
	rootCmd.AddCommand(infoCMD)
}