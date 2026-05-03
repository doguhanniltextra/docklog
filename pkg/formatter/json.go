package formatter

import (
	"encoding/json"

	"github.com/doguhanniltextra/docklog/pkg/types"
)

type JsonFormatter struct{}

func NewJsonFormatter() *JsonFormatter {
	return &JsonFormatter{}
}

func (j *JsonFormatter) Format(msg types.LogMessage) string {
	b, _ := json.Marshal(msg)
	return string(b) + "\n"
}
