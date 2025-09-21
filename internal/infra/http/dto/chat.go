package dto

import (
	"fmt"
	"net/http"
	"strings"
)

type SetPresenceRequest struct {
	Phone string `json:"phone" binding:"required" example:"5511999999999"`
	State string `json:"state" binding:"required" example:"available"`
	Media string `json:"media,omitempty" example:""`
}

func (r SetPresenceRequest) Validate() error {
	if strings.TrimSpace(r.Phone) == "" {
		return fmt.Errorf("phone is required")
	}
	if strings.TrimSpace(r.State) == "" {
		return fmt.Errorf("state is required")
	}
	validStates := []string{"available", "unavailable", "composing", "recording", "paused"}
	for _, validState := range validStates {
		if r.State == validState {
			return nil
		}
	}
	return fmt.Errorf("invalid state, must be one of: %s", strings.Join(validStates, ", "))
}

type SetDisappearingTimerRequest struct {
	JID   string `json:"jid" binding:"required" example:"5511999999999@s.whatsapp.net"`
	Timer string `json:"timer" binding:"required" example:"24h"`
}

func (r SetDisappearingTimerRequest) Validate() error {
	if strings.TrimSpace(r.JID) == "" {
		return fmt.Errorf("jid is required")
	}
	if strings.TrimSpace(r.Timer) == "" {
		return fmt.Errorf("timer is required")
	}
	validTimers := []string{"off", "0", "24h", "7d", "90d"}
	for _, validTimer := range validTimers {
		if r.Timer == validTimer {
			return nil
		}
	}
	return fmt.Errorf("invalid timer, must be one of: %s", strings.Join(validTimers, ", "))
}

type ListChatsRequest struct {
	Type string `json:"type,omitempty" example:"all"`
}

func (r ListChatsRequest) Validate() error {
	if r.Type == "" {
		return nil
	}
	validTypes := []string{"all", "groups", "contacts"}
	for _, validType := range validTypes {
		if r.Type == validType {
			return nil
		}
	}
	return fmt.Errorf("invalid type, must be one of: %s", strings.Join(validTypes, ", "))
}

type GetChatInfoRequest struct {
	JID string `json:"jid" binding:"required" example:"5511999999999@s.whatsapp.net"`
}

func (r GetChatInfoRequest) Validate() error {
	if strings.TrimSpace(r.JID) == "" {
		return fmt.Errorf("jid is required")
	}
	return nil
}

type PinChatRequest struct {
	JID    string `json:"jid" binding:"required" example:"5511999999999@s.whatsapp.net"`
	Pinned bool   `json:"pinned" example:"true"`
}

func (r PinChatRequest) Validate() error {
	if strings.TrimSpace(r.JID) == "" {
		return fmt.Errorf("jid is required")
	}
	return nil
}

type MuteChatRequest struct {
	JID      string `json:"jid" binding:"required" example:"5511999999999@s.whatsapp.net"`
	Muted    bool   `json:"muted" example:"true"`
	Duration string `json:"duration,omitempty" example:"8h"`
}

func (r MuteChatRequest) Validate() error {
	if strings.TrimSpace(r.JID) == "" {
		return fmt.Errorf("jid is required")
	}
	if r.Muted && r.Duration != "" {
		validDurations := []string{"1h", "8h", "1w", "forever"}
		for _, validDuration := range validDurations {
			if r.Duration == validDuration {
				return nil
			}
		}
		return fmt.Errorf("invalid duration, must be one of: %s", strings.Join(validDurations, ", "))
	}
	return nil
}

type ArchiveChatRequest struct {
	JID      string `json:"jid" binding:"required" example:"5511999999999@s.whatsapp.net"`
	Archived bool   `json:"archived" example:"true"`
}

func (r ArchiveChatRequest) Validate() error {
	if strings.TrimSpace(r.JID) == "" {
		return fmt.Errorf("jid is required")
	}
	return nil
}

type ChatErrorResponse struct {
	Code    string `json:"code" example:"INVALID_JID"`
	Message string `json:"message" example:"Invalid JID format"`
	Details string `json:"details" example:"JID must include domain"`
}

type ChatData struct {
	Phone   string `json:"phone"`
	JID     string `json:"jid,omitempty"`
	Action  string `json:"action"`
	Message string `json:"message,omitempty"`
}

type ChatResponse struct {
	Success bool               `json:"success"`
	Code    int                `json:"code"`
	Data    *ChatData          `json:"data,omitempty"`
	Error   *ChatErrorResponse `json:"error,omitempty"`
}

type ChatInfo struct {
	JID         string `json:"jid" example:"5511999999999@s.whatsapp.net"`
	Name        string `json:"name" example:"João Silva"`
	Type        string `json:"type" example:"contact"`
	LastMessage string `json:"last_message,omitempty" example:"Hello!"`
	Timestamp   string `json:"timestamp,omitempty" example:"2023-01-01T00:00:00Z"`
	UnreadCount int    `json:"unread_count" example:"3"`
	Pinned      bool   `json:"pinned" example:"false"`
	Muted       bool   `json:"muted" example:"false"`
	Archived    bool   `json:"archived" example:"false"`
}

type ListChatsResponse struct {
	Success bool               `json:"success"`
	Code    int                `json:"code"`
	Data    *ListChatsData     `json:"data,omitempty"`
	Error   *ChatErrorResponse `json:"error,omitempty"`
}

type ListChatsData struct {
	Chats []ChatInfo `json:"chats"`
	Type  string     `json:"type"`
	Count int        `json:"count"`
}

type GetChatInfoResponse struct {
	Success bool               `json:"success"`
	Code    int                `json:"code"`
	Data    *ChatInfo          `json:"data,omitempty"`
	Error   *ChatErrorResponse `json:"error,omitempty"`
}

type ChatHistoryData struct {
	MessageID string `json:"message_id" example:"msg_123"`
	From      string `json:"from" example:"5511999999999@s.whatsapp.net"`
	To        string `json:"to" example:"5511888888888@s.whatsapp.net"`
	Type      string `json:"type" example:"text"`
	Content   string `json:"content" example:"Hello!"`
	Timestamp int64  `json:"timestamp" example:"1640995200"`
	FromMe    bool   `json:"from_me" example:"false"`
}

type ChatHistoryResponseData struct {
	Phone    string            `json:"phone"`
	Messages []ChatHistoryData `json:"messages"`
	Count    int               `json:"count"`
	Limit    int               `json:"limit"`
}

type ChatHistoryResponse struct {
	Success bool                     `json:"success"`
	Code    int                      `json:"code"`
	Data    *ChatHistoryResponseData `json:"data,omitempty"`
	Error   *ChatErrorResponse       `json:"error,omitempty"`
}

func NewChatErrorResponse(code int, errorCode, message, details string) *ChatResponse {
	return &ChatResponse{
		Success: false,
		Code:    code,
		Error: &ChatErrorResponse{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}
}

func NewChatSuccessResponse(phone, jid, action string) *ChatResponse {
	return &ChatResponse{
		Success: true,
		Code:    http.StatusOK,
		Data: &ChatData{
			Phone:  phone,
			JID:    jid,
			Action: action,
		},
	}
}

func NewListChatsErrorResponse(code int, errorCode, message, details string) *ListChatsResponse {
	return &ListChatsResponse{
		Success: false,
		Code:    code,
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
		Code:    http.StatusOK,
		Data: &ListChatsData{
			Chats: chats,
			Type:  chatType,
			Count: len(chats),
		},
	}
}

func NewGetChatInfoErrorResponse(code int, errorCode, message, details string) *GetChatInfoResponse {
	return &GetChatInfoResponse{
		Success: false,
		Code:    code,
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
		Code:    http.StatusOK,
		Data:    &chatInfo,
	}
}
