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
	p := NewRedactProcessor()
	msg := &types.LogMessage{Message: "User email is admin@company.com"}
	got, _ := p.Process(msg)
	if got.Message != "User email is ***" {
		t.Errorf("expected redacted message, got %q", got.Message)
	}
}

func TestDedupeProcessor(t *testing.T) {
	p := NewDedupeProcessor()
	msg1 := &types.LogMessage{Message: "duplicate"}
	msg2 := &types.LogMessage{Message: "duplicate"}
	msg3 := &types.LogMessage{Message: "new"}

	if _, keep := p.Process(msg1); !keep {
		t.Error("expected first message to be kept")
	}
	if _, keep := p.Process(msg2); keep {
		t.Error("expected second duplicate message to be dropped")
	}
	if _, keep := p.Process(msg3); !keep {
		t.Error("expected third new message to be kept")
	}
}
