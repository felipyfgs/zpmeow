package dto

import (
	"errors"
	"net/url"
	"strings"
	"time"
)

type SessionInfo struct {
	ID         string    `json:"id" example:"default"`
	Name       string    `json:"name" example:"default"`
	Status     string    `json:"status" example:"connected"`
	DeviceJID  string    `json:"device_jid" example:"5511999999999@s.whatsapp.net"`
	ProxyURL   string    `json:"proxy_url,omitempty" example:"http://proxy.example.com:8080"`
	WebhookURL string    `json:"webhook_url,omitempty" example:"https://webhook.example.com/whatsapp"`
	Events     []string  `json:"events,omitempty" example:"message,status"`
	ApiKey     string    `json:"api_key" example:"550e8400-e29b-41d4-a716-446655440000"`
	CreatedAt  time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt  time.Time `json:"updated_at" example:"2023-01-01T00:00:00Z"`
}

type CreateSessionRequest struct {
	Name       string `json:"name" validate:"required,min=1,max=50" binding:"required" example:"default"`
	WebhookURL string `json:"webhook_url,omitempty" validate:"omitempty,webhook_url" example:"https://webhook.example.com/whatsapp"`
	Events     string `json:"events,omitempty" validate:"omitempty,max=500" example:"message,status"`
	ProxyURL   string `json:"proxy_url,omitempty" validate:"omitempty,url" example:"http://proxy.example.com:8080"`
}

func (r *CreateSessionRequest) Validate() error {
	name := strings.TrimSpace(r.Name)
	if name == "" {
		return errors.New("name is required")
	}
	if len(name) > 50 {
		return errors.New("name must not exceed 50 characters")
	}
	if r.WebhookURL != "" {
		if _, err := url.ParseRequestURI(r.WebhookURL); err != nil {
			return errors.New("invalid webhook_url")
		}
	}
	if r.ProxyURL != "" {
		if _, err := url.ParseRequestURI(r.ProxyURL); err != nil {
			return errors.New("invalid proxy_url")
		}
	}
	return nil
}

type PairPhoneRequest struct {
	Phone string `json:"phone" validate:"required,phone_number" binding:"required" example:"5511999999999"`
}

func (r *PairPhoneRequest) Validate() error {
	phone := strings.TrimSpace(r.Phone)
	if phone == "" {
		return errors.New("phone is required")
	}
	return nil
}

type SessionResponse struct {
	Success   bool        `json:"success"`
	Code      int         `json:"code"`
	Data      SessionData `json:"data"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

type SessionData struct {
	SessionID string        `json:"session_id" example:"session_123"`
	Action    string        `json:"action" example:"create"`
	Status    string        `json:"status" example:"success"`
	Timestamp time.Time     `json:"timestamp" example:"2023-01-01T00:00:00Z"`
	Session   *SessionInfo  `json:"session,omitempty"`
	Sessions  []SessionInfo `json:"sessions,omitempty"`
	QRCode    string        `json:"qr_code,omitempty"`
}

type CreateSessionResponse struct {
	Success   bool              `json:"success"`
	Code      int               `json:"code"`
	Data      SessionCreateData `json:"data"`
	Error     *ErrorInfo        `json:"error,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

type SessionCreateData struct {
	Action    string       `json:"action" example:"create"`
	Status    string       `json:"status" example:"success"`
	Timestamp time.Time    `json:"timestamp" example:"2023-01-01T00:00:00Z"`
	Session   *SessionInfo `json:"session"`
}

type ConnectSessionResponse struct {
	Success   bool               `json:"success"`
	Code      int                `json:"code"`
	Data      SessionConnectData `json:"data"`
	Error     *ErrorInfo         `json:"error,omitempty"`
	Timestamp time.Time          `json:"timestamp"`
}

type SessionConnectData struct {
	SessionID  string                 `json:"session_id" example:"session_123"`
	Action     string                 `json:"action" example:"connect"`
	Status     string                 `json:"status" example:"success"`
	Timestamp  time.Time              `json:"timestamp" example:"2023-01-01T00:00:00Z"`
	Session    *SessionInfo           `json:"session"`
	Connection *SessionConnectionInfo `json:"connection"`
	QRCode     string                 `json:"qr_code,omitempty"`
}

type SessionConnectionInfo struct {
	QRCode      string `json:"qr_code,omitempty"`
	Connected   bool   `json:"connected"`
	IsConnected bool   `json:"is_connected"`
}

type PairPhoneResponse struct {
	Success   bool                  `json:"success"`
	Code      int                   `json:"code"`
	Data      PairPhoneResponseData `json:"data"`
	Error     *ErrorInfo            `json:"error,omitempty"`
	Timestamp time.Time             `json:"timestamp"`
}

type PairPhoneResponseData struct {
	SessionID string    `json:"session_id" example:"session_123"`
	Action    string    `json:"action" example:"pair"`
	Status    string    `json:"status" example:"success"`
	Timestamp time.Time `json:"timestamp" example:"2023-01-01T00:00:00Z"`
	Phone     string    `json:"phone" example:"5511999999999"`
	Code      string    `json:"code,omitempty" example:"123456"`
}

type SessionStatusResponse struct {
	Success   bool                      `json:"success"`
	Code      int                       `json:"code"`
	Data      SessionStatusResponseData `json:"data"`
	Error     *ErrorInfo                `json:"error,omitempty"`
	Timestamp time.Time                 `json:"timestamp"`
}

type SessionStatusResponseData struct {
	SessionID     string    `json:"session_id" example:"session_123"`
	Action        string    `json:"action" example:"status"`
	Status        string    `json:"status" example:"success"`
	Timestamp     time.Time `json:"timestamp" example:"2023-01-01T00:00:00Z"`
	Name          string    `json:"name" example:"My Session"`
	SessionStatus string    `json:"session_status" example:"connected"`
	DeviceJID     string    `json:"device_jid,omitempty" example:"5511999999999@s.whatsapp.net"`
	IsConnected   bool      `json:"is_connected"`
	ClientStatus  string    `json:"client_status" example:"connected"`
	CreatedAt     time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt     time.Time `json:"updated_at" example:"2023-01-01T00:00:00Z"`
}

type SessionErrorResponse struct {
	Success   bool       `json:"success"`
	Code      int        `json:"code"`
	Error     *ErrorInfo `json:"error"`
	Timestamp time.Time  `json:"timestamp"`
}

type SessionListResponse struct {
	Success   bool        `json:"success"`
	Code      int         `json:"code"`
	Data      SessionData `json:"data"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

func NewSessionSuccessResponse(sessionID, action string, data interface{}) *SessionResponse {
	response := &SessionResponse{
		Success: true,
		Code:    200,
		Data: SessionData{
			SessionID: sessionID,
			Action:    action,
			Status:    "success",
			Timestamp: time.Now(),
		},
		Timestamp: time.Now(),
	}

	switch v := data.(type) {
	case *SessionInfo:
		response.Data.Session = v
	case []SessionInfo:
		response.Data.Sessions = v
	case string:
		response.Data.QRCode = v
	}

	return response
}

func NewSessionErrorResponse(code int, errorCode, message, details string) *SessionResponse {
	return &SessionResponse{
		Success: false,
		Code:    code,
		Data: SessionData{
			Status:    "error",
			Timestamp: time.Now(),
		},
		Error: &ErrorInfo{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now(),
	}
}

func NewCreateSessionSuccessResponse(sessionInfo *SessionInfo) *CreateSessionResponse {
	return &CreateSessionResponse{
		Success: true,
		Code:    201,
		Data: SessionCreateData{
			Action:    "create",
			Status:    "success",
			Timestamp: time.Now(),
			Session:   sessionInfo,
		},
		Timestamp: time.Now(),
	}
}

func NewConnectSessionSuccessResponse(sessionInfo *SessionInfo, connectionInfo *SessionConnectionInfo, qrCode string) *ConnectSessionResponse {
	return &ConnectSessionResponse{
		Success: true,
		Code:    200,
		Data: SessionConnectData{
			SessionID:  sessionInfo.ID,
			Action:     "connect",
			Status:     "success",
			Timestamp:  time.Now(),
			Session:    sessionInfo,
			Connection: connectionInfo,
			QRCode:     qrCode,
		},
		Timestamp: time.Now(),
	}
}

func NewPairPhoneSuccessResponse(sessionID, phone, code string) *PairPhoneResponse {
	return &PairPhoneResponse{
		Success: true,
		Code:    200,
		Data: PairPhoneResponseData{
			SessionID: sessionID,
			Action:    "pair",
			Status:    "success",
			Timestamp: time.Now(),
			Phone:     phone,
			Code:      code,
		},
		Timestamp: time.Now(),
	}
}

func NewSessionStatusSuccessResponse(data SessionStatusResponseData) *SessionStatusResponse {
	return &SessionStatusResponse{
		Success:   true,
		Code:      200,
		Data:      data,
		Timestamp: time.Now(),
	}
}
