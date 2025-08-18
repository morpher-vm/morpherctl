package config

import (
	"fmt"

	"morpherctl/internal/config"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration file",
	Long:  `Initialize a new configuration file with default values.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		return initConfig()
	},
}

func initConfig() error {
	// Initialize configuration manager.
	configMgr := config.NewManager(configFile)

	// Initialize configuration.
	if err := configMgr.Init(); err != nil {
		return fmt.Errorf("failed to initialize configuration: %w", err)
	}

	fmt.Printf("Configuration file initialized: %s\n", configMgr.GetConfigFile())
	return nil
}
