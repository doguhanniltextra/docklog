package processor

import (
	"regexp"
	"testing"
	"github.com/doguhanniltextra/docklog/pkg/types"
)

func TestFilterProcessor(t *testing.T) {
	tests := []struct {
		name       string
		filter     string
		exclude    string
		regex      *regexp.Regexp
		input      string
		wantOutput bool
	}{
		{
			name:       "Include match",
			filter:     "error",
			input:      "this is an error",
			wantOutput: true,
		},
		{
			name:       "Include mismatch",
			filter:     "error",
			input:      "this is fine",
			wantOutput: false,
		},
		{
			name:       "Exclude match",
			exclude:    "spam",
			input:      "this is spam",
			wantOutput: false,
		},
		{
			name:       "Regex match",
			regex:      regexp.MustCompile(`^\[\d+\]`),
			input:      "[123] log",
			wantOutput: true,
		},
		{
			name:       "Filter and Exclude both set (Keep)",
			filter:     "success",
			exclude:    "fail",
			input:      "the operation was a success",
			wantOutput: true,
		},
		{
			name:       "Filter and Exclude both set (Drop due to exclude)",
			filter:     "success",
			exclude:    "fail",
			input:      "the operation was a success but with a fail",
			wantOutput: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewFilterProcessor(tt.filter, tt.exclude, tt.regex)
			msg := &types.LogMessage{Message: tt.input}
			_, keep := p.Process(msg)
			if keep != tt.wantOutput {
				t.Errorf("FilterProcessor.Process() keep = %v, want %v", keep, tt.wantOutput)
			}
		})
	}
}

func TestRedactProcessor(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Redact email",
			input: "contact admin@company.com",
			want:  "contact ***",
		},
		{
			name:  "Redact IPv4",
			input: "connection from 192.168.1.1",
			want:  "connection from ***",
		},
		{
			name:  "Redact Bearer Token",
			input: "Authorization: Bearer abc123def456",
			want:  "Authorization: ***",
		},
		{
			name:  "Redact API Key",
			input: "api_key: secret-value-123",
			want:  "***",
		},
	}

	p := NewRedactProcessor()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &types.LogMessage{Message: tt.input}
			got, _ := p.Process(msg)
			if got.Message != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got.Message)
			}
		})
	}
}

func TestDedupeProcessor(t *testing.T) {
	p := NewDedupeProcessor()
	
	// Test single container deduplication
	msg1 := &types.LogMessage{ContainerName: "c1", Message: "duplicate"}
	msg2 := &types.LogMessage{ContainerName: "c1", Message: "duplicate"}
	msg3 := &types.LogMessage{ContainerName: "c1", Message: "new"}

	if _, keep := p.Process(msg1); !keep {
		t.Error("expected first message to be kept")
	}
	if _, keep := p.Process(msg2); keep {
		t.Error("expected second duplicate message to be dropped")
	}
	if _, keep := p.Process(msg3); !keep {
		t.Error("expected third new message to be kept")
	}

	// Test cross-container isolation (same message, different container)
	msg4 := &types.LogMessage{ContainerName: "c2", Message: "new"}
	if _, keep := p.Process(msg4); !keep {
		t.Error("expected message from different container to be kept even if content is same as c1's last message")
	}
}


func TestLevelProcessor(t *testing.T) {
	tests := []struct {
		name          string
		allowedLevels []string
		input         string
		wantKeep      bool
	}{
		{
			name:          "Match error level",
			allowedLevels: []string{"ERROR"},
			input:         "ERROR: something went wrong",
			wantKeep:      true,
		},
		{
			name:          "Mismatch level",
			allowedLevels: []string{"ERROR"},
			input:         "INFO: business as usual",
			wantKeep:      false,
		},
		{
			name:          "Case insensitive match",
			allowedLevels: []string{"warn"},
			input:         "WARNING: check this",
			wantKeep:      true,
		},
		{
			name:          "Multiple levels match",
			allowedLevels: []string{"INFO", "DEBUG"},
			input:         "debug log message",
			wantKeep:      true,
		},
		{
			name:          "No levels specified (pass all)",
			allowedLevels: []string{},
			input:         "any message",
			wantKeep:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewLevelProcessor(tt.allowedLevels)
			msg := &types.LogMessage{Message: tt.input}
			_, keep := p.Process(msg)
			if keep != tt.wantKeep {
				t.Errorf("LevelProcessor.Process() keep = %v, want %v", keep, tt.wantKeep)
			}
		})
	}
}

func TestLimitProcessor(t *testing.T) {
	doneCalled := false
	onDone := func() {
		doneCalled = true
	}

	p := NewLimitProcessor(2, onDone)
	msg := &types.LogMessage{Message: "msg"}

	// First message
	if _, keep := p.Process(msg); !keep {
		t.Error("expected first message to be kept")
	}
	// Second message
	if _, keep := p.Process(msg); !keep {
		t.Error("expected second message to be kept")
	}
	// Third message (should be dropped)
	if _, keep := p.Process(msg); keep {
		t.Error("expected third message to be dropped")
	}

	if !doneCalled {
		t.Error("expected done callback to be called")
	}
}

