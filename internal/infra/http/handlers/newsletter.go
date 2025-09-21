package handlers

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"zpmeow/internal/application"
	"zpmeow/internal/infra/http/dto"
	"zpmeow/internal/infra/wmeow"

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

	_ = count
	_ = before
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

	err := h.wmeowService.NewsletterMarkViewed(c.Request.Context(), sessionID, newsletterJID, req.ServerIDs)
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

	if req.MediaData == "" || req.MediaType == "" {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "MISSING_REQUIRED_FIELDS",
				Message: "data and media_type are required",
			},
		})
		return
	}

	data, err := base64.StdEncoding.DecodeString(req.MediaData)
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
			Error: &dto.NewsletterErrorResponse{
				Code:    "SESSION_NOT_CONNECTED",
				Message: "Session not found or not connected",
			},
		})
		return
	}

	if inviteKey == "" {
		c.JSON(http.StatusBadRequest, dto.NewsletterInfoResponse{
			Success: false,
			Error: &dto.NewsletterErrorResponse{
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
			Error: &dto.NewsletterErrorResponse{
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
		CreatedAt:       time.Unix(info.CreatedAt, 0),
		IsVerified:      info.IsVerified,
		SubscriberCount: info.Subscribers,
	}

	c.JSON(http.StatusOK, dto.NewsletterInfoResponse{
		Success: true,
		Data:    &result,
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
func (h *NewsletterHandler) CreateNewsletter(c *gin.Context) {
	sessionID := c.Param("sessionId")

	resolvedSessionID, err := h.resolveSessionID(c, sessionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.CreateNewsletterResponse{
			Success: false,
			Error:   &dto.NewsletterErrorResponse{Code: "ERROR", Message: "Session not found: " + err.Error()},
		})
		return
	}

	if !h.wmeowService.IsClientConnected(resolvedSessionID) {
		c.JSON(http.StatusBadRequest, dto.CreateNewsletterResponse{
			Success: false,
			Error:   &dto.NewsletterErrorResponse{Code: "ERROR", Message: "Session not connected"},
		})
		return
	}

	var req dto.CreateNewsletterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.CreateNewsletterResponse{
			Success: false,
			Error:   &dto.NewsletterErrorResponse{Code: "ERROR", Message: "Invalid request format: " + err.Error()},
		})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, dto.CreateNewsletterResponse{
			Success: false,
			Error:   &dto.NewsletterErrorResponse{Code: "ERROR", Message: "Newsletter name is required"},
		})
		return
	}

	resp, err := h.wmeowService.CreateNewsletter(c.Request.Context(), resolvedSessionID, req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.CreateNewsletterResponse{
			Success: false,
			Error:   &dto.NewsletterErrorResponse{Code: "ERROR", Message: "Failed to create newsletter: " + err.Error()},
		})
		return
	}

	result := &dto.NewsletterInfo{
		JID:             resp.JID,
		Name:            resp.Name,
		Description:     resp.Description,
		CreatedAt:       time.Unix(resp.Timestamp, 0),
		SubscriberCount: 0,
	}

	c.JSON(http.StatusCreated, dto.CreateNewsletterResponse{
		Success: true,
		Data:    result,
	})
}

func (h *NewsletterHandler) GetNewsletter(c *gin.Context) {
	sessionID := c.Param("sessionId")
	newsletterJID := c.Param("newsletterId")

	if !h.wmeowService.IsClientConnected(sessionID) {
		c.JSON(http.StatusBadRequest, dto.NewsletterInfoResponse{
			Success: false,
			Error: &dto.NewsletterErrorResponse{
				Code:    "SESSION_NOT_CONNECTED",
				Message: "Session not found or not connected",
			},
		})
		return
	}

	if newsletterJID == "" {
		c.JSON(http.StatusBadRequest, dto.NewsletterInfoResponse{
			Success: false,
			Error:   &dto.NewsletterErrorResponse{Code: "ERROR", Message: "Newsletter JID is required"},
		})
		return
	}

	info, err := h.wmeowService.GetNewsletterInfo(c.Request.Context(), sessionID, newsletterJID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewsletterInfoResponse{
			Success: false,
			Error:   &dto.NewsletterErrorResponse{Code: "ERROR", Message: "Failed to get newsletter info: " + err.Error()},
		})
		return
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

	c.JSON(http.StatusOK, dto.NewsletterInfoResponse{
		Success: true,
		Data:    result,
	})
}

func (h *NewsletterHandler) ListNewsletters(c *gin.Context) {
	sessionID := c.Param("sessionId")

	if !h.wmeowService.IsClientConnected(sessionID) {
		c.JSON(http.StatusBadRequest, dto.NewsletterListResponse{
			Success: false,
			Error:   &dto.NewsletterErrorResponse{Code: "ERROR", Message: "Session not found or not connected"},
		})
		return
	}

	newsletters, err := h.wmeowService.GetSubscribedNewsletters(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewsletterListResponse{
			Success: false,
			Error:   &dto.NewsletterErrorResponse{Code: "ERROR", Message: "Failed to get newsletters: " + err.Error()},
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

	c.JSON(http.StatusOK, dto.NewsletterListResponse{
		Success: true,
		Data:    listResult,
	})
}

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

func (h *NewsletterHandler) SendNewsletterMessage(c *gin.Context) {
	sessionID := c.Param("sessionId")
	newsletterJID := c.Param("newsletterId")

	if !h.wmeowService.IsClientConnected(sessionID) {
		c.JSON(http.StatusBadRequest, dto.SendNewsletterMessageResponse{
			Success: false,
			Error:   &dto.NewsletterErrorResponse{Code: "ERROR", Message: "Session not found or not connected"},
		})
		return
	}

	var req dto.SendNewsletterMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.SendNewsletterMessageResponse{
			Success: false,
			Error:   &dto.NewsletterErrorResponse{Code: "ERROR", Message: "Invalid request body: " + err.Error()},
		})
		return
	}

	if req.Message == "" && req.MediaData == "" {
		c.JSON(http.StatusBadRequest, dto.SendNewsletterMessageResponse{
			Success: false,
			Error:   &dto.NewsletterErrorResponse{Code: "ERROR", Message: "Either message or media_data is required"},
		})
		return
	}

	err := h.wmeowService.SendNewsletterMessage(c.Request.Context(), sessionID, newsletterJID, req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.SendNewsletterMessageResponse{
			Success: false,
			Error:   &dto.NewsletterErrorResponse{Code: "ERROR", Message: "Failed to send newsletter message: " + err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, dto.SendNewsletterMessageResponse{
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

	_ = c.Query("count")
	_ = c.Query("before")

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
