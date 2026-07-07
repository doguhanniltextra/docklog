// Package sink defines the final destinations for processed log messages.
// It supports multiple output types like the terminal console and physical files.
package sink

import (
	"errors"
	"fmt"
	"os"

	"github.com/doguhanniltextra/docklog/pkg/formatter"
	"github.com/doguhanniltextra/docklog/pkg/types"
)

// Sink defines the contract for any component that serves as a final
// destination for log messages. Sinks are responsible for the physical
// writing of formatted strings to their respective backends.
type Sink interface {
	// Write formats and delivers the log message to the underlying destination.
	Write(msg types.LogMessage) error
	// Close performs any necessary cleanup (e.g., closing file handles).
	Close() error
}

// ConsoleSink writes formatted log messages directly to standard output (Stdout).
type ConsoleSink struct {
	// formatter is the strategy used to stringify the message before printing.
	formatter formatter.Formatter
}

// NewConsoleSink creates a new ConsoleSink with the specified formatter.
func NewConsoleSink(f formatter.Formatter) *ConsoleSink {
	return &ConsoleSink{formatter: f}
}

// Write outputs the message to the console.
func (s *ConsoleSink) Write(msg types.LogMessage) error {
	_, err := fmt.Fprint(os.Stdout, s.formatter.Format(msg))
	return err
}

// Close is a no-op for the console sink.
func (s *ConsoleSink) Close() error { return nil }

// FileSink persists log messages to a physical file on disk.
type FileSink struct {
	// file is the underlying OS file handle.
	file *os.File
	// formatter is used to format the message (usually a non-colored variant for files).
	formatter formatter.Formatter
}

// NewFileSink opens or creates a file at the specified path for appending logs.
func NewFileSink(path string, f formatter.Formatter) (*FileSink, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %q: %w", path, err)
	}
	return &FileSink{file: file, formatter: f}, nil
}

// Write appends the formatted message to the end of the file.
func (s *FileSink) Write(msg types.LogMessage) error {
	_, err := s.file.WriteString(s.formatter.Format(msg))
	return err
}

// Close closes the underlying file handle.
func (s *FileSink) Close() error {
	return s.file.Close()
}

// MultiSink allows broadcasting a single log message to multiple destinations.
// This is used when logs need to be simultaneously printed to the screen
// and persisted to a file.
type MultiSink struct {
	// sinks is the list of active output destinations.
	sinks []Sink
}

// NewMultiSink creates a composite sink from a variadic list of sinks.
func NewMultiSink(sinks ...Sink) *MultiSink {
	return &MultiSink{sinks: sinks}
}

// Write delivers the message to all child sinks. If any sink fails,
// it returns the error immediately.
func (m *MultiSink) Write(msg types.LogMessage) error {
	for _, s := range m.sinks {
		if err := s.Write(msg); err != nil {
			return err
		}
	}
	return nil
}

// Close gracefully shuts down all child sinks.
func (m *MultiSink) Close() error {
	var errs []error
	for _, s := range m.sinks {
		if err := s.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
