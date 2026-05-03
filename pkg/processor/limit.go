package processor

import (
	"sync/atomic"

	"github.com/doguhanniltextra/docklog/pkg/types"
)

// LimitProcessor stops processing log messages after a certain threshold is reached.
// It is useful for implementing commands like 'run' which should exit after N logs.
type LimitProcessor struct {
	// max is the maximum number of logs allowed to pass.
	max int64
	// count is the thread-safe counter of processed logs.
	count int64
	// done is an optional callback executed when the limit is reached.
	done func()
}

// NewLimitProcessor creates a new LimitProcessor with the specified maximum count.
// If max is <= 0, the limit is disabled.
func NewLimitProcessor(max int, onLimitReached func()) *LimitProcessor {
	return &LimitProcessor{
		max:  int64(max),
		done: onLimitReached,
	}
}

// Process increments the internal counter and checks against the maximum.
// If the limit is exceeded, it executes the 'done' callback and drops the message.
func (p *LimitProcessor) Process(msg *types.LogMessage) (*types.LogMessage, bool) {
	if p.max <= 0 {
		return msg, true
	}

	current := atomic.AddInt64(&p.count, 1)
	if current > p.max {
		if p.done != nil {
			p.done()
		}
		return nil, false
	}

	return msg, true
}
