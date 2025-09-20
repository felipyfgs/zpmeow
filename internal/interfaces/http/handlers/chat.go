package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"zpmeow/internal/application"
	"zpmeow/internal/infra/wmeow"
	"zpmeow/internal/interfaces/dto"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	sessionService *application.SessionApp
	wmeowService   wmeow.WameowService
}

func NewChatHandler(sessionService *application.SessionApp, wmeowService wmeow.WameowService) *ChatHandler {
	return &ChatHandler{
		sessionService: sessionService,
		wmeowService:   wmeowService,
	}
}

// @Summary		Set presence in chat
// @Description	Set user presence state in a specific chat (composing, available, etc.)
// @Tags			Chat
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string					true	"Session ID"
// @Param			request		body		dto.SetPresenceRequest	true	"Presence request"
// @Success		200			{object}	dto.ChatResponse
// @Failure		400			{object}	dto.ChatResponse
// @Failure		500			{object}	dto.ChatResponse
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/chat/presence [post]
func (h *ChatHandler) SetPresence(c *gin.Context) {
	sessionID := c.Param("sessionId")

	var req dto.SetPresenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewChatErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
		return
	}

	if req.State == "" {
		c.JSON(http.StatusBadRequest, dto.NewChatErrorResponse(
			http.StatusBadRequest,
			"MISSING_STATE",
			"State is required",
			"Valid states: available, unavailable, composing, recording, paused",
		))
		return
	}

	ctx := c.Request.Context()
	err := h.wmeowService.SetPresence(ctx, sessionID, req.Phone, req.State, req.Media)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewChatErrorResponse(
			http.StatusInternalServerError,
			"SET_PRESENCE_FAILED",
			"Failed to set presence",
			err.Error(),
		))
		return
	}

	response := dto.NewChatSuccessResponse(req.Phone, "", "set_presence")
	c.JSON(http.StatusOK, response)
}

// @Summary		Download image
// @Description	Download image media from a message
// @Tags			Chat
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string						true	"Session ID"
// @Param			request		body		dto.DownloadMediaRequest	true	"Download request"
// @Success		200			{object}	dto.MediaDownloadResponse
// @Failure		400			{object}	dto.ChatResponse
// @Failure		500			{object}	dto.ChatResponse
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/chat/download/image [post]
func (h *ChatHandler) DownloadImage(c *gin.Context) {
	h.downloadMedia(c, "image")
}

// @Summary		Download video
// @Description	Download video media from a message
// @Tags			Chat
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string						true	"Session ID"
// @Param			request		body		dto.DownloadMediaRequest	true	"Download request"
// @Success		200			{object}	dto.MediaDownloadResponse
// @Failure		400			{object}	dto.ChatResponse
// @Failure		500			{object}	dto.ChatResponse
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/chat/download/video [post]
func (h *ChatHandler) DownloadVideo(c *gin.Context) {
	h.downloadMedia(c, "video")
}

// @Summary		Download audio
// @Description	Download audio media from a message
// @Tags			Chat
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string						true	"Session ID"
// @Param			request		body		dto.DownloadMediaRequest	true	"Download request"
// @Success		200			{object}	dto.MediaDownloadResponse
// @Failure		400			{object}	dto.ChatResponse
// @Failure		500			{object}	dto.ChatResponse
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/chat/download/audio [post]
func (h *ChatHandler) DownloadAudio(c *gin.Context) {
	h.downloadMedia(c, "audio")
}

// @Summary		Download document
// @Description	Download document media from a message
// @Tags			Chat
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string						true	"Session ID"
// @Param			request		body		dto.DownloadMediaRequest	true	"Download request"
// @Success		200			{object}	dto.MediaDownloadResponse
// @Failure		400			{object}	dto.ChatResponse
// @Failure		500			{object}	dto.ChatResponse
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/chat/download/document [post]
func (h *ChatHandler) DownloadDocument(c *gin.Context) {
	h.downloadMedia(c, "document")
}

func (h *ChatHandler) downloadMedia(c *gin.Context, mediaType string) {
	sessionID := c.Param("sessionId")

	var req dto.DownloadMediaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewChatErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
		return
	}

	if req.MessageID == "" {
		c.JSON(http.StatusBadRequest, dto.NewChatErrorResponse(
			http.StatusBadRequest,
			"MISSING_MESSAGE_ID",
			"Message ID is required",
			"",
		))
		return
	}

	ctx := c.Request.Context()
	data, mimeType, err := h.wmeowService.DownloadMedia(ctx, sessionID, req.MessageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewChatErrorResponse(
			http.StatusInternalServerError,
			"DOWNLOAD_FAILED",
			"Failed to download media",
			err.Error(),
		))
		return
	}

	response := &dto.MediaDownloadResponse{
		Success:   true,
		Code:      http.StatusOK,
		MessageID: req.MessageID,
		MediaType: mediaType,
		MimeType:  mimeType,
		Data:      data,
		Size:      len(data),
	}

	c.JSON(http.StatusOK, response)
}

// @Summary		Get chat history
// @Description	Get chat history for a specific contact
// @Tags			Chat
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string	true	"Session ID"
// @Param			phone		query		string	true	"Phone number"
// @Param			limit		query		int		false	"Limit of messages (default: 50, max: 1000)"
// @Success		200			{object}	dto.ChatHistoryResponse
// @Failure		400			{object}	dto.ChatResponse
// @Failure		500			{object}	dto.ChatResponse
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/chat/history [get]
func (h *ChatHandler) GetChatHistory(c *gin.Context) {
	phone := c.Query("phone")
	limitStr := c.DefaultQuery("limit", "50")

	if phone == "" {
		c.JSON(http.StatusBadRequest, dto.NewChatErrorResponse(
			http.StatusBadRequest,
			"MISSING_PHONE",
			"Phone number is required",
			"",
		))
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewChatErrorResponse(
			http.StatusBadRequest,
			"INVALID_LIMIT",
			"Invalid limit parameter",
			err.Error(),
		))
		return
	}

	response := &dto.ChatHistoryResponse{
		Success: true,
		Code:    http.StatusOK,
		Data: dto.ChatHistoryResponseData{
			Phone:    phone,
			Messages: []dto.ChatHistoryData{},
			Count:    0,
			Limit:    limit,
		},
	}

	c.JSON(http.StatusOK, response)
}

// @Summary		Set disappearing timer
// @Description	Set disappearing timer for messages in a chat
// @Tags			Chat
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string								true	"Session ID"
// @Param			request		body		dto.SetDisappearingTimerRequest		true	"Disappearing timer request"
// @Success		200			{object}	dto.ChatResponse
// @Failure		400			{object}	dto.ChatResponse
// @Failure		500			{object}	dto.ChatResponse
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/chat/disappearing-timer [post]
func (h *ChatHandler) SetDisappearingTimer(c *gin.Context) {
	sessionID := c.Param("sessionId")

	var req dto.SetDisappearingTimerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewChatErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
		return
	}

	if req.JID == "" {
		c.JSON(http.StatusBadRequest, dto.NewChatErrorResponse(
			http.StatusBadRequest,
			"MISSING_JID",
			"JID is required",
			"",
		))
		return
	}

	if req.Timer == "" {
		c.JSON(http.StatusBadRequest, dto.NewChatErrorResponse(
			http.StatusBadRequest,
			"MISSING_TIMER",
			"Timer is required",
			"Valid values: off, 24h, 7d, 90d",
		))
		return
	}

	var timer time.Duration
	switch strings.ToLower(req.Timer) {
	case "off", "0":
		timer = 0
	case "24h":
		timer = 24 * time.Hour
	case "7d":
		timer = 7 * 24 * time.Hour
	case "90d":
		timer = 90 * 24 * time.Hour
	default:
		c.JSON(http.StatusBadRequest, dto.NewChatErrorResponse(
			http.StatusBadRequest,
			"INVALID_TIMER",
			"Invalid timer value",
			"Valid values: off, 24h, 7d, 90d",
		))
		return
	}

	ctx := c.Request.Context()
	err := h.wmeowService.SetDisappearingTimer(ctx, sessionID, req.JID, timer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewChatErrorResponse(
			http.StatusInternalServerError,
			"SET_TIMER_FAILED",
			"Failed to set disappearing timer",
			err.Error(),
		))
		return
	}

	response := dto.NewChatSuccessResponse(req.JID, "", "set_disappearing_timer")
	c.JSON(http.StatusOK, response)
}

// @Summary		List chats
// @Description	List all chats (groups and/or contacts) for a session
// @Tags			Chat
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string					true	"Session ID"
// @Param			request		body		dto.ListChatsRequest	false	"List chats request"
// @Success		200			{object}	dto.ListChatsResponse
// @Failure		400			{object}	dto.ListChatsResponse
// @Failure		500			{object}	dto.ListChatsResponse
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/chat/list [post]
func (h *ChatHandler) ListChats(c *gin.Context) {
	sessionID := c.Param("sessionId")

	var req dto.ListChatsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Type = "all"
	}

	if req.Type == "" {
		req.Type = "all"
	}

	validTypes := map[string]bool{
		"all":      true,
		"groups":   true,
		"contacts": true,
	}

	if !validTypes[req.Type] {
		c.JSON(http.StatusBadRequest, dto.NewListChatsErrorResponse(
			http.StatusBadRequest,
			"INVALID_TYPE",
			"Invalid chat type",
			"Valid types: all, groups, contacts",
		))
		return
	}

	ctx := c.Request.Context()
	chats, err := h.wmeowService.ListChats(ctx, sessionID, req.Type)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewListChatsErrorResponse(
			http.StatusInternalServerError,
			"LIST_CHATS_FAILED",
			"Failed to list chats",
			err.Error(),
		))
		return
	}

	dtoChats := make([]dto.ChatInfo, len(chats))
	for i, chat := range chats {
		dtoChats[i] = dto.ChatInfo{
			JID:         chat.JID,
			Name:        chat.Name,
			Type:        chat.Type,
			LastMessage: chat.LastMessage,
			Timestamp:   chat.LastSeen,
			UnreadCount: chat.UnreadCount,
			Pinned:      chat.IsPinned,
			Muted:       chat.IsMuted,
			Archived:    chat.IsArchived,
		}
	}

	response := dto.NewListChatsSuccessResponse(dtoChats, req.Type)
	c.JSON(http.StatusOK, response)
}

// @Summary		Get chat info
// @Description	Get detailed information about a specific chat (group or contact)
// @Tags			Chat
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string						true	"Session ID"
// @Param			request		body		dto.GetChatInfoRequest		true	"Get chat info request"
// @Success		200			{object}	dto.GetChatInfoResponse
// @Failure		400			{object}	dto.GetChatInfoResponse
// @Failure		500			{object}	dto.GetChatInfoResponse
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/chat/info [post]
func (h *ChatHandler) GetChatInfo(c *gin.Context) {
	sessionID := c.Param("sessionId")

	var req dto.GetChatInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewGetChatInfoErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
		return
	}

	if req.JID == "" {
		c.JSON(http.StatusBadRequest, dto.NewGetChatInfoErrorResponse(
			http.StatusBadRequest,
			"MISSING_JID",
			"JID is required",
			"",
		))
		return
	}

	ctx := c.Request.Context()
	chatInfo, err := h.wmeowService.GetChatInfo(ctx, sessionID, req.JID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewGetChatInfoErrorResponse(
			http.StatusInternalServerError,
			"GET_CHAT_INFO_FAILED",
			"Failed to get chat info",
			err.Error(),
		))
		return
	}

	dtoChatInfo := dto.ChatInfo{
		JID:         chatInfo.JID,
		Name:        chatInfo.Name,
		Type:        chatInfo.Type,
		LastMessage: chatInfo.LastMessage,
		Timestamp:   chatInfo.LastSeen,
		UnreadCount: chatInfo.UnreadCount,
		Pinned:      chatInfo.IsPinned,
		Muted:       chatInfo.IsMuted,
		Archived:    chatInfo.IsArchived,
	}

	response := dto.NewGetChatInfoSuccessResponse(dtoChatInfo)
	c.JSON(http.StatusOK, response)
}

// @Summary		Pin/unpin chat
// @Description	Pin or unpin a chat to keep it at the top of the chat list
// @Tags			Chat
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string					true	"Session ID"
// @Param			request		body		dto.PinChatRequest		true	"Pin chat request"
// @Success		200			{object}	dto.ChatResponse
// @Failure		400			{object}	dto.ChatResponse
// @Failure		500			{object}	dto.ChatResponse
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/chat/pin [post]
func (h *ChatHandler) PinChat(c *gin.Context) {
	sessionID := c.Param("sessionId")

	var req dto.PinChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewChatErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
		return
	}

	if req.JID == "" {
		c.JSON(http.StatusBadRequest, dto.NewChatErrorResponse(
			http.StatusBadRequest,
			"MISSING_JID",
			"JID is required",
			"",
		))
		return
	}

	ctx := c.Request.Context()
	err := h.wmeowService.PinChat(ctx, sessionID, req.JID, req.Pinned)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewChatErrorResponse(
			http.StatusInternalServerError,
			"PIN_CHAT_FAILED",
			"Failed to pin/unpin chat",
			err.Error(),
		))
		return
	}

	action := "unpin_chat"
	if req.Pinned {
		action = "pin_chat"
	}

	response := dto.NewChatSuccessResponse(req.JID, "", action)
	c.JSON(http.StatusOK, response)
}

// @Summary		Mute/unmute chat
// @Description	Mute or unmute a chat for a specified duration
// @Tags			Chat
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string					true	"Session ID"
// @Param			request		body		dto.MuteChatRequest		true	"Mute chat request"
// @Success		200			{object}	dto.ChatResponse
// @Failure		400			{object}	dto.ChatResponse
// @Failure		500			{object}	dto.ChatResponse
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/chat/mute [post]
func (h *ChatHandler) MuteChat(c *gin.Context) {
	sessionID := c.Param("sessionId")

	var req dto.MuteChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewChatErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
		return
	}

	if req.JID == "" {
		c.JSON(http.StatusBadRequest, dto.NewChatErrorResponse(
			http.StatusBadRequest,
			"MISSING_JID",
			"JID is required",
			"",
		))
		return
	}

	var duration time.Duration
	if req.Muted && req.Duration != "" {
		switch strings.ToLower(req.Duration) {
		case "1h":
			duration = 1 * time.Hour
		case "8h":
			duration = 8 * time.Hour
		case "1w":
			duration = 7 * 24 * time.Hour
		case "forever", "":
			duration = 0 // 0 means forever
		default:
			c.JSON(http.StatusBadRequest, dto.NewChatErrorResponse(
				http.StatusBadRequest,
				"INVALID_DURATION",
				"Invalid mute duration",
				"Valid durations: 1h, 8h, 1w, forever",
			))
			return
		}
	}

	ctx := c.Request.Context()
	err := h.wmeowService.MuteChat(ctx, sessionID, req.JID, req.Muted, duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewChatErrorResponse(
			http.StatusInternalServerError,
			"MUTE_CHAT_FAILED",
			"Failed to mute/unmute chat",
			err.Error(),
		))
		return
	}

	action := "unmute_chat"
	if req.Muted {
		action = "mute_chat"
	}

	response := dto.NewChatSuccessResponse(req.JID, "", action)
	c.JSON(http.StatusOK, response)
}

// @Summary		Archive/unarchive chat
// @Description	Archive or unarchive a chat (archiving automatically unpins the chat)
// @Tags			Chat
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string						true	"Session ID"
// @Param			request		body		dto.ArchiveChatRequest		true	"Archive chat request"
// @Success		200			{object}	dto.ChatResponse
// @Failure		400			{object}	dto.ChatResponse
// @Failure		500			{object}	dto.ChatResponse
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/chat/archive [post]
func (h *ChatHandler) ArchiveChat(c *gin.Context) {
	sessionID := c.Param("sessionId")

	var req dto.ArchiveChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewChatErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
		return
	}

	if req.JID == "" {
		c.JSON(http.StatusBadRequest, dto.NewChatErrorResponse(
			http.StatusBadRequest,
			"MISSING_JID",
			"JID is required",
			"",
		))
		return
	}

	ctx := c.Request.Context()
	err := h.wmeowService.ArchiveChat(ctx, sessionID, req.JID, req.Archived)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewChatErrorResponse(
			http.StatusInternalServerError,
			"ARCHIVE_CHAT_FAILED",
			"Failed to archive/unarchive chat",
			err.Error(),
		))
		return
	}

	action := "unarchive_chat"
	if req.Archived {
		action = "archive_chat"
	}

	response := dto.NewChatSuccessResponse(req.JID, "", action)
	c.JSON(http.StatusOK, response)
}
