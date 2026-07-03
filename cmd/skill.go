package cmd

import (
	"fmt"
	"os"

	"github.com/kingnathanal/pictl/internal/skill"
	"github.com/spf13/cobra"
)

var forceSkillInstall bool

var skillCMD = &cobra.Command{
	Use:   "skill",
	Short: "Manage the AI skill for using pictl",
	Args:  cobra.NoArgs,
}

var skillInstallCMD = &cobra.Command{
	Use:   "install",
	Short: "Install the pictl AI skill into supported project directories",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		workingDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting current directory: %w", err)
		}

		results, err := skill.InstallAll(workingDir, forceSkillInstall)
		if err != nil {
			return fmt.Errorf("installing pictl skill: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "pictl skill v%s\n", skill.Version)
		for _, result := range results {
			fmt.Fprintf(cmd.OutOrStdout(), "%-9s %s\n", result.Action, result.Path)
		}
		return nil
	},
}

var skillStatusCMD = &cobra.Command{
	Use:   "status",
	Short: "Show the installation and version status of the pictl AI skill",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		workingDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting current directory: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Bundled pictl skill: v%s\n", skill.Version)
		for _, result := range skill.InspectAll(workingDir) {
			installed := result.InstalledVersion
			if installed == "" {
				installed = "-"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%-18s %-18s installed=%-8s %s\n", result.Target, result.State, installed, result.Detail)
		}
		return nil
	},
}

func init() {
	skillInstallCMD.Flags().BoolVar(&forceSkillInstall, "force", false, "Replace existing skill files with different content")
	skillCMD.AddCommand(skillInstallCMD)
	skillCMD.AddCommand(skillStatusCMD)
	rootCmd.AddCommand(skillCMD)
}
