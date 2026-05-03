// Package formatter provides various strategies for transforming structured
// LogMessage objects into strings suitable for different output targets.
package formatter

import (
	"github.com/doguhanniltextra/docklog/pkg/types"
)

// Formatter defines the contract for any component that needs to convert
// a LogMessage into a printable format. Implementations can focus on
// human readability, machine parsing, or specific external log protocols.
type Formatter interface {
	// Format takes a LogMessage and returns its string representation.
	// It must be thread-safe if shared across multiple goroutines.
	Format(msg types.LogMessage) string
}
