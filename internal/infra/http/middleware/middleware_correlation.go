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

type contextKey string

const correlationIDContextKey contextKey = "correlation_id"

func CorrelationIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		correlationID := c.Get(CorrelationIDHeader)

		if correlationID == "" {
			correlationID = generateCorrelationID()
		}

		ctx := context.WithValue(c.Context(), correlationIDContextKey, correlationID)
		c.SetUserContext(ctx)

		c.Set(CorrelationIDHeader, correlationID)

		c.Locals(CorrelationIDKey, correlationID)

		return c.Next()
	}
}

func generateCorrelationID() string {
	id := uuid.New().String()
	return id[:8]
}

func GetCorrelationID(ctx context.Context) string {
	if id, ok := ctx.Value(CorrelationIDKey).(string); ok {
		return id
	}
	return ""
}

func GetCorrelationIDFromFiber(c *fiber.Ctx) string {
	if id := c.Locals(CorrelationIDKey); id != nil {
		if correlationID, ok := id.(string); ok {
			return correlationID
		}
	}
	return ""
}
