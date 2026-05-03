package source

import (
	"context"

	"github.com/doguhanniltextra/docklog/pkg/types"
)

// LogSource defines the interface for components that produce log messages.
type LogSource interface {
	Run(ctx context.Context, logChan chan<- types.LogMessage) error
}
