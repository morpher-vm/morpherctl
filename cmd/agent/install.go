package agent

import (
	"fmt"
	"os"

	"morpherctl/internal/agent"
	"morpherctl/internal/config"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install [version]",
	Short: "Install morpher agent",
	Long:  `Install morpher agent with specified version or latest version.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		return installAgent(args)
	},
}

func installAgent(args []string) error {
	version := "latest"
	if len(args) > 0 {
		version = args[0]
	}

	// Get controller IP from config.
	configMgr := config.NewManager("")
	controllerIP, err := configMgr.GetString("controller.ip")
	if err != nil {
		return fmt.Errorf("failed to get controller IP from config: %w", err)
	}

	fmt.Printf("Installing morpher agent %s...\n", version)
	fmt.Printf("Controller IP: %s\n", controllerIP)

	// Create agent instance.
	ag := agent.NewAgent(controllerIP)

	// Run preflight checks.
	fmt.Print("Running preflight checks... ")
	if err := runPreflight(); err != nil {
		fmt.Println("FAILED")
		return fmt.Errorf("preflight failed: %w", err)
	}
	fmt.Println("OK")

	// Install agent.
	fmt.Print("Installing agent... ")
	if err := ag.Install(version); err != nil {
		fmt.Println("FAILED")
		return fmt.Errorf("installation failed: %w", err)
	}
	fmt.Println("OK")

	fmt.Println("Installation completed successfully")
	return nil
}

func runPreflight() error {
	// Check if running as root or with sudo.
	if os.Geteuid() != 0 {
		return fmt.Errorf("installation requires root privileges")
	}

	// Check if systemctl is available.
	if _, err := os.Stat("/run/systemd/system"); os.IsNotExist(err) {
		return fmt.Errorf("systemd is not available")
	}

	return nil
}
