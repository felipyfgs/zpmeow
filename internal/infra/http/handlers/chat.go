package handlers

import (
	"strconv"
	"strings"
	"time"

	"zpmeow/internal/application"
	"zpmeow/internal/infra/http/dto"
	"zpmeow/internal/infra/wmeow"

	"github.com/gofiber/fiber/v2"
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

// SetPresence godoc
// @Summary Set chat presence
// @Description Sets presence status (typing, recording, etc.) for a chat
// @Tags Chat
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.SetPresenceRequest true "Presence request"
// @Success 200 {object} dto.ChatResponse "Presence set successfully"
// @Failure 400 {object} dto.ChatResponse "Invalid request data"
// @Failure 401 {object} dto.ChatResponse "Unauthorized - Invalid API key" "Invalid request data"
// @Failure 404 {object} dto.ChatResponse "Session not found"
// @Failure 500 {object} dto.ChatResponse "Failed to set presence"
// @Router /session/{sessionId}/chat/presence [post]
// @Router /session/{sessionId}/presences/typing [post]
// @Router /session/{sessionId}/presences/recording [post]
func (h *ChatHandler) SetPresence(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	var req dto.SetPresenceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewChatErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	if req.State == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewChatErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_STATE",
			"State is required",
			"Valid states: available, unavailable, composing, recording, paused",
		))
	}

	ctx := c.Context()
	err := h.wmeowService.SetPresence(ctx, sessionID, req.Phone, req.State, req.Media)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewChatErrorResponse(
			fiber.StatusInternalServerError,
			"SET_PRESENCE_FAILED",
			"Failed to set presence",
			err.Error(),
		))
	}

	response := dto.NewChatSuccessResponse(req.Phone, "", "set_presence")
	return c.Status(fiber.StatusOK).JSON(response)
}

func (h *ChatHandler) DownloadImage(c *fiber.Ctx) error {
	return h.downloadMedia(c, "image")
}

func (h *ChatHandler) DownloadVideo(c *fiber.Ctx) error {
	return h.downloadMedia(c, "video")
}

func (h *ChatHandler) DownloadAudio(c *fiber.Ctx) error {
	return h.downloadMedia(c, "audio")
}

func (h *ChatHandler) DownloadDocument(c *fiber.Ctx) error {
	return h.downloadMedia(c, "document")
}

func (h *ChatHandler) downloadMedia(c *fiber.Ctx, mediaType string) error {
	sessionID := c.Params("sessionId")

	var req dto.DownloadMediaRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewChatErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	if req.MessageID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewChatErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_MESSAGE_ID",
			"Message ID is required",
			"",
		))
	}

	ctx := c.Context()
	data, mimeType, err := h.wmeowService.DownloadMedia(ctx, sessionID, req.MessageID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewChatErrorResponse(
			fiber.StatusInternalServerError,
			"DOWNLOAD_FAILED",
			"Failed to download media",
			err.Error(),
		))
	}

	response := &dto.MediaDownloadResponse{
		Success: true,
		Code:    fiber.StatusOK,
		Data: &dto.MediaDownloadData{
			MediaID:  req.MessageID,
			Type:     mediaType,
			MimeType: mimeType,
			Data:     data,
			Size:     int64(len(data)),
		},
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// GetChatHistory godoc
// @Summary Get chat history
// @Description Retrieves message history for a specific chat
// @Tags Chat
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param phone query string true "Phone number or JID"
// @Param limit query int false "Number of messages to retrieve" default(50)
// @Success 200 {object} dto.ChatResponse "Chat history"
// @Failure 400 {object} dto.ChatResponse "Invalid request data"
// @Failure 401 {object} dto.ChatResponse "Unauthorized - Invalid API key" "Invalid request data"
// @Failure 404 {object} dto.ChatResponse "Session not found"
// @Failure 500 {object} dto.ChatResponse "Failed to get chat history"
// @Router /session/{sessionId}/chat/history [get]
func (h *ChatHandler) GetChatHistory(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")
	phone := c.Query("phone")
	limitStr := c.Query("limit", "50")

	if phone == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewChatErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_PHONE",
			"Phone number is required",
			"",
		))
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewChatErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_LIMIT",
			"Invalid limit parameter",
			err.Error(),
		))
	}

	ctx := c.Context()
	req := application.GetChatHistoryRequest{
		SessionID: sessionID,
		Phone:     phone,
		Limit:     limit,
		Offset:    0,
	}

	result, err := h.chatService.GetChatHistory(ctx, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewChatErrorResponse(
			fiber.StatusInternalServerError,
			"GET_CHAT_HISTORY_FAILED",
			"Failed to get chat history",
			err.Error(),
		))
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
		Code:    fiber.StatusOK,
		Data: &dto.ChatHistoryResponseData{
			Phone:    phone,
			Messages: messages,
			Count:    result.Count,
			Limit:    limit,
		},
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func (h *ChatHandler) SetDisappearingTimer(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	var req dto.SetDisappearingTimerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewChatErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	if req.JID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewChatErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_JID",
			"JID is required",
			"",
		))
	}

	if req.Timer == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewChatErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_TIMER",
			"Timer is required",
			"Valid values: off, 24h, 7d, 90d",
		))
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
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewChatErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_TIMER",
			"Invalid timer value",
			"Valid values: off, 24h, 7d, 90d",
		))
	}

	ctx := c.Context()
	err := h.wmeowService.SetDisappearingTimer(ctx, sessionID, req.JID, timer)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewChatErrorResponse(
			fiber.StatusInternalServerError,
			"SET_TIMER_FAILED",
			"Failed to set disappearing timer",
			err.Error(),
		))
	}

	response := dto.NewChatSuccessResponse(req.JID, "", "set_disappearing_timer")
	return c.Status(fiber.StatusOK).JSON(response)
}

// ListChats godoc
// @Summary List chats
// @Description Retrieves a list of chats for a session
// @Tags Chat
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.ListChatsRequest false "List chats request"
// @Success 200 {object} dto.ChatResponse "Chats list"
// @Failure 400 {object} dto.ChatResponse "Invalid request data"
// @Failure 401 {object} dto.ChatResponse "Unauthorized - Invalid API key" "Invalid request data"
// @Failure 404 {object} dto.ChatResponse "Session not found"
// @Failure 500 {object} dto.ChatResponse "Failed to list chats"
// @Router /session/{sessionId}/chat/list [post]
func (h *ChatHandler) ListChats(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	var req dto.ListChatsRequest
	if err := c.BodyParser(&req); err != nil {
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
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewListChatsErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_TYPE",
			"Invalid chat type",
			"Valid types: all, groups, contacts",
		))
	}

	ctx := c.Context()
	chats, err := h.wmeowService.ListChats(ctx, sessionID, req.Type)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewListChatsErrorResponse(
			fiber.StatusInternalServerError,
			"LIST_CHATS_FAILED",
			"Failed to list chats",
			err.Error(),
		))
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
	return c.Status(fiber.StatusOK).JSON(response)
}

func (h *ChatHandler) GetChatInfo(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	var req dto.GetChatInfoRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGetChatInfoErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	if req.JID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGetChatInfoErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_JID",
			"JID is required",
			"",
		))
	}

	ctx := c.Context()
	chatInfo, err := h.wmeowService.GetChatInfo(ctx, sessionID, req.JID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewGetChatInfoErrorResponse(
			fiber.StatusInternalServerError,
			"GET_CHAT_INFO_FAILED",
			"Failed to get chat info",
			err.Error(),
		))
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
	return c.Status(fiber.StatusOK).JSON(response)
}

func (h *ChatHandler) PinChat(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	var req dto.PinChatRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewChatErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	if req.JID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewChatErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_JID",
			"JID is required",
			"",
		))
	}

	ctx := c.Context()
	err := h.wmeowService.PinChat(ctx, sessionID, req.JID, req.Pinned)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewChatErrorResponse(
			fiber.StatusInternalServerError,
			"PIN_CHAT_FAILED",
			"Failed to pin/unpin chat",
			err.Error(),
		))
	}

	action := "unpin_chat"
	if req.Pinned {
		action = "pin_chat"
	}

	response := dto.NewChatSuccessResponse(req.JID, "", action)
	return c.Status(fiber.StatusOK).JSON(response)
}

func (h *ChatHandler) MuteChat(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	var req dto.MuteChatRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewChatErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	if req.JID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewChatErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_JID",
			"JID is required",
			"",
		))
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
			return c.Status(fiber.StatusBadRequest).JSON(dto.NewChatErrorResponse(
				fiber.StatusBadRequest,
				"INVALID_DURATION",
				"Invalid mute duration",
				"Valid durations: 1h, 8h, 1w, forever",
			))
		}
	}

	ctx := c.Context()
	err := h.wmeowService.MuteChat(ctx, sessionID, req.JID, req.Muted, duration)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewChatErrorResponse(
			fiber.StatusInternalServerError,
			"MUTE_CHAT_FAILED",
			"Failed to mute/unmute chat",
			err.Error(),
		))
	}

	action := "unmute_chat"
	if req.Muted {
		action = "mute_chat"
	}

	response := dto.NewChatSuccessResponse(req.JID, "", action)
	return c.Status(fiber.StatusOK).JSON(response)
}

func (h *ChatHandler) ArchiveChat(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	var req dto.ArchiveChatRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewChatErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	if req.JID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewChatErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_JID",
			"JID is required",
			"",
		))
	}

	ctx := c.Context()
	err := h.wmeowService.ArchiveChat(ctx, sessionID, req.JID, req.Archived)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewChatErrorResponse(
			fiber.StatusInternalServerError,
			"ARCHIVE_CHAT_FAILED",
			"Failed to archive/unarchive chat",
			err.Error(),
		))
	}

	action := "unarchive_chat"
	if req.Archived {
		action = "archive_chat"
	}

	response := dto.NewChatSuccessResponse(req.JID, "", action)
	return c.Status(fiber.StatusOK).JSON(response)
}
