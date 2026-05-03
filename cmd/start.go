package cmd

import (
	"fmt"
	"regexp"

	"github.com/doguhanniltextra/docklog/pkg/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// buildConfig acts as the configuration translation layer.
// It reads the globally resolved state from Viper (which transparently merges
// CLI flags, environment variables, and the .docklog.yaml configuration file)
// and maps it into the internal config.Config structure.
// Returns an error if any regex filters provided by the user are invalid.
func buildConfig() (*config.Config, error) {
	cfg := config.DefaultConfig()
	if filter := viper.GetString("filter"); filter != "" {
		cfg.Filter = filter
	}
	if exclude := viper.GetString("exclude"); exclude != "" {
		cfg.Exclude = exclude
	}
	if regexStr := viper.GetString("regex"); regexStr != "" {
		r, err := regexp.Compile(regexStr)
		if err != nil {
			return nil, fmt.Errorf("invalid regex: %v", err)
		}
		cfg.RegexFilter = r
	}
	if contStr := viper.GetString("container"); contStr != "" {
		r, err := regexp.Compile(contStr)
		if err != nil {
			return nil, fmt.Errorf("invalid container regex: %v", err)
		}
		cfg.ContainerFilter = r
	}
	if since := viper.GetString("since"); since != "" {
		cfg.Since = since
	}
	if out := viper.GetString("output"); out != "" {
		cfg.Output = out
	}
	if tail := viper.GetString("tail"); tail != "" {
		cfg.TailLines = tail
	}
	if viper.GetBool("dedupe") {
		cfg.Deduplicate = true
	}
	if viper.GetBool("json") {
		cfg.JsonOutput = true
	}
	if viper.GetBool("redact") {
		cfg.Redact = true
	}
	if viper.GetBool("timestamps") {
		cfg.ShowTimestamps = true
	}
	return cfg, nil
}

// startCmd represents the primary entry point for the standard logging behavior.
// When invoked, it establishes a connection to the Docker daemon, initializes
// the aggregator, and blocks until interrupted via SIGINT/SIGTERM.
//
// All flags attached to startCmd are marked as Persistent, allowing child
// commands (e.g., smell-error) to inherit them and override specific behaviors
// without duplicating CLI flag registration logic.
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the real-time log aggregator",
	Long: `Connects to the Docker socket, discovers running containers, 
and streams their logs to the terminal.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Docklog is starting. Press Ctrl+C to exit.")

		cfg, err := buildConfig()
		if err != nil {
			return err
		}

		return runPipeline(cfg)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
