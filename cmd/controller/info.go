package controller

import (
	"context"
	"fmt"

	"morpherctl/internal/controller"

	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Get controller information",
	Long:  `Get detailed information about the morpher controller.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		return getControllerInfo()
	},
}

func getControllerInfo() error {
	// Create controller client.
	client, timeout, err := controller.CreateControllerClient()
	if err != nil {
		return fmt.Errorf("failed to create controller client: %w", err)
	}

	fmt.Printf("Getting controller information: %s\n", client.GetBaseURL())

	// Get controller info.
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	response, err := client.GetInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get controller info: %w", err)
	}

	// Display response.
	if response.Success {
		fmt.Println("Successfully retrieved controller information")

		// Display detailed information if available.
		if response.Result != nil {
			fmt.Println("\nController Details:")
			fmt.Printf("  OS: %s %s (%s)\n",
				response.Result.OS.Name,
				response.Result.OS.PlatformVersion,
				response.Result.OS.KernelVersion)
			fmt.Printf("  Go Version: %s\n", response.Result.GoVersion)
			fmt.Printf("  Uptime: %s\n", response.Result.UpTime)
		} else {
			fmt.Println("  No detailed information available")
		}
	} else {
		fmt.Printf("Failed to get controller information: %d\n", response.StatusCode)
	}

	return nil
}
