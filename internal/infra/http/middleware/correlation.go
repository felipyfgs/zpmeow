package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	CorrelationIDHeader = "X-Correlation-ID"
	CorrelationIDKey    = "correlation_id"
)

// CorrelationIDMiddleware adds a correlation ID to each request
func CorrelationIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if correlation ID already exists in header
		correlationID := c.GetHeader(CorrelationIDHeader)
		
		// If not provided, generate a new one
		if correlationID == "" {
			correlationID = generateCorrelationID()
		}
		
		// Add to context
		ctx := context.WithValue(c.Request.Context(), CorrelationIDKey, correlationID)
		c.Request = c.Request.WithContext(ctx)
		
		// Add to response header
		c.Header(CorrelationIDHeader, correlationID)
		
		// Add to gin context for easy access
		c.Set(CorrelationIDKey, correlationID)
		
		c.Next()
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

// GetCorrelationIDFromGin extracts correlation ID from gin context
func GetCorrelationIDFromGin(c *gin.Context) string {
	if id, exists := c.Get(CorrelationIDKey); exists {
		if correlationID, ok := id.(string); ok {
			return correlationID
		}
	}
	return ""
}
