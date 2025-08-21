package agent

import (
	"github.com/spf13/cobra"
)

var AgentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage morpher agents",
	Long:  `Manage morpher agents including installation, configuration, and lifecycle management.`,
}

func init() {
	// Add subcommands.
	AgentCmd.AddCommand(installCmd, uninstallCmd, upgradeCmd, verifyCmd, statusCmd)
}
