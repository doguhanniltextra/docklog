package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List currently running containers",
	Run: func(cmd *cobra.Command, args []string) {
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			fmt.Printf("Error creating Docker client: %v\n", err)
			return
		}

		containers, err := cli.ContainerList(context.Background(), container.ListOptions{})
		if err != nil {
			fmt.Printf("Error listing containers: %v\n", err)
			return
		}

		if len(containers) == 0 {
			fmt.Println("No running containers found.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "CONTAINER ID\tIMAGE\tNAME\tSTATUS")
		for _, c := range containers {
			name := "N/A"
			if len(c.Names) > 0 {
				name = c.Names[0][1:] // Remove leading slash
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", c.ID[:12], c.Image, name, c.Status)
		}
		w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
