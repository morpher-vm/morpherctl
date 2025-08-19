package controller

import (
	"context"
	"fmt"

	"morpherctl/internal/controller"

	"github.com/spf13/cobra"
)

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Ping the controller",
	Long:  `Send a ping request to the morpher controller to check connectivity.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		return pingController()
	},
}

func pingController() error {
	// Create controller client.
	client, timeout, err := controller.CreateControllerClient()
	if err != nil {
		return fmt.Errorf("failed to create controller client: %w", err)
	}

	fmt.Printf("Sending ping request to controller: %s\n", client.GetBaseURL())

	// Send ping request.
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	response, err := client.Ping(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to controller: %w", err)
	}

	// Display response.
	if response.Success {
		fmt.Println("Controller responded successfully")
		if response.ResponseTime != "" {
			fmt.Printf("Response time: %v\n", response.ResponseTime)
		}
	} else {
		fmt.Printf("Controller responded but with unexpected status code: %d\n", response.StatusCode)
	}

	return nil
}
