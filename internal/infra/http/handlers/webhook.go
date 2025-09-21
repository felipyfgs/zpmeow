package handlers

import (
	"fmt"
	"strings"
	"time"

	"zpmeow/internal/application"
	"zpmeow/internal/infra/http/dto"
	"zpmeow/internal/infra/wmeow"

	"github.com/gofiber/fiber/v2"
)

type WebhookHandler struct {
	*BaseHandler
	sessionService *application.SessionApp
	webhookApp     *application.WebhookApp
	wmeowService   wmeow.WameowService
}

func NewWebhookHandler(sessionService *application.SessionApp, webhookApp *application.WebhookApp, wmeowService wmeow.WameowService) *WebhookHandler {
	return &WebhookHandler{
		BaseHandler:    NewBaseHandler("webhook-handler"),
		sessionService: sessionService,
		webhookApp:     webhookApp,
		wmeowService:   wmeowService,
	}
}

func (h *WebhookHandler) resolveSessionID(c *fiber.Ctx, sessionIDOrName string) (string, error) {
	if h.sessionService == nil {
		return sessionIDOrName, nil
	}

	ctx := c.Context()
	session, err := h.sessionService.GetSession(ctx, sessionIDOrName)
	if err != nil {
		return "", err
	}

	return session.SessionID().Value(), nil
}

// SetWebhook godoc
// @Summary Set webhook URL
// @Description Sets or updates the webhook URL and events for a session
// @Tags Webhooks
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.RegisterWebhookRequest true "Webhook configuration"
// @Success 200 {object} dto.WebhookResponse "Webhook set successfully"
// @Failure 400 {object} dto.WebhookResponse "Invalid request data"
// @Failure 401 {object} dto.WebhookResponse "Unauthorized - Invalid API key" "Invalid request data"
// @Failure 404 {object} dto.WebhookResponse "Session not found"
// @Failure 500 {object} dto.WebhookResponse "Failed to set webhook"
// @Router /session/{sessionId}/webhook [post]
func (h *WebhookHandler) SetWebhook(c *fiber.Ctx) error {
	sessionIDOrName := c.Params("sessionId")

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON( dto.WebhookResponse{
			Success: false,
			Code:    fiber.StatusNotFound,
			Data:    &dto.WebhookResponseData{},
			Error: &dto.ErrorInfo{
				Code:    "SESSION_NOT_FOUND",
				Message: "Session not found: " + err.Error(),
			},
		})
	}

	var req dto.RegisterWebhookRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON( dto.WebhookResponse{
			Success: false,
			Code:    fiber.StatusBadRequest,
			Data:    &dto.WebhookResponseData{},
			Error: &dto.ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body: " + err.Error(),
			},
		})
	}

	validEvents := make([]string, 0)
	allValidEvents, err := h.webhookApp.ListEvents(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON( dto.WebhookResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "VALIDATION_ERROR",
				Message: "Failed to get valid events",
				Details: err.Error(),
			},
		})
	}

	for _, event := range req.Events {
		if h.isValidEvent(event, allValidEvents) {
			validEvents = append(validEvents, event)
		} else {
			return c.Status(fiber.StatusBadRequest).JSON( dto.WebhookResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "INVALID_EVENT",
					Message: fmt.Sprintf("Invalid event type: %s", event),
				},
			})
		}
	}

	if len(validEvents) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON( dto.WebhookResponse{
			Success: false,
			Code:    fiber.StatusBadRequest,
			Data:    &dto.WebhookResponseData{},
			Error: &dto.ErrorInfo{
				Code:    "INVALID_EVENTS",
				Message: fmt.Sprintf("No valid events provided. Valid events include: %s", strings.Join(allValidEvents, ", ")),
			},
		})
	}

	err = h.webhookApp.SetWebhook(c.Context(), sessionID, req.URL, validEvents)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON( dto.WebhookResponse{
			Success: false,
			Code:    fiber.StatusBadRequest,
			Data:    &dto.WebhookResponseData{},
			Error: &dto.ErrorInfo{
				Code:    "VALIDATION_FAILED",
				Message: "Failed to register webhook: " + err.Error(),
			},
		})
	}

	err = h.wmeowService.UpdateSessionSubscriptions(sessionID, validEvents)
	if err != nil {
		h.logger.Warnf("Failed to update session subscriptions for %s: %v", sessionID, err)
	}

	return c.Status(fiber.StatusCreated).JSON( dto.StandardWebhookCreateResponse{
		Success: true,
		Code:    fiber.StatusCreated,
		Data: &dto.StandardWebhookData{
			CreatedAt: time.Now(),
			Events:    validEvents,
			SessionID: sessionIDOrName,
			Status:    "active",
			URL:       req.URL,
		},
	})
}

// GetWebhook godoc
// @Summary Get webhook configuration
// @Description Retrieves the current webhook URL and events for a session
// @Tags Webhooks
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Success 200 {object} dto.WebhookResponse "Webhook configuration"
// @Failure 404 {object} dto.WebhookResponse "Session not found"
// @Failure 500 {object} dto.WebhookResponse "Failed to get webhook"
// @Router /session/{sessionId}/webhook [get]
func (h *WebhookHandler) GetWebhook(c *fiber.Ctx) error {
	sessionIDOrName := c.Params("sessionId")

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON( dto.WebhookResponse{
			Success: false,
			Code:    fiber.StatusNotFound,
			Data:    &dto.WebhookResponseData{},
			Error: &dto.ErrorInfo{
				Code:    "SESSION_NOT_FOUND",
				Message: "Session not found: " + err.Error(),
			},
		})
	}

	webhookURL, events, err := h.webhookApp.GetWebhook(c.Context(), sessionID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON( dto.WebhookResponse{
			Success: false,
			Code:    fiber.StatusInternalServerError,
			Data:    &dto.WebhookResponseData{},
			Error: &dto.ErrorInfo{
				Code:    "GET_FAILED",
				Message: "Failed to get webhook: " + err.Error(),
			},
		})
	}

	if webhookURL == "" {
		return c.Status(fiber.StatusNotFound).JSON( dto.WebhookResponse{
			Success: false,
			Code:    fiber.StatusNotFound,
			Data:    &dto.WebhookResponseData{},
			Error: &dto.ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "No webhook configured for this session",
			},
		})
	}

	return c.Status(fiber.StatusOK).JSON( dto.StandardWebhookResponse{
		Success: true,
		Code:    fiber.StatusOK,
		Data: &dto.StandardWebhookData{
			CreatedAt: time.Now(),
			Events:    events,
			SessionID: sessionIDOrName,
			Status:    "active",
			URL:       webhookURL,
		},
	})
}

// ListEvents godoc
// @Summary List supported webhook events
// @Description Retrieves a list of all supported webhook events
// @Tags Webhooks
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Success 200 {object} dto.SupportedEventsResponse "Supported events list"
// @Failure 500 {object} dto.SupportedEventsResponse "Failed to get events"
// @Router /session/{sessionId}/webhooks/events [get]
func (h *WebhookHandler) ListEvents(c *fiber.Ctx) error {
	events, err := h.webhookApp.ListEvents(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON( fiber.Map{
			"error":   "Failed to get supported events",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON( dto.SupportedEventsResponse{
		Success: true,
		Code:    fiber.StatusOK,
		Status:  fiber.StatusOK,
		Message: "Supported events retrieved successfully",
		Data: &dto.SupportedEventsData{
			Events: events,
			Count:  len(events),
		},
	})
}

func (h *WebhookHandler) isValidEvent(event string, validEvents []string) bool {
	for _, validEvent := range validEvents {
		if event == validEvent {
			return true
		}
	}
	return false
}
