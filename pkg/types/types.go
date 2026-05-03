package types

import "time"

// LogMessage encapsulates a single line of log output retrieved from a container.
type LogMessage struct {
	// ContainerName is the canonical name of the Docker container.
	ContainerName string `json:"container"`
	// Message is the raw log string output by the container.
	Message string `json:"message"`
	// IsError is true if the message was read from the container's STDERR stream.
	IsError bool `json:"is_error"`
	// Timestamp is the exact time the message was aggregated locally.
	Timestamp time.Time `json:"timestamp"`
}
