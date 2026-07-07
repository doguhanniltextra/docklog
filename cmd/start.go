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

	// String fields — Viper resolves from flags/env/config file with correct defaults
	cfg.Filter = viper.GetString("filter")
	cfg.Exclude = viper.GetString("exclude")
	cfg.Since = viper.GetString("since")
	cfg.Output = viper.GetString("output")
	cfg.TailLines = viper.GetString("tail")

	// Boolean fields — zero-value (false) is already the default in DefaultConfig
	cfg.Deduplicate = viper.GetBool("dedupe")
	cfg.JsonOutput = viper.GetBool("json")
	cfg.Redact = viper.GetBool("redact")
	cfg.ShowTimestamps = viper.GetBool("timestamps")
	cfg.NoColor = viper.GetBool("no-color")

	// Integer fields
	cfg.BufferLength = viper.GetInt("buffer")

	// Regex fields — these need validation, extracted into a helper
	if err := compileRegexFlag("regex", &cfg.RegexFilter); err != nil {
		return nil, err
	}
	if err := compileRegexFlag("container", &cfg.ContainerFilter); err != nil {
		return nil, err
	}

	return cfg, nil
}

// compileRegexFlag reads a Viper string key and compiles it into a *regexp.Regexp.
// If the key is empty, the target is left as nil. Returns a descriptive error
// if the pattern is invalid.
func compileRegexFlag(key string, target **regexp.Regexp) error {
	pattern := viper.GetString(key)
	if pattern == "" {
		return nil
	}
	r, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid --%s regex %q: %v", key, pattern, err)
	}
	*target = r
	return nil
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
