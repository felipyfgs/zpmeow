package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"zpmeow/internal/infra/database"
	"zpmeow/internal/interfaces/dto"
)

type HealthHandler struct {
	*BaseHandler
	db *sqlx.DB
}

func NewHealthHandler(db *sqlx.DB) *HealthHandler {
	return &HealthHandler{
		BaseHandler: NewBaseHandler("health-handler"),
		db:          db,
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

func (h *HealthHandler) sendSuccessResponse(c *gin.Context, status, message, version string, dependencies map[string]string) {
	data := HealthData{
		Status:       status,
		Message:      message,
		Version:      version,
		Service:      "meow",
		Timestamp:    time.Now(),
		Dependencies: dependencies,
	}
	h.SendSuccessResponse(c, http.StatusOK, data)
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

	return dependencies
}

// @Summary		Health check endpoint
// @Description	Returns the health status of the service using standardized response format
// @Tags			Health
// @Accept			json
// @Produce		json
// @Success		200	{object}	HealthStandardResponse	"Service is healthy"
// @Failure		503	{object}	HealthStandardResponse	"Service is unhealthy"
// @Router			/health [get]
func (h *HealthHandler) Health(c *gin.Context) {
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
		h.sendSuccessResponse(c, "ok", "Service is healthy", "1.0.0", dependencies)
		h.logger.Infof("Health check completed successfully")
	} else {
		h.SendErrorResponse(c, http.StatusServiceUnavailable, "UNHEALTHY", "Service is unhealthy", nil)
		h.logger.Warnf("Health check failed - some dependencies are unhealthy")
	}
}

// @Summary		Ping endpoint
// @Description	Simple ping endpoint that returns pong using standardized response format
// @Tags			Health
// @Accept			json
// @Produce		json
// @Success		200	{object}	HealthStandardResponse	"Pong response"
// @Router			/ping [get]
func (h *HealthHandler) Ping(c *gin.Context) {
	h.logger.Infof("Ping requested")

	h.sendSuccessResponse(c, "ok", "pong", "", nil)
	h.logger.Infof("Ping completed successfully")
}

// @Summary		Metrics endpoint
// @Description	Returns application metrics and performance data
// @Tags			Health
// @Accept			json
// @Produce		json
// @Success		200	{object}	map[string]interface{}	"Application metrics"
// @Router			/metrics [get]
func (h *HealthHandler) Metrics(c *gin.Context) {
	h.logger.Infof("Metrics requested")

	metrics := map[string]interface{}{
		"status":    "metrics not implemented yet",
		"timestamp": time.Now().Unix(),
	}
	h.SendSuccessResponse(c, http.StatusOK, metrics)

	h.logger.Infof("Metrics completed successfully")
}

// @Summary		Reset metrics endpoint
// @Description	Resets all application metrics
// @Tags			Health
// @Accept			json
// @Produce		json
// @Success		200	{object}	HealthStandardResponse	"Metrics reset successfully"
// @Router			/metrics/reset [post]
func (h *HealthHandler) ResetMetrics(c *gin.Context) {
	h.logger.Infof("Metrics reset requested")

	h.sendSuccessResponse(c, "ok", "Metrics reset not implemented yet", "", nil)

	h.logger.Infof("Metrics reset completed successfully")
}
