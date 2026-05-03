package processor

import (
	"sync"

	"github.com/doguhanniltextra/docklog/pkg/types"
)

type DedupeProcessor struct {
	lastMsg map[string]string
	mu      sync.Mutex
}

func NewDedupeProcessor() *DedupeProcessor {
	return &DedupeProcessor{
		lastMsg: make(map[string]string),
	}
}

func (d *DedupeProcessor) Process(msg *types.LogMessage) (*types.LogMessage, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if last, exists := d.lastMsg[msg.ContainerName]; exists && last == msg.Message {
		return nil, false
	}
	d.lastMsg[msg.ContainerName] = msg.Message
	return msg, true
}
