package sink

import (
	"os"
	"testing"

	"github.com/doguhanniltextra/docklog/pkg/types"
)

// MockFormatter just returns the message as is.
type MockFormatter struct{}

func (m *MockFormatter) Format(msg types.LogMessage) string {
	return msg.Message
}

func TestFileSink(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "docklog-test-*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	s, err := NewFileSink(tmpFile.Name(), &MockFormatter{})
	if err != nil {
		t.Fatalf("failed to create FileSink: %v", err)
	}

	msg := types.LogMessage{Message: "test log entry\n"}
	if err := s.Write(msg); err != nil {
		t.Errorf("failed to write to FileSink: %v", err)
	}
	s.Close()

	// Read back and verify
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "test log entry\n" {
		t.Errorf("expected %q, got %q", "test log entry\n", string(content))
	}
}

func TestMultiSink(t *testing.T) {
	sink1 := &mockSink{}
	sink2 := &mockSink{}
	m := NewMultiSink(sink1, sink2)

	msg := types.LogMessage{Message: "multi-test"}
	if err := m.Write(msg); err != nil {
		t.Errorf("failed to write to MultiSink: %v", err)
	}

	if len(sink1.msgs) != 1 || sink1.msgs[0].Message != "multi-test" {
		t.Error("sink1 did not receive correct message")
	}
	if len(sink2.msgs) != 1 || sink2.msgs[0].Message != "multi-test" {
		t.Error("sink2 did not receive correct message")
	}

	m.Close()
	if !sink1.closed || !sink2.closed {
		t.Error("child sinks were not closed")
	}
}

type mockSink struct {
	msgs   []types.LogMessage
	closed bool
}

func (m *mockSink) Write(msg types.LogMessage) error {
	m.msgs = append(m.msgs, msg)
	return nil
}

func (m *mockSink) Close() error {
	m.closed = true
	return nil
}
