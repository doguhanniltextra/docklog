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
	viper.BindPFlag("container", rootCmd.PersistentFlags().Lookup("container"))
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
