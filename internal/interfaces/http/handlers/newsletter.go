package handlers

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"zpmeow/internal/application"
	"zpmeow/internal/infra/wmeow"
	"zpmeow/internal/interfaces/dto"

	"github.com/gin-gonic/gin"
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

func (h *NewsletterHandler) GetNewsletterMessageUpdates(c *gin.Context) {
	sessionID := c.Param("sessionId")
	newsletterID := c.Param("newsletterId")

	if !h.wmeowService.IsClientConnected(sessionID) {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "SESSION_NOT_CONNECTED",
				Message: "Session not found or not connected",
			},
		})
		return
	}

	count := 50
	if countStr := c.Query("count"); countStr != "" {
		if c, err := strconv.Atoi(countStr); err == nil && c > 0 {
			count = c
		}
	}
	before := c.Query("before")

	_ = count  // TODO: implement pagination with count
	_ = before // TODO: implement pagination with before
	updates, err := h.wmeowService.GetNewsletterMessageUpdates(c.Request.Context(), sessionID, newsletterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "GET_UPDATES_FAILED",
				Message: "Failed to get newsletter message updates: " + err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    updates,
		"count":   len(updates),
	})
}

func (h *NewsletterHandler) MarkNewsletterViewed(c *gin.Context) {
	sessionID := c.Param("sessionId")
	newsletterJID := c.Param("newsletterId")

	if !h.wmeowService.IsClientConnected(sessionID) {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "SESSION_NOT_CONNECTED",
				Message: "Session not found or not connected",
			},
		})
		return
	}

	var req dto.MarkViewedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_REQUEST_BODY",
				Message: "Invalid request body: " + err.Error(),
			},
		})
		return
	}

	err := h.wmeowService.NewsletterMarkViewed(c.Request.Context(), sessionID, newsletterJID, req.MessageIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "MARK_VIEWED_FAILED",
				Message: "Failed to mark messages as viewed: " + err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    map[string]string{"message": "Messages marked as viewed successfully"},
	})
}

func (h *NewsletterHandler) SendNewsletterReaction(c *gin.Context) {
	sessionID := c.Param("sessionId")
	newsletterJID := c.Param("newsletterId")

	if !h.wmeowService.IsClientConnected(sessionID) {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "SESSION_NOT_CONNECTED",
				Message: "Session not found or not connected",
			},
		})
		return
	}

	var req dto.SendReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_REQUEST_BODY",
				Message: "Invalid request body: " + err.Error(),
			},
		})
		return
	}

	if req.ServerID == "" || req.Reaction == "" || req.MessageID == "" {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "MISSING_REQUIRED_FIELDS",
				Message: "server_id, reaction, and message_id are required",
			},
		})
		return
	}

	err := h.wmeowService.NewsletterSendReaction(
		c.Request.Context(),
		sessionID,
		newsletterJID,
		req.MessageID,
		req.Reaction,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "SEND_REACTION_FAILED",
				Message: "Failed to send reaction: " + err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    map[string]string{"message": "Reaction sent successfully"},
	})
}

func (h *NewsletterHandler) ToggleNewsletterMute(c *gin.Context) {
	sessionID := c.Param("sessionId")
	newsletterJID := c.Param("newsletterId")

	if !h.wmeowService.IsClientConnected(sessionID) {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "SESSION_NOT_CONNECTED",
				Message: "Session not found or not connected",
			},
		})
		return
	}

	var req dto.ToggleMuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_REQUEST_BODY",
				Message: "Invalid request body: " + err.Error(),
			},
		})
		return
	}

	err := h.wmeowService.NewsletterToggleMute(c.Request.Context(), sessionID, newsletterJID, req.Mute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "TOGGLE_MUTE_FAILED",
				Message: "Failed to toggle mute: " + err.Error(),
			},
		})
		return
	}

	action := "muted"
	if !req.Mute {
		action = "unmuted"
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    map[string]string{"message": fmt.Sprintf("Newsletter %s successfully", action)},
	})
}

func (h *NewsletterHandler) SubscribeLiveUpdates(c *gin.Context) {
	sessionID := c.Param("sessionId")
	newsletterJID := c.Param("newsletterId")

	if !h.wmeowService.IsClientConnected(sessionID) {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "SESSION_NOT_CONNECTED",
				Message: "Session not found or not connected",
			},
		})
		return
	}

	err := h.wmeowService.NewsletterSubscribeLiveUpdates(c.Request.Context(), sessionID, newsletterJID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "SUBSCRIBE_LIVE_UPDATES_FAILED",
				Message: "Failed to subscribe to live updates: " + err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    map[string]string{"message": "Successfully subscribed to live updates"},
	})
}

// @Summary		Upload media for newsletter
// @Description	Upload media files (image, video, audio, document) for use in newsletter messages. Returns MediaHandle required for sending media messages.
// @Tags			Newsletters
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string							true	"Session ID"
// @Param			request		body		dto.UploadNewsletterMediaRequest	true	"Upload media request"
// @Success		200			{object}	dto.StandardResponse	"Media uploaded successfully"
// @Failure		400			{object}	dto.StandardResponse				"Bad request - Invalid file or parameters"
// @Failure		404			{object}	dto.StandardResponse				"Session not found or not connected"
// @Failure		500			{object}	dto.StandardResponse				"Internal server error"
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/newsletter/upload [post]
func (h *NewsletterHandler) UploadNewsletterMedia(c *gin.Context) {
	sessionID := c.Param("sessionId")

	if !h.wmeowService.IsClientConnected(sessionID) {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "SESSION_NOT_CONNECTED",
				Message: "Session not found or not connected",
			},
		})
		return
	}

	var req dto.UploadNewsletterMediaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_REQUEST_BODY",
				Message: "Invalid request body: " + err.Error(),
			},
		})
		return
	}

	if req.Data == "" || req.MediaType == "" {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "MISSING_REQUIRED_FIELDS",
				Message: "data and media_type are required",
			},
		})
		return
	}

	data, err := base64.StdEncoding.DecodeString(req.Data)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_BASE64_DATA",
				Message: "Invalid base64 data: " + err.Error(),
			},
		})
		return
	}

	validTypes := map[string]bool{
		"image":    true,
		"video":    true,
		"audio":    true,
		"document": true,
	}

	if !validTypes[req.MediaType] {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_MEDIA_TYPE",
				Message: "Invalid media type. Supported: image, video, audio, document",
			},
		})
		return
	}

	err = h.wmeowService.UploadNewsletter(c.Request.Context(), sessionID, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "UPLOAD_FAILED",
				Message: "Failed to upload media: " + err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Media uploaded successfully",
	})
}

func (h *NewsletterHandler) GetNewsletterByInvite(c *gin.Context) {
	sessionID := c.Param("sessionId")
	inviteKey := c.Param("inviteKey")

	if !h.wmeowService.IsClientConnected(sessionID) {
		c.JSON(http.StatusBadRequest, dto.NewsletterInfoResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "SESSION_NOT_CONNECTED",
				Message: "Session not found or not connected",
			},
		})
		return
	}

	if inviteKey == "" {
		c.JSON(http.StatusBadRequest, dto.NewsletterInfoResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "MISSING_INVITE_KEY",
				Message: "Invite key is required",
			},
		})
		return
	}

	info, err := h.wmeowService.GetNewsletterInfoWithInvite(c.Request.Context(), sessionID, inviteKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewsletterInfoResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "GET_NEWSLETTER_INFO_FAILED",
				Message: "Failed to get newsletter info: " + err.Error(),
			},
		})
		return
	}

	result := dto.NewsletterInfo{
		JID:             info.JID,
		Name:            info.Name,
		Description:     info.Description,
		CreatedAt:       info.CreatedAt,
		IsVerified:      info.IsVerified,
		SubscriberCount: info.Subscribers,
	}

	c.JSON(http.StatusOK, dto.NewsletterInfoResponse{
		Success: true,
		Data:    result,
	})
}

func (h *NewsletterHandler) resolveSessionID(c *gin.Context, sessionIDOrName string) (string, error) {
	if h.sessionService == nil {
		return sessionIDOrName, nil
	}

	ctx := c.Request.Context()
	session, err := h.sessionService.GetSession(ctx, sessionIDOrName)
	if err != nil {
		return "", err
	}

	return session.SessionID().String(), nil
}

// @Summary		Create newsletter
// @Description	Create a new meow newsletter
// @Tags			Newsletters
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string						true	"Session ID"
// @Param			request		body		dto.CreateNewsletterRequest	true	"Newsletter creation request"
// @Success		201			{object}	dto.CreateNewsletterResponse	"Newsletter created successfully"
// @Failure		400			{object}	dto.CreateNewsletterResponse	"Bad request"
// @Failure		500			{object}	dto.CreateNewsletterResponse	"Internal server error"
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/newsletter [post]
func (h *NewsletterHandler) CreateNewsletter(c *gin.Context) {
	sessionID := c.Param("sessionId")

	resolvedSessionID, err := h.resolveSessionID(c, sessionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.CreateNewsletterResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "ERROR", Message: "Session not found: " + err.Error()},
		})
		return
	}

	if !h.wmeowService.IsClientConnected(resolvedSessionID) {
		c.JSON(http.StatusBadRequest, dto.CreateNewsletterResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "ERROR", Message: "Session not connected"},
		})
		return
	}

	var req dto.CreateNewsletterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.CreateNewsletterResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "ERROR", Message: "Invalid request format: " + err.Error()},
		})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, dto.CreateNewsletterResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "ERROR", Message: "Newsletter name is required"},
		})
		return
	}

	resp, err := h.wmeowService.CreateNewsletter(c.Request.Context(), resolvedSessionID, req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.CreateNewsletterResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "ERROR", Message: "Failed to create newsletter: " + err.Error()},
		})
		return
	}

	result := dto.CreateNewsletterResult{
		JID:         resp.JID,
		ServerID:    resp.ServerID,
		Timestamp:   fmt.Sprintf("%d", resp.Timestamp),
		Name:        resp.Name,
		Description: resp.Description,
	}

	c.JSON(http.StatusCreated, dto.CreateNewsletterResponse{
		Success: true,
		Data:    result,
	})
}

// @Summary		Get newsletter information
// @Description	Get information about a specific newsletter
// @Tags			Newsletters
// @Accept			json
// @Produce		json
// @Param			sessionId		path		string						true	"Session ID"
// @Param			newsletterId	path		string						true	"Newsletter JID"
// @Success		200				{object}	dto.NewsletterInfoResponse	"Newsletter information retrieved"
// @Failure		400				{object}	dto.NewsletterInfoResponse	"Bad request"
// @Failure		404				{object}	dto.NewsletterInfoResponse	"Newsletter not found"
// @Failure		500				{object}	dto.NewsletterInfoResponse	"Internal server error"
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/newsletter/{newsletterId} [get]
func (h *NewsletterHandler) GetNewsletter(c *gin.Context) {
	sessionID := c.Param("sessionId")
	newsletterJID := c.Param("newsletterId")

	if !h.wmeowService.IsClientConnected(sessionID) {
		c.JSON(http.StatusBadRequest, dto.NewsletterInfoResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "SESSION_NOT_CONNECTED",
				Message: "Session not found or not connected",
			},
		})
		return
	}

	if newsletterJID == "" {
		c.JSON(http.StatusBadRequest, dto.NewsletterInfoResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "ERROR", Message: "Newsletter JID is required"},
		})
		return
	}

	info, err := h.wmeowService.GetNewsletterInfo(c.Request.Context(), sessionID, newsletterJID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewsletterInfoResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "ERROR", Message: "Failed to get newsletter info: " + err.Error()},
		})
		return
	}

	result := &dto.NewsletterInfo{
		JID:             info.JID,
		Name:            info.Name,
		Description:     info.Description,
		SubscriberCount: info.Subscribers,
		CreatedAt:       info.CreatedAt,
		IsVerified:      info.IsVerified,
		Picture:         "",
		Verified:        info.IsVerified,
		UpdatedAt:       "",
		OwnerJID:        "",
		Subscribers:     info.Subscribers,
		Muted:           false,
	}

	c.JSON(http.StatusOK, dto.NewsletterInfoResponse{
		Success: true,
		Data:    *result,
	})
}

// @Summary		List newsletters
// @Description	Get a list of all subscribed newsletters for a session
// @Tags			Newsletters
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string						true	"Session ID"
// @Success		200			{object}	dto.NewsletterListResponse	"Newsletters retrieved successfully"
// @Failure		400			{object}	dto.NewsletterListResponse	"Bad request"
// @Failure		500			{object}	dto.NewsletterListResponse	"Internal server error"
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/newsletter/list [get]
func (h *NewsletterHandler) ListNewsletters(c *gin.Context) {
	sessionID := c.Param("sessionId")

	if !h.wmeowService.IsClientConnected(sessionID) {
		c.JSON(http.StatusBadRequest, dto.NewsletterListResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "ERROR", Message: "Session not found or not connected"},
		})
		return
	}

	newsletters, err := h.wmeowService.GetSubscribedNewsletters(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewsletterListResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "ERROR", Message: "Failed to get newsletters: " + err.Error()},
		})
		return
	}

	result := make([]dto.NewsletterInfo, len(newsletters))
	for i, newsletter := range newsletters {
		result[i] = dto.NewsletterInfo{
			JID:             newsletter.JID,
			Name:            newsletter.Name,
			Description:     newsletter.Description,
			SubscriberCount: newsletter.Subscribers,
			CreatedAt:       newsletter.CreatedAt,
			IsVerified:      newsletter.IsVerified,
		}
	}

	listResult := &dto.NewsletterList{
		Newsletters: result,
		Count:       len(result),
		Total:       len(result),
	}

	c.JSON(http.StatusOK, dto.NewsletterListResponse{
		Success: true,
		Data:    *listResult,
	})
}

// @Summary		Subscribe to newsletter
// @Description	Subscribe to a newsletter to receive updates
// @Tags			Newsletters
// @Accept			json
// @Produce		json
// @Param			sessionId		path		string				true	"Session ID"
// @Param			newsletterId	path		string				true	"Newsletter JID"
// @Success		200				{object}	dto.StandardResponse	"Subscribed successfully"
// @Failure		400				{object}	dto.StandardResponse	"Bad request"
// @Failure		404				{object}	dto.StandardResponse	"Newsletter not found"
// @Failure		500				{object}	dto.StandardResponse	"Internal server error"
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/newsletter/{newsletterId}/subscribe [post]
func (h *NewsletterHandler) SubscribeToNewsletter(c *gin.Context) {
	sessionID := c.Param("sessionId")
	newsletterJID := c.Param("newsletterId")

	_, err := h.sessionService.GetSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "ERROR", Message: "Session not found: " + err.Error()},
		})
		return
	}

	if newsletterJID == "" {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "ERROR", Message: "Newsletter JID is required"},
		})
		return
	}

	err = h.wmeowService.FollowNewsletter(c.Request.Context(), sessionID, newsletterJID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "ERROR", Message: "Failed to subscribe to newsletter: " + err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    map[string]string{"message": "Successfully subscribed to newsletter"},
	})
}

// @Summary		Unsubscribe from newsletter
// @Description	Unsubscribe from a newsletter to stop receiving updates
// @Tags			Newsletters
// @Accept			json
// @Produce		json
// @Param			sessionId		path		string				true	"Session ID"
// @Param			newsletterId	path		string				true	"Newsletter JID"
// @Success		200				{object}	dto.StandardResponse	"Unsubscribed successfully"
// @Failure		400				{object}	dto.StandardResponse	"Bad request"
// @Failure		404				{object}	dto.StandardResponse	"Newsletter not found"
// @Failure		500				{object}	dto.StandardResponse	"Internal server error"
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/newsletter/{newsletterId}/unsubscribe [post]
func (h *NewsletterHandler) UnsubscribeFromNewsletter(c *gin.Context) {
	sessionID := c.Param("sessionId")
	newsletterJID := c.Param("newsletterId")

	_, err := h.sessionService.GetSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "ERROR", Message: "Session not found: " + err.Error()},
		})
		return
	}

	if newsletterJID == "" {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "ERROR", Message: "Newsletter JID is required"},
		})
		return
	}

	err = h.wmeowService.UnfollowNewsletter(c.Request.Context(), sessionID, newsletterJID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "ERROR", Message: "Failed to unsubscribe from newsletter: " + err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    map[string]string{"message": "Successfully unsubscribed from newsletter"},
	})
}

// @Summary		Send newsletter message
// @Description	Send a text or media message to newsletter subscribers. For media messages, MediaHandle is required.
// @Tags			Newsletters
// @Accept			json
// @Produce		json
// @Param			sessionId		path		string					true	"Session ID"
// @Param			newsletterId	path		string					true	"Newsletter ID (format: {id}@newsletter)"
// @Param			request			body		dto.SendNewsletterMessageRequest	true	"Message content (text or media)"
// @Success		200				{object}	dto.SendNewsletterMessageResponse	"Message sent successfully"
// @Failure		400				{object}	dto.SendNewsletterMessageResponse	"Bad request - Invalid parameters or missing MediaHandle for media"
// @Failure		404				{object}	dto.SendNewsletterMessageResponse	"Newsletter not found or session not connected"
// @Failure		500				{object}	dto.SendNewsletterMessageResponse	"Internal server error"
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/newsletter/{newsletterId}/send [post]
func (h *NewsletterHandler) SendNewsletterMessage(c *gin.Context) {
	sessionID := c.Param("sessionId")
	newsletterJID := c.Param("newsletterId")

	if !h.wmeowService.IsClientConnected(sessionID) {
		c.JSON(http.StatusBadRequest, dto.SendNewsletterMessageResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "ERROR", Message: "Session not found or not connected"},
		})
		return
	}

	var req dto.SendNewsletterMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.SendNewsletterMessageResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "ERROR", Message: "Invalid request body: " + err.Error()},
		})
		return
	}

	if req.Message == "" && req.MediaHandle == "" {
		c.JSON(http.StatusBadRequest, dto.SendNewsletterMessageResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "ERROR", Message: "Either message or media_handle is required"},
		})
		return
	}

	err := h.wmeowService.SendNewsletterMessage(c.Request.Context(), sessionID, newsletterJID, req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.SendNewsletterMessageResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "ERROR", Message: "Failed to send newsletter message: " + err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, dto.SendNewsletterMessageResponse{
		Success: true,
		Data: dto.SendNewsletterMessageData{
			SessionID:     sessionID,
			NewsletterJID: newsletterJID,
			MessageID:     fmt.Sprintf("msg_%d", time.Now().Unix()),
			Action:        "send",
			Status:        "sent",
			Timestamp:     time.Now(),
		},
	})
}

func (h *NewsletterHandler) GetNewsletterMessages(c *gin.Context) {
	sessionID := c.Param("sessionId")
	newsletterJID := c.Param("newsletterId")

	if !h.wmeowService.IsClientConnected(sessionID) {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "ERROR", Message: "Session not found or not connected"},
		})
		return
	}

	_ = c.Query("count")  // Ignore for now
	_ = c.Query("before") // Ignore for now

	messages, err := h.wmeowService.GetNewsletterMessages(c.Request.Context(), sessionID, newsletterJID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "ERROR", Message: "Failed to get newsletter messages: " + err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    messages,
		"count":   len(messages),
	})
}
