package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"morpherctl/cmd/version"
)

var rootCmd = &cobra.Command{
	Use:           "morpherctl",
	Short:         "CLI tool for managing morpher agents that perform VM migrations",
	Long:          `A command-line tool for managing morpher agents that perform VM migrations.`,
	SilenceErrors: true,
	SilenceUsage:  true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		rootCmd.PrintErrf("Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(version.VersionCmd)
}
