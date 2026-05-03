package cmd

import (
	"fmt"

	"github.com/doguhanniltextra/docklog/pkg/config"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Docklog",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Docklog %s\n", config.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
