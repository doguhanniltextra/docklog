package docker

import (
	"bufio"
	"context"
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

type DockerSource struct {
	cli        *client.Client
	cfg        *config.Config
	containers map[string]context.CancelFunc
	mu         sync.Mutex
}

func NewDockerSource(cfg *config.Config) (*DockerSource, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &DockerSource{
		cli:        cli,
		cfg:        cfg,
		containers: make(map[string]context.CancelFunc),
	}, nil
}

func (s *DockerSource) Run(ctx context.Context, logChan chan<- logtypes.LogMessage) error {
	// List current running containers
	containers, err := s.cli.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return err
	}

	for _, c := range containers {
		s.startContainerStream(ctx, c.ID, s.getContainerName(c.Names), logChan)
	}

	// Listen for events
	msgs, errs := s.cli.Events(ctx, types.EventsOptions{})
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
					s.startContainerStream(ctx, msg.ID, name, logChan)
				} else if msg.Action == "die" {
					s.stopContainerStream(msg.ID)
				}
			}
		}
	}
}

func (s *DockerSource) getContainerName(names []string) string {
	if len(names) > 0 {
		return strings.TrimPrefix(names[0], "/")
	}
	return "unknown"
}

func (s *DockerSource) startContainerStream(ctx context.Context, id, name string, logChan chan<- logtypes.LogMessage) {
	if s.cfg.ContainerFilter != nil && !s.cfg.ContainerFilter.MatchString(name) {
		return
	}

	s.mu.Lock()
	if _, exists := s.containers[id]; exists {
		s.mu.Unlock()
		return
	}
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

		stdoutR, stdoutW := io.Pipe()
		stderrR, stderrW := io.Pipe()

		go s.readStream(stdoutR, name, false, logChan)
		go s.readStream(stderrR, name, true, logChan)

		_, _ = stdcopy.StdCopy(stdoutW, stderrW, reader)

		stdoutW.Close()
		stderrW.Close()
	}()
}

func (s *DockerSource) stopContainerStream(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if cancel, exists := s.containers[id]; exists {
		cancel()
		delete(s.containers, id)
	}
}

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
