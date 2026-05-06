package formatter

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/doguhanniltextra/docklog/pkg/types"
)

func TestJsonFormatter(t *testing.T) {
	f := NewJsonFormatter()

	msg := types.LogMessage{
		ContainerName: "test-container",
		Message:       "Hello JSON",
		Timestamp:     time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		IsError:       true,
	}

	output := f.Format(msg)

	if !strings.HasSuffix(output, "\n") {
		t.Error("expected output to end with a newline")
	}

	var decoded types.LogMessage
	err := json.Unmarshal([]byte(output), &decoded)
	if err != nil {
		t.Fatalf("failed to unmarshal JSON output: %v", err)
	}

	if decoded.ContainerName != msg.ContainerName {
		t.Errorf("expected container name %q, got %q", msg.ContainerName, decoded.ContainerName)
	}
	if decoded.Message != msg.Message {
		t.Errorf("expected message %q, got %q", msg.Message, decoded.Message)
	}
	if decoded.IsError != msg.IsError {
		t.Errorf("expected isError %v, got %v", msg.IsError, decoded.IsError)
	}
}
