package processor

import (
	"regexp"

	"github.com/doguhanniltextra/docklog/pkg/types"
)

// defaultRedactPatterns defines common sensitive information patterns to be redacted.
var defaultRedactPatterns = []*regexp.Regexp{
	regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),             // Email
	regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`),                                // IPv4
	regexp.MustCompile(`(?i)bearer\s+[A-Za-z0-9\-\._~\+/]+=*`),                       // Bearer Token
	regexp.MustCompile(`(?i)(api[_-]?key)\s*[:=]\s*['"]?[A-Za-z0-9\-\._~\+/]+['"]?`), // API Key
}

// RedactProcessor automatically masks sensitive information in log messages.
// It uses a set of pre-defined regular expressions to find and replace data like
// emails, IP addresses, and authentication tokens with asterisks.
type RedactProcessor struct {
	// patterns is the list of regular expressions used for redaction.
	patterns []*regexp.Regexp
}

// NewRedactProcessor creates a new RedactProcessor with default redaction rules.
func NewRedactProcessor() *RedactProcessor {
	return &RedactProcessor{
		patterns: defaultRedactPatterns,
	}
}

// Process scans the log message text and replaces any sensitive matches with "***".
// It always returns true as it only modifies the message content and never drops it.
func (r *RedactProcessor) Process(msg *types.LogMessage) (*types.LogMessage, bool) {
	text := msg.Message
	for _, pattern := range r.patterns {
		text = pattern.ReplaceAllString(text, "***")
	}
	msg.Message = text
	return msg, true
}
