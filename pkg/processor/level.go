package processor

import (
	"strings"

	"github.com/doguhanniltextra/docklog/pkg/types"
)

type LevelProcessor struct {
	allowedLevels []string
}

func NewLevelProcessor(allowedLevels []string) *LevelProcessor {
	// Normalize levels to uppercase
	var normalized []string
	for _, l := range allowedLevels {
		normalized = append(normalized, strings.ToUpper(l))
	}
	return &LevelProcessor{
		allowedLevels: normalized,
	}
}

func (p *LevelProcessor) Process(msg *types.LogMessage) (*types.LogMessage, bool) {
	if len(p.allowedLevels) == 0 {
		return msg, true
	}

	upperMsg := strings.ToUpper(msg.Message)
	found := false
	for _, level := range p.allowedLevels {
		if strings.Contains(upperMsg, level) {
			found = true
			break
		}
	}

	return msg, found
}
