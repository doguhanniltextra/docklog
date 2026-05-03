package formatter

import (
	"fmt"
	"sync"

	"github.com/doguhanniltextra/docklog/pkg/types"
	"github.com/fatih/color"
)

type HumanFormatter struct {
	timeFormat string
	colors     []*color.Color
	contColors map[string]*color.Color
	colorIdx   int
	mu         sync.Mutex
	errColor   *color.Color
	infoColor  *color.Color
}

func NewHumanFormatter(timeFormat string, colors []*color.Color) *HumanFormatter {
	return &HumanFormatter{
		timeFormat: timeFormat,
		colors:     colors,
		contColors: make(map[string]*color.Color),
		errColor:   color.New(color.FgRed, color.Bold),
		infoColor:  color.New(color.FgHiWhite),
	}
}

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
