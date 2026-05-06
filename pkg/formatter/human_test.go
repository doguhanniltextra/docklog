package formatter

import (
	"strings"
	"testing"
	"time"

	"github.com/doguhanniltextra/docklog/pkg/types"
	"github.com/fatih/color"
)

func TestHumanFormatter(t *testing.T) {
	colors := []*color.Color{
		color.New(color.FgCyan),
		color.New(color.FgGreen),
	}
	f := NewHumanFormatter("15:04:05", colors)

	msg := types.LogMessage{
		ContainerName: "test-container",
		Message:       "Hello World",
		Timestamp:     time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		IsError:       false,
	}

	output := f.Format(msg)

	// Check if output contains key parts
	if !strings.Contains(output, "INFO") {
		t.Errorf("expected output to contain INFO, got %q", output)
	}
	if !strings.Contains(output, "12:00:00") {
		t.Errorf("expected output to contain timestamp, got %q", output)
	}
	if !strings.Contains(output, "test-container") {
		t.Errorf("expected output to contain container name, got %q", output)
	}
	if !strings.Contains(output, "Hello World") {
		t.Errorf("expected output to contain message, got %q", output)
	}

	// Test error formatting
	msgErr := types.LogMessage{
		ContainerName: "test-container",
		Message:       "Fatal Error",
		Timestamp:     time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		IsError:       true,
	}
	outputErr := f.Format(msgErr)
	if !strings.Contains(outputErr, "ERROR") {
		t.Errorf("expected output to contain ERROR, got %q", outputErr)
	}

	// Test color consistency
	color1 := f.getColor("container-1")
	color2 := f.getColor("container-1")
	if color1 != color2 {
		t.Error("expected same color for same container")
	}

	color3 := f.getColor("container-2")
	if color1 == color3 {
		t.Error("expected different colors for different containers (with enough palette)")
	}
}
