package middleware

import (
	"strings"
	"time"

	"zpmeow/internal/infra/logging"

	"github.com/gofiber/fiber/v2"
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

func Logger() fiber.Handler {
	httpLogger := logging.GetLogger().Sub("http")

	return func(c *fiber.Ctx) error {
		start := time.Now()
		path := c.Path()

		// Skip logging for certain paths
		if shouldSkipLogging(path) {
			return c.Next()
		}

		// Process request
		err := c.Next()

		// Log after request is processed
		latency := time.Since(start)
		status := c.Response().StatusCode()

		entry := HTTPLogEntry{
			Method:   c.Method(),
			Path:     path,
			Status:   status,
			Latency:  latency.String(),
			ClientIP: c.IP(),
			Level:    determineLogLevel(status),
		}

		// Add error if present
		if err != nil {
			entry.Error = err.Error()
		}

		// Add user agent for non-static resources
		if !isStaticResource(path) {
			entry.UserAgent = c.Get("User-Agent")
		}

		// Extract correlation ID from context
		correlationID := GetCorrelationID(c.Context())

		LogHTTPRequest(httpLogger, entry, correlationID)

		return err
	}
}

func LogHTTPRequest(httpLogger logging.Logger, entry HTTPLogEntry, correlationID string) {
	logEntry := httpLogger.With().
		Str("method", entry.Method).
		Str("path", entry.Path).
		Int("status", entry.Status).
		Str("latency", entry.Latency).
		Str("client_ip", entry.ClientIP)

	if correlationID != "" {
		logEntry = logEntry.Str("correlation_id", correlationID)
	}

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
