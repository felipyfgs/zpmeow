package middleware

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func RequestValidationMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if string(c.Method()) == "POST" || string(c.Method()) == "PUT" || string(c.Method()) == "PATCH" {
			contentType := c.Get("Content-Type")
			if contentType == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   "Missing Content-Type header",
					"message": "Content-Type header is required for this request",
					"code":    fiber.StatusBadRequest,
				})
			}

			if strings.HasPrefix(c.Path(), "/api/") && !strings.Contains(contentType, "application/json") {
				return c.Status(fiber.StatusUnsupportedMediaType).JSON(fiber.Map{
					"error":   "Unsupported Content-Type",
					"message": "Content-Type must be application/json for API endpoints",
					"code":    fiber.StatusUnsupportedMediaType,
				})
			}

			if len(c.Body()) > 10*1024*1024 {
				return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
					"error":   "Request too large",
					"message": "Request body exceeds maximum size of 10MB",
					"code":    fiber.StatusRequestEntityTooLarge,
				})
			}

			if strings.Contains(c.Get("Content-Type"), "application/json") && len(c.Body()) > 0 {
				if err := validateJSONStructure(c); err != nil {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"error":   "Invalid JSON",
						"message": fmt.Sprintf("Request body contains invalid JSON: %s", err.Error()),
						"code":    fiber.StatusBadRequest,
					})
				}
			}
		}

		return c.Next()
	}
}

func validateJSONStructure(c *fiber.Ctx) error {
	body := c.Body()
	if len(body) == 0 {
		return nil
	}

	var js json.RawMessage
	if err := json.Unmarshal(body, &js); err != nil {
		return fmt.Errorf("invalid JSON structure: %w", err)
	}

	return nil
}

func ContentSecurityMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline'; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data: https:; " +
			"font-src 'self'; " +
			"connect-src 'self'; " +
			"frame-ancestors 'none'"
		c.Set("Content-Security-Policy", csp)

		if strings.HasPrefix(c.Path(), "/api/") {
			c.Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
			c.Set("Pragma", "no-cache")
			c.Set("Expires", "0")
		}

		return c.Next()
	}
}

func APIVersionMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		apiVersion := c.Get("API-Version")
		if apiVersion == "" {
			apiVersion = "v1"
		}

		supportedVersions := []string{"v1"}
		isSupported := false
		for _, version := range supportedVersions {
			if apiVersion == version {
				isSupported = true
				break
			}
		}

		if !isSupported {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Unsupported API version",
				"message": fmt.Sprintf("API version '%s' is not supported. Supported versions: %v", apiVersion, supportedVersions),
				"code":    fiber.StatusBadRequest,
			})
		}

		c.Set("API-Version", apiVersion)
		return c.Next()
	}
}

func CompressionMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		acceptEncoding := c.Get("Accept-Encoding")
		if strings.Contains(acceptEncoding, "gzip") {
			c.Set("Content-Encoding", "gzip")
		}
		return c.Next()
	}
}

func RequestSizeMiddleware(maxSize int64) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if int64(len(c.Body())) > maxSize {
			return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
				"error":   "Request too large",
				"message": fmt.Sprintf("Request body exceeds maximum size of %d bytes", maxSize),
				"code":    fiber.StatusRequestEntityTooLarge,
			})
		}
		return c.Next()
	}
}

func MethodOverrideMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if string(c.Method()) == "POST" {
			override := c.Get("X-HTTP-Method-Override")
			if override != "" {
				c.Set("X-Original-Method", string(c.Method()))
				c.Set("X-Override-Method", override)
			}
		}
		return c.Next()
	}
}

func CacheControlMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if strings.HasPrefix(c.Path(), "/api/") {
			c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Set("Pragma", "no-cache")
			c.Set("Expires", "0")
		} else {
			c.Set("Cache-Control", "public, max-age=3600")
		}
		return c.Next()
	}
}
