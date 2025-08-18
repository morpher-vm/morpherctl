package config

import (
	"fmt"

	"morpherctl/internal/config"

	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set configuration value",
	Long:  `Set a configuration key-value pair.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(_ *cobra.Command, args []string) error {
		return setConfig(args[0], args[1])
	},
}

func setConfig(key, value string) error {
	// Initialize configuration manager.
	configMgr := config.NewManager(configFile)

	// Set configuration value.
	if err := configMgr.Set(key, value); err != nil {
		return fmt.Errorf("failed to set configuration value: %w", err)
	}

	fmt.Printf("Configuration updated: %s = %s\n", key, value)
	return nil
}
