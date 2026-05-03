// Package processor defines the pipeline components responsible for modifying,
// filtering, or enriching log messages as they flow through the system.
// It follows a modular design where multiple processors can be chained together.
package processor

import (
	"regexp"
	"strings"

	"github.com/doguhanniltextra/docklog/pkg/types"
)

// FilterProcessor implements the logic for dropping or keeping log messages
// based on user-defined criteria. It supports literal substring matching,
// exclusion patterns, and complex regular expressions.
type FilterProcessor struct {
	// filter is the lowercase substring that must be present in the message.
	filter string
	// exclude is the lowercase substring that must NOT be present in the message.
	exclude string
	// regexFilter is an optional compiled regular expression for fine-grained filtering.
	regexFilter *regexp.Regexp
}

// NewFilterProcessor creates a new instance of FilterProcessor.
// It pre-processes the filter and exclude strings by converting them to lowercase
// to ensure case-insensitive matching during the processing phase.
func NewFilterProcessor(filter, exclude string, regexFilter *regexp.Regexp) *FilterProcessor {
	return &FilterProcessor{
		filter:      strings.ToLower(filter),
		exclude:     strings.ToLower(exclude),
		regexFilter: regexFilter,
	}
}

// Process evaluates a log message against the configured filters.
// It returns the message and 'true' if it passes all filters, or 'nil' and 'false'
// if it should be dropped.
//
// The evaluation order is:
// 1. Exclusion (Literal match)
// 2. Regular Expression match
// 3. Inclusion (Literal match)
func (f *FilterProcessor) Process(msg *types.LogMessage) (*types.LogMessage, bool) {
	textLower := strings.ToLower(msg.Message)

	if f.exclude != "" && strings.Contains(textLower, f.exclude) {
		return nil, false
	}

	if f.regexFilter != nil && !f.regexFilter.MatchString(msg.Message) {
		return nil, false
	}

	if f.filter != "" && !strings.Contains(textLower, f.filter) {
		return nil, false
	}

	return msg, true
}
