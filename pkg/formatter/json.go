package formatter

import (
	"encoding/json"

	"github.com/doguhanniltextra/docklog/pkg/types"
)

// JsonFormatter transforms a LogMessage into a single-line JSON string.
// This is ideal for piping output into tools like 'jq' or sending logs
// to centralized logging systems like ELK or Datadog.
type JsonFormatter struct{}

// NewJsonFormatter creates a new instance of JsonFormatter.
func NewJsonFormatter() *JsonFormatter {
	return &JsonFormatter{}
}

// Format marshals the LogMessage into JSON and appends a newline.
// It ignores any errors during marshaling and returns an empty string if it fails.
func (j *JsonFormatter) Format(msg types.LogMessage) string {
	b, err := json.Marshal(msg)
	if err != nil {
		return ""
	}
	return string(b) + "\n"
}
