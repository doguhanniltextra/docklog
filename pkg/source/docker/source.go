// Package docker implements the LogSource interface for the Docker Engine.
// It handles container discovery, event monitoring, and log stream multiplexing
// specifically for Docker environments.
package docker

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/doguhanniltextra/docklog/pkg/config"
	logtypes "github.com/doguhanniltextra/docklog/pkg/types"
)

// DockerSource manages the lifecycle of log streams from Docker containers.
// It uses the Docker SDK to interact with the Docker daemon.
type DockerSource struct {
	// cli is the negotiated Docker API client.
	cli *client.Client
	// cfg contains the filtering and collection parameters.
	cfg *config.Config
	// containers tracks active cancellation functions for individual log streams.
	containers map[string]context.CancelFunc
	// mu synchronizes access to the internal containers map.
	mu sync.Mutex
}

// NewDockerSource creates a new DockerSource and initializes the Docker client.
// It automatically detects the API version of the local Docker daemon.
func NewDockerSource(cfg *config.Config) (*DockerSource, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to init docker client for source: %w", err)
	}

	return &DockerSource{
		cli:        cli,
		cfg:        cfg,
		containers: make(map[string]context.CancelFunc),
	}, nil
}

// Run starts the Docker log collection process.
// It first attaches to all currently running containers, then enters a loop
// listening for Docker events (start/die) to dynamically adjust the monitored set.
func (s *DockerSource) Run(ctx context.Context, logChan chan<- logtypes.LogMessage) error {
	// Initial discovery of running containers
	containers, err := s.cli.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return fmt.Errorf("initial container list failed: %w", err)
	}

	for _, c := range containers {
		s.startContainerStream(ctx, c.ID, s.getContainerName(c.Names), logChan)
	}

	// Dynamic discovery via event stream
	msgs, errs := s.cli.Events(ctx, types.EventsOptions{})
	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-errs:
			if err != nil {
				return fmt.Errorf("docker source event error: %w", err)
			}
		case msg := <-msgs:
			if msg.Type == events.ContainerEventType {
				if msg.Action == "start" {
					name := msg.Actor.Attributes["name"]
					s.startContainerStream(ctx, msg.ID, name, logChan)
				} else if msg.Action == "die" {
					s.stopContainerStream(msg.ID)
				}
			}
		}
	}
}

// getContainerName strips the leading slash from Docker container names.
func (s *DockerSource) getContainerName(names []string) string {
	if len(names) > 0 {
		return strings.TrimPrefix(names[0], "/")
	}
	return "unknown"
}

// startContainerStream attaches to a container's logs and pumps them into logChan.
// It respects the --container regex filter provided in the configuration.
func (s *DockerSource) startContainerStream(ctx context.Context, id, name string, logChan chan<- logtypes.LogMessage) {
	// Early exit if container name doesn't match the user-defined filter
	if s.cfg.ContainerFilter != nil && !s.cfg.ContainerFilter.MatchString(name) {
		return
	}

	s.mu.Lock()
	if _, exists := s.containers[id]; exists {
		s.mu.Unlock()
		return
	}
	// Context used to stop this specific stream when the container dies or app exits
	streamCtx, cancel := context.WithCancel(ctx)
	s.containers[id] = cancel
	s.mu.Unlock()

	go func() {
		defer s.stopContainerStream(id)

		options := container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     true,
			Tail:       s.cfg.TailLines,
			Timestamps: s.cfg.ShowTimestamps,
		}

		// Handle time-based log retrieval
		if s.cfg.Since != "" {
			if dur, err := time.ParseDuration(s.cfg.Since); err == nil {
				options.Since = strconv.FormatInt(time.Now().Add(-dur).Unix(), 10)
			} else {
				options.Since = s.cfg.Since
			}
		}

		reader, err := s.cli.ContainerLogs(streamCtx, id, options)
		if err != nil {
			return
		}
		defer reader.Close()

		// Docker's log protocol multiplexes Stdout and Stderr into a single stream.
		// We use stdcopy to demultiplex them into separate pipes for processing.
		stdoutR, stdoutW := io.Pipe()
		stderrR, stderrW := io.Pipe()

		go s.readStream(stdoutR, name, false, logChan)
		go s.readStream(stderrR, name, true, logChan)

		_, _ = stdcopy.StdCopy(stdoutW, stderrW, reader)

		stdoutW.Close()
		stderrW.Close()
	}()
}

// stopContainerStream cancels the stream context and removes it from the tracking map.
func (s *DockerSource) stopContainerStream(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if cancel, exists := s.containers[id]; exists {
		cancel()
		delete(s.containers, id)
	}
}

// readStream scans a demultiplexed reader (Stdout or Stderr) and pushes messages
// to the output channel.
func (s *DockerSource) readStream(r io.Reader, name string, isError bool, logChan chan<- logtypes.LogMessage) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		logChan <- logtypes.LogMessage{
			ContainerName: name,
			Message:       scanner.Text(),
			IsError:       isError,
			Timestamp:     time.Now(),
		}
	}
}
