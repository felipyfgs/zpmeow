package middleware

import (
	"strings"

	"zpmeow/internal/infra/logging"

	"github.com/gin-gonic/gin"
)

type LogLevel int

const (
	LogLevelInfo LogLevel = iota
	LogLevelWarn
	LogLevelError
)

type HTTPLogEntry struct {
	Method    string   `json:"method"`
	Path      string   `json:"path"`
	Status    int      `json:"status"`
	Latency   string   `json:"latency"`
	ClientIP  string   `json:"client_ip"`
	UserAgent string   `json:"user_agent,omitempty"`
	Error     string   `json:"error,omitempty"`
	Level     LogLevel `json:"level"`
}

func Logger() gin.HandlerFunc {
	httpLogger := logging.GetLogger().Sub("http")

	return gin.LoggerWithFormatter(func(params gin.LogFormatterParams) string {

		if shouldSkipLogging(params.Path) {
			return ""
		}

		entry := CreateHTTPLogEntry(params)
		LogHTTPRequest(httpLogger, entry)

		return ""
	})
}

func CreateHTTPLogEntry(params gin.LogFormatterParams) HTTPLogEntry {
	entry := HTTPLogEntry{
		Method:   params.Method,
		Path:     params.Path,
		Status:   params.StatusCode,
		Latency:  params.Latency.String(),
		ClientIP: params.ClientIP,
		Level:    determineLogLevel(params.StatusCode),
	}

	if params.ErrorMessage != "" {
		entry.Error = params.ErrorMessage
	}

	if !isStaticResource(params.Path) && params.Request != nil {
		entry.UserAgent = params.Request.UserAgent()
	}

	return entry
}

func LogHTTPRequest(httpLogger logging.Logger, entry HTTPLogEntry) {
	logEntry := httpLogger.With().
		Str("method", entry.Method).
		Str("path", entry.Path).
		Int("status", entry.Status).
		Str("latency", entry.Latency).
		Str("client_ip", entry.ClientIP)

	if entry.Error != "" {
		logEntry = logEntry.Str("error", entry.Error)
	}
	if entry.UserAgent != "" {
		logEntry = logEntry.Str("user_agent", entry.UserAgent)
	}

	switch entry.Level {
	case LogLevelError:
		logEntry.Logger().Error("HTTP Request")
	case LogLevelWarn:
		logEntry.Logger().Warn("HTTP Request")
	default:
		logEntry.Logger().Info("HTTP Request")
	}
}

var skipLogPaths = map[string]bool{
	"/ping":        true,
	"/health":      true,
	"/favicon.ico": true,
}

func shouldSkipLogging(path string) bool {
	return skipLogPaths[path]
}

func determineLogLevel(statusCode int) LogLevel {
	switch {
	case statusCode >= 500:
		return LogLevelError
	case statusCode >= 400:
		return LogLevelWarn
	default:
		return LogLevelInfo
	}
}

var (
	staticPrefixes = []string{
		"/swagger/",
		"/static/",
		"/assets/",
	}

	staticExtensions = map[string]bool{
		".css":   true,
		".js":    true,
		".png":   true,
		".jpg":   true,
		".jpeg":  true,
		".gif":   true,
		".ico":   true,
		".svg":   true,
		".woff":  true,
		".woff2": true,
		".ttf":   true,
		".eot":   true,
	}
)

func isStaticResource(path string) bool {

	for _, prefix := range staticPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	for ext := range staticExtensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}

	return false
}
