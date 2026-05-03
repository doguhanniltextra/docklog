package processor

import (
	"sync"

	"github.com/doguhanniltextra/docklog/pkg/types"
)

// DedupeProcessor prevents consecutive identical log messages from the same container
// from cluttering the output. This is particularly useful for containers that
// emit repetitive "heartbeat" or "waiting" logs.
type DedupeProcessor struct {
	// lastMsg maps container names to their last emitted log content.
	lastMsg map[string]string
	// mu synchronizes access to the lastMsg map.
	mu sync.Mutex
}

// NewDedupeProcessor creates a new DedupeProcessor with an empty tracking map.
func NewDedupeProcessor() *DedupeProcessor {
	return &DedupeProcessor{
		lastMsg: make(map[string]string),
	}
}

// Process checks if the current message is identical to the last one seen
// for this specific container. If it is, it returns false to drop the message.
func (d *DedupeProcessor) Process(msg *types.LogMessage) (*types.LogMessage, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if last, exists := d.lastMsg[msg.ContainerName]; exists && last == msg.Message {
		return nil, false
	}
	d.lastMsg[msg.ContainerName] = msg.Message
	return msg, true
}
