package aggregator

import (
	"context"
	"testing"
	"time"

	"github.com/doguhanniltextra/docklog/pkg/processor"
	"github.com/doguhanniltextra/docklog/pkg/types"
)

// MockSource sends a fixed number of messages and then finishes.
type MockSource struct {
	messages []types.LogMessage
}

func (m *MockSource) Run(ctx context.Context, logChan chan<- types.LogMessage) error {
	for _, msg := range m.messages {
		select {
		case <-ctx.Done():
			return nil
		case logChan <- msg:
		}
	}
	// Give the pipeline a moment to process before returning
	time.Sleep(10 * time.Millisecond)
	return nil
}

// MockProcessor modifies the message or drops it.
type MockProcessor struct {
	suffix string
	drop   bool
}

func (m *MockProcessor) Process(msg *types.LogMessage) (*types.LogMessage, bool) {
	if m.drop {
		return nil, false
	}
	msg.Message += m.suffix
	return msg, true
}

// MockSink collects written messages.
type MockSink struct {
	written []types.LogMessage
}

func (m *MockSink) Write(msg types.LogMessage) error {
	m.written = append(m.written, msg)
	return nil
}

func (m *MockSink) Close() error { return nil }

func TestPipeline(t *testing.T) {
	t.Run("Basic flow", func(t *testing.T) {
		src := &MockSource{
			messages: []types.LogMessage{
				{Message: "msg1"},
				{Message: "msg2"},
			},
		}
		proc := &MockProcessor{suffix: "-processed"}
		snk := &MockSink{}

		pipeline := NewPipeline(src, []processor.Processor{proc}, snk, 10)
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		err := pipeline.Run(ctx)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if len(snk.written) != 2 {
			t.Errorf("expected 2 messages written, got %d", len(snk.written))
		}
		if snk.written[0].Message != "msg1-processed" {
			t.Errorf("expected msg1-processed, got %s", snk.written[0].Message)
		}
	})

	t.Run("Processor drop", func(t *testing.T) {
		src := &MockSource{
			messages: []types.LogMessage{
				{Message: "keep"},
				{Message: "drop"},
			},
		}
		// A processor that drops messages with "drop"
		dropProc := &mockDropProcessor{}
		snk := &MockSink{}

		pipeline := NewPipeline(src, []processor.Processor{dropProc}, snk, 10)
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		pipeline.Run(ctx)

		if len(snk.written) != 1 {
			t.Errorf("expected 1 message written, got %d", len(snk.written))
		}
		if snk.written[0].Message != "keep" {
			t.Errorf("expected keep, got %s", snk.written[0].Message)
		}
	})
}

type mockDropProcessor struct{}

func (m *mockDropProcessor) Process(msg *types.LogMessage) (*types.LogMessage, bool) {
	if msg.Message == "drop" {
		return nil, false
	}
	return msg, true
}
