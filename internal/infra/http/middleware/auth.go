package middleware

import (
	"context"
	"fmt"
	"net/http"

	"zpmeow/internal/config"
	"zpmeow/internal/domain/session"
	"zpmeow/internal/infra/logging"

	"github.com/gin-gonic/gin"
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

func (a *AuthMiddleware) AuthenticateGlobal() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := a.extractAPIKey(c)
		if apiKey == "" {
			a.logger.Warn("Missing API key in request")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API key required"})
			c.Abort()
			return
		}

		if apiKey == a.config.GetAuth().GetGlobalAPIKey() {
			a.logger.Debug("Global API key authenticated successfully")
			c.Set("auth_type", "global")
			c.Set("api_key", apiKey)
			c.Next()
			return
		}

		a.logger.Warn("Invalid global API key provided")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
		c.Abort()
	}
}

func (a *AuthMiddleware) AuthenticateSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := a.extractAPIKey(c)
		if apiKey == "" {
			a.logger.Warn("Missing API key in request")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API key required"})
			c.Abort()
			return
		}

		if apiKey == a.config.GetAuth().GetGlobalAPIKey() {
			a.logger.Debug("Global API key authenticated for session access")
			c.Set("auth_type", "global")
			c.Set("api_key", apiKey)
			c.Next()
			return
		}

		session, err := a.sessionRepo.GetByApiKey(context.Background(), apiKey)
		if err != nil {
			if err == fmt.Errorf("session not found") {
				a.logger.Warn("Invalid session API key provided")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
			} else {
				a.logger.Error("Error validating session API key: " + err.Error())
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication error"})
			}
			c.Abort()
			return
		}

		a.logger.Debug("Session API key authenticated successfully for session: " + session.SessionID().Value() + " (" + session.Name().Value() + ")")
		c.Set("auth_type", "session")
		c.Set("api_key", apiKey)
		c.Set("session_id", session.SessionID())
		c.Set("session", session)
		c.Next()
	}
}

func (a *AuthMiddleware) AuthenticateAny() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := a.extractAPIKey(c)
		if apiKey == "" {
			a.logger.Warn("Missing API key in request")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API key required"})
			c.Abort()
			return
		}

		if apiKey == a.config.GetAuth().GetGlobalAPIKey() {
			a.logger.Debug("Global API key authenticated")
			c.Set("auth_type", "global")
			c.Set("api_key", apiKey)
			c.Next()
			return
		}

		session, err := a.sessionRepo.GetByApiKey(context.Background(), apiKey)
		if err != nil {
			if err == fmt.Errorf("session not found") {
				a.logger.Warn("Invalid API key provided")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
			} else {
				a.logger.Error("Error validating API key: " + err.Error())
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication error"})
			}
			c.Abort()
			return
		}

		a.logger.Debug("Session API key authenticated for session: " + session.SessionID().Value() + " (" + session.Name().Value() + ")")
		c.Set("auth_type", "session")
		c.Set("api_key", apiKey)
		c.Set("session_id", session.SessionID())
		c.Set("session", session)
		c.Next()
	}
}

func (a *AuthMiddleware) extractAPIKey(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		return authHeader
	}
	return ""
}

func GetAuthenticatedSession(c *gin.Context) (*session.Session, bool) {
	if sessionData, exists := c.Get("session"); exists {
		if s, ok := sessionData.(*session.Session); ok {
			return s, true
		}
	}
	return nil, false
}

func GetAuthType(c *gin.Context) string {
	if authType, exists := c.Get("auth_type"); exists {
		if t, ok := authType.(string); ok {
			return t
		}
	}
	return ""
}

func IsGlobalAuth(c *gin.Context) bool {
	return GetAuthType(c) == "global"
}

func IsSessionAuth(c *gin.Context) bool {
	return GetAuthType(c) == "session"
}
