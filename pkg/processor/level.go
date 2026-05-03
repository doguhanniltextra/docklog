package processor

import (
	"strings"

	"github.com/doguhanniltextra/docklog/pkg/types"
)

// LevelProcessor filters log messages based on their severity level (e.g., INFO, ERROR).
// It performs a case-insensitive check for the existence of the level string
// within the log message.
type LevelProcessor struct {
	// allowedLevels is a slice of uppercase level names that are allowed to pass.
	allowedLevels []string
}

// NewLevelProcessor creates a new LevelProcessor with the specified allowed levels.
// All provided levels are automatically normalized to uppercase for consistent matching.
func NewLevelProcessor(allowedLevels []string) *LevelProcessor {
	var normalized []string
	for _, l := range allowedLevels {
		normalized = append(normalized, strings.ToUpper(l))
	}
	return &LevelProcessor{
		allowedLevels: normalized,
	}
}

// Process checks if the log message contains any of the allowed severity levels.
// If allowedLevels is empty, all messages are passed through.
// The matching is case-insensitive.
func (p *LevelProcessor) Process(msg *types.LogMessage) (*types.LogMessage, bool) {
	// If no levels specified, pass everything
	if len(p.allowedLevels) == 0 {
		return msg, true
	}

	upperMsg := strings.ToUpper(msg.Message)
	for _, level := range p.allowedLevels {
		if strings.Contains(upperMsg, level) {
			return msg, true
		}
	}

	return nil, false
}
