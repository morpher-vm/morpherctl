package config

import (
	"github.com/spf13/cobra"
)

var (
	configFile string
)

var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage morpherctl configuration",
	Long:  `Manage morpherctl configuration including initialization, setting, getting, and displaying config values.`,
}

func init() {
	// Add subcommands.
	ConfigCmd.AddCommand(initCmd, setCmd, getCmd, showCmd)

	// Set configuration file path.
	ConfigCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is $HOME/.morpherctl/config.yaml)")
}
