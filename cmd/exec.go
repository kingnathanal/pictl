package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/kingnathanal/pictl/internal/config"
	"github.com/kingnathanal/pictl/internal/ssh"
)


var execCMD = &cobra.Command{
	Use:   "exec [command]",
	Short: "Run an arbitrary shell command on all cluster nodes",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig(getConfigPath())
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		command := args[0]
		nodes := config.FilterNodes(cfg.Nodes, targetNode, targetRole)
		results := ssh.ExecAll(nodes, cfg.SSHKeyPath, command)
		ssh.PrintExecResults(results)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(execCMD)
}