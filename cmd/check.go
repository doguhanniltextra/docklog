package cmd

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Tests the Docker connection",
	Long:  `Attempts to connect to the local Docker daemon and checks its status.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Looking for Docker daemon...")

		// Create Docker client from environment variables (default settings)
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			fmt.Printf("Failed to create Docker Client: %v\n", err)
			return
		}
		defer cli.Close()

		// Send a "Ping" to the Docker Engine
		ping, err := cli.Ping(context.Background())
		if err != nil {
			fmt.Printf("Could not connect to Docker. Is Docker Desktop/Engine running?\nError: %v\n", err)
			return
		}

		fmt.Printf(" Success! Docker API Version: %s\n", ping.APIVersion)
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
