package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"zpmeow/internal/application"
	"zpmeow/internal/infra/http/dto"
	"zpmeow/internal/infra/wmeow"

	"github.com/gin-gonic/gin"
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

func (h *WebhookHandler) resolveSessionID(c *gin.Context, sessionIDOrName string) (string, error) {
	if h.sessionService == nil {
		return sessionIDOrName, nil
	}

	ctx := c.Request.Context()
	session, err := h.sessionService.GetSession(ctx, sessionIDOrName)
	if err != nil {
		return "", err
	}

	return session.SessionID().Value(), nil
}

func (h *WebhookHandler) SetWebhook(c *gin.Context) {
	sessionIDOrName := c.Param("sessionId")

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.WebhookResponse{
			Success: false,
			Code:    http.StatusNotFound,
			Data:    &dto.WebhookResponseData{},
			Error: &dto.ErrorInfo{
				Code:    "SESSION_NOT_FOUND",
				Message: "Session not found: " + err.Error(),
			},
		})
		return
	}

	var req dto.RegisterWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.WebhookResponse{
			Success: false,
			Code:    http.StatusBadRequest,
			Data:    &dto.WebhookResponseData{},
			Error: &dto.ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body: " + err.Error(),
			},
		})
		return
	}

	validEvents := make([]string, 0)
	allValidEvents, err := h.webhookApp.ListEvents(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.WebhookResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "VALIDATION_ERROR",
				Message: "Failed to get valid events",
				Details: err.Error(),
			},
		})
		return
	}

	for _, event := range req.Events {
		if h.isValidEvent(event, allValidEvents) {
			validEvents = append(validEvents, event)
		} else {
			c.JSON(http.StatusBadRequest, dto.WebhookResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "INVALID_EVENT",
					Message: fmt.Sprintf("Invalid event type: %s", event),
				},
			})
			return
		}
	}

	if len(validEvents) == 0 {
		c.JSON(http.StatusBadRequest, dto.WebhookResponse{
			Success: false,
			Code:    http.StatusBadRequest,
			Data:    &dto.WebhookResponseData{},
			Error: &dto.ErrorInfo{
				Code:    "INVALID_EVENTS",
				Message: fmt.Sprintf("No valid events provided. Valid events include: %s", strings.Join(allValidEvents, ", ")),
			},
		})
		return
	}

	err = h.webhookApp.SetWebhook(c.Request.Context(), sessionID, req.URL, validEvents)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.WebhookResponse{
			Success: false,
			Code:    http.StatusBadRequest,
			Data:    &dto.WebhookResponseData{},
			Error: &dto.ErrorInfo{
				Code:    "VALIDATION_FAILED",
				Message: "Failed to register webhook: " + err.Error(),
			},
		})
		return
	}

	err = h.wmeowService.UpdateSessionSubscriptions(sessionID, validEvents)
	if err != nil {
		h.logger.Warnf("Failed to update session subscriptions for %s: %v", sessionID, err)
	}

	c.JSON(http.StatusCreated, dto.StandardWebhookCreateResponse{
		Success: true,
		Code:    http.StatusCreated,
		Data: &dto.StandardWebhookData{
			CreatedAt: time.Now(),
			Events:    validEvents,
			SessionID: sessionIDOrName,
			Status:    "active",
			URL:       req.URL,
		},
	})
}

func (h *WebhookHandler) GetWebhook(c *gin.Context) {
	sessionIDOrName := c.Param("sessionId")

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.WebhookResponse{
			Success: false,
			Code:    http.StatusNotFound,
			Data:    &dto.WebhookResponseData{},
			Error: &dto.ErrorInfo{
				Code:    "SESSION_NOT_FOUND",
				Message: "Session not found: " + err.Error(),
			},
		})
		return
	}

	webhookURL, events, err := h.webhookApp.GetWebhook(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.WebhookResponse{
			Success: false,
			Code:    http.StatusInternalServerError,
			Data:    &dto.WebhookResponseData{},
			Error: &dto.ErrorInfo{
				Code:    "GET_FAILED",
				Message: "Failed to get webhook: " + err.Error(),
			},
		})
		return
	}

	if webhookURL == "" {
		c.JSON(http.StatusNotFound, dto.WebhookResponse{
			Success: false,
			Code:    http.StatusNotFound,
			Data:    &dto.WebhookResponseData{},
			Error: &dto.ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "No webhook configured for this session",
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.StandardWebhookResponse{
		Success: true,
		Code:    http.StatusOK,
		Data: &dto.StandardWebhookData{
			CreatedAt: time.Now(),
			Events:    events,
			SessionID: sessionIDOrName,
			Status:    "active",
			URL:       webhookURL,
		},
	})
}

func (h *WebhookHandler) ListEvents(c *gin.Context) {
	events, err := h.webhookApp.ListEvents(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get supported events",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.SupportedEventsResponse{
		Success: true,
		Code:    http.StatusOK,
		Status:  http.StatusOK,
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
