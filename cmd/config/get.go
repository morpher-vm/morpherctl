package config

import (
	"fmt"

	"morpherctl/internal/config"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get configuration value",
	Long:  `Get a configuration value by key.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		return getConfig(args[0])
	},
}

func getConfig(key string) error {
	// Initialize configuration manager.
	configMgr := config.NewManager(configFile)

	// Get configuration value.
	value, err := configMgr.Get(key)
	if err != nil {
		return fmt.Errorf("failed to get configuration value: %w", err)
	}

	fmt.Printf("%s = %v\n", key, value)
	return nil
}
