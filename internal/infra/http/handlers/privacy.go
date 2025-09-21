package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

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

// SetAllPrivacySettings godoc
// @Summary Set all privacy settings
// @Description Sets all privacy settings for a WhatsApp session
// @Tags Privacy
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.SetAllPrivacySettingsRequest true "Privacy settings request"
// @Success 200 {object} dto.PrivacySettingsResponse "Privacy settings updated"
// @Failure 400 {object} dto.PrivacySettingsResponse "Invalid request data"
// @Failure 401 {object} dto.PrivacySettingsResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.PrivacySettingsResponse "Session not found"
// @Failure 500 {object} dto.PrivacySettingsResponse "Failed to set privacy settings"
// @Router /session/{sessionId}/privacy/set [put]
func (h *PrivacyHandler) SetAllPrivacySettings(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	var req dto.SetAllPrivacySettingsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.PrivacySettingsResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request format",
				Details: err.Error(),
			},
		})
	}

	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.PrivacySettingsResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "VALIDATION_ERROR",
				Message: "Invalid privacy settings",
				Details: err.Error(),
			},
		})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 60*time.Second)
	defer cancel()

	currentSettings, err := h.wmeowService.GetPrivacySettings(ctx, sessionID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.PrivacySettingsResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "PRIVACY_FETCH_ERROR",
				Message: "Failed to fetch current privacy settings",
				Details: err.Error(),
			},
		})
	}

	if req.GroupsAddMe != "" {
		err := h.wmeowService.SetPrivacySetting(ctx, sessionID, "groupadd", req.GroupsAddMe)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(dto.PrivacySettingsResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "PRIVACY_UPDATE_ERROR",
					Message: "Failed to update group add privacy",
					Details: err.Error(),
				},
			})
		}
		currentSettings.GroupsAddMe = req.GroupsAddMe
	}

	if req.LastSeen != "" {
		err := h.wmeowService.SetPrivacySetting(ctx, sessionID, "last", req.LastSeen)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(dto.PrivacySettingsResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "PRIVACY_UPDATE_ERROR",
					Message: "Failed to update last seen privacy",
					Details: err.Error(),
				},
			})
		}
		currentSettings.LastSeen = req.LastSeen
	}

	if req.Status != "" {
		err := h.wmeowService.SetPrivacySetting(ctx, sessionID, "status", req.Status)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(dto.PrivacySettingsResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "PRIVACY_UPDATE_ERROR",
					Message: "Failed to update status privacy",
					Details: err.Error(),
				},
			})
		}
		currentSettings.Status = req.Status
	}

	if req.ProfilePhoto != "" {
		err := h.wmeowService.SetPrivacySetting(ctx, sessionID, "profile", req.ProfilePhoto)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(dto.PrivacySettingsResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "PRIVACY_UPDATE_ERROR",
					Message: "Failed to update profile privacy",
					Details: err.Error(),
				},
			})
		}
		currentSettings.ProfilePhoto = req.ProfilePhoto
	}

	value := "none"
	if req.ReadReceipts {
		value = "all"
	}
	err = h.wmeowService.SetPrivacySetting(ctx, sessionID, "readreceipts", value)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.PrivacySettingsResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "PRIVACY_UPDATE_ERROR",
				Message: "Failed to update read receipts privacy",
				Details: err.Error(),
			},
		})
	}
	currentSettings.ReadReceipts = value == "all"

	if req.CallsAddMe != "" {
		err := h.wmeowService.SetPrivacySetting(ctx, sessionID, "calladd", req.CallsAddMe)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(dto.PrivacySettingsResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "PRIVACY_UPDATE_ERROR",
					Message: "Failed to update call privacy",
					Details: err.Error(),
				},
			})
		}
		currentSettings.CallsAddMe = req.CallsAddMe
	}

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

	return c.Status(fiber.StatusOK).JSON(dto.PrivacySettingsResponse{
		Success: true,
		Data:    data,
	})
}

// GetBlocklist godoc
// @Summary Get blocked contacts
// @Description Retrieves the list of blocked contacts for a WhatsApp session
// @Tags Privacy
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Success 200 {object} dto.BlocklistResponse "Blocked contacts list"
// @Failure 400 {object} dto.BlocklistResponse "Invalid request data"
// @Failure 401 {object} dto.BlocklistResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.BlocklistResponse "Session not found"
// @Failure 500 {object} dto.BlocklistResponse "Failed to get blocklist"
// @Router /session/{sessionId}/privacy/blocklist [get]
func (h *PrivacyHandler) GetBlocklist(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")
	if err := validateSessionID(sessionID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.BlocklistResponse{
			Success: false,
			Error:   createValidationError("session ID", err.Error()),
		})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()

	blocklist, err := h.wmeowService.GetBlocklist(ctx, sessionID)
	if err != nil {
		if strings.Contains(err.Error(), "client not found") {
			return c.Status(fiber.StatusNotFound).JSON(dto.BlocklistResponse{
				Success: false,
				Error:   createNotFoundError("Session"),
			})
		} else if strings.Contains(err.Error(), "not connected") {
			return c.Status(fiber.StatusServiceUnavailable).JSON(dto.BlocklistResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "SESSION_NOT_CONNECTED",
					Message: "Session is not connected",
					Details: "Please ensure the meow session is connected before performing this operation",
				},
			})
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(dto.BlocklistResponse{
				Success: false,
				Error:   createInternalError("fetch blocklist", err.Error()),
			})
		}
	}

	jids := blocklist

	data := &dto.BlockedContactsData{
		SessionID:       sessionID,
		BlockedContacts: jids,
		Count:           len(jids),
	}

	return c.Status(fiber.StatusOK).JSON(dto.BlocklistResponse{
		Success: true,
		Data:    data,
	})
}

// UpdateBlocklist godoc
// @Summary Update blocked contacts
// @Description Blocks or unblocks contacts for a WhatsApp session
// @Tags Privacy
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.UpdateBlocklistRequest true "Update blocklist request"
// @Success 200 {object} dto.BlocklistResponse "Blocklist updated successfully"
// @Failure 400 {object} dto.BlocklistResponse "Invalid request data"
// @Failure 401 {object} dto.BlocklistResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.BlocklistResponse "Session not found"
// @Failure 500 {object} dto.BlocklistResponse "Failed to update blocklist"
// @Router /session/{sessionId}/privacy/blocklist [put]
func (h *PrivacyHandler) UpdateBlocklist(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")
	if err := validateSessionID(sessionID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.BlocklistResponse{
			Success: false,
			Error:   createValidationError("session ID", err.Error()),
		})
	}

	var req dto.UpdateBlocklistRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.BlocklistResponse{
			Success: false,
			Error:   createValidationError("request body", err.Error()),
		})
	}

	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.BlocklistResponse{
			Success: false,
			Error:   createValidationError("request", err.Error()),
		})
	}

	if err := validateBlocklistAction(req.Action); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.BlocklistResponse{
			Success: false,
			Error:   createValidationError("action", err.Error()),
		})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()

	err := h.wmeowService.UpdateBlocklist(ctx, sessionID, req.Action, req.Contacts)
	if err != nil {
		if strings.Contains(err.Error(), "client not found") {
			return c.Status(fiber.StatusNotFound).JSON(dto.BlocklistResponse{
				Success: false,
				Error:   createNotFoundError("Session"),
			})
		} else if strings.Contains(err.Error(), "not connected") {
			return c.Status(fiber.StatusServiceUnavailable).JSON(dto.BlocklistResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "SESSION_NOT_CONNECTED",
					Message: "Session is not connected",
					Details: "Please ensure the meow session is connected before performing this operation",
				},
			})
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(dto.BlocklistResponse{
				Success: false,
				Error:   createInternalError(fmt.Sprintf("%s contact", req.Action), err.Error()),
			})
		}
	}

	data := &dto.BlockedContactsData{
		SessionID:       sessionID,
		BlockedContacts: req.Contacts,
		Count:           len(req.Contacts),
	}

	return c.Status(fiber.StatusOK).JSON(dto.BlocklistResponse{
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

// FindPrivacySettings godoc
// @Summary Find privacy settings
// @Description Retrieves current privacy settings for a WhatsApp session
// @Tags Privacy
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.FindPrivacySettingsRequest false "Privacy settings query"
// @Success 200 {object} dto.PrivacySettingsResponse "Privacy settings"
// @Failure 400 {object} dto.PrivacySettingsResponse "Invalid request data"
// @Failure 401 {object} dto.PrivacySettingsResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.PrivacySettingsResponse "Session not found"
// @Failure 500 {object} dto.PrivacySettingsResponse "Failed to get privacy settings"
// @Router /session/{sessionId}/privacy/find [post]
func (h *PrivacyHandler) FindPrivacySettings(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	var req dto.FindPrivacySettingsRequest
	if err := c.BodyParser(&req); err != nil {
		req.Settings = []string{}
	}

	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()

	allSettings, err := h.wmeowService.GetPrivacySettings(ctx, sessionID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.PrivacySettingsResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "PRIVACY_FETCH_ERROR",
				Message: "Failed to fetch privacy settings",
				Details: err.Error(),
			},
		})
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

		return c.Status(fiber.StatusOK).JSON(dto.PrivacySettingsResponse{
			Success: true,
			Data:    data,
		})
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

	return c.Status(fiber.StatusOK).JSON(dto.PrivacySettingsResponse{
		Success: true,
		Data:    filteredData,
	})
}
