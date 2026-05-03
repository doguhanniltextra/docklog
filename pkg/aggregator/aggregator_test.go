package aggregator

import (
	"bytes"
	"regexp"
	"testing"
	"time"

	"docklog/pkg/config"
)

func TestReadStream(t *testing.T) {
	tests := []struct {
		name       string
		cfg        *config.Config
		input      string
		isError    bool
		wantOutput []string
	}{
		{
			name:       "No filters, multiple lines",
			cfg:        &config.Config{},
			input:      "line 1\nline 2\nline 3\n",
			isError:    false,
			wantOutput: []string{"line 1", "line 2", "line 3"},
		},
		{
			name: "Filter by keyword (case insensitive)",
			cfg: &config.Config{
				Filter: "ERROR",
			},
			input:      "this is fine\nthis has an eRRoR\nerror here too\n",
			isError:    false,
			wantOutput: []string{"this has an eRRoR", "error here too"},
		},
		{
			name: "Exclude by keyword",
			cfg: &config.Config{
				Exclude: "ignore",
			},
			input:      "process 1 ok\nprocess 2 IGNORE this\nprocess 3 ok\n",
			isError:    false,
			wantOutput: []string{"process 1 ok", "process 3 ok"},
		},
		{
			name: "Filter by regex",
			cfg: &config.Config{
				RegexFilter: regexp.MustCompile(`^\[\d+\] .*`),
			},
			input:      "[123] process started\n[abc] invalid log\n[456] process ended\n",
			isError:    false,
			wantOutput: []string{"[123] process started", "[456] process ended"},
		},
		{
			name: "Combine filter and exclude",
			cfg: &config.Config{
				Filter:  "INFO",
				Exclude: "spam",
			},
			input:      "INFO: starting\nINFO: spam message\nWARN: something else\nINFO: ended\n",
			isError:    false,
			wantOutput: []string{"INFO: starting", "INFO: ended"},
		},
		{
			name: "Redact sensitive information",
			cfg: &config.Config{
				Redact: true,
			},
			input:      "User email is admin@company.com\nUser IP is 192.168.1.1\nAuth: Bearer abcdef12345=\nAPI Key is api_key: 'sk_test_123'\nNormal log message\n",
			isError:    false,
			wantOutput: []string{"User email is ***", "User IP is ***", "Auth: ***", "API Key is ***", "Normal log message"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock the aggregator with a buffered channel
			a := &Aggregator{
				cfg:     tt.cfg,
				logChan: make(chan LogMessage, 10),
			}

			reader := bytes.NewBufferString(tt.input)

			// Execute the method synchronously
			a.readStream(reader, "test-container", tt.isError)

			close(a.logChan)

			// Collect output
			var gotOutput []string
			for msg := range a.logChan {
				gotOutput = append(gotOutput, msg.Message)

				if msg.ContainerName != "test-container" {
					t.Errorf("expected container name 'test-container', got %q", msg.ContainerName)
				}
				if msg.IsError != tt.isError {
					t.Errorf("expected isError %v, got %v", tt.isError, msg.IsError)
				}
				if time.Since(msg.Timestamp) > time.Second {
					t.Errorf("expected recent timestamp, got %v", msg.Timestamp)
				}
			}

			// Verify results
			if len(gotOutput) != len(tt.wantOutput) {
				t.Fatalf("expected %d messages, got %d (%v)", len(tt.wantOutput), len(gotOutput), gotOutput)
			}

			for i, want := range tt.wantOutput {
				if gotOutput[i] != want {
					t.Errorf("at index %d, expected %q, got %q", i, want, gotOutput[i])
				}
			}
		})
	}
}

func TestGetContainerName(t *testing.T) {
	a := &Aggregator{}

	tests := []struct {
		names []string
		want  string
	}{
		{names: []string{"/my-container"}, want: "my-container"},
		{names: []string{"/my-container", "/another-name"}, want: "my-container"},
		{names: []string{}, want: "unknown"},
		{names: nil, want: "unknown"},
	}

	for _, tt := range tests {
		got := a.getContainerName(tt.names)
		if got != tt.want {
			t.Errorf("getContainerName(%v) = %v; want %v", tt.names, got, tt.want)
		}
	}
}
