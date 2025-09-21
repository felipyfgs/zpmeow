package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"zpmeow/internal/application"
	"zpmeow/internal/infra/http/dto"
	"zpmeow/internal/infra/wmeow"
)

type PrivacyHandler struct {
	sessionService *application.SessionApp
	wmeowService   wmeow.WameowService
}

func NewPrivacyHandler(sessionService *application.SessionApp, wmeowService wmeow.WameowService) *PrivacyHandler {
	return &PrivacyHandler{
		sessionService: sessionService,
		wmeowService:   wmeowService,
	}
}

// @Summary		Set multiple privacy settings
// @Description	Set multiple privacy settings in a single request
// @Tags			Privacy
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string								true	"Session ID"
// @Param			request		body		dto.SetAllPrivacySettingsRequest	true	"Privacy settings request"
// @Success		200			{object}	dto.PrivacySettingsResponse
// @Failure		400			{object}	dto.PrivacySettingsResponse
// @Failure		404			{object}	dto.PrivacySettingsResponse
// @Failure		500			{object}	dto.PrivacySettingsResponse
// @Router			/session/{sessionId}/privacy/set [put]
func (h *PrivacyHandler) SetAllPrivacySettings(c *gin.Context) {
	sessionID := c.Param("sessionId")

	var req dto.SetAllPrivacySettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.PrivacySettingsResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request format",
				Details: err.Error(),
			},
		})
		return
	}

	// Validate the request
	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, dto.PrivacySettingsResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "VALIDATION_ERROR",
				Message: "Invalid privacy settings",
				Details: err.Error(),
			},
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	currentSettings, err := h.wmeowService.GetPrivacySettings(ctx, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.PrivacySettingsResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "PRIVACY_FETCH_ERROR",
				Message: "Failed to fetch current privacy settings",
				Details: err.Error(),
			},
		})
		return
	}

	if req.GroupsAddMe != "" {
		err := h.wmeowService.SetPrivacySetting(ctx, sessionID, "groupadd", req.GroupsAddMe)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.PrivacySettingsResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "PRIVACY_UPDATE_ERROR",
					Message: "Failed to update group add privacy",
					Details: err.Error(),
				},
			})
			return
		}
		currentSettings.GroupsAddMe = req.GroupsAddMe
	}

	if req.LastSeen != "" {
		err := h.wmeowService.SetPrivacySetting(ctx, sessionID, "last", req.LastSeen)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.PrivacySettingsResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "PRIVACY_UPDATE_ERROR",
					Message: "Failed to update last seen privacy",
					Details: err.Error(),
				},
			})
			return
		}
		currentSettings.LastSeen = req.LastSeen
	}

	if req.Status != "" {
		err := h.wmeowService.SetPrivacySetting(ctx, sessionID, "status", req.Status)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.PrivacySettingsResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "PRIVACY_UPDATE_ERROR",
					Message: "Failed to update status privacy",
					Details: err.Error(),
				},
			})
			return
		}
		currentSettings.Status = req.Status
	}

	if req.ProfilePhoto != "" {
		err := h.wmeowService.SetPrivacySetting(ctx, sessionID, "profile", req.ProfilePhoto)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.PrivacySettingsResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "PRIVACY_UPDATE_ERROR",
					Message: "Failed to update profile privacy",
					Details: err.Error(),
				},
			})
			return
		}
		currentSettings.ProfilePhoto = req.ProfilePhoto
	}

	// ReadReceipts is a boolean, so we always set it
	value := "none"
	if req.ReadReceipts {
		value = "all"
	}
	err = h.wmeowService.SetPrivacySetting(ctx, sessionID, "readreceipts", value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.PrivacySettingsResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "PRIVACY_UPDATE_ERROR",
				Message: "Failed to update read receipts privacy",
				Details: err.Error(),
			},
		})
		return
	}
	currentSettings.ReadReceipts = value == "all"

	if req.CallsAddMe != "" {
		err := h.wmeowService.SetPrivacySetting(ctx, sessionID, "calladd", req.CallsAddMe)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.PrivacySettingsResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "PRIVACY_UPDATE_ERROR",
					Message: "Failed to update call privacy",
					Details: err.Error(),
				},
			})
			return
		}
		currentSettings.CallsAddMe = req.CallsAddMe
	}

	// Online privacy setting is not supported in this implementation

	data := &dto.PrivacySettingsData{
		SessionID:         sessionID,
		LastSeen:          currentSettings.LastSeen,
		ProfilePhoto:      currentSettings.ProfilePhoto,
		Status:            currentSettings.Status,
		ReadReceipts:      currentSettings.ReadReceipts,
		GroupsAddMe:       currentSettings.GroupsAddMe,
		CallsAddMe:        currentSettings.CallsAddMe,
		DisappearingChats: currentSettings.DisappearingMessages == "on",
		UpdatedAt:         time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, dto.PrivacySettingsResponse{
		Success: true,
		Data:    data,
	})
}

// @Summary		Get blocked contacts list
// @Description	Get the list of blocked contacts (blocklist)
// @Tags			Privacy
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string	true	"Session ID"
// @Success		200			{object}	dto.BlocklistResponse
// @Failure		400			{object}	dto.BlocklistResponse
// @Failure		404			{object}	dto.BlocklistResponse
// @Failure		500			{object}	dto.BlocklistResponse
// @Router			/session/{sessionId}/privacy/blocklist [get]
func (h *PrivacyHandler) GetBlocklist(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if err := validateSessionID(sessionID); err != nil {
		c.JSON(http.StatusBadRequest, dto.BlocklistResponse{
			Success: false,
			Error:   createValidationError("session ID", err.Error()),
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	blocklist, err := h.wmeowService.GetBlocklist(ctx, sessionID)
	if err != nil {
		if strings.Contains(err.Error(), "client not found") {
			c.JSON(http.StatusNotFound, dto.BlocklistResponse{
				Success: false,
				Error:   createNotFoundError("Session"),
			})
		} else if strings.Contains(err.Error(), "not connected") {
			c.JSON(http.StatusServiceUnavailable, dto.BlocklistResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "SESSION_NOT_CONNECTED",
					Message: "Session is not connected",
					Details: "Please ensure the meow session is connected before performing this operation",
				},
			})
		} else {
			c.JSON(http.StatusInternalServerError, dto.BlocklistResponse{
				Success: false,
				Error:   createInternalError("fetch blocklist", err.Error()),
			})
		}
		return
	}

	jids := blocklist

	data := &dto.BlockedContactsData{
		SessionID:       sessionID,
		BlockedContacts: jids,
		Count:           len(jids),
	}

	c.JSON(http.StatusOK, dto.BlocklistResponse{
		Success: true,
		Data:    data,
	})
}

// @Summary		Update blocklist (block/unblock contact)
// @Description	Block or unblock a contact
// @Tags			Privacy
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string						true	"Session ID"
// @Param			request		body		dto.UpdateBlocklistRequest	true	"Update blocklist request"
// @Success		200			{object}	dto.BlocklistResponse
// @Failure		400			{object}	dto.BlocklistResponse
// @Failure		404			{object}	dto.BlocklistResponse
// @Failure		500			{object}	dto.BlocklistResponse
// @Router			/session/{sessionId}/privacy/blocklist [put]
func (h *PrivacyHandler) UpdateBlocklist(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if err := validateSessionID(sessionID); err != nil {
		c.JSON(http.StatusBadRequest, dto.BlocklistResponse{
			Success: false,
			Error:   createValidationError("session ID", err.Error()),
		})
		return
	}

	var req dto.UpdateBlocklistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.BlocklistResponse{
			Success: false,
			Error:   createValidationError("request body", err.Error()),
		})
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, dto.BlocklistResponse{
			Success: false,
			Error:   createValidationError("request", err.Error()),
		})
		return
	}

	if err := validateBlocklistAction(req.Action); err != nil {
		c.JSON(http.StatusBadRequest, dto.BlocklistResponse{
			Success: false,
			Error:   createValidationError("action", err.Error()),
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	err := h.wmeowService.UpdateBlocklist(ctx, sessionID, req.Action, req.Contacts)
	if err != nil {
		if strings.Contains(err.Error(), "client not found") {
			c.JSON(http.StatusNotFound, dto.BlocklistResponse{
				Success: false,
				Error:   createNotFoundError("Session"),
			})
		} else if strings.Contains(err.Error(), "not connected") {
			c.JSON(http.StatusServiceUnavailable, dto.BlocklistResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "SESSION_NOT_CONNECTED",
					Message: "Session is not connected",
					Details: "Please ensure the meow session is connected before performing this operation",
				},
			})
		} else {
			c.JSON(http.StatusInternalServerError, dto.BlocklistResponse{
				Success: false,
				Error:   createInternalError(fmt.Sprintf("%s contact", req.Action), err.Error()),
			})
		}
		return
	}

	data := &dto.BlockedContactsData{
		SessionID:       sessionID,
		BlockedContacts: req.Contacts,
		Count:           len(req.Contacts),
	}

	c.JSON(http.StatusOK, dto.BlocklistResponse{
		Success: true,
		Data:    data,
	})
}

func validateSessionID(sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}

	if len(sessionID) != 36 {
		return fmt.Errorf("session ID must be a valid UUID")
	}

	return nil
}

func validateBlocklistAction(action string) error {
	if action == "" {
		return fmt.Errorf("action cannot be empty")
	}

	if action != "block" && action != "unblock" {
		return fmt.Errorf("action must be 'block' or 'unblock', got '%s'", action)
	}

	return nil
}

func createValidationError(field, message string) *dto.ErrorInfo {
	return &dto.ErrorInfo{
		Code:    "VALIDATION_ERROR",
		Message: fmt.Sprintf("Invalid %s", field),
		Details: message,
	}
}

func createInternalError(operation, details string) *dto.ErrorInfo {
	return &dto.ErrorInfo{
		Code:    "INTERNAL_ERROR",
		Message: fmt.Sprintf("Failed to %s", operation),
		Details: details,
	}
}

func createNotFoundError(resource string) *dto.ErrorInfo {
	return &dto.ErrorInfo{
		Code:    "NOT_FOUND",
		Message: fmt.Sprintf("%s not found", resource),
		Details: fmt.Sprintf("The requested %s could not be found", strings.ToLower(resource)),
	}
}

// @Summary		Find specific privacy settings
// @Description	Get specific privacy settings or all settings if none specified
// @Tags			Privacy
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string							true	"Session ID"
// @Param			request		body		dto.FindPrivacySettingsRequest	false	"Find privacy settings request (optional)"
// @Success		200			{object}	dto.PrivacySettingsResponse
// @Failure		400			{object}	dto.PrivacySettingsResponse
// @Failure		404			{object}	dto.PrivacySettingsResponse
// @Failure		500			{object}	dto.PrivacySettingsResponse
// @Router			/session/{sessionId}/privacy/find [post]
func (h *PrivacyHandler) FindPrivacySettings(c *gin.Context) {
	sessionID := c.Param("sessionId")

	var req dto.FindPrivacySettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Settings = []string{}
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	allSettings, err := h.wmeowService.GetPrivacySettings(ctx, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.PrivacySettingsResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "PRIVACY_FETCH_ERROR",
				Message: "Failed to fetch privacy settings",
				Details: err.Error(),
			},
		})
		return
	}

	if len(req.Settings) == 0 {
		data := &dto.PrivacySettingsData{
			SessionID:         sessionID,
			LastSeen:          allSettings.LastSeen,
			ProfilePhoto:      allSettings.ProfilePhoto,
			Status:            allSettings.Status,
			ReadReceipts:      allSettings.ReadReceipts,
			GroupsAddMe:       allSettings.GroupsAddMe,
			CallsAddMe:        allSettings.CallsAddMe,
			DisappearingChats: allSettings.DisappearingMessages == "on",
			UpdatedAt:         time.Now().Format(time.RFC3339),
		}

		c.JSON(http.StatusOK, dto.PrivacySettingsResponse{
			Success: true,
			Data:    data,
		})
		return
	}

	filteredData := &dto.PrivacySettingsData{
		SessionID: sessionID,
		UpdatedAt: time.Now().Format(time.RFC3339),
	}

	for _, setting := range req.Settings {
		switch setting {
		case "groups_add_me":
			filteredData.GroupsAddMe = allSettings.GroupsAddMe
		case "last_seen":
			filteredData.LastSeen = allSettings.LastSeen
		case "status":
			filteredData.Status = allSettings.Status
		case "profile_photo":
			filteredData.ProfilePhoto = allSettings.ProfilePhoto
		case "read_receipts":
			filteredData.ReadReceipts = allSettings.ReadReceipts
		case "calls_add_me":
			filteredData.CallsAddMe = allSettings.CallsAddMe
		}
	}

	c.JSON(http.StatusOK, dto.PrivacySettingsResponse{
		Success: true,
		Data:    filteredData,
	})
}
