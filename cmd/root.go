// Package cmd contains the CLI layer of the Docklog application.
// It leverages the Cobra and Viper frameworks to provide robust command-line
// interfaces, POSIX-compliant flag parsing, and hierarchical configuration
// management (environment variables, config files, flags).
package cmd

import (
	"fmt"
	"os"

	"github.com/doguhanniltextra/docklog/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "docklog",
	Short:   "A real-time terminal log aggregator for Docker",
	Version: config.Version,
	Long: `Docklog is a CLI application that discovers running Docker containers 
and aggregates their stdout/stderr streams into a single, color-coded, 
and filterable terminal view.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.docklog.yaml or ./.docklog.yaml)")
	rootCmd.PersistentFlags().StringP("container", "c", "", "Filter containers by a regular expression")
	rootCmd.PersistentFlags().StringP("filter", "f", "", "Filter logs by a specific keyword (e.g., ERROR)")
	rootCmd.PersistentFlags().StringP("exclude", "x", "", "Exclude logs containing a specific keyword")
	rootCmd.PersistentFlags().StringP("regex", "r", "", "Filter logs using a regular expression")
	rootCmd.PersistentFlags().StringP("since", "s", "", "Show logs since timestamp (e.g. 2013-01-02T13:23:37) or relative (e.g. 42m for 42 minutes)")
	rootCmd.PersistentFlags().StringP("output", "o", "", "Output logs to a file (e.g., logs.txt)")
	rootCmd.PersistentFlags().StringP("tail", "t", "10", "Number of lines to show from the end of the logs")
	rootCmd.PersistentFlags().BoolP("dedupe", "d", false, "Prevent consecutive identical logs from spamming")
	rootCmd.PersistentFlags().Bool("json", false, "Output logs in JSON format")
	rootCmd.PersistentFlags().Bool("redact", false, "Redact sensitive information (emails, IPv4, tokens)")
	rootCmd.PersistentFlags().Bool("timestamps", false, "Show timestamps from Docker daemon")

	viper.BindPFlag("container", rootCmd.PersistentFlags().Lookup("container"))
	viper.BindPFlag("filter", rootCmd.PersistentFlags().Lookup("filter"))
	viper.BindPFlag("exclude", rootCmd.PersistentFlags().Lookup("exclude"))
	viper.BindPFlag("regex", rootCmd.PersistentFlags().Lookup("regex"))
	viper.BindPFlag("since", rootCmd.PersistentFlags().Lookup("since"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("tail", rootCmd.PersistentFlags().Lookup("tail"))
	viper.BindPFlag("dedupe", rootCmd.PersistentFlags().Lookup("dedupe"))
	viper.BindPFlag("json", rootCmd.PersistentFlags().Lookup("json"))
	viper.BindPFlag("redact", rootCmd.PersistentFlags().Lookup("redact"))
	viper.BindPFlag("timestamps", rootCmd.PersistentFlags().Lookup("timestamps"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".docklog")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
