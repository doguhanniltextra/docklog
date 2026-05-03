package sink

import (
	"fmt"
	"os"

	"github.com/doguhanniltextra/docklog/pkg/formatter"
	"github.com/doguhanniltextra/docklog/pkg/types"
)

// Sink defines the interface for log output destinations.
type Sink interface {
	Write(msg types.LogMessage) error
	Close() error
}

type ConsoleSink struct {
	formatter formatter.Formatter
}

func NewConsoleSink(f formatter.Formatter) *ConsoleSink {
	return &ConsoleSink{formatter: f}
}

func (s *ConsoleSink) Write(msg types.LogMessage) error {
	fmt.Print(s.formatter.Format(msg))
	return nil
}

func (s *ConsoleSink) Close() error { return nil }

type FileSink struct {
	file      *os.File
	formatter formatter.Formatter
}

func NewFileSink(path string, f formatter.Formatter) (*FileSink, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return &FileSink{file: file, formatter: f}, nil
}

func (s *FileSink) Write(msg types.LogMessage) error {
	_, err := s.file.WriteString(s.formatter.Format(msg))
	return err
}

func (s *FileSink) Close() error {
	return s.file.Close()
}

type MultiSink struct {
	sinks []Sink
}

func NewMultiSink(sinks ...Sink) *MultiSink {
	return &MultiSink{sinks: sinks}
}

func (m *MultiSink) Write(msg types.LogMessage) error {
	for _, s := range m.sinks {
		if err := s.Write(msg); err != nil {
			return err
		}
	}
	return nil
}

func (m *MultiSink) Close() error {
	for _, s := range m.sinks {
		s.Close()
	}
	return nil
}
