package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorGreen  = "\033[32m"
	ColorCyan   = "\033[36m"
	ColorGray   = "\033[90m"
	ColorWhite  = "\033[97m"
	ColorBold   = "\033[1m"
	ColorDim    = "\033[2m"
)

var levelColors = map[string]string{
	"debug": ColorGray,
	"info":  ColorGreen,
	"warn":  ColorYellow,
	"error": ColorRed,
	"fatal": ColorRed + ColorBold,
	"panic": ColorRed + ColorBold,
}

var levelAbbrev = map[string]string{
	"debug": "DBG",
	"info":  "INF",
	"warn":  "WRN",
	"error": "ERR",
	"fatal": "FTL",
	"panic": "PNC",
}

type ConsoleWriter struct {
	Out     io.Writer
	NoColor bool
}

func (w *ConsoleWriter) Write(p []byte) (n int, err error) {
	var event map[string]interface{}

	if err := parseJSON(p, &event); err != nil {
		_, writeErr := w.Out.Write(p)
		if writeErr != nil {
			return 0, writeErr
		}
		return len(p), nil
	}

	formatted := w.formatEvent(event)

	_, writeErr := w.Out.Write([]byte(formatted))
	if writeErr != nil {
		return 0, writeErr
	}

	return len(p), nil // Return original length to satisfy zerolog
}

func (w *ConsoleWriter) formatEvent(event map[string]interface{}) string {
	var parts []string

	if timestamp, ok := event["time"].(string); ok {
		if t, err := time.Parse(time.RFC3339, timestamp); err == nil {
			timeStr := t.Format("02-01-2006 15:04:05")
			parts = append(parts, timeStr)
		}
	}

	level := "info"
	if l, ok := event["level"].(string); ok {
		level = l
	}

	levelStr := levelAbbrev[level]
	if levelStr == "" {
		levelStr = strings.ToUpper(level)[:3]
	}

	if !w.NoColor {
		color := levelColors[level]
		levelStr = color + levelStr + ColorReset
	}
	parts = append(parts, levelStr)

	message := ""
	if msg, ok := event["message"].(string); ok {
		message = msg
	}
	parts = append(parts, message)

	context := w.formatContext(event)
	if context != "" {
		parts = append(parts, context)
	}

	return strings.Join(parts, " ") + "\n"
}

func (w *ConsoleWriter) formatContext(event map[string]interface{}) string {
	var contextParts []string

	skipFields := map[string]bool{
		"time":    true,
		"level":   true,
		"message": true,
		"caller":  true,
	}

	fieldOrder := []string{"module", "action", "event", "userId", "jid", "to", "from", "type", "traceId", "correlationId", "plan", "ts", "payload", "reason", "error"}

	for _, field := range fieldOrder {
		if value, exists := event[field]; exists {
			valueStr := w.formatValue(field, value)
			contextParts = append(contextParts, fmt.Sprintf("%s=%s", field, valueStr))
		}
	}

	for key, value := range event {
		if skipFields[key] {
			continue
		}

		found := false
		for _, field := range fieldOrder {
			if key == field {
				found = true
				break
			}
		}
		if found {
			continue
		}

		valueStr := w.formatValue(key, value)
		contextParts = append(contextParts, fmt.Sprintf("%s=%s", key, valueStr))
	}

	return strings.Join(contextParts, " ")
}

func (w *ConsoleWriter) formatValue(key string, value interface{}) string {
	switch v := value.(type) {
	case string:
		if key == "payload" && len(v) > 100 {
			return fmt.Sprintf(`"%s..."`, v[:97])
		}
		if strings.Contains(v, " ") || strings.Contains(v, "=") {
			return fmt.Sprintf(`"%s"`, v)
		}
		return v
	case int, int64, float64:
		return fmt.Sprintf("%v", v)
	case bool:
		return strconv.FormatBool(v)
	case time.Time:
		return v.Format(time.RFC3339)
	default:
		return fmt.Sprintf("%v", value)
	}
}

func isIDField(key string) bool {
	key = strings.ToLower(key)
	return strings.Contains(key, "id") ||
		strings.Contains(key, "uuid") ||
		strings.Contains(key, "hash") ||
		strings.Contains(key, "token") ||
		strings.Contains(key, "session")
}

func parseJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func FormatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%.2fμs", float64(d.Nanoseconds())/1000)
	}
	if d < time.Second {
		return fmt.Sprintf("%.2fms", float64(d.Nanoseconds())/1000000)
	}
	if d < time.Minute {
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
	return d.String()
}

func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
