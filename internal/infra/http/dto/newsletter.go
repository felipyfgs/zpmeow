package dto

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

type CreateNewsletterRequest struct {
	Name        string `json:"name" binding:"required" example:"My Newsletter"`
	Description string `json:"description,omitempty" example:"This is my newsletter"`
}

func (r CreateNewsletterRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if len(r.Name) > 100 {
		return fmt.Errorf("name must not exceed 100 characters")
	}
	if len(r.Description) > 500 {
		return fmt.Errorf("description must not exceed 500 characters")
	}
	return nil
}

type UpdateNewsletterRequest struct {
	Name        string `json:"name,omitempty" example:"Updated Newsletter"`
	Description string `json:"description,omitempty" example:"Updated description"`
}

func (r UpdateNewsletterRequest) Validate() error {
	if r.Name != "" && len(r.Name) > 100 {
		return fmt.Errorf("name must not exceed 100 characters")
	}
	if len(r.Description) > 500 {
		return fmt.Errorf("description must not exceed 500 characters")
	}
	return nil
}

type SubscribeNewsletterRequest struct {
	NewsletterJID string `json:"newsletter_jid" binding:"required" example:"120363025246125486@newsletter"`
}

func (r SubscribeNewsletterRequest) Validate() error {
	if strings.TrimSpace(r.NewsletterJID) == "" {
		return fmt.Errorf("newsletter_jid is required")
	}
	if !strings.Contains(r.NewsletterJID, "@newsletter") {
		return fmt.Errorf("invalid newsletter JID format")
	}
	return nil
}

type UnsubscribeNewsletterRequest struct {
	NewsletterJID string `json:"newsletter_jid" binding:"required" example:"120363025246125486@newsletter"`
}

func (r UnsubscribeNewsletterRequest) Validate() error {
	if strings.TrimSpace(r.NewsletterJID) == "" {
		return fmt.Errorf("newsletter_jid is required")
	}
	if !strings.Contains(r.NewsletterJID, "@newsletter") {
		return fmt.Errorf("invalid newsletter JID format")
	}
	return nil
}

type NewsletterErrorResponse struct {
	Code    string `json:"code" example:"NEWSLETTER_NOT_FOUND"`
	Message string `json:"message" example:"Newsletter not found"`
	Details string `json:"details" example:"Newsletter with JID '120363025246125486@newsletter' not found"`
}

type NewsletterInfo struct {
	JID             string    `json:"jid" example:"120363025246125486@newsletter"`
	Name            string    `json:"name" example:"My Newsletter"`
	Description     string    `json:"description,omitempty" example:"This is my newsletter"`
	Owner           string    `json:"owner,omitempty" example:"5511999999999@s.whatsapp.net"`
	CreatedAt       time.Time `json:"created_at,omitempty" example:"2023-01-01T12:00:00Z"`
	SubscriberCount int       `json:"subscriber_count" example:"100"`
	IsSubscribed    bool      `json:"is_subscribed" example:"true"`
	IsVerified      bool      `json:"is_verified" example:"false"`
}

type NewsletterResponse struct {
	Success bool                     `json:"success"`
	Code    int                      `json:"code"`
	Data    *NewsletterResponseData  `json:"data,omitempty"`
	Error   *NewsletterErrorResponse `json:"error,omitempty"`
}

type NewsletterResponseData struct {
	SessionId  string          `json:"session_id"`
	Action     string          `json:"action"`
	Status     string          `json:"status"`
	Message    string          `json:"message,omitempty"`
	Newsletter *NewsletterInfo `json:"newsletter,omitempty"`
	Timestamp  time.Time       `json:"timestamp"`
}

type NewsletterListResponse struct {
	Success bool                     `json:"success"`
	Code    int                      `json:"code"`
	Data    *NewsletterListData      `json:"data,omitempty"`
	Error   *NewsletterErrorResponse `json:"error,omitempty"`
}

type NewsletterListData struct {
	SessionId   string           `json:"session_id"`
	Newsletters []NewsletterInfo `json:"newsletters"`
	Count       int              `json:"count"`
	Total       int              `json:"total"`
}

type NewsletterActionResponse struct {
	Success bool                     `json:"success"`
	Code    int                      `json:"code"`
	Data    *NewsletterActionData    `json:"data,omitempty"`
	Error   *NewsletterErrorResponse `json:"error,omitempty"`
}

type NewsletterActionData struct {
	SessionId     string    `json:"session_id"`
	NewsletterJID string    `json:"newsletter_jid"`
	Action        string    `json:"action"`
	Status        string    `json:"status"`
	Message       string    `json:"message,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
}

func NewNewsletterErrorResponse(code int, errorCode, message, details string) *NewsletterResponse {
	return &NewsletterResponse{
		Success: false,
		Code:    code,
		Error: &NewsletterErrorResponse{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}
}

func NewNewsletterSuccessResponse(sessionID, action, message string, newsletter *NewsletterInfo) *NewsletterResponse {
	return &NewsletterResponse{
		Success: true,
		Code:    http.StatusOK,
		Data: &NewsletterResponseData{
			SessionId:  sessionID,
			Action:     action,
			Status:     "success",
			Message:    message,
			Newsletter: newsletter,
			Timestamp:  time.Now(),
		},
	}
}

func NewNewsletterListSuccessResponse(sessionID string, newsletters []NewsletterInfo) *NewsletterListResponse {
	return &NewsletterListResponse{
		Success: true,
		Code:    http.StatusOK,
		Data: &NewsletterListData{
			SessionId:   sessionID,
			Newsletters: newsletters,
			Count:       len(newsletters),
			Total:       len(newsletters),
		},
	}
}

func NewNewsletterListErrorResponse(code int, errorCode, message, details string) *NewsletterListResponse {
	return &NewsletterListResponse{
		Success: false,
		Code:    code,
		Error: &NewsletterErrorResponse{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}
}

func NewNewsletterActionSuccessResponse(sessionID, newsletterJID, action, message string) *NewsletterActionResponse {
	return &NewsletterActionResponse{
		Success: true,
		Code:    http.StatusOK,
		Data: &NewsletterActionData{
			SessionId:     sessionID,
			NewsletterJID: newsletterJID,
			Action:        action,
			Status:        "success",
			Message:       message,
			Timestamp:     time.Now(),
		},
	}
}

func NewNewsletterActionErrorResponse(code int, errorCode, message, details string) *NewsletterActionResponse {
	return &NewsletterActionResponse{
		Success: false,
		Code:    code,
		Error: &NewsletterErrorResponse{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}
}

type MarkViewedRequest struct {
	NewsletterJID string   `json:"newsletter_jid" binding:"required" example:"120363025246125486@newsletter"`
	ServerIDs     []string `json:"server_ids" binding:"required" example:"[\"1\", \"2\", \"3\"]"`
}

func (r MarkViewedRequest) Validate() error {
	if strings.TrimSpace(r.NewsletterJID) == "" {
		return fmt.Errorf("newsletter_jid is required")
	}
	if len(r.ServerIDs) == 0 {
		return fmt.Errorf("server_ids list cannot be empty")
	}
	return nil
}

type SendReactionRequest struct {
	NewsletterJID string `json:"newsletter_jid" binding:"required" example:"120363025246125486@newsletter"`
	ServerID      string `json:"server_id" binding:"required" example:"1"`
	MessageID     string `json:"message_id" binding:"required" example:"msg123"`
	Reaction      string `json:"reaction" binding:"required" example:"👍"`
}

func (r SendReactionRequest) Validate() error {
	if strings.TrimSpace(r.NewsletterJID) == "" {
		return fmt.Errorf("newsletter_jid is required")
	}
	if strings.TrimSpace(r.ServerID) == "" {
		return fmt.Errorf("server_id is required")
	}
	if strings.TrimSpace(r.MessageID) == "" {
		return fmt.Errorf("message_id is required")
	}
	if strings.TrimSpace(r.Reaction) == "" {
		return fmt.Errorf("reaction is required")
	}
	return nil
}

type ToggleMuteRequest struct {
	NewsletterJID string `json:"newsletter_jid" binding:"required" example:"120363025246125486@newsletter"`
	Mute          bool   `json:"mute" example:"true"`
}

func (r ToggleMuteRequest) Validate() error {
	if strings.TrimSpace(r.NewsletterJID) == "" {
		return fmt.Errorf("newsletter_jid is required")
	}
	return nil
}

type UploadNewsletterMediaRequest struct {
	NewsletterJID string `json:"newsletter_jid" binding:"required" example:"120363025246125486@newsletter"`
	MediaData     string `json:"media_data" binding:"required" example:"base64_encoded_media"`
	MediaType     string `json:"media_type" binding:"required" example:"image/jpeg"`
	Caption       string `json:"caption,omitempty" example:"Media caption"`
}

func (r UploadNewsletterMediaRequest) Validate() error {
	if strings.TrimSpace(r.NewsletterJID) == "" {
		return fmt.Errorf("newsletter_jid is required")
	}
	if strings.TrimSpace(r.MediaData) == "" {
		return fmt.Errorf("media_data is required")
	}
	if strings.TrimSpace(r.MediaType) == "" {
		return fmt.Errorf("media_type is required")
	}
	return nil
}

type NewsletterInfoResponse struct {
	Success bool                     `json:"success"`
	Code    int                      `json:"code"`
	Data    *NewsletterInfo          `json:"data,omitempty"`
	Error   *NewsletterErrorResponse `json:"error,omitempty"`
}

type CreateNewsletterResponse struct {
	Success bool                     `json:"success"`
	Code    int                      `json:"code"`
	Data    *NewsletterInfo          `json:"data,omitempty"`
	Error   *NewsletterErrorResponse `json:"error,omitempty"`
}

type SendNewsletterMessageRequest struct {
	NewsletterJID string `json:"newsletter_jid" binding:"required" example:"120363025246125486@newsletter"`
	Message       string `json:"message" binding:"required" example:"Hello newsletter subscribers!"`
	MediaType     string `json:"media_type,omitempty" example:"text"`
	MediaData     string `json:"media_data,omitempty" example:"base64_encoded_media"`
}

func (r SendNewsletterMessageRequest) Validate() error {
	if strings.TrimSpace(r.NewsletterJID) == "" {
		return fmt.Errorf("newsletter_jid is required")
	}
	if strings.TrimSpace(r.Message) == "" && strings.TrimSpace(r.MediaData) == "" {
		return fmt.Errorf("either message or media_data is required")
	}
	return nil
}

type SendNewsletterMessageResponse struct {
	Success bool                     `json:"success"`
	Code    int                      `json:"code"`
	Data    *NewsletterMessageData   `json:"data,omitempty"`
	Error   *NewsletterErrorResponse `json:"error,omitempty"`
}

type NewsletterMessageData struct {
	SessionId     string    `json:"session_id"`
	NewsletterJID string    `json:"newsletter_jid"`
	MessageID     string    `json:"message_id"`
	ServerID      string    `json:"server_id"`
	Status        string    `json:"status"`
	Timestamp     time.Time `json:"timestamp"`
}
