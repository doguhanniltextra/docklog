package processor

import (
	"sync/atomic"

	"github.com/doguhanniltextra/docklog/pkg/types"
)

type LimitProcessor struct {
	max   int64
	count int64
	done  func()
}

func NewLimitProcessor(max int, onLimitReached func()) *LimitProcessor {
	return &LimitProcessor{
		max:  int64(max),
		done: onLimitReached,
	}
}

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
