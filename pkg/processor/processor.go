package processor

import (
	"github.com/doguhanniltextra/docklog/pkg/types"
)

// Processor defines the interface for components that can modify or filter log messages.
type Processor interface {
	// Process handles a single log message. 
	// Returns the processed message and a boolean indicating if the message should continue through the pipeline.
	// If it returns false, the message is dropped.
	Process(msg *types.LogMessage) (*types.LogMessage, bool)
}
