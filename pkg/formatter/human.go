package formatter

import (
	"fmt"
	"sync"

	"github.com/doguhanniltextra/docklog/pkg/types"
	"github.com/fatih/color"
)

// HumanFormatter converts LogMessages into a visually rich, color-coded string
// optimized for terminal consumption. It maintains state to ensure that
// container names are consistently colored throughout the session.
type HumanFormatter struct {
	// timeFormat is the Go time layout string (e.g., "15:04:05").
	timeFormat string
	// colors is the available palette of terminal colors.
	colors []*color.Color
	// contColors stores assigned colors per container to maintain visual consistency.
	contColors map[string]*color.Color
	// colorIdx tracks the next color to assign from the palette.
	colorIdx int
	// mu protects the color maps and index during concurrent formatting.
	mu sync.Mutex
	// errColor is the styling applied to ERROR level logs.
	errColor *color.Color
	// infoColor is the styling applied to INFO level logs.
	infoColor *color.Color
}

// NewHumanFormatter initializes a HumanFormatter with a specific time layout
// and a color palette.
func NewHumanFormatter(timeFormat string, colors []*color.Color) *HumanFormatter {
	return &HumanFormatter{
		timeFormat: timeFormat,
		colors:     colors,
		contColors: make(map[string]*color.Color),
		errColor:   color.New(color.FgRed, color.Bold),
		infoColor:  color.New(color.FgHiWhite),
	}
}

// getColor retrieves the existing color for a container or assigns a new one
// if it's the first time the container is being seen.
func (h *HumanFormatter) getColor(containerName string) *color.Color {
	h.mu.Lock()
	defer h.mu.Unlock()
	if c, ok := h.contColors[containerName]; ok {
		return c
	}

	c := h.colors[h.colorIdx%len(h.colors)]
	h.colorIdx++
	h.contColors[containerName] = c
	return c
}

// Format applies colorization and layout to a LogMessage.
// The output format is: [LEVEL][TIMESTAMP] container="name" message...
func (h *HumanFormatter) Format(msg types.LogMessage) string {
	cColor := h.getColor(msg.ContainerName)
	timeStr := msg.Timestamp.Format(h.timeFormat)

	level := "INFO"
	msgColor := h.infoColor
	if msg.IsError {
		level = "ERROR"
		msgColor = h.errColor
	}

	return fmt.Sprintf("%s[%s] %s %s\n",
		msgColor.Sprint(level),
		timeStr,
		cColor.Sprintf("container=%q", msg.ContainerName),
		msg.Message,
	)
}
