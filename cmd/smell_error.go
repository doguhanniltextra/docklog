package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// smellErrorCmd is a highly specialized child command of "start".
// It is designed specifically to track down application panics, timeouts, and
// critical failures across all (or a filtered subset) of running containers.
//
// By inheriting configuration from startCmd, it respects global bounds (e.g. --container)
// but forcibly overrides specific internal behaviors:
//  1. Enforces a case-insensitive substring filter for "error".
//  2. Enables Deduplication to ensure that rapid error loops do not overwhelm
//     the terminal output and bury the root cause.
var smellErrorCmd = &cobra.Command{
	Use:   "smell-error",
	Short: "Finds and aggregates ERROR logs without spam",
	Long:  `Automatically filters for ERRORs and prevents identical errors from spamming the terminal consecutively.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Looking for ERRORs... Spam filter is active. Press Ctrl+C to exit.")

		cfg, err := buildConfig()
		if err != nil {
			return err
		}
		cfg.Filter = "error"   // Auto filter for error
		cfg.Deduplicate = true // Enable our new spam filter

		return runPipeline(cfg)
	},
}

func init() {
	rootCmd.AddCommand(smellErrorCmd)
}
