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

	resolvedSessionID, err := h.resolveSessionID(c, sessionID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.CreateNewsletterResponse{
			Success: false,
			Error:   &dto.NewsletterErrorResponse{Code: "ERROR", Message: "Session not found: " + err.Error()},
		})
	}

	if !h.wmeowService.IsClientConnected(resolvedSessionID) {
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

	resp, err := h.wmeowService.CreateNewsletter(c.Context(), resolvedSessionID, req.Name, req.Description)
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
		SessionID:   sessionID,
		Newsletters: result,
		Count:       len(result),
		Total:       len(result),
	}

	return c.Status(fiber.StatusOK).JSON(dto.NewsletterListResponse{
		Success: true,
		Data:    listResult,
	})
}

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
			SessionID:     sessionID,
			NewsletterJID: newsletterJID,
			MessageID:     fmt.Sprintf("msg_%d", time.Now().Unix()),
			ServerID:      fmt.Sprintf("srv_%d", time.Now().Unix()),
			Status:        "sent",
			Timestamp:     time.Now(),
		},
	})
}

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
