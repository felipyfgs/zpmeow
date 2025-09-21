package middleware

import (
	"context"
	"fmt"

	"zpmeow/internal/config"
	"zpmeow/internal/domain/session"
	"zpmeow/internal/infra/logging"

	"github.com/gofiber/fiber/v2"
)

type AuthMiddleware struct {
	config      *config.Config
	sessionRepo session.Repository
	logger      logging.Logger
}

func NewAuthMiddleware(config *config.Config, sessionRepo session.Repository, logger logging.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		config:      config,
		sessionRepo: sessionRepo,
		logger:      logger,
	}
}

func (a *AuthMiddleware) AuthenticateGlobal() fiber.Handler {
	return func(c *fiber.Ctx) error {
		apiKey := a.extractAPIKey(c)
		if apiKey == "" {
			a.logger.Warn("Missing API key in request")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "API key required"})
		}

		if apiKey == a.config.GetAuth().GetGlobalAPIKey() {
			a.logger.Debug("Global API key authenticated successfully")
			c.Locals("auth_type", "global")
			c.Locals("api_key", apiKey)
			return c.Next()
		}

		a.logger.Warn("Invalid global API key provided")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid API key"})
	}
}

func (a *AuthMiddleware) AuthenticateSession() fiber.Handler {
	return func(c *fiber.Ctx) error {
		apiKey := a.extractAPIKey(c)
		if apiKey == "" {
			a.logger.Warn("Missing API key in request")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "API key required"})
		}

		if apiKey == a.config.GetAuth().GetGlobalAPIKey() {
			a.logger.Debug("Global API key authenticated for session access")
			c.Locals("auth_type", "global")
			c.Locals("api_key", apiKey)
			return c.Next()
		}

		session, err := a.sessionRepo.GetByApiKey(context.Background(), apiKey)
		if err != nil {
			if err == fmt.Errorf("session not found") {
				a.logger.Warn("Invalid session API key provided")
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid API key"})
			} else {
				a.logger.Error("Error validating session API key: " + err.Error())
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Authentication error"})
			}
		}

		a.logger.Debug("Session API key authenticated successfully for session: " + session.SessionID().Value() + " (" + session.Name().Value() + ")")
		c.Locals("auth_type", "session")
		c.Locals("api_key", apiKey)
		c.Locals("session_id", session.SessionID())
		c.Locals("session", session)
		return c.Next()
	}
}

func (a *AuthMiddleware) AuthenticateAny() fiber.Handler {
	return func(c *fiber.Ctx) error {
		apiKey := a.extractAPIKey(c)
		if apiKey == "" {
			a.logger.Warn("Missing API key in request")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "API key required"})
		}

		if apiKey == a.config.GetAuth().GetGlobalAPIKey() {
			a.logger.Debug("Global API key authenticated")
			c.Locals("auth_type", "global")
			c.Locals("api_key", apiKey)
			return c.Next()
		}

		session, err := a.sessionRepo.GetByApiKey(context.Background(), apiKey)
		if err != nil {
			if err == fmt.Errorf("session not found") {
				a.logger.Warn("Invalid API key provided")
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid API key"})
			} else {
				a.logger.Error("Error validating API key: " + err.Error())
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Authentication error"})
			}
		}

		a.logger.Debug("Session API key authenticated for session: " + session.SessionID().Value() + " (" + session.Name().Value() + ")")
		c.Locals("auth_type", "session")
		c.Locals("api_key", apiKey)
		c.Locals("session_id", session.SessionID())
		c.Locals("session", session)
		return c.Next()
	}
}

func (a *AuthMiddleware) extractAPIKey(c *fiber.Ctx) string {
	// Try X-API-Key header first (preferred)
	apiKey := c.Get("X-API-Key")
	if apiKey != "" {
		return apiKey
	}

	// Fallback to Authorization header
	authHeader := c.Get("Authorization")
	if authHeader != "" {
		return authHeader
	}

	return ""
}

func GetAuthenticatedSession(c *fiber.Ctx) (*session.Session, bool) {
	if sessionData := c.Locals("session"); sessionData != nil {
		if s, ok := sessionData.(*session.Session); ok {
			return s, true
		}
	}
	return nil, false
}

func GetAuthType(c *fiber.Ctx) string {
	if authType := c.Locals("auth_type"); authType != nil {
		if t, ok := authType.(string); ok {
			return t
		}
	}
	return ""
}

func IsGlobalAuth(c *fiber.Ctx) bool {
	return GetAuthType(c) == "global"
}

func IsSessionAuth(c *fiber.Ctx) bool {
	return GetAuthType(c) == "session"
}
