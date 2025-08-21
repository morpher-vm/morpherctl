package agent

import (
	"fmt"
	"os"

	"morpherctl/internal/agent"
	"morpherctl/internal/config"

	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall morpher agent",
	Long:  `Remove morpher agent and all associated files from the system.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		return uninstallAgent()
	},
}

func uninstallAgent() error {
	fmt.Println("Uninstalling morpher agent...")

	// Get controller IP from config.
	configMgr := config.NewManager("")
	controllerIP, err := configMgr.GetString("controller.ip")
	if err != nil {
		return fmt.Errorf("failed to get controller IP from config: %w", err)
	}

	// Create agent instance.
	ag := agent.NewAgent(controllerIP)

	// Check if agent is installed.
	status, err := ag.Status()
	if err != nil {
		return fmt.Errorf("failed to check agent status: %w", err)
	}

	if !status.Installed {
		return fmt.Errorf("agent is not installed")
	}

	// Confirm uninstallation.
	fmt.Print("Are you sure you want to uninstall the agent? (y/N): ")
	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		fmt.Println("Failed to read input, proceeding with uninstallation")
	} else if response != "y" && response != "Y" {
		fmt.Println("Uninstallation cancelled")
		return nil
	}

	// Run preflight checks.
	fmt.Print("Running preflight checks... ")
	if err := runUninstallPreflight(); err != nil {
		fmt.Println("FAILED")
		return fmt.Errorf("preflight failed: %w", err)
	}
	fmt.Println("OK")

	// Uninstall agent.
	fmt.Print("Uninstalling agent... ")
	if err := ag.Uninstall(); err != nil {
		fmt.Println("FAILED")
		return fmt.Errorf("uninstallation failed: %w", err)
	}
	fmt.Println("OK")

	fmt.Println("Uninstallation completed successfully")
	return nil
}

func runUninstallPreflight() error {
	// Check if running as root or with sudo.
	if os.Geteuid() != 0 {
		return fmt.Errorf("uninstallation requires root privileges")
	}

	// Check if systemctl is available.
	if _, err := os.Stat("/run/systemd/system"); os.IsNotExist(err) {
		return fmt.Errorf("systemd is not available")
	}

	return nil
}
