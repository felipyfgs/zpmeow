package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func RequestValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			contentType := c.GetHeader("Content-Type")
			if contentType == "" {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error":   "Missing Content-Type header",
					"message": "Content-Type header is required for this request",
					"code":    http.StatusBadRequest,
				})
				return
			}

			if strings.HasPrefix(c.Request.URL.Path, "/api/") && !strings.Contains(contentType, "application/json") {
				c.AbortWithStatusJSON(http.StatusUnsupportedMediaType, gin.H{
					"error":   "Unsupported Content-Type",
					"message": "Content-Type must be application/json for API endpoints",
					"code":    http.StatusUnsupportedMediaType,
				})
				return
			}
		}

		if c.Request.ContentLength > 10*1024*1024 {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
				"error":   "Request too large",
				"message": "Request body exceeds maximum size of 10MB",
				"code":    http.StatusRequestEntityTooLarge,
			})
			return
		}

		if strings.Contains(c.GetHeader("Content-Type"), "application/json") && c.Request.ContentLength > 0 {
			if err := validateJSONStructure(c); err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error":   "Invalid JSON",
					"message": fmt.Sprintf("Request body contains invalid JSON: %s", err.Error()),
					"code":    http.StatusBadRequest,
				})
				return
			}
		}

		c.Next()
	}
}

func validateJSONStructure(c *gin.Context) error {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}

	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	var js json.RawMessage
	if err := json.Unmarshal(body, &js); err != nil {
		return fmt.Errorf("invalid JSON structure: %w", err)
	}

	return nil
}

func ContentSecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")

		c.Header("X-Frame-Options", "DENY")

		c.Header("X-XSS-Protection", "1; mode=block")

		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline'; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data: https:; " +
			"font-src 'self'; " +
			"connect-src 'self'; " +
			"frame-ancestors 'none'"
		c.Header("Content-Security-Policy", csp)

		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Header("Cache-Control", "no-store, no-cache, must-revalidate, private")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}

		c.Next()
	}
}

func APIVersionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiVersion := c.GetHeader("API-Version")
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
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":              "Unsupported API version",
				"message":            fmt.Sprintf("API version '%s' is not supported", apiVersion),
				"supported_versions": supportedVersions,
				"code":               http.StatusBadRequest,
			})
			return
		}

		c.Set("api_version", apiVersion)
		c.Header("API-Version", apiVersion)

		c.Next()
	}
}

func CompressionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		acceptEncoding := c.GetHeader("Accept-Encoding")
		if strings.Contains(acceptEncoding, "gzip") {
			c.Header("Content-Encoding", "gzip")
		}

		c.Next()
	}
}

func RequestSizeMiddleware(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
				"error":   "Request too large",
				"message": fmt.Sprintf("Request body exceeds maximum size of %d bytes", maxSize),
				"code":    http.StatusRequestEntityTooLarge,
			})
			return
		}

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)

		c.Next()
	}
}

func MethodOverrideMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "POST" {
			method := c.GetHeader("X-HTTP-Method-Override")
			if method != "" {
				c.Request.Method = strings.ToUpper(method)
			}
		}

		c.Next()
	}
}

func CacheControlMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		if strings.HasPrefix(path, "/api/") {
			c.Header("Cache-Control", "no-store, no-cache, must-revalidate, private")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		} else if strings.HasPrefix(path, "/static/") {
			c.Header("Cache-Control", "public, max-age=3600")
		} else if path == "/health" || path == "/ping" {
			c.Header("Cache-Control", "public, max-age=60")
		}

		c.Next()
	}
}
