package agent

import (
	"fmt"

	"morpherctl/internal/agent"
	"morpherctl/internal/config"

	"github.com/spf13/cobra"
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify morpher agent installation",
	Long:  `Verify that morpher agent is properly installed and configured.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		return verifyAgent()
	},
}

func verifyAgent() error {
	fmt.Println("Verifying morpher agent installation...")

	// Get controller IP from config.
	configMgr := config.NewManager("")
	controllerIP, err := configMgr.GetString("controller.ip")
	if err != nil {
		return fmt.Errorf("failed to get controller IP from config: %w", err)
	}

	// Create agent instance.
	ag := agent.NewAgent(controllerIP)

	// Verify installation.
	result, err := ag.Verify()
	if err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}

	// Display results.
	fmt.Println("\nVerification Results:")
	fmt.Println("=====================")

	// Binary installation.
	if result.BinaryInstalled {
		fmt.Printf("Binary:     INSTALLED (%s)\n", result.BinaryPath)
	} else {
		fmt.Println("Binary:     NOT INSTALLED")
	}

	// Service status.
	if result.ServiceStatus != "" {
		fmt.Printf("Service:    %s\n", result.ServiceStatus)
	} else {
		fmt.Println("Service:    NOT FOUND")
	}

	// Service enabled.
	if result.ServiceEnabled {
		fmt.Println("Enabled:    YES")
	} else {
		fmt.Println("Enabled:    NO")
	}

	// Overall status.
	fmt.Println("\nOverall Status:")
	switch {
	case result.BinaryInstalled && result.ServiceRunning && result.ServiceEnabled:
		fmt.Println("PASS - Agent is properly installed and running")
	case result.BinaryInstalled && result.ServiceRunning:
		fmt.Println("WARNING - Agent is running but not enabled")
	case result.BinaryInstalled:
		fmt.Println("WARNING - Agent is installed but service is not running")
	default:
		fmt.Println("FAIL - Agent is not properly installed")
	}

	return nil
}
