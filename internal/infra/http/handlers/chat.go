package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"zpmeow/internal/application"
	"zpmeow/internal/infra/http/dto"
	"zpmeow/internal/infra/wmeow"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	chatService  *application.ChatApp
	wmeowService wmeow.WameowService
}

func NewChatHandler(chatService *application.ChatApp, wmeowService wmeow.WameowService) *ChatHandler {
	return &ChatHandler{
		chatService:  chatService,
		wmeowService: wmeowService,
	}
}

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

func (h *ChatHandler) DownloadImage(c *gin.Context) {
	h.downloadMedia(c, "image")
}

func (h *ChatHandler) DownloadVideo(c *gin.Context) {
	h.downloadMedia(c, "video")
}

func (h *ChatHandler) DownloadAudio(c *gin.Context) {
	h.downloadMedia(c, "audio")
}

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
		Success: true,
		Code:    http.StatusOK,
		Data: &dto.MediaDownloadData{
			MediaID:  req.MessageID,
			Type:     mediaType,
			MimeType: mimeType,
			Data:     data,
			Size:     int64(len(data)),
		},
	}

	c.JSON(http.StatusOK, response)
}

func (h *ChatHandler) GetChatHistory(c *gin.Context) {
	sessionID := c.Param("sessionId")
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

	ctx := c.Request.Context()
	req := application.GetChatHistoryRequest{
		SessionID: sessionID,
		Phone:     phone,
		Limit:     limit,
		Offset:    0,
	}

	result, err := h.chatService.GetChatHistory(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewChatErrorResponse(
			http.StatusInternalServerError,
			"GET_CHAT_HISTORY_FAILED",
			"Failed to get chat history",
			err.Error(),
		))
		return
	}

	var messages []dto.ChatHistoryData
	for _, message := range result.Messages {
		timestamp := int64(0)
		if message.Timestamp != "" {
			timestamp = time.Now().Unix()
		}

		messages = append(messages, dto.ChatHistoryData{
			MessageID: message.ID,
			From:      message.FromJID,
			To:        message.ChatJID,
			Type:      message.Type,
			Content:   message.Content,
			Timestamp: timestamp,
			FromMe:    message.IsFromMe,
		})
	}

	response := &dto.ChatHistoryResponse{
		Success: true,
		Code:    http.StatusOK,
		Data: &dto.ChatHistoryResponseData{
			Phone:    phone,
			Messages: messages,
			Count:    result.Count,
			Limit:    limit,
		},
	}

	c.JSON(http.StatusOK, response)
}

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
		timestamp := ""
		if !chat.LastSeen.IsZero() {
			timestamp = chat.LastSeen.Format(time.RFC3339)
		}
		dtoChats[i] = dto.ChatInfo{
			JID:         chat.JID,
			Name:        chat.Name,
			Type:        chat.Type,
			LastMessage: chat.LastMessage,
			Timestamp:   timestamp,
			UnreadCount: chat.UnreadCount,
			Pinned:      chat.IsPinned,
			Muted:       chat.IsMuted,
			Archived:    chat.IsArchived,
		}
	}

	response := dto.NewListChatsSuccessResponse(dtoChats, req.Type)
	c.JSON(http.StatusOK, response)
}

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

	timestamp := ""
	if !chatInfo.LastSeen.IsZero() {
		timestamp = chatInfo.LastSeen.Format(time.RFC3339)
	}
	dtoChatInfo := dto.ChatInfo{
		JID:         chatInfo.JID,
		Name:        chatInfo.Name,
		Type:        chatInfo.Type,
		LastMessage: chatInfo.LastMessage,
		Timestamp:   timestamp,
		UnreadCount: chatInfo.UnreadCount,
		Pinned:      chatInfo.IsPinned,
		Muted:       chatInfo.IsMuted,
		Archived:    chatInfo.IsArchived,
	}

	response := dto.NewGetChatInfoSuccessResponse(dtoChatInfo)
	c.JSON(http.StatusOK, response)
}

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
			duration = 0
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
