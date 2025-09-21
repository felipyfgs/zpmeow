package dto

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)


type SetWebhookRequest struct {
	URL    string   `json:"url" binding:"required" example:"https://example.com/webhook"`
	Events []string `json:"events,omitempty" example:"[\"message\", \"status\"]"`
}

func (r SetWebhookRequest) Validate() error {
	if strings.TrimSpace(r.URL) == "" {
		return fmt.Errorf("url is required")
	}
	if !strings.HasPrefix(r.URL, "http://") && !strings.HasPrefix(r.URL, "https://") {
		return fmt.Errorf("url must be a valid HTTP or HTTPS URL")
	}
	return nil
}

type GetWebhookRequest struct {
	SessionID string `json:"session_id,omitempty"`
}

type DeleteWebhookRequest struct {
	SessionID string `json:"session_id,omitempty"`
}

type TestWebhookRequest struct {
	URL     string                 `json:"url" binding:"required" example:"https://example.com/webhook"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

func (r TestWebhookRequest) Validate() error {
	if strings.TrimSpace(r.URL) == "" {
		return fmt.Errorf("url is required")
	}
	if !strings.HasPrefix(r.URL, "http://") && !strings.HasPrefix(r.URL, "https://") {
		return fmt.Errorf("url must be a valid HTTP or HTTPS URL")
	}
	return nil
}

type RegisterWebhookRequest struct {
	URL    string   `json:"url" binding:"required" example:"https://example.com/webhook"`
	Events []string `json:"events,omitempty" example:"[\"message\", \"status\"]"`
}

func (r RegisterWebhookRequest) Validate() error {
	if strings.TrimSpace(r.URL) == "" {
		return fmt.Errorf("url is required")
	}
	if !strings.HasPrefix(r.URL, "http://") && !strings.HasPrefix(r.URL, "https://") {
		return fmt.Errorf("url must be a valid HTTP or HTTPS URL")
	}
	return nil
}


type WebhookErrorResponse struct {
	Code    string `json:"code" example:"WEBHOOK_NOT_FOUND"`
	Message string `json:"message" example:"Webhook not found"`
	Details string `json:"details" example:"No webhook configured for this session"`
}

type WebhookInfo struct {
	SessionID string    `json:"session_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	URL       string    `json:"url" example:"https://example.com/webhook"`
	Events    []string  `json:"events,omitempty" example:"[\"message\", \"status\"]"`
	IsActive  bool      `json:"is_active" example:"true"`
	CreatedAt time.Time `json:"created_at" example:"2023-01-01T12:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2023-01-01T12:00:00Z"`
}

type WebhookResponse struct {
	Success bool                  `json:"success"`
	Code    int                   `json:"code"`
	Status  int                   `json:"status,omitempty"`
	Message string                `json:"message,omitempty"`
	Data    *WebhookResponseData  `json:"data,omitempty"`
	Error   *WebhookErrorResponse `json:"error,omitempty"`
}

type WebhookResponseData struct {
	SessionID string       `json:"session_id"`
	Action    string       `json:"action,omitempty"`
	Status    string       `json:"status"`
	Message   string       `json:"message,omitempty"`
	Webhook   *WebhookInfo `json:"webhook,omitempty"`
	Timestamp time.Time    `json:"timestamp"`
	CreatedAt time.Time    `json:"created_at,omitempty"`
	Events    []string     `json:"events,omitempty"`
	URL       string       `json:"url,omitempty"`
}

type WebhookListResponse struct {
	Success bool                  `json:"success"`
	Code    int                   `json:"code"`
	Data    *WebhookListData      `json:"data,omitempty"`
	Error   *WebhookErrorResponse `json:"error,omitempty"`
}

type WebhookListData struct {
	Webhooks []WebhookInfo `json:"webhooks"`
	Count    int           `json:"count"`
}

type WebhookTestResponse struct {
	Success bool                  `json:"success"`
	Code    int                   `json:"code"`
	Data    *WebhookTestData      `json:"data,omitempty"`
	Error   *WebhookErrorResponse `json:"error,omitempty"`
}

type WebhookTestData struct {
	URL          string    `json:"url"`
	Status       string    `json:"status"`
	ResponseCode int       `json:"response_code,omitempty"`
	ResponseTime string    `json:"response_time,omitempty"`
	Message      string    `json:"message,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
}


func NewWebhookErrorResponse(code int, errorCode, message, details string) *WebhookResponse {
	return &WebhookResponse{
		Success: false,
		Code:    code,
		Error: &WebhookErrorResponse{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}
}

func NewWebhookSuccessResponse(sessionID, action, message string, webhook *WebhookInfo) *WebhookResponse {
	return &WebhookResponse{
		Success: true,
		Code:    http.StatusOK,
		Data: &WebhookResponseData{
			SessionID: sessionID,
			Action:    action,
			Status:    "success",
			Message:   message,
			Webhook:   webhook,
			Timestamp: time.Now(),
		},
	}
}

func NewWebhookListSuccessResponse(webhooks []WebhookInfo) *WebhookListResponse {
	return &WebhookListResponse{
		Success: true,
		Code:    http.StatusOK,
		Data: &WebhookListData{
			Webhooks: webhooks,
			Count:    len(webhooks),
		},
	}
}

func NewWebhookListErrorResponse(code int, errorCode, message, details string) *WebhookListResponse {
	return &WebhookListResponse{
		Success: false,
		Code:    code,
		Error: &WebhookErrorResponse{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}
}

func NewWebhookTestSuccessResponse(url, status, responseTime, message string, responseCode int) *WebhookTestResponse {
	return &WebhookTestResponse{
		Success: true,
		Code:    http.StatusOK,
		Data: &WebhookTestData{
			URL:          url,
			Status:       status,
			ResponseCode: responseCode,
			ResponseTime: responseTime,
			Message:      message,
			Timestamp:    time.Now(),
		},
	}
}

func NewWebhookTestErrorResponse(code int, errorCode, message, details string) *WebhookTestResponse {
	return &WebhookTestResponse{
		Success: false,
		Code:    code,
		Error: &WebhookErrorResponse{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}
}


type WebhookEvent struct {
	Type      string                 `json:"type" example:"message"`
	SessionID string                 `json:"session_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Timestamp time.Time              `json:"timestamp" example:"2023-01-01T12:00:00Z"`
	Data      map[string]interface{} `json:"data"`
}

type MessageWebhookEvent struct {
	Type      string    `json:"type" example:"message"`
	SessionID string    `json:"session_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Timestamp time.Time `json:"timestamp" example:"2023-01-01T12:00:00Z"`
	Data      struct {
		MessageID string `json:"message_id" example:"msg_123"`
		From      string `json:"from" example:"5511999999999@s.whatsapp.net"`
		To        string `json:"to" example:"5511888888888@s.whatsapp.net"`
		Type      string `json:"type" example:"text"`
		Content   string `json:"content" example:"Hello, World!"`
		FromMe    bool   `json:"from_me" example:"false"`
	} `json:"data"`
}

type StatusWebhookEvent struct {
	Type      string    `json:"type" example:"status"`
	SessionID string    `json:"session_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Timestamp time.Time `json:"timestamp" example:"2023-01-01T12:00:00Z"`
	Data      struct {
		Status    string `json:"status" example:"connected"`
		Message   string `json:"message,omitempty" example:"Session connected successfully"`
		DeviceJID string `json:"device_jid,omitempty" example:"5511999999999.0:1@s.whatsapp.net"`
	} `json:"data"`
}

type StandardWebhookCreateResponse = WebhookResponse
type StandardWebhookData = WebhookResponseData
type StandardWebhookResponse = WebhookResponse

type SupportedEventsResponse struct {
	Success bool                  `json:"success"`
	Code    int                   `json:"code"`
	Status  int                   `json:"status,omitempty"`
	Message string                `json:"message,omitempty"`
	Data    *SupportedEventsData  `json:"data,omitempty"`
	Error   *WebhookErrorResponse `json:"error,omitempty"`
}

type SupportedEventsData struct {
	Events []string `json:"events"`
	Count  int      `json:"count"`
}
