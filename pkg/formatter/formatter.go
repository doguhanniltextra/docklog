package formatter

import (
	"github.com/doguhanniltextra/docklog/pkg/types"
)

// Formatter defines the interface for converting a LogMessage into a printable string.
type Formatter interface {
	Format(msg types.LogMessage) string
}
