package config

import (
	"fmt"

	"morpherctl/internal/config"

	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show all configuration",
	Long:  `Display all current configuration values.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		return showConfig()
	},
}

func showConfig() error {
	// Initialize configuration manager.
	configMgr := config.NewManager(configFile)

	// Get all configuration values.
	allConfig, err := configMgr.GetAll()
	if err != nil {
		return fmt.Errorf("failed to get all configuration values: %w", err)
	}

	fmt.Println("Current configuration:")
	for key, value := range allConfig {
		fmt.Printf("  %s = %v\n", key, value)
	}
	return nil
}
