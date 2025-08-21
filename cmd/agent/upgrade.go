package agent

import (
	"fmt"
	"os"
	"os/exec"

	"morpherctl/internal/agent"
	"morpherctl/internal/config"

	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade [version]",
	Short: "Upgrade morpher agent",
	Long:  `Upgrade morpher agent to a new version.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		return upgradeAgent(args)
	},
}

func upgradeAgent(args []string) error {
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

	fmt.Printf("Upgrading morpher agent to %s...\n", version)
	fmt.Printf("Controller IP: %s\n", controllerIP)

	// Create agent instance.
	ag := agent.NewAgent(controllerIP)

	// Check current status.
	fmt.Print("Checking current installation... ")
	status, err := ag.Status()
	if err != nil {
		fmt.Println("FAILED")
		return fmt.Errorf("failed to check agent status: %w", err)
	}

	if !status.Installed {
		fmt.Println("FAILED")
		return fmt.Errorf("agent is not installed")
	}
	fmt.Println("OK")

	// Run preflight checks.
	fmt.Print("Running preflight checks... ")
	if err := runUpgradePreflight(); err != nil {
		fmt.Println("FAILED")
		return fmt.Errorf("preflight failed: %w", err)
	}
	fmt.Println("OK")

	// Upgrade agent.
	fmt.Print("Upgrading agent... ")
	if err := ag.Upgrade(version); err != nil {
		fmt.Println("FAILED")
		return fmt.Errorf("upgrade failed: %w", err)
	}
	fmt.Println("OK")

	fmt.Println("Upgrade completed successfully")
	return nil
}

func runUpgradePreflight() error {
	// Check if sudo is available.
	if _, err := exec.LookPath("sudo"); err != nil {
		return fmt.Errorf("sudo is not available - upgrade requires sudo privileges")
	}

	// Check if systemd is available (Linux) or launchd (macOS).
	if _, err := os.Stat("/run/systemd/system"); os.IsNotExist(err) {
		// Check if launchd is available (macOS).
		if _, err := exec.LookPath("launchctl"); err != nil {
			return fmt.Errorf("neither systemd nor launchd is available - unsupported platform")
		}
		fmt.Println("Using launchd (macOS)")
	} else {
		fmt.Println("Using systemd (Linux)")
	}

	// Check if user can run sudo commands.
	if os.Geteuid() != 0 {
		fmt.Print("Checking sudo privileges... ")
		cmd := exec.Command("sudo", "-n", "true")
		if err := cmd.Run(); err != nil {
			fmt.Println("FAILED")
			return fmt.Errorf("sudo privileges required - please run with 'sudo morpherctl agent upgrade' or ensure sudo access")
		}
		fmt.Println("OK")
	}

	return nil
}
