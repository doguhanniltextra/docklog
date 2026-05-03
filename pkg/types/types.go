// Package types defines the core data structures shared across the docklog ecosystem.
// Centralizing these types prevents circular dependencies between the aggregator,
// processors, and formatters.
package types

import "time"

// LogMessage encapsulates a single line of log output retrieved from a container.
// It serves as the primary data transfer object (DTO) within the logging pipeline,
// carrying both the raw content and essential metadata for downstream processing.
type LogMessage struct {
	// ContainerName is the canonical name of the Docker container (without the leading slash).
	ContainerName string `json:"container"`
	// Message is the raw log string output by the container, stripped of Docker's
	// binary protocol headers.
	Message string `json:"message"`
	// IsError indicates if the message originated from the container's STDERR stream.
	// This allows formatters to apply specific styling (e.g., red text).
	IsError bool `json:"is_error"`
	// Timestamp is the moment the log line was captured by the docklog aggregator.
	Timestamp time.Time `json:"timestamp"`
}
