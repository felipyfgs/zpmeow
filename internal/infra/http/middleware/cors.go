package middleware

import (
	"time"

	"zpmeow/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func CORS(cfg config.CORSConfigProvider) fiber.Handler {
	corsConfig := cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-API-Key,X-Correlation-ID",
		ExposeHeaders:    "X-Correlation-ID",
		AllowCredentials: cfg.GetAllowCredentials(),
		MaxAge:           int(time.Duration(cfg.GetMaxAge()) * time.Second / time.Second),
	}

	if !cfg.GetAllowAllOrigins() && len(cfg.GetAllowOrigins()) > 0 {
		corsConfig.AllowOrigins = ""
		for i, origin := range cfg.GetAllowOrigins() {
			if i > 0 {
				corsConfig.AllowOrigins += ","
			}
			corsConfig.AllowOrigins += origin
		}
	}

	if len(cfg.GetAllowMethods()) > 0 {
		corsConfig.AllowMethods = ""
		for i, method := range cfg.GetAllowMethods() {
			if i > 0 {
				corsConfig.AllowMethods += ","
			}
			corsConfig.AllowMethods += method
		}
	}

	if len(cfg.GetAllowHeaders()) > 0 {
		corsConfig.AllowHeaders = ""
		for i, header := range cfg.GetAllowHeaders() {
			if i > 0 {
				corsConfig.AllowHeaders += ","
			}
			corsConfig.AllowHeaders += header
		}
	}

	if len(cfg.GetExposeHeaders()) > 0 {
		corsConfig.ExposeHeaders = ""
		for i, header := range cfg.GetExposeHeaders() {
			if i > 0 {
				corsConfig.ExposeHeaders += ","
			}
			corsConfig.ExposeHeaders += header
		}
	}

	return cors.New(corsConfig)
}
