package middleware

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	CorrelationIDHeader = "X-Correlation-ID"
	CorrelationIDKey    = "correlation_id"
)

// CorrelationIDMiddleware adds a correlation ID to each request
func CorrelationIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if correlation ID already exists in header
		correlationID := c.Get(CorrelationIDHeader)

		// If not provided, generate a new one
		if correlationID == "" {
			correlationID = generateCorrelationID()
		}

		// Add to context
		ctx := context.WithValue(c.Context(), CorrelationIDKey, correlationID)
		c.SetUserContext(ctx)

		// Add to response header
		c.Set(CorrelationIDHeader, correlationID)

		// Add to fiber context for easy access
		c.Locals(CorrelationIDKey, correlationID)

		return c.Next()
	}
}

// generateCorrelationID generates a short correlation ID
func generateCorrelationID() string {
	// Generate a UUID and take first 8 characters for brevity
	id := uuid.New().String()
	return id[:8]
}

// GetCorrelationID extracts correlation ID from context
func GetCorrelationID(ctx context.Context) string {
	if id, ok := ctx.Value(CorrelationIDKey).(string); ok {
		return id
	}
	return ""
}

// GetCorrelationIDFromFiber extracts correlation ID from fiber context
func GetCorrelationIDFromFiber(c *fiber.Ctx) string {
	if id := c.Locals(CorrelationIDKey); id != nil {
		if correlationID, ok := id.(string); ok {
			return correlationID
		}
	}
	return ""
}
