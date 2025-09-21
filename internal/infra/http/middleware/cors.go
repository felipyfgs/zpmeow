package middleware

import (
	"time"

	"zpmeow/internal/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORS(cfg config.CORSConfigProvider) gin.HandlerFunc {
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = cfg.GetAllowAllOrigins()
	corsConfig.AllowOrigins = cfg.GetAllowOrigins()
	corsConfig.AllowHeaders = cfg.GetAllowHeaders()
	corsConfig.AllowMethods = cfg.GetAllowMethods()
	corsConfig.ExposeHeaders = cfg.GetExposeHeaders()
	corsConfig.AllowCredentials = cfg.GetAllowCredentials()
	corsConfig.MaxAge = time.Duration(cfg.GetMaxAge()) * time.Second

	return cors.New(corsConfig)
}
