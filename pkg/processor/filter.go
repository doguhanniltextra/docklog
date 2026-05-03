package processor

import (
	"regexp"
	"strings"

	"github.com/doguhanniltextra/docklog/pkg/types"
)

type FilterProcessor struct {
	filter      string
	exclude     string
	regexFilter *regexp.Regexp
}

func NewFilterProcessor(filter, exclude string, regexFilter *regexp.Regexp) *FilterProcessor {
	return &FilterProcessor{
		filter:      strings.ToLower(filter),
		exclude:     strings.ToLower(exclude),
		regexFilter: regexFilter,
	}
}

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
