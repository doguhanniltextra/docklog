package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/doguhanniltextra/docklog/pkg/aggregator"
	"github.com/doguhanniltextra/docklog/pkg/config"
	"github.com/doguhanniltextra/docklog/pkg/formatter"
	"github.com/doguhanniltextra/docklog/pkg/processor"
	"github.com/doguhanniltextra/docklog/pkg/sink"
	"github.com/doguhanniltextra/docklog/pkg/source/docker"
	"github.com/spf13/cobra"
)

func runPipeline(cfg *config.Config) error {
	// Initialize Source
	src, err := docker.NewDockerSource(cfg)
	if err != nil {
		return err
	}

	// Initialize Processors
	var processors []processor.Processor
	processors = append(processors, processor.NewFilterProcessor(cfg.Filter, cfg.Exclude, cfg.RegexFilter))

	if cfg.Redact {
		processors = append(processors, processor.NewRedactProcessor())
	}

	if cfg.Deduplicate {
		processors = append(processors, processor.NewDedupeProcessor())
	}

	// Initialize Formatter
	var fmttr formatter.Formatter
	if cfg.JsonOutput {
		fmttr = formatter.NewJsonFormatter()
	} else {
		fmttr = formatter.NewHumanFormatter(cfg.TimeFormat, cfg.Colors)
	}

	// Initialize Sink
	var sinks []sink.Sink
	sinks = append(sinks, sink.NewConsoleSink(fmttr))

	if cfg.Output != "" {
		var fileFmttr formatter.Formatter
		if cfg.JsonOutput {
			fileFmttr = formatter.NewJsonFormatter()
		} else {
			fileFmttr = formatter.NewHumanFormatter(cfg.TimeFormat, cfg.Colors)
		}
		fSink, err := sink.NewFileSink(cfg.Output, fileFmttr)
		if err != nil {
			return err
		}
		sinks = append(sinks, fSink)
	}

	snk := sink.NewMultiSink(sinks...)

	// Initialize and Run Pipeline
	pipe := aggregator.NewPipeline(src, processors, snk, cfg.BufferLength)

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

	return pipe.Run(ctx)
}

// runCmd represents the 'run' command, which acts as a modern alias for 'start'.
// It leverages the same underlying pipeline architecture.
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Starts the real-time log aggregator (alias for start)",
	Long: `Connects to the Docker socket, discovers running containers, 
and streams their logs to the terminal using the modular pipeline architecture.`,
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
	rootCmd.AddCommand(runCmd)
}
