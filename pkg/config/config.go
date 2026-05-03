package config

import (
	"regexp"

	"github.com/fatih/color"
)

// Version is the current version of the application.
// It is defined here to allow build-time injection using -ldflags.
var Version = "v1.1.0"

// Config defines the operational parameters for the Log Aggregator.
// It acts as the central configuration object that determines how logs are
// collected, filtered, and presented to the user.
//
// By utilizing a unified Config struct, the Aggregator is decoupled from the CLI layer,
// allowing it to be easily tested or embedded in other systems.
type Config struct {
	// Filter defines a simple substring to match against log lines (case-insensitive).
	// If set, only lines containing this substring will be emitted.
	Filter string

	// Exclude defines a substring that will cause log lines to be dropped.
	// This is evaluated before other filtering mechanisms to drop noisy logs early.
	Exclude string

	// RegexFilter specifies a regular expression used to filter log lines.
	// If set, only lines matching the regular expression will be emitted.
	// This provides fine-grained control for power-users over standard substring matching.
	RegexFilter *regexp.Regexp

	// ContainerFilter specifies a regular expression used to filter which containers
	// are monitored. If a container's name does not match this regex, its log stream
	// is never initialized, saving goroutines and memory.
	ContainerFilter *regexp.Regexp

	// TailLines defines the number of lines to show from the end of the logs
	// when a new container stream is initialized (e.g., "10", "all").
	TailLines string

	// Since restricts log collection to logs generated after a specific timestamp
	// or duration (e.g., "2013-01-02T13:23:37" or "42m").
	Since string

	// Output is the file path where logs should be persisted. If empty, logs
	// are only printed to stdout.
	Output string

	// JsonOutput indicates whether the aggregator should format the output as
	// structured JSON strings instead of human-readable colored text.
	// Highly recommended for machine-processing or integrations with ELK/Splunk.
	JsonOutput bool

	// TimeFormat specifies the Go time format string used when displaying
	// human-readable timestamps (default: "15:04:05.000").
	TimeFormat string

	// BufferLength controls the size of the internal log channel.
	// A larger buffer prevents blocking the Docker stream readers during
	// bursts of high log volume.
	BufferLength int

	// Colors holds the palette used to color-code container names in the terminal.
	// The aggregator cycles through this palette to assign consistent colors
	// to individual containers.
	Colors []*color.Color

	// Deduplicate enables the spam-prevention mechanism.
	// When true, consecutive identical log messages from the same container
	// are dropped to keep the terminal output clean.
	Deduplicate bool

	// Redact enables the automatic redaction of sensitive information.
	// When true, emails, IPv4 addresses, and tokens are replaced with "***".
	Redact bool

	// ShowTimestamps indicates whether the Docker log streams should include
	// RFC3339 timestamps. Note that this is different from the local aggregation
	// timestamp shown in the human-readable output.
	ShowTimestamps bool
}

// DefaultConfig instantiates and returns the default configuration for the aggregator.
// It initializes sensible defaults for terminal output, including a standard color palette
// and a generous buffer size to handle high-throughput environments.
func DefaultConfig() *Config {
	return &Config{
		Filter:         "",
		TailLines:      "10",
		TimeFormat:     "15:04:05.000",
		BufferLength:   1000,
		ShowTimestamps: false,
		Colors: []*color.Color{
			color.New(color.FgCyan),
			color.New(color.FgGreen),
			color.New(color.FgYellow),
			color.New(color.FgMagenta),
			color.New(color.FgBlue),
			color.New(color.FgHiCyan),
			color.New(color.FgHiGreen),
			color.New(color.FgHiYellow),
			color.New(color.FgHiMagenta),
		},
	}
}
