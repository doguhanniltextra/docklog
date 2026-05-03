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

	"github.com/doguhanniltextra/docklog/pkg/config"
	"github.com/doguhanniltextra/docklog/pkg/types"
)

var redactPatterns = []*regexp.Regexp{
	regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),             // Email
	regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`),                                // IPv4
	regexp.MustCompile(`(?i)bearer\s+[A-Za-z0-9\-\._~\+/]+=*`),                       // Bearer Token
	regexp.MustCompile(`(?i)(api[_-]?key)\s*[:=]\s*['"]?[A-Za-z0-9\-\._~\+/]+['"]?`), // API Key
}

// ContainerStatus represents the monitoring state of a specific Docker container.
// It is used for UI feedback (e.g., in the 'list' command) to show which containers
// are currently being watched versus those being ignored by filters.
type ContainerStatus struct {
	// ID is the shortened (12-char) Docker container ID.
	ID string
	// Name is the human-readable container name.
	Name string
	// Image is the Docker image being run.
	Image string
	// Status is the raw status string from Docker (e.g., "Up 5 minutes").
	Status string
	// Uptime is a processed version of the status indicating how long it's been running.
	Uptime string
	// IsMatched indicates if this container passes the global --container regex filter.
	IsMatched bool
}

// Aggregator manages the lifecycle of concurrent log streams originating from
// multiple Docker containers. It orchestrates the Docker Client API calls,
// handles event-driven container discovery (start/die), and manages a centralized
// pipeline for log formatting and output.
type Aggregator struct {
	// cli is the official Docker Engine API client.
	cli *client.Client
	// logChan is the central multiplexed channel where all log lines from all
	// containers are gathered before being sent to the printer.
	logChan chan types.LogMessage
	// cfg holds the operational parameters (filters, formats, etc.).
	cfg *config.Config
	// containers tracks active log streams and their cancellation functions,
	// keyed by the 64-character Docker container ID.
	containers map[string]context.CancelFunc
	// mu protects concurrent access to the 'containers' and 'contColors' maps.
	mu sync.Mutex
	// colorIndex is used to rotate through the color palette for new containers.
	colorIndex int
	// contColors maps container names to specific terminal colors for consistency.
	contColors map[string]*color.Color
}

// New initializes a new Aggregator instance using the provided configuration.
// It negotiates the Docker API version with the local Docker daemon automatically.
// Returns an error if the Docker Engine is unreachable or the client fails to initialize.
func New(cfg *config.Config) (*Aggregator, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	return &Aggregator{
		cli:        cli,
		logChan:    make(chan types.LogMessage, cfg.BufferLength),
		cfg:        cfg,
		containers: make(map[string]context.CancelFunc),
		contColors: make(map[string]*color.Color),
	}, nil
}

// getColor retrieves or assigns a unique color for a container name.
// This ensures that all log lines from a specific container share the same color
// in the terminal, even if the stream is interrupted and restarted.
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
//
// Execution Flow:
// 1. Spawns the printer goroutine to process the log channel.
// 2. Lists all currently running containers and attaches log streams to those
//    that match the configuration filters.
// 3. Subscribes to Docker Engine events to automatically handle 'start' and 'die'
//    events, ensuring new containers are monitored immediately.
func (a *Aggregator) Start(ctx context.Context) error {
	// Start the printer
	go a.printer(ctx)

	// List current running containers
	containers, err := a.cli.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
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
				return fmt.Errorf("docker event stream error: %w", err)
			}
		case msg := <-msgs:
			if msg.Type == events.ContainerEventType {
				if msg.Action == "start" {
					// Extract name from Actor attributes for efficiency
					name := msg.Actor.Attributes["name"]
					a.startContainerStream(ctx, msg.ID, name)
				} else if msg.Action == "die" {
					// Clean up resources immediately when container stops
					a.stopContainerStream(msg.ID)
				}
			}
		}
	}
}

// ListContainers performs a one-shot discovery of all running containers and
// evaluates them against the aggregator's configuration. It returns a list of
// ContainerStatus objects describing which containers would be monitored.
// This is primarily used by the 'list' command for configuration previewing.
func (a *Aggregator) ListContainers(ctx context.Context) ([]ContainerStatus, error) {
	containers, err := a.cli.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers for discovery: %w", err)
	}

	var results []ContainerStatus
	for _, c := range containers {
		name := a.getContainerName(c.Names)
		isMatched := true
		if a.cfg.ContainerFilter != nil && !a.cfg.ContainerFilter.MatchString(name) {
			isMatched = false
		}

		results = append(results, ContainerStatus{
			ID:        c.ID[:12],
			Name:      name,
			Image:     c.Image,
			Status:    c.Status,
			Uptime:    c.Status, // Docker's Status field often contains uptime info
			IsMatched: isMatched,
		})
	}
	return results, nil
}

// getContainerName extracts the primary name of a container, stripping the
// leading slash provided by the Docker API.
func (a *Aggregator) getContainerName(names []string) string {
	if len(names) > 0 {
		return strings.TrimPrefix(names[0], "/")
	}
	return "unknown"
}

// startContainerStream establishes a connection to a specific container's log stream.
// If a ContainerFilter is active in the configuration, it evaluates the container name
// and immediately returns if the container does not match, preventing resource allocation.
//
// For matched containers, it spawns a worker goroutine that multiplexes Stdout and Stderr
// into the central log channel.
func (a *Aggregator) startContainerStream(ctx context.Context, id, name string) {
	if a.cfg.ContainerFilter != nil && !a.cfg.ContainerFilter.MatchString(name) {
		return
	}

	a.mu.Lock()
	if _, exists := a.containers[id]; exists {
		a.mu.Unlock()
		return
	}
	// Create a sub-context so we can stop this specific stream independently
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
			Timestamps: a.cfg.ShowTimestamps,
		}

		// Handle 'Since' duration/timestamp logic
		if a.cfg.Since != "" {
			if dur, err := time.ParseDuration(a.cfg.Since); err == nil {
				options.Since = strconv.FormatInt(time.Now().Add(-dur).Unix(), 10)
			} else {
				options.Since = a.cfg.Since
			}
		}

		reader, err := a.cli.ContainerLogs(streamCtx, id, options)
		if err != nil {
			return
		}
		defer reader.Close()

		// Docker multiplexes Stdout and Stderr in a custom binary protocol.
		// We use Pipes to feed these streams into our standard scanner-based reader.
		stdoutR, stdoutW := io.Pipe()
		stderrR, stderrW := io.Pipe()

		go a.readStream(stdoutR, name, false)
		go a.readStream(stderrR, name, true)

		// StdCopy decodes the Docker multiplexed stream into the separate pipes.
		_, _ = stdcopy.StdCopy(stdoutW, stderrW, reader)

		stdoutW.Close()
		stderrW.Close()
	}()
}

// stopContainerStream gracefully terminates a container's log monitoring goroutine
// by invoking its context cancellation function.
func (a *Aggregator) stopContainerStream(id string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if cancel, exists := a.containers[id]; exists {
		cancel()
		delete(a.containers, id)
	}
}

// readStream scans an incoming raw byte stream and applies the processing pipeline.
// It handles filtering (substring, regex, exclusion), redaction of sensitive data,
// and publishes valid messages to the central log channel.
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
//
// Lifecycle: It runs until the provided context is canceled.
func (a *Aggregator) printer(ctx context.Context) {
	// Pre-initialize colors for common log levels
	errColor := color.New(color.FgRed, color.Bold)
	infoColor := color.New(color.FgHiWhite)

	// Handle persistent file output if configured
	var outFile *os.File
	if a.cfg.Output != "" {
		var err error
		// Open in Append mode to prevent overwriting existing logs
		outFile, err = os.OpenFile(a.cfg.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Printf("Warning: Could not open output file %s: %v\n", a.cfg.Output, err)
		} else {
			defer outFile.Close()
		}
	}

	// lastMsg tracks the previous line per container to facilitate deduplication.
	lastMsg := make(map[string]string)

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-a.logChan:
			// Apply deduplication logic early in the print cycle
			if a.cfg.Deduplicate {
				if last, exists := lastMsg[msg.ContainerName]; exists && last == msg.Message {
					continue
				}
				lastMsg[msg.ContainerName] = msg.Message
			}

			// Resolve presentation details
			cColor := a.getColor(msg.ContainerName)
			timeStr := msg.Timestamp.Format(a.cfg.TimeFormat)

			level := "INFO"
			msgColor := infoColor
			if msg.IsError {
				level = "ERROR"
				msgColor = errColor
			}

			// Scenario 1: Structured JSON Output
			if a.cfg.JsonOutput {
				b, _ := json.Marshal(msg)
				jsonLine := string(b) + "\n"
				fmt.Print(jsonLine)
				if outFile != nil {
					outFile.WriteString(jsonLine)
				}
				continue
			}

			// Scenario 2: Human-Readable Terminal Output
			// Format: [LEVEL] [TIMESTAMP] container="name" message...
			line := fmt.Sprintf("%s[%s] %s %s\n",
				msgColor.Sprint(level),
				timeStr,
				cColor.Sprintf("container=%q", msg.ContainerName),
				msg.Message,
			)
			fmt.Print(line)

			// Mirror to file if configured, stripping ANSI color codes
			if outFile != nil {
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
