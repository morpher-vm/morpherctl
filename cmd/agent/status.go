package agent

import (
	"fmt"

	"morpherctl/internal/agent"
	"morpherctl/internal/config"

	"github.com/spf13/cobra"
)

var (
	detailed bool
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show morpher agent status",
	Long:  `Display the current status of morpher agent installation and service.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		return showStatus()
	},
}

func init() {
	statusCmd.Flags().BoolVar(&detailed, "detailed", false, "Show detailed status information")
}

func showStatus() error {
	fmt.Println("Checking morpher agent status...")

	// Get controller IP from config.
	configMgr := config.NewManager("")
	controllerIP, err := configMgr.GetString("controller.ip")
	if err != nil {
		return fmt.Errorf("failed to get controller IP from config: %w", err)
	}

	// Create agent instance.
	ag := agent.NewAgent(controllerIP)

	// Get status.
	status, err := ag.Status()
	if err != nil {
		return fmt.Errorf("failed to get agent status: %w", err)
	}

	// Display status.
	fmt.Println("\nAgent Status:")
	fmt.Println("=============")

	// Installation status.
	if status.Installed {
		fmt.Printf("Installed:  YES (%s)\n", status.InstallPath)
	} else {
		fmt.Println("Installed:  NO")
	}

	// Service status.
	if status.ServiceStatus != "" {
		fmt.Printf("Service:    %s\n", status.ServiceStatus)
	} else {
		fmt.Println("Service:    NOT FOUND")
	}

	// Service enabled.
	if status.ServiceEnabled {
		fmt.Println("Enabled:    YES")
	} else {
		fmt.Println("Enabled:    NO")
	}

	// Detailed information if requested.
	if detailed && status.Installed {
		fmt.Println("\nDetailed Information:")
		fmt.Println("=====================")
		fmt.Printf("Binary Path:    %s\n", status.InstallPath)
		fmt.Printf("Controller IP:  %s\n", controllerIP)
		fmt.Printf("Architecture:   %s\n", ag.Architecture)
	}

	// Summary.
	fmt.Println("\nSummary:")
	switch {
	case status.Installed && status.ServiceStatus == "active" && status.ServiceEnabled:
		fmt.Println("Agent is running and properly configured")
	case status.Installed && status.ServiceStatus == "active":
		fmt.Println("Agent is running but not enabled at boot")
	case status.Installed:
		fmt.Println("Agent is installed but service is not running")
	default:
		fmt.Println("Agent is not installed")
	}

	return nil
}
