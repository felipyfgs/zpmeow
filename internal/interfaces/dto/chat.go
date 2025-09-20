package dto

import (
	"time"
)

type SetPresenceRequest struct {
	Phone string `json:"phone,omitempty" example:"5511999999999"`
	State string `json:"state" binding:"required" example:"available"` // available, unavailable, composing, recording, paused
	Media string `json:"media,omitempty" example:""`                   // Optional media type for composing state
}

type DownloadMediaRequest struct {
	MessageID string `json:"message_id" binding:"required" example:"msg_123"`
}

type ListChatsRequest struct {
	Type string `json:"type,omitempty" example:"all"` // "all", "groups", "contacts"
}

type GetChatInfoRequest struct {
	JID string `json:"jid" binding:"required" example:"5511999999999@s.meow.net"`
}

type SetDisappearingTimerRequest struct {
	JID   string `json:"jid" binding:"required" example:"5511999999999@s.meow.net"`
	Timer string `json:"timer" binding:"required" example:"24h"` // "off", "24h", "7d", "90d"
}

type PinChatRequest struct {
	JID    string `json:"jid" binding:"required" example:"5511999999999@s.meow.net"`
	Pinned bool   `json:"pinned" example:"true"`
}

type MuteChatRequest struct {
	JID      string `json:"jid" binding:"required" example:"5511999999999@s.meow.net"`
	Muted    bool   `json:"muted" example:"true"`
	Duration string `json:"duration,omitempty" example:"8h"` // "1h", "8h", "1w", "forever" (only when muted=true)
}

type ArchiveChatRequest struct {
	JID      string `json:"jid" binding:"required" example:"5511999999999@s.meow.net"`
	Archived bool   `json:"archived" example:"true"`
}

type ChatResponse struct {
	Success bool               `json:"success"`
	Code    int                `json:"code"`
	Data    ChatData           `json:"data"`
	Error   *ChatErrorResponse `json:"error,omitempty"`
}

type ChatInfo struct {
	JID         string    `json:"jid" example:"5511999999999@s.meow.net"`
	Name        string    `json:"name" example:"João Silva"`
	Type        string    `json:"type" example:"contact"` // "contact", "group"
	LastMessage string    `json:"last_message,omitempty" example:"Hello!"`
	Timestamp   time.Time `json:"timestamp,omitempty" example:"2023-01-01T00:00:00Z"`
	UnreadCount int       `json:"unread_count" example:"3"`
	Pinned      bool      `json:"pinned" example:"false"`
	Muted       bool      `json:"muted" example:"false"`
	Archived    bool      `json:"archived" example:"false"`
}

type ListChatsResponse struct {
	Success bool               `json:"success" example:"true"`
	Code    int                `json:"code" example:"200"`
	Data    ListChatsData      `json:"data"`
	Error   *ChatErrorResponse `json:"error,omitempty"`
}

type ListChatsData struct {
	Chats []ChatInfo `json:"chats"`
	Count int        `json:"count" example:"25"`
	Type  string     `json:"type" example:"all"`
}

type GetChatInfoResponse struct {
	Success bool               `json:"success" example:"true"`
	Code    int                `json:"code" example:"200"`
	Data    ChatInfo           `json:"data"`
	Error   *ChatErrorResponse `json:"error,omitempty"`
}

type ChatData struct {
	Phone     string    `json:"phone" example:"5511999999999"`
	MessageID string    `json:"message_id,omitempty" example:"msg_123"`
	Action    string    `json:"action" example:"mark_read"`
	Status    string    `json:"status" example:"success"`
	Timestamp time.Time `json:"timestamp" example:"2023-01-01T00:00:00Z"`
}

type ChatErrorResponse struct {
	Code    string `json:"code" example:"INVALID_PHONE"`
	Message string `json:"message" example:"Invalid phone number format"`
	Details string `json:"details,omitempty" example:"Phone number must include country code"`
}

type MediaDownloadResponse struct {
	Success   bool   `json:"success" example:"true"`
	Code      int    `json:"code" example:"200"`
	MessageID string `json:"message_id" example:"msg_123"`
	MediaType string `json:"media_type" example:"image"`
	MimeType  string `json:"mime_type" example:"image/jpeg"`
	Data      []byte `json:"data"` // Base64 encoded media data
	Size      int    `json:"size" example:"1024"`
}

type ChatHistoryResponse struct {
	Success bool                    `json:"success" example:"true"`
	Code    int                     `json:"code" example:"200"`
	Data    ChatHistoryResponseData `json:"data"`
	Error   *ChatErrorResponse      `json:"error,omitempty"`
}

type ChatHistoryResponseData struct {
	Phone    string            `json:"phone" example:"5511999999999"`
	Messages []ChatHistoryData `json:"messages"`
	Count    int               `json:"count" example:"10"`
	Limit    int               `json:"limit" example:"50"`
}

type ChatHistoryData struct {
	MessageID   string    `json:"message_id" example:"msg_123456789"`
	Phone       string    `json:"phone" example:"5511999999999"`
	FromPhone   string    `json:"from_phone" example:"5511888888888"`
	MessageType string    `json:"message_type" example:"text"`
	Content     string    `json:"content" example:"Hello, World!"`
	MediaURL    string    `json:"media_url,omitempty" example:"https://example.com/image.jpg"`
	Timestamp   time.Time `json:"timestamp" example:"2023-01-01T00:00:00Z"`
	IsFromMe    bool      `json:"is_from_me" example:"false"`
}

func NewChatSuccessResponse(phone, messageID, action string) *ChatResponse {
	return &ChatResponse{
		Success: true,
		Code:    200,
		Data: ChatData{
			Phone:     phone,
			MessageID: messageID,
			Action:    action,
			Status:    "success",
			Timestamp: time.Now(),
		},
	}
}

func NewChatErrorResponse(code int, errorCode, message, details string) *ChatResponse {
	return &ChatResponse{
		Success: false,
		Code:    code,
		Data: ChatData{
			Status:    "error",
			Timestamp: time.Now(),
		},
		Error: &ChatErrorResponse{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}
}

func NewListChatsSuccessResponse(chats []ChatInfo, chatType string) *ListChatsResponse {
	return &ListChatsResponse{
		Success: true,
		Code:    200,
		Data: ListChatsData{
			Chats: chats,
			Count: len(chats),
			Type:  chatType,
		},
	}
}

func NewListChatsErrorResponse(code int, errorCode, message, details string) *ListChatsResponse {
	return &ListChatsResponse{
		Success: false,
		Code:    code,
		Data: ListChatsData{
			Chats: []ChatInfo{},
			Count: 0,
		},
		Error: &ChatErrorResponse{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}
}

func NewGetChatInfoSuccessResponse(chatInfo ChatInfo) *GetChatInfoResponse {
	return &GetChatInfoResponse{
		Success: true,
		Code:    200,
		Data:    chatInfo,
	}
}

func NewGetChatInfoErrorResponse(code int, errorCode, message, details string) *GetChatInfoResponse {
	return &GetChatInfoResponse{
		Success: false,
		Code:    code,
		Data:    ChatInfo{},
		Error: &ChatErrorResponse{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}
}
