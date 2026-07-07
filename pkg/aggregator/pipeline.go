// Package aggregator provides the core orchestration logic for the log processing
// pipeline. It coordinates the flow of messages from sources to sinks through processors.
package aggregator

import (
	"context"

	"github.com/doguhanniltextra/docklog/pkg/processor"
	"github.com/doguhanniltextra/docklog/pkg/sink"
	"github.com/doguhanniltextra/docklog/pkg/source"
	"github.com/doguhanniltextra/docklog/pkg/types"
)

// Pipeline orchestrates the flow of log messages from a Source, 
// through a series of Processors, to a Sink.
type Pipeline struct {
	source     source.LogSource
	processors []processor.Processor
	sink       sink.Sink
	bufferLen  int
}

// NewPipeline creates a new log processing pipeline.
func NewPipeline(src source.LogSource, procs []processor.Processor, snk sink.Sink, bufferLen int) *Pipeline {
	return &Pipeline{
		source:     src,
		processors: procs,
		sink:       snk,
		bufferLen:  bufferLen,
	}
}

// Run starts the pipeline and blocks until the context is cancelled or an error occurs.
func (p *Pipeline) Run(ctx context.Context) error {
	logChan := make(chan types.LogMessage, p.bufferLen)
	errChan := make(chan error, 1)

	// Close sink when pipeline exits
	defer p.sink.Close()

	// Start source
	go func() {
		errChan <- p.source.Run(ctx, logChan)
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-errChan:
			if err != nil {
				return err
			}
			return nil
		case msg := <-logChan:
			currentMsg := &msg
			keep := true

			// Apply processors in order
			for _, proc := range p.processors {
				var processed *types.LogMessage
				processed, keep = proc.Process(currentMsg)
				if !keep {
					break
				}
				currentMsg = processed
			}

			if keep {
				if err := p.sink.Write(*currentMsg); err != nil {
					return err
				}
			}
		}
	}
}
