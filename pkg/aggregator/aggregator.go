// Package aggregator is responsible for interfacing with the Docker Engine API
// to discover running containers, establish multiplexed log streams, and
// aggregate output across multiple containers. It provides mechanisms to filter,
// format, and route log streams synchronously or asynchronously.
package aggregator

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/fatih/color"

	"docklog/pkg/config"
)

var redactPatterns = []*regexp.Regexp{
	regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),             // Email
	regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`),                                // IPv4
	regexp.MustCompile(`(?i)bearer\s+[A-Za-z0-9\-\._~\+/]+=*`),                       // Bearer Token
	regexp.MustCompile(`(?i)(api[_-]?key)\s*[:=]\s*['"]?[A-Za-z0-9\-\._~\+/]+['"]?`), // API Key
}

// LogMessage encapsulates a single line of log output retrieved from a container.
// It includes structural metadata required for JSON formatting and filtering.
type LogMessage struct {
	// ContainerName is the canonical name of the Docker container.
	ContainerName string `json:"container"`
	// Message is the raw log string output by the container (excluding Docker headers).
	Message string `json:"message"`
	// IsError is true if the message was read from the container's STDERR stream.
	IsError bool `json:"is_error"`
	// Timestamp is the exact time the message was aggregated locally.
	Timestamp time.Time `json:"timestamp"`
}

// Aggregator manages the lifecycle of concurrent log streams originating from
// multiple Docker containers. It orchestrates the Docker Client API calls,
// handles event-driven container discovery (start/die), and manages a centralized
// pipeline for log formatting and output.
type Aggregator struct {
	cli        *client.Client
	logChan    chan LogMessage
	cfg        *config.Config
	containers map[string]context.CancelFunc
	mu         sync.Mutex
	colorIndex int
	contColors map[string]*color.Color
}

// New initializes a new Aggregator instance using the provided configuration.
// It negotiates the Docker API version with the local Docker daemon automatically.
// Returns an error if the Docker Engine is unreachable.
func New(cfg *config.Config) (*Aggregator, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &Aggregator{
		cli:        cli,
		logChan:    make(chan LogMessage, cfg.BufferLength),
		cfg:        cfg,
		containers: make(map[string]context.CancelFunc),
		contColors: make(map[string]*color.Color),
	}, nil
}

func (a *Aggregator) getColor(containerName string) *color.Color {
	a.mu.Lock()
	defer a.mu.Unlock()
	if c, ok := a.contColors[containerName]; ok {
		return c
	}

	c := a.cfg.Colors[a.colorIndex%len(a.cfg.Colors)]
	a.colorIndex++
	a.contColors[containerName] = c
	return c
}

// Start begins the log aggregation process. It blocks until the provided context
// is canceled or an unrecoverable Docker event stream error occurs.
// It initializes existing containers and listens for Docker Engine events
// to dynamically attach to new containers or detach from dying ones.
func (a *Aggregator) Start(ctx context.Context) error {
	// Start the printer
	go a.printer(ctx)

	// List current running containers
	containers, err := a.cli.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return err
	}

	for _, c := range containers {
		a.startContainerStream(ctx, c.ID, a.getContainerName(c.Names))
	}

	// Listen for events
	msgs, errs := a.cli.Events(ctx, types.EventsOptions{})
	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-errs:
			if err != nil {
				return err
			}
		case msg := <-msgs:
			if msg.Type == events.ContainerEventType {
				if msg.Action == "start" {
					name := msg.Actor.Attributes["name"]
					a.startContainerStream(ctx, msg.ID, name)
				} else if msg.Action == "die" {
					a.stopContainerStream(msg.ID)
				}
			}
		}
	}
}

func (a *Aggregator) getContainerName(names []string) string {
	if len(names) > 0 {
		return strings.TrimPrefix(names[0], "/")
	}
	return "unknown"
}

// startContainerStream establishes a connection to a specific container's log stream.
// If a ContainerFilter is active in the configuration, it evaluates the container name
// and immediately returns if the container does not match, preventing resource allocation.
func (a *Aggregator) startContainerStream(ctx context.Context, id, name string) {
	if a.cfg.ContainerFilter != nil && !a.cfg.ContainerFilter.MatchString(name) {
		return
	}

	a.mu.Lock()
	if _, exists := a.containers[id]; exists {
		a.mu.Unlock()
		return
	}
	streamCtx, cancel := context.WithCancel(ctx)
	a.containers[id] = cancel
	a.mu.Unlock()

	go func() {
		defer a.stopContainerStream(id)

		options := container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     true,
			Tail:       a.cfg.TailLines,
		}

		if a.cfg.Since != "" {
			if dur, err := time.ParseDuration(a.cfg.Since); err == nil {
				// It's a duration like "5m", convert to Unix timestamp string
				options.Since = strconv.FormatInt(time.Now().Add(-dur).Unix(), 10)
			} else {
				// Otherwise, just pass it as is (might be a timestamp)
				options.Since = a.cfg.Since
			}
		}

		reader, err := a.cli.ContainerLogs(streamCtx, id, options)
		if err != nil {
			return
		}
		defer reader.Close()

		stdoutR, stdoutW := io.Pipe()
		stderrR, stderrW := io.Pipe()

		go a.readStream(stdoutR, name, false)
		go a.readStream(stderrR, name, true)

		_, err = stdcopy.StdCopy(stdoutW, stderrW, reader)

		stdoutW.Close()
		stderrW.Close()
	}()
}

func (a *Aggregator) stopContainerStream(id string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if cancel, exists := a.containers[id]; exists {
		cancel()
		delete(a.containers, id)
	}
}

// readStream scans the incoming raw byte stream from Docker (Stdout or Stderr)
// and applies all configured filtering mechanisms (Regex, Substring, Exclude).
// Messages that pass the filters are wrapped in a LogMessage and published
// to the aggregator's central log channel.
func (a *Aggregator) readStream(r io.Reader, name string, isError bool) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		text := scanner.Text()

		if a.cfg.Exclude != "" && strings.Contains(strings.ToLower(text), strings.ToLower(a.cfg.Exclude)) {
			continue
		}

		if a.cfg.RegexFilter != nil && !a.cfg.RegexFilter.MatchString(text) {
			continue
		}

		filterLower := strings.ToLower(a.cfg.Filter)
		if filterLower != "" && !strings.Contains(strings.ToLower(text), filterLower) {
			continue
		}

		if a.cfg.Redact {
			for _, pattern := range redactPatterns {
				text = pattern.ReplaceAllString(text, "***")
			}
		}

		a.logChan <- LogMessage{
			ContainerName: name,
			Message:       text,
			IsError:       isError,
			Timestamp:     time.Now(),
		}
	}
}

// printer is a long-running goroutine responsible for consuming messages
// from the central log channel. It performs deduplication, structural formatting
// (JSON or Human-readable), colorization, and multi-sink routing (Stdout vs File).
func (a *Aggregator) printer(ctx context.Context) {
	errColor := color.New(color.FgRed, color.Bold)
	infoColor := color.New(color.FgHiWhite)

	var outFile *os.File
	if a.cfg.Output != "" {
		var err error
		outFile, err = os.OpenFile(a.cfg.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Printf("Warning: Could not open output file %s: %v\n", a.cfg.Output, err)
		} else {
			defer outFile.Close()
		}
	}

	// Keep track of the last message seen per container to prevent spam
	lastMsg := make(map[string]string)

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-a.logChan:
			// Deduplication check
			if a.cfg.Deduplicate {
				if last, exists := lastMsg[msg.ContainerName]; exists && last == msg.Message {
					continue // Skip if identical to the previous message from the same container
				}
				lastMsg[msg.ContainerName] = msg.Message
			}

			cColor := a.getColor(msg.ContainerName)
			timeStr := msg.Timestamp.Format(a.cfg.TimeFormat)

			level := "INFO"
			msgColor := infoColor
			if msg.IsError {
				level = "ERROR"
				msgColor = errColor
			}

			if a.cfg.JsonOutput {
				b, _ := json.Marshal(msg)
				jsonLine := string(b) + "\n"
				fmt.Print(jsonLine)
				if outFile != nil {
					outFile.WriteString(jsonLine)
				}
				continue
			}

			// Format with fixed width for typical short messages, or allow overflow
			// Let's just output it Spire-style
			line := fmt.Sprintf("%s[%s] %s %s\n",
				msgColor.Sprint(level),
				timeStr,
				cColor.Sprintf("container=%q", msg.ContainerName),
				msg.Message,
			)
			fmt.Print(line)

			if outFile != nil {
				// Strip colors for file output
				cleanLine := fmt.Sprintf("%s[%s] container=%q %s\n",
					level,
					timeStr,
					msg.ContainerName,
					msg.Message,
				)
				outFile.WriteString(cleanLine)
			}
		}
	}
}
