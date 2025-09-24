package handlers

import (
	"context"
	"time"

	"zpmeow/internal/application/ports"
	"zpmeow/internal/infra/database"
	"zpmeow/internal/infra/http/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

type HealthHandler struct {
	*BaseHandler
	db    *sqlx.DB
	cache ports.CacheManager
}

func NewHealthHandler(db *sqlx.DB) *HealthHandler {
	return &HealthHandler{
		BaseHandler: NewBaseHandler("health-handler"),
		db:          db,
		cache:       nil,
	}
}

func NewHealthHandlerWithCache(db *sqlx.DB, cache ports.CacheManager) *HealthHandler {
	return &HealthHandler{
		BaseHandler: NewBaseHandler("health-handler"),
		db:          db,
		cache:       cache,
	}
}

type HealthData struct {
	Status       string            `json:"status" example:"ok"`
	Message      string            `json:"message" example:"Service is healthy"`
	Version      string            `json:"version,omitempty" example:"1.0.0"`
	Service      string            `json:"service" example:"meow"`
	Timestamp    time.Time         `json:"timestamp"`
	Dependencies map[string]string `json:"dependencies,omitempty"`
}

type HealthStandardResponse = dto.StandardResponse

func (h *HealthHandler) sendSuccessResponse(c *fiber.Ctx, status, message, version string, dependencies map[string]string) error {
	data := HealthData{
		Status:       status,
		Message:      message,
		Version:      version,
		Service:      "meow",
		Timestamp:    time.Now(),
		Dependencies: dependencies,
	}
	return h.SendSuccessResponse(c, fiber.StatusOK, data)
}

func (h *HealthHandler) checkDependencies() map[string]string {
	dependencies := make(map[string]string)

	if h.db != nil {
		if err := database.HealthCheck(h.db); err != nil {
			dependencies["database"] = "unhealthy: " + err.Error()
		} else {
			dependencies["database"] = "healthy"
		}
	} else {
		dependencies["database"] = "not configured"
	}

	if h.cache != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := h.cache.Ping(ctx); err != nil {
			dependencies["cache"] = "unhealthy: " + err.Error()
		} else {
			dependencies["cache"] = "healthy"
		}
	} else {
		dependencies["cache"] = "not configured"
	}

	return dependencies
}

// Health godoc
// @Summary Health check endpoint
// @Description Performs a comprehensive health check of the service and its dependencies
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} dto.StandardResponse{data=HealthData} "Service is healthy"
// @Failure 503 {object} dto.StandardResponse "Service is unhealthy"
// @Router /health [get]
func (h *HealthHandler) Health(c *fiber.Ctx) error {
	h.logger.Infof("Health check requested")

	dependencies := h.checkDependencies()

	allHealthy := true
	for _, status := range dependencies {
		if status != "healthy" {
			allHealthy = false
			break
		}
	}

	if allHealthy {
		h.logger.Infof("Health check completed successfully")
		return h.sendSuccessResponse(c, "ok", "Service is healthy", "1.0.0", dependencies)
	} else {
		h.logger.Warnf("Health check failed - some dependencies are unhealthy")
		return h.SendErrorResponse(c, fiber.StatusServiceUnavailable, "UNHEALTHY", "Service is unhealthy", nil)
	}
}
