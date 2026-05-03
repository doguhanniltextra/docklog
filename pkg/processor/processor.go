// Package processor provides a pipeline-based architecture for manipulating and
// filtering log messages as they flow from the source to the sinks.
package processor

import (
	"github.com/doguhanniltextra/docklog/pkg/types"
)

// Processor defines the interface for components that can modify or filter log messages.
// Implementations can be chained together to create complex log processing pipelines.
type Processor interface {
	// Process handles a single log message.
	// It returns the (potentially modified) message and a boolean 'keep'.
	// If 'keep' is false, the message is dropped and will not proceed further
	// in the processing pipeline.
	Process(msg *types.LogMessage) (processed *types.LogMessage, keep bool)
}
