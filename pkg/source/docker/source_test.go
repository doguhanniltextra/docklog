package docker

import (
	"bytes"
	"testing"
	"time"

	logtypes "github.com/doguhanniltextra/docklog/pkg/types"
)

func TestReadStream(t *testing.T) {
	s := &DockerSource{}
	logChan := make(chan logtypes.LogMessage, 10)

	input := "line 1\nline 2\n"
	reader := bytes.NewBufferString(input)

	go s.readStream(reader, "test-container", false, logChan)

	select {
	case msg := <-logChan:
		if msg.Message != "line 1" {
			t.Errorf("expected 'line 1', got %q", msg.Message)
		}
		if msg.ContainerName != "test-container" {
			t.Errorf("expected 'test-container', got %q", msg.ContainerName)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for log message")
	}

	select {
	case msg := <-logChan:
		if msg.Message != "line 2" {
			t.Errorf("expected 'line 2', got %q", msg.Message)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for second log message")
	}
}
