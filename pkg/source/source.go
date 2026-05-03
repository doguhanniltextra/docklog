// Package source defines the abstraction for log emitters.
// It allows docklog to potentially support multiple log providers beyond Docker,
// such as Kubernetes, local files, or remote syslog streams.
package source

import (
	"context"

	"github.com/doguhanniltextra/docklog/pkg/types"
)

// LogSource defines the contract for any component that generates log messages.
// Implementations are expected to be long-running and push messages into the
// provided channel until the context is canceled.
type LogSource interface {
	// Run starts the log collection process and streams messages into logChan.
	// It should block until the context is canceled or a fatal error occurs.
	Run(ctx context.Context, logChan chan<- types.LogMessage) error
}
