package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/doguhanniltextra/docklog/pkg/source/docker"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List currently running containers and their monitoring status",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := buildConfig()
		if err != nil {
			return err
		}

		src, err := docker.NewDockerSource(cfg)
		if err != nil {
			return err
		}

		containers, err := src.ListContainers(context.Background())
		if err != nil {
			return fmt.Errorf("error listing containers: %v", err)
		}

		if len(containers) == 0 {
			fmt.Println("No running containers found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "CONTAINER ID\tIMAGE\tNAME\tSTATUS\tDOCKLOG")

		green := color.New(color.FgGreen, color.Bold)
		gray := color.New(color.FgHiBlack)

		for _, c := range containers {
			status := gray.Sprint("IGNORED")
			if c.IsMatched {
				status = green.Sprint("WATCHING")
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", c.ID, c.Image, c.Name, c.Status, status)
		}
		w.Flush()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
