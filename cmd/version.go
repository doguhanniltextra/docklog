package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version = "v1.1.0"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Docklog",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Docklog %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
