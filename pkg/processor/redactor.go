package processor

import (
	"regexp"

	"github.com/doguhanniltextra/docklog/pkg/types"
)

var defaultRedactPatterns = []*regexp.Regexp{
	regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),             // Email
	regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`),                                // IPv4
	regexp.MustCompile(`(?i)bearer\s+[A-Za-z0-9\-\._~\+/]+=*`),                       // Bearer Token
	regexp.MustCompile(`(?i)(api[_-]?key)\s*[:=]\s*['"]?[A-Za-z0-9\-\._~\+/]+['"]?`), // API Key
}

type RedactProcessor struct {
	patterns []*regexp.Regexp
}

func NewRedactProcessor() *RedactProcessor {
	return &RedactProcessor{
		patterns: defaultRedactPatterns,
	}
}

func (r *RedactProcessor) Process(msg *types.LogMessage) (*types.LogMessage, bool) {
	text := msg.Message
	for _, pattern := range r.patterns {
		text = pattern.ReplaceAllString(text, "***")
	}
	msg.Message = text
	return msg, true
}
