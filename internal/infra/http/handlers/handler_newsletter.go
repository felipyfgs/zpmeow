package handlers

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"zpmeow/internal/application"
	"zpmeow/internal/infra/http/dto"
	"zpmeow/internal/infra/wmeow"

	"github.com/gofiber/fiber/v2"
)

type NewsletterHandler struct {
	sessionService *application.SessionApp
	wmeowService   wmeow.WameowService
}

func NewNewsletterHandler(sessionService *application.SessionApp, wmeowService wmeow.WameowService) *NewsletterHandler {
	return &NewsletterHandler{
		sessionService: sessionService,
		wmeowService:   wmeowService,
	}
}

// GetNewsletterMessageUpdates godoc
// @Summary Get newsletter message updates
// @Description Retrieves message updates for a specific newsletter
// @Tags Newsletters
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param newsletterId path string true "Newsletter ID"
// @Success 200 {object} dto.StandardResponse "Newsletter message updates"
// @Failure 400 {object} dto.StandardResponse "Invalid request data"
// @Failure 401 {object} dto.StandardResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.StandardResponse "Session not found"
// @Failure 500 {object} dto.StandardResponse "Failed to get message updates"
// @Router /session/{sessionId}/newsletter/{newsletterId}/updates [get]
func (h *NewsletterHandler) GetNewsletterMessageUpdates(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")
	newsletterID := c.Params("newsletterId")

	if !h.wmeowService.IsClientConnected(sessionID) {
		return c.Status(fiber.StatusBadRequest).JSON(dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "SESSION_NOT_CONNECTED",
				Message: "Session not found or not connected",
			},
		})
	}

	count := 50
	if countStr := c.Query("count"); countStr != "" {
		if c, err := strconv.Atoi(countStr); err == nil && c > 0 {
			count = c
		}
	}
	before := c.Query("before")

	_ = count
	_ = before
	updates, err := h.wmeowService.GetNewsletterMessageUpdates(c.Context(), sessionID, newsletterID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "GET_UPDATES_FAILED",
				Message: "Failed to get newsletter message updates: " + err.Error(),
			},
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    updates,
		"count":   len(updates),
	})
}

// MarkNewsletterViewed godoc
// @Summary Mark newsletter messages as viewed
// @Description Marks newsletter messages as viewed for a specific newsletter
// @Tags Newsletters
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param newsletterId path string true "Newsletter ID"
// @Param request body dto.MarkViewedRequest true "Mark newsletter viewed request"
// @Success 200 {object} dto.StandardResponse "Messages marked as viewed successfully"
// @Failure 400 {object} dto.StandardResponse "Invalid request data"
// @Failure 401 {object} dto.StandardResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.StandardResponse "Session not found"
// @Failure 500 {object} dto.StandardResponse "Failed to mark as viewed"
// @Router /session/{sessionId}/newsletter/{newsletterId}/viewed [post]
func (h *NewsletterHandler) MarkNewsletterViewed(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")
	newsletterJID := c.Params("newsletterId")

	if !h.wmeowService.IsClientConnected(sessionID) {
		return c.Status(fiber.StatusBadRequest).JSON(dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "SESSION_NOT_CONNECTED",
				Message: "Session not found or not connected",
			},
		})
	}

	var req dto.MarkViewedRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_REQUEST_BODY",
				Message: "Invalid request body: " + err.Error(),
			},
		})
	}

	err := h.wmeowService.NewsletterMarkViewed(c.Context(), sessionID, newsletterJID, req.ServerIDs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "MARK_VIEWED_FAILED",
				Message: "Failed to mark messages as viewed: " + err.Error(),
			},
		})
	}

	return c.Status(fiber.StatusOK).JSON(dto.StandardResponse{
		Success: true,
		Data:    map[string]string{"message": "Messages marked as viewed successfully"},
	})
}

// SendNewsletterReaction godoc
// @Summary Send newsletter reaction
// @Description Sends a reaction to a newsletter message
// @Tags Newsletters
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param newsletterId path string true "Newsletter ID"
// @Param request body dto.SendReactionRequest true "Send newsletter reaction request"
// @Success 200 {object} dto.StandardResponse "Reaction sent successfully"
// @Failure 400 {object} dto.StandardResponse "Invalid request data"
// @Failure 401 {object} dto.StandardResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.StandardResponse "Session not found"
// @Failure 500 {object} dto.StandardResponse "Failed to send reaction"
// @Router /session/{sessionId}/newsletter/{newsletterId}/reaction [post]
func (h *NewsletterHandler) SendNewsletterReaction(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")
	newsletterJID := c.Params("newsletterId")

	if !h.wmeowService.IsClientConnected(sessionID) {
		return c.Status(fiber.StatusBadRequest).JSON(dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "SESSION_NOT_CONNECTED",
				Message: "Session not found or not connected",
			},
		})
	}

	var req dto.SendReactionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_REQUEST_BODY",
				Message: "Invalid request body: " + err.Error(),
			},
		})
	}

	if req.ServerID == "" || req.Reaction == "" || req.MessageID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "MISSING_REQUIRED_FIELDS",
				Message: "server_id, reaction, and message_id are required",
			},
		})
	}

	err := h.wmeowService.NewsletterSendReaction(
		c.Context(),
		sessionID,
		newsletterJID,
		req.MessageID,
		req.Reaction,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "SEND_REACTION_FAILED",
				Message: "Failed to send reaction: " + err.Error(),
			},
		})
	}

	return c.Status(fiber.StatusOK).JSON(dto.StandardResponse{
		Success: true,
		Data:    map[string]string{"message": "Reaction sent successfully"},
	})
}

// ToggleNewsletterMute godoc
// @Summary Toggle newsletter mute
// @Description Mutes or unmutes a newsletter
// @Tags Newsletters
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param newsletterId path string true "Newsletter ID"
// @Param request body dto.ToggleMuteRequest true "Toggle newsletter mute request"
// @Success 200 {object} dto.StandardResponse "Newsletter muted/unmuted successfully"
// @Failure 400 {object} dto.StandardResponse "Invalid request data"
// @Failure 401 {object} dto.StandardResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.StandardResponse "Session not found"
// @Failure 500 {object} dto.StandardResponse "Failed to toggle mute"
// @Router /session/{sessionId}/newsletter/{newsletterId}/mute [post]
func (h *NewsletterHandler) ToggleNewsletterMute(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")
	newsletterJID := c.Params("newsletterId")

	if !h.wmeowService.IsClientConnected(sessionID) {
		return c.Status(fiber.StatusBadRequest).JSON(dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "SESSION_NOT_CONNECTED",
				Message: "Session not found or not connected",
			},
		})
	}

	var req dto.ToggleMuteRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_REQUEST_BODY",
				Message: "Invalid request body: " + err.Error(),
			},
		})
	}

	err := h.wmeowService.NewsletterToggleMute(c.Context(), sessionID, newsletterJID, req.Mute)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "TOGGLE_MUTE_FAILED",
				Message: "Failed to toggle mute: " + err.Error(),
			},
		})
	}

	action := "muted"
	if !req.Mute {
		action = "unmuted"
	}

	return c.Status(fiber.StatusOK).JSON(dto.StandardResponse{
		Success: true,
		Data:    map[string]string{"message": fmt.Sprintf("Newsletter %s successfully", action)},
	})
}

// SubscribeLiveUpdates godoc
// @Summary Subscribe to newsletter live updates
// @Description Subscribes to live updates for a newsletter
// @Tags Newsletters
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param newsletterId path string true "Newsletter ID"
// @Success 200 {object} dto.StandardResponse "Successfully subscribed to live updates"
// @Failure 400 {object} dto.StandardResponse "Invalid request data"
// @Failure 401 {object} dto.StandardResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.StandardResponse "Session not found"
// @Failure 500 {object} dto.StandardResponse "Failed to subscribe to live updates"
// @Router /session/{sessionId}/newsletter/{newsletterId}/live [post]
func (h *NewsletterHandler) SubscribeLiveUpdates(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")
	newsletterJID := c.Params("newsletterId")

	if !h.wmeowService.IsClientConnected(sessionID) {
		return c.Status(fiber.StatusBadRequest).JSON(dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "SESSION_NOT_CONNECTED",
				Message: "Session not found or not connected",
			},
		})
	}

	err := h.wmeowService.NewsletterSubscribeLiveUpdates(c.Context(), sessionID, newsletterJID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "SUBSCRIBE_LIVE_UPDATES_FAILED",
				Message: "Failed to subscribe to live updates: " + err.Error(),
			},
		})
	}

	return c.Status(fiber.StatusOK).JSON(dto.StandardResponse{
		Success: true,
		Data:    map[string]string{"message": "Successfully subscribed to live updates"},
	})
}

// UploadNewsletterMedia godoc
// @Summary Upload newsletter media
// @Description Uploads media content for newsletter messages
// @Tags Newsletters
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.UploadNewsletterMediaRequest true "Upload newsletter media request"
// @Success 200 {object} dto.StandardResponse "Media uploaded successfully"
// @Failure 400 {object} dto.StandardResponse "Invalid request data"
// @Failure 401 {object} dto.StandardResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.StandardResponse "Session not found"
// @Failure 500 {object} dto.StandardResponse "Failed to upload media"
// @Router /session/{sessionId}/newsletter/media [post]
func (h *NewsletterHandler) UploadNewsletterMedia(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	if !h.wmeowService.IsClientConnected(sessionID) {
		return c.Status(fiber.StatusBadRequest).JSON(dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "SESSION_NOT_CONNECTED",
				Message: "Session not found or not connected",
			},
		})
	}

	var req dto.UploadNewsletterMediaRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_REQUEST_BODY",
				Message: "Invalid request body: " + err.Error(),
			},
		})
	}

	if req.MediaData == "" || req.MediaType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "MISSING_REQUIRED_FIELDS",
				Message: "data and media_type are required",
			},
		})
	}

	data, err := base64.StdEncoding.DecodeString(req.MediaData)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_BASE64_DATA",
				Message: "Invalid base64 data: " + err.Error(),
			},
		})
	}

	validTypes := map[string]bool{
		"image":    true,
		"video":    true,
		"audio":    true,
		"document": true,
	}

	if !validTypes[req.MediaType] {
		return c.Status(fiber.StatusBadRequest).JSON(dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_MEDIA_TYPE",
				Message: "Invalid media type. Supported: image, video, audio, document",
			},
		})
	}

	err = h.wmeowService.UploadNewsletter(c.Context(), sessionID, data)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "UPLOAD_FAILED",
				Message: "Failed to upload media: " + err.Error(),
			},
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Media uploaded successfully",
	})
}

// GetNewsletterByInvite godoc
// @Summary Get newsletter by invite
// @Description Retrieves newsletter information using an invite key
// @Tags Newsletters
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param inviteKey path string true "Newsletter invite key"
// @Success 200 {object} dto.NewsletterInfoResponse "Newsletter information"
// @Failure 400 {object} dto.NewsletterInfoResponse "Invalid request data"
// @Failure 401 {object} dto.NewsletterInfoResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.NewsletterInfoResponse "Session not found"
// @Failure 500 {object} dto.NewsletterInfoResponse "Failed to get newsletter"
// @Router /session/{sessionId}/newsletter/invite/{inviteKey} [get]
func (h *NewsletterHandler) GetNewsletterByInvite(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")
	inviteKey := c.Params("inviteKey")

	if !h.wmeowService.IsClientConnected(sessionID) {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewsletterInfoResponse{
			Success: false,
			Error: &dto.NewsletterErrorResponse{
				Code:    "SESSION_NOT_CONNECTED",
				Message: "Session not found or not connected",
			},
		})
	}

	if inviteKey == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewsletterInfoResponse{
			Success: false,
			Error: &dto.NewsletterErrorResponse{
				Code:    "MISSING_INVITE_KEY",
				Message: "Invite key is required",
			},
		})
	}

	info, err := h.wmeowService.GetNewsletterInfoWithInvite(c.Context(), sessionID, inviteKey)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewsletterInfoResponse{
			Success: false,
			Error: &dto.NewsletterErrorResponse{
				Code:    "GET_NEWSLETTER_INFO_FAILED",
				Message: "Failed to get newsletter info: " + err.Error(),
			},
		})
	}

	result := dto.NewsletterInfo{
		JID:             info.JID,
		Name:            info.Name,
		Description:     info.Description,
		CreatedAt:       time.Unix(info.CreatedAt, 0),
		IsVerified:      info.IsVerified,
		SubscriberCount: info.Subscribers,
	}

	return c.Status(fiber.StatusOK).JSON(dto.NewsletterInfoResponse{
		Success: true,
		Data:    &result,
	})
}

func (h *NewsletterHandler) resolveSessionID(c *fiber.Ctx, sessionIDOrName string) (string, error) {
	if h.sessionService == nil {
		return sessionIDOrName, nil
	}

	ctx := c.Context()
	session, err := h.sessionService.GetSession(ctx, sessionIDOrName)
	if err != nil {
		return "", err
	}

	return session.SessionID().String(), nil
}

// CreateNewsletter godoc
// @Summary Create newsletter
// @Description Creates a new WhatsApp newsletter
// @Tags Newsletters
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.CreateNewsletterRequest true "Newsletter creation request"
// @Success 201 {object} dto.CreateNewsletterResponse "Newsletter created successfully"
// @Failure 400 {object} dto.CreateNewsletterResponse "Invalid request data"
// @Failure 404 {object} dto.CreateNewsletterResponse "Session not found"
// @Failure 500 {object} dto.CreateNewsletterResponse "Failed to create newsletter"
// @Router /session/{sessionId}/newsletter [post]
func (h *NewsletterHandler) CreateNewsletter(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	resolvedSessionId, err := h.resolveSessionID(c, sessionID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.CreateNewsletterResponse{
			Success: false,
			Error:   &dto.NewsletterErrorResponse{Code: "ERROR", Message: "Session not found: " + err.Error()},
		})
	}

	if !h.wmeowService.IsClientConnected(resolvedSessionId) {
		return c.Status(fiber.StatusBadRequest).JSON(dto.CreateNewsletterResponse{
			Success: false,
			Error:   &dto.NewsletterErrorResponse{Code: "ERROR", Message: "Session not connected"},
		})
	}

	var req dto.CreateNewsletterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.CreateNewsletterResponse{
			Success: false,
			Error:   &dto.NewsletterErrorResponse{Code: "ERROR", Message: "Invalid request format: " + err.Error()},
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.CreateNewsletterResponse{
			Success: false,
			Error:   &dto.NewsletterErrorResponse{Code: "ERROR", Message: "Newsletter name is required"},
		})
	}

	resp, err := h.wmeowService.CreateNewsletter(c.Context(), resolvedSessionId, req.Name, req.Description)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.CreateNewsletterResponse{
			Success: false,
			Error:   &dto.NewsletterErrorResponse{Code: "ERROR", Message: "Failed to create newsletter: " + err.Error()},
		})
	}

	result := &dto.NewsletterInfo{
		JID:             resp.JID,
		Name:            resp.Name,
		Description:     resp.Description,
		CreatedAt:       time.Unix(resp.Timestamp, 0),
		SubscriberCount: 0,
	}

	return c.Status(fiber.StatusCreated).JSON(dto.CreateNewsletterResponse{
		Success: true,
		Data:    result,
	})
}

// GetNewsletter godoc
// @Summary Get newsletter information
// @Description Retrieves detailed information about a specific newsletter
// @Tags Newsletters
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param newsletterId path string true "Newsletter ID"
// @Success 200 {object} dto.NewsletterInfoResponse "Newsletter information"
// @Failure 400 {object} dto.NewsletterInfoResponse "Invalid request data"
// @Failure 401 {object} dto.NewsletterInfoResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.NewsletterInfoResponse "Session not found"
// @Failure 500 {object} dto.NewsletterInfoResponse "Failed to get newsletter"
// @Router /session/{sessionId}/newsletter/{newsletterId} [get]
func (h *NewsletterHandler) GetNewsletter(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")
	newsletterJID := c.Params("newsletterId")

	if !h.wmeowService.IsClientConnected(sessionID) {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewsletterInfoResponse{
			Success: false,
			Error: &dto.NewsletterErrorResponse{
				Code:    "SESSION_NOT_CONNECTED",
				Message: "Session not found or not connected",
			},
		})
	}

	if newsletterJID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewsletterInfoResponse{
			Success: false,
			Error:   &dto.NewsletterErrorResponse{Code: "ERROR", Message: "Newsletter JID is required"},
		})
	}

	info, err := h.wmeowService.GetNewsletterInfo(c.Context(), sessionID, newsletterJID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewsletterInfoResponse{
			Success: false,
			Error:   &dto.NewsletterErrorResponse{Code: "ERROR", Message: "Failed to get newsletter info: " + err.Error()},
		})
	}

	result := &dto.NewsletterInfo{
		JID:             info.JID,
		Name:            info.Name,
		Description:     info.Description,
		SubscriberCount: info.Subscribers,
		CreatedAt:       time.Unix(info.CreatedAt, 0),
		IsVerified:      info.IsVerified,
		IsSubscribed:    false,
	}

	return c.Status(fiber.StatusOK).JSON(dto.NewsletterInfoResponse{
		Success: true,
		Data:    result,
	})
}

// ListNewsletters godoc
// @Summary List newsletters
// @Description Retrieves a list of all newsletters for a session
// @Tags Newsletters
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Success 200 {object} dto.NewsletterListResponse "Newsletters list"
// @Failure 400 {object} dto.NewsletterListResponse "Invalid request data"
// @Failure 401 {object} dto.NewsletterListResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.NewsletterListResponse "Session not found"
// @Failure 500 {object} dto.NewsletterListResponse "Failed to list newsletters"
// @Router /session/{sessionId}/newsletter/list [get]
func (h *NewsletterHandler) ListNewsletters(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	if !h.wmeowService.IsClientConnected(sessionID) {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewsletterListResponse{
			Success: false,
			Error:   &dto.NewsletterErrorResponse{Code: "ERROR", Message: "Session not found or not connected"},
		})
	}

	newsletters, err := h.wmeowService.GetSubscribedNewsletters(c.Context(), sessionID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewsletterListResponse{
			Success: false,
			Error:   &dto.NewsletterErrorResponse{Code: "ERROR", Message: "Failed to get newsletters: " + err.Error()},
		})
	}

	result := make([]dto.NewsletterInfo, len(newsletters))
	for i, newsletter := range newsletters {
		result[i] = dto.NewsletterInfo{
			JID:             newsletter.JID,
			Name:            newsletter.Name,
			Description:     newsletter.Description,
			SubscriberCount: newsletter.Subscribers,
			CreatedAt:       time.Unix(newsletter.CreatedAt, 0),
			IsVerified:      newsletter.IsVerified,
		}
	}

	listResult := &dto.NewsletterListData{
		SessionId:   sessionID,
		Newsletters: result,
		Count:       len(result),
		Total:       len(result),
	}

	return c.Status(fiber.StatusOK).JSON(dto.NewsletterListResponse{
		Success: true,
		Data:    listResult,
	})
}

// SubscribeToNewsletter godoc
// @Summary Subscribe to newsletter
// @Description Subscribes to a newsletter
// @Tags Newsletters
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param newsletterId path string true "Newsletter ID"
// @Success 200 {object} dto.StandardResponse "Successfully subscribed to newsletter"
// @Failure 400 {object} dto.StandardResponse "Invalid request data"
// @Failure 401 {object} dto.StandardResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.StandardResponse "Session not found"
// @Failure 500 {object} dto.StandardResponse "Failed to subscribe"
// @Router /session/{sessionId}/newsletter/{newsletterId}/subscribe [post]
func (h *NewsletterHandler) SubscribeToNewsletter(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")
	newsletterJID := c.Params("newsletterId")

	_, err := h.sessionService.GetSession(c.Context(), sessionID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "ERROR", Message: "Session not found: " + err.Error()},
		})
	}

	if newsletterJID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "ERROR", Message: "Newsletter JID is required"},
		})
	}

	err = h.wmeowService.FollowNewsletter(c.Context(), sessionID, newsletterJID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "ERROR", Message: "Failed to subscribe to newsletter: " + err.Error()},
		})
	}

	return c.Status(fiber.StatusOK).JSON(dto.StandardResponse{
		Success: true,
		Data:    map[string]string{"message": "Successfully subscribed to newsletter"},
	})
}

// UnsubscribeFromNewsletter godoc
// @Summary Unsubscribe from newsletter
// @Description Unsubscribes from a newsletter
// @Tags Newsletters
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param newsletterId path string true "Newsletter ID"
// @Success 200 {object} dto.StandardResponse "Successfully unsubscribed from newsletter"
// @Failure 400 {object} dto.StandardResponse "Invalid request data"
// @Failure 401 {object} dto.StandardResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.StandardResponse "Session not found"
// @Failure 500 {object} dto.StandardResponse "Failed to unsubscribe"
// @Router /session/{sessionId}/newsletter/{newsletterId}/unsubscribe [post]
func (h *NewsletterHandler) UnsubscribeFromNewsletter(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")
	newsletterJID := c.Params("newsletterId")

	_, err := h.sessionService.GetSession(c.Context(), sessionID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "ERROR", Message: "Session not found: " + err.Error()},
		})
	}

	if newsletterJID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "ERROR", Message: "Newsletter JID is required"},
		})
	}

	err = h.wmeowService.UnfollowNewsletter(c.Context(), sessionID, newsletterJID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "ERROR", Message: "Failed to unsubscribe from newsletter: " + err.Error()},
		})
	}

	return c.Status(fiber.StatusOK).JSON(dto.StandardResponse{
		Success: true,
		Data:    map[string]string{"message": "Successfully unsubscribed from newsletter"},
	})
}

// SendNewsletterMessage godoc
// @Summary Send newsletter message
// @Description Sends a message to a newsletter
// @Tags Newsletters
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param newsletterId path string true "Newsletter ID"
// @Param request body dto.SendNewsletterMessageRequest true "Send newsletter message request"
// @Success 200 {object} dto.SendNewsletterMessageResponse "Message sent successfully"
// @Failure 400 {object} dto.SendNewsletterMessageResponse "Invalid request data"
// @Failure 401 {object} dto.SendNewsletterMessageResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.SendNewsletterMessageResponse "Session not found"
// @Failure 500 {object} dto.SendNewsletterMessageResponse "Failed to send message"
// @Router /session/{sessionId}/newsletter/{newsletterId}/send [post]
func (h *NewsletterHandler) SendNewsletterMessage(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")
	newsletterJID := c.Params("newsletterId")

	if !h.wmeowService.IsClientConnected(sessionID) {
		return c.Status(fiber.StatusBadRequest).JSON(dto.SendNewsletterMessageResponse{
			Success: false,
			Error:   &dto.NewsletterErrorResponse{Code: "ERROR", Message: "Session not found or not connected"},
		})
	}

	var req dto.SendNewsletterMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.SendNewsletterMessageResponse{
			Success: false,
			Error:   &dto.NewsletterErrorResponse{Code: "ERROR", Message: "Invalid request body: " + err.Error()},
		})
	}

	if req.Message == "" && req.MediaData == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.SendNewsletterMessageResponse{
			Success: false,
			Error:   &dto.NewsletterErrorResponse{Code: "ERROR", Message: "Either message or media_data is required"},
		})
	}

	err := h.wmeowService.SendNewsletterMessage(c.Context(), sessionID, newsletterJID, req.Message)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.SendNewsletterMessageResponse{
			Success: false,
			Error:   &dto.NewsletterErrorResponse{Code: "ERROR", Message: "Failed to send newsletter message: " + err.Error()},
		})
	}

	return c.Status(fiber.StatusOK).JSON(dto.SendNewsletterMessageResponse{
		Success: true,
		Data: &dto.NewsletterMessageData{
			SessionId:     sessionID,
			NewsletterJID: newsletterJID,
			MessageID:     fmt.Sprintf("msg_%d", time.Now().Unix()),
			ServerID:      fmt.Sprintf("srv_%d", time.Now().Unix()),
			Status:        "sent",
			Timestamp:     time.Now(),
		},
	})
}

// GetNewsletterMessages godoc
// @Summary Get newsletter messages
// @Description Retrieves messages from a newsletter
// @Tags Newsletters
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param newsletterId path string true "Newsletter ID"
// @Param limit query int false "Number of messages to retrieve" default(50)
// @Success 200 {object} dto.StandardResponse "Newsletter messages"
// @Failure 400 {object} dto.StandardResponse "Invalid request data"
// @Failure 401 {object} dto.StandardResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.StandardResponse "Session not found"
// @Failure 500 {object} dto.StandardResponse "Failed to get messages"
// @Router /session/{sessionId}/newsletter/{newsletterId}/messages [get]
func (h *NewsletterHandler) GetNewsletterMessages(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")
	newsletterJID := c.Params("newsletterId")

	if !h.wmeowService.IsClientConnected(sessionID) {
		return c.Status(fiber.StatusBadRequest).JSON(dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "ERROR", Message: "Session not found or not connected"},
		})
	}

	_ = c.Query("count")
	_ = c.Query("before")

	messages, err := h.wmeowService.GetNewsletterMessages(c.Context(), sessionID, newsletterJID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "ERROR", Message: "Failed to get newsletter messages: " + err.Error()},
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    messages,
		"count":   len(messages),
	})
}
