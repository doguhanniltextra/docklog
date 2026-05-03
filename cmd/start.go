package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"syscall"

	"docklog/pkg/aggregator"
	"docklog/pkg/config"

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

		agg, err := aggregator.New(cfg)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Handle graceful shutdown
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigCh
			fmt.Println("\nShutting down docklog...")
			cancel()
		}()

		return agg.Start(ctx)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.PersistentFlags().StringP("filter", "f", "", "Filter logs by a specific keyword (e.g., ERROR)")
	startCmd.PersistentFlags().StringP("exclude", "x", "", "Exclude logs containing a specific keyword")
	startCmd.PersistentFlags().StringP("regex", "r", "", "Filter logs using a regular expression")
	startCmd.PersistentFlags().StringP("container", "c", "", "Filter containers by a regular expression")
	startCmd.PersistentFlags().StringP("since", "s", "", "Show logs since timestamp (e.g. 2013-01-02T13:23:37) or relative (e.g. 42m for 42 minutes)")
	startCmd.PersistentFlags().StringP("output", "o", "", "Output logs to a file (e.g., logs.txt)")
	startCmd.PersistentFlags().StringP("tail", "t", "10", "Number of lines to show from the end of the logs")
	startCmd.PersistentFlags().BoolP("dedupe", "d", false, "Prevent consecutive identical logs from spamming")
	startCmd.PersistentFlags().Bool("json", false, "Output logs in JSON format")
	startCmd.PersistentFlags().Bool("redact", false, "Redact sensitive information (emails, IPv4, tokens)")

	viper.BindPFlag("filter", startCmd.PersistentFlags().Lookup("filter"))
	viper.BindPFlag("exclude", startCmd.PersistentFlags().Lookup("exclude"))
	viper.BindPFlag("regex", startCmd.PersistentFlags().Lookup("regex"))
	viper.BindPFlag("container", startCmd.PersistentFlags().Lookup("container"))
	viper.BindPFlag("since", startCmd.PersistentFlags().Lookup("since"))
	viper.BindPFlag("output", startCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("tail", startCmd.PersistentFlags().Lookup("tail"))
	viper.BindPFlag("dedupe", startCmd.PersistentFlags().Lookup("dedupe"))
	viper.BindPFlag("json", startCmd.PersistentFlags().Lookup("json"))
	viper.BindPFlag("redact", startCmd.PersistentFlags().Lookup("redact"))
}
