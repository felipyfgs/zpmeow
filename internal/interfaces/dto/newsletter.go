package dto

import (
	"fmt"
	"time"
)

type NewsletterInfo struct {
	JID             string `json:"jid"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	SubscriberCount int    `json:"subscriberCount"`
	CreatedAt       int64  `json:"createdAt"`
	IsVerified      bool   `json:"isVerified"`
	Picture         string `json:"picture,omitempty"`
	Verified        bool   `json:"verified"`
	UpdatedAt       string `json:"updated_at,omitempty"`
	OwnerJID        string `json:"owner_jid,omitempty"`
	Subscribers     int    `json:"subscribers"`
	Muted           bool   `json:"muted"`
}

type NewsletterList struct {
	Newsletters []NewsletterInfo `json:"newsletters"`
	Total       int              `json:"total"`
	Count       int              `json:"count"` // Additional field referenced in handlers
}

type NewsletterMessage struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
	MediaURL  string `json:"mediaUrl,omitempty"`
	MediaType string `json:"mediaType,omitempty"`
}

type NewsletterMessages struct {
	Messages []NewsletterMessage `json:"messages"`
	Total    int                 `json:"total"`
}

type GetNewsletterMessageUpdatesRequest struct {
	NewsletterJID string `json:"newsletter_jid" binding:"required" example:"120363025246125486@newsletter"`
}

type MarkViewedRequest struct {
	NewsletterJID string   `json:"newsletter_jid" binding:"required" example:"120363025246125486@newsletter"`
	MessageIDs    []string `json:"message_ids" binding:"required" example:"[\"msg1\", \"msg2\"]"`
}

type SendReactionRequest struct {
	NewsletterJID string `json:"newsletter_jid" binding:"required" example:"120363025246125486@newsletter"`
	MessageID     string `json:"message_id" binding:"required" example:"msg123"`
	ServerID      string `json:"server_id" binding:"required" example:"server123"`
	Reaction      string `json:"reaction" binding:"required" example:"👍"`
}

type ToggleMuteRequest struct {
	NewsletterJID string `json:"newsletter_jid" binding:"required" example:"120363025246125486@newsletter"`
	Mute          bool   `json:"mute" example:"true"`
}

type CreateNewsletterRequest struct {
	Name        string `json:"name" binding:"required" example:"My Newsletter"`
	Description string `json:"description,omitempty" example:"Newsletter description"`
	Picture     string `json:"picture,omitempty"` // Base64 encoded image
}

type SendNewsletterMessageRequest struct {
	NewsletterJID string `json:"newsletter_jid" binding:"required" example:"120363025246125486@newsletter"`
	Message       string `json:"message,omitempty" example:"Hello subscribers!"`
	MediaHandle   string `json:"media_handle,omitempty" example:"media_123"`
	MediaType     string `json:"media_type,omitempty" example:"image"`
}

type GetNewsletterMessagesRequest struct {
	NewsletterJID string `json:"newsletter_jid" binding:"required" example:"120363025246125486@newsletter"`
	Limit         int    `json:"limit,omitempty" example:"50"`
	Before        string `json:"before,omitempty" example:"msg_123"`
}

type UploadNewsletterMediaRequest struct {
	Data      string `json:"data" binding:"required"` // Base64 encoded media
	MediaType string `json:"media_type" binding:"required" example:"image"`
	Filename  string `json:"filename,omitempty" example:"image.jpg"`
}

type NewsletterResponse struct {
	Success bool           `json:"success"`
	Code    int            `json:"code"`
	Data    NewsletterData `json:"data"`
	Error   *ErrorInfo     `json:"error,omitempty"`
}

type NewsletterInfoResponse struct {
	Success bool           `json:"success"`
	Code    int            `json:"code"`
	Data    NewsletterInfo `json:"data"`
	Error   *ErrorInfo     `json:"error,omitempty"`
}

type NewsletterListResponse struct {
	Success bool           `json:"success"`
	Code    int            `json:"code"`
	Data    NewsletterList `json:"data"`
	Error   *ErrorInfo     `json:"error,omitempty"`
}

type CreateNewsletterResponse struct {
	Success bool                   `json:"success"`
	Code    int                    `json:"code"`
	Data    CreateNewsletterResult `json:"data"`
	Error   *ErrorInfo             `json:"error,omitempty"`
}

type CreateNewsletterResult struct {
	JID         string `json:"jid"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ServerID    string `json:"server_id"`
	Timestamp   string `json:"timestamp"`
}

type SendNewsletterMessageResponse struct {
	Success bool                      `json:"success"`
	Code    int                       `json:"code"`
	Data    SendNewsletterMessageData `json:"data"`
	Error   *ErrorInfo                `json:"error,omitempty"`
}

type NewsletterData struct {
	SessionID     string        `json:"sessionId"`
	Action        string        `json:"action"`
	Status        string        `json:"status"`
	Timestamp     time.Time     `json:"timestamp"`
	NewsletterJID string        `json:"newsletterJid,omitempty"`
	Updates       []interface{} `json:"updates,omitempty"`
	Message       string        `json:"message,omitempty"`
}

type SendNewsletterMessageData struct {
	SessionID     string    `json:"session_id"`
	NewsletterJID string    `json:"newsletter_jid"`
	MessageID     string    `json:"message_id,omitempty"`
	Action        string    `json:"action"`
	Status        string    `json:"status"`
	Timestamp     time.Time `json:"timestamp"`
}

func (r *GetNewsletterMessageUpdatesRequest) Validate() error {
	if r.NewsletterJID == "" {
		return fmt.Errorf("newsletter JID is required")
	}
	return nil
}

func (r *MarkViewedRequest) Validate() error {
	if r.NewsletterJID == "" {
		return fmt.Errorf("newsletter JID is required")
	}
	if len(r.MessageIDs) == 0 {
		return fmt.Errorf("at least one message ID is required")
	}
	return nil
}

func (r *SendReactionRequest) Validate() error {
	if r.NewsletterJID == "" {
		return fmt.Errorf("newsletter JID is required")
	}
	if r.MessageID == "" {
		return fmt.Errorf("message ID is required")
	}
	if r.ServerID == "" {
		return fmt.Errorf("server ID is required")
	}
	if r.Reaction == "" {
		return fmt.Errorf("reaction is required")
	}
	return nil
}

func (r *ToggleMuteRequest) Validate() error {
	if r.NewsletterJID == "" {
		return fmt.Errorf("newsletter JID is required")
	}
	return nil
}

func NewNewsletterSuccessResponse(sessionID, action string, updates []interface{}) *NewsletterResponse {
	return &NewsletterResponse{
		Success: true,
		Code:    200,
		Data: NewsletterData{
			SessionID: sessionID,
			Action:    action,
			Status:    "success",
			Timestamp: time.Now(),
			Updates:   updates,
		},
	}
}

func NewNewsletterErrorResponse(code int, errorCode, message, details string) *NewsletterResponse {
	return &NewsletterResponse{
		Success: false,
		Code:    code,
		Data: NewsletterData{
			Status:    "error",
			Timestamp: time.Now(),
		},
		Error: &ErrorInfo{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}
}

func NewNewsletterOperationResponse(sessionID, action, message string) *NewsletterResponse {
	return &NewsletterResponse{
		Success: true,
		Code:    200,
		Data: NewsletterData{
			SessionID: sessionID,
			Action:    action,
			Status:    "success",
			Timestamp: time.Now(),
			Message:   message,
		},
	}
}
