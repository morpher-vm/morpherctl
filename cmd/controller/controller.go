package controller

import (
	"github.com/spf13/cobra"
)

var ControllerCmd = &cobra.Command{
	Use:   "controller",
	Short: "Manage morpher controller",
	Long:  `Manage morpher controller including ping, status, and info operations.`,
}

func init() {
	ControllerCmd.AddCommand(pingCmd, infoCmd)
}
