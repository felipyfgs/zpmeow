package dto

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

type CreateSessionRequest struct {
	Name   string `json:"name" binding:"required" example:"my-session"`
	ApiKey string `json:"api_key,omitempty" example:"sk-1234567890abcdef"`
}

func (r CreateSessionRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if len(r.Name) < 3 {
		return fmt.Errorf("name must be at least 3 characters long")
	}
	if len(r.Name) > 50 {
		return fmt.Errorf("name must not exceed 50 characters")
	}
	return nil
}

type UpdateSessionRequest struct {
	Name   string `json:"name,omitempty" example:"updated-session"`
	ApiKey string `json:"api_key,omitempty" example:"sk-1234567890abcdef"`
}

func (r UpdateSessionRequest) Validate() error {
	if r.Name != "" {
		if len(r.Name) < 3 {
			return fmt.Errorf("name must be at least 3 characters long")
		}
		if len(r.Name) > 50 {
			return fmt.Errorf("name must not exceed 50 characters")
		}
	}
	return nil
}

type PairPhoneRequest struct {
	Phone string `json:"phone" binding:"required" example:"5511999999999"`
}

func (r PairPhoneRequest) Validate() error {
	if strings.TrimSpace(r.Phone) == "" {
		return fmt.Errorf("phone is required")
	}
	phone := strings.TrimSpace(r.Phone)
	if len(phone) < 10 {
		return fmt.Errorf("phone number must be at least 10 digits")
	}
	for _, char := range phone {
		if char < '0' || char > '9' {
			return fmt.Errorf("phone number must contain only digits")
		}
	}
	return nil
}

type PairCodeRequest struct {
	Code string `json:"code" binding:"required" example:"123456"`
}

func (r PairCodeRequest) Validate() error {
	if strings.TrimSpace(r.Code) == "" {
		return fmt.Errorf("code is required")
	}
	code := strings.TrimSpace(r.Code)
	if len(code) != 6 {
		return fmt.Errorf("code must be exactly 6 digits")
	}
	for _, char := range code {
		if char < '0' || char > '9' {
			return fmt.Errorf("code must contain only digits")
		}
	}
	return nil
}

type ErrorInfo struct {
	Code    string `json:"code" example:"SESSION_NOT_FOUND"`
	Message string `json:"message" example:"Session not found"`
	Details string `json:"details,omitempty" example:"Session with ID 'test' does not exist"`
}

type SessionInfo struct {
	ID        string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name      string    `json:"name" example:"my-session"`
	Status    string    `json:"status" example:"connected"`
	DeviceJID string    `json:"device_jid,omitempty" example:"5511999999999.0:1@s.whatsapp.net"`
	ApiKey    string    `json:"api_key,omitempty" example:"sk-1234567890abcdef"`
	CreatedAt time.Time `json:"created_at" example:"2023-01-01T12:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2023-01-01T12:00:00Z"`
}

type SessionData struct {
	SessionID string        `json:"session_id,omitempty"`
	Action    string        `json:"action,omitempty"`
	Status    string        `json:"status"`
	Timestamp time.Time     `json:"timestamp"`
	Session   *SessionInfo  `json:"session,omitempty"`
	Sessions  []SessionInfo `json:"sessions,omitempty"`
	QRCode    string        `json:"qr_code,omitempty"`
	Phone     string        `json:"phone,omitempty"`
	Code      string        `json:"code,omitempty"`
}

type SessionResponse struct {
	Success bool        `json:"success"`
	Code    int         `json:"code"`
	Data    SessionData `json:"data"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

type SessionListResponse struct {
	Success bool          `json:"success"`
	Code    int           `json:"code"`
	Data    []SessionInfo `json:"data"`
	Error   *ErrorInfo    `json:"error,omitempty"`
}

type CreateSessionResponse struct {
	Success bool               `json:"success"`
	Code    int                `json:"code"`
	Data    *SessionCreateData `json:"data"`
	Error   *ErrorInfo         `json:"error,omitempty"`
}

type QRCodeResponse struct {
	Success bool       `json:"success"`
	Code    int        `json:"code"`
	Data    QRCodeData `json:"data"`
	Error   *ErrorInfo `json:"error,omitempty"`
}

type QRCodeData struct {
	SessionID string    `json:"session_id"`
	QRCode    string    `json:"qr_code"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

type PairResponse struct {
	Success bool       `json:"success"`
	Code    int        `json:"code"`
	Data    PairData   `json:"data"`
	Error   *ErrorInfo `json:"error,omitempty"`
}

type PairData struct {
	SessionID string    `json:"session_id"`
	Phone     string    `json:"phone,omitempty"`
	Status    string    `json:"status"`
	Message   string    `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
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
	}
}

func NewSessionSuccessResponse(sessionID, action string, data interface{}) *SessionResponse {
	response := &SessionResponse{
		Success: true,
		Code:    http.StatusOK,
		Data: SessionData{
			SessionID: sessionID,
			Action:    action,
			Status:    "success",
			Timestamp: time.Now(),
		},
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

func NewCreateSessionSuccessResponse(sessionInfo *SessionInfo) *CreateSessionResponse {
	return &CreateSessionResponse{
		Success: true,
		Code:    http.StatusCreated,
		Data: &SessionCreateData{
			Action:    "create",
			Status:    "success",
			Timestamp: time.Now(),
			Session:   sessionInfo,
		},
	}
}

func NewCreateSessionErrorResponse(code int, errorCode, message, details string) *CreateSessionResponse {
	return &CreateSessionResponse{
		Success: false,
		Code:    code,
		Error: &ErrorInfo{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}
}

func NewQRCodeSuccessResponse(sessionID, qrCode string) *QRCodeResponse {
	return &QRCodeResponse{
		Success: true,
		Code:    http.StatusOK,
		Data: QRCodeData{
			SessionID: sessionID,
			QRCode:    qrCode,
			Status:    "qr_generated",
			Timestamp: time.Now(),
		},
	}
}

func NewQRCodeErrorResponse(code int, errorCode, message, details string) *QRCodeResponse {
	return &QRCodeResponse{
		Success: false,
		Code:    code,
		Error: &ErrorInfo{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}
}

func NewPairSuccessResponse(sessionID, phone, status, message string) *PairResponse {
	return &PairResponse{
		Success: true,
		Code:    http.StatusOK,
		Data: PairData{
			SessionID: sessionID,
			Phone:     phone,
			Status:    status,
			Message:   message,
			Timestamp: time.Now(),
		},
	}
}

func NewPairErrorResponse(code int, errorCode, message, details string) *PairResponse {
	return &PairResponse{
		Success: false,
		Code:    code,
		Error: &ErrorInfo{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}
}

func NewSessionListSuccessResponse(sessions []SessionInfo) *SessionListResponse {
	return &SessionListResponse{
		Success: true,
		Code:    http.StatusOK,
		Data:    sessions,
	}
}

func NewSessionListErrorResponse(code int, errorCode, message, details string) *SessionListResponse {
	return &SessionListResponse{
		Success: false,
		Code:    code,
		Error: &ErrorInfo{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}
}

type SessionCreateData struct {
	SessionID string       `json:"session_id,omitempty"`
	Name      string       `json:"name,omitempty"`
	Action    string       `json:"action,omitempty"`
	Status    string       `json:"status"`
	Timestamp time.Time    `json:"timestamp"`
	Session   *SessionInfo `json:"session"`
}

type SessionConnectionInfo struct {
	Status      string    `json:"status"`
	Connected   bool      `json:"connected"`
	IsConnected bool      `json:"is_connected"`
	LastSeen    time.Time `json:"last_seen,omitempty"`
	DeviceJID   string    `json:"device_jid,omitempty"`
	QRCode      string    `json:"qr_code,omitempty"`
}

type ConnectSessionResponse struct {
	Success bool                `json:"success"`
	Code    int                 `json:"code"`
	Data    *SessionConnectData `json:"data,omitempty"`
	Error   *ErrorInfo          `json:"error,omitempty"`
}

type SessionConnectData struct {
	SessionID  string                 `json:"session_id"`
	Action     string                 `json:"action,omitempty"`
	Status     string                 `json:"status"`
	Timestamp  time.Time              `json:"timestamp"`
	Connection *SessionConnectionInfo `json:"connection,omitempty"`
	QRCode     string                 `json:"qr_code,omitempty"`
	Session    *SessionInfo           `json:"session,omitempty"`
}

type PairPhoneResponse struct {
	Success bool                   `json:"success"`
	Code    int                    `json:"code"`
	Data    *PairPhoneResponseData `json:"data,omitempty"`
	Error   *ErrorInfo             `json:"error,omitempty"`
}

type PairPhoneResponseData struct {
	SessionID string    `json:"session_id"`
	Phone     string    `json:"phone"`
	Action    string    `json:"action,omitempty"`
	Status    string    `json:"status"`
	Message   string    `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Code      string    `json:"code,omitempty"`
}

type SessionStatusResponse struct {
	Success bool                       `json:"success"`
	Code    int                        `json:"code"`
	Data    *SessionStatusResponseData `json:"data,omitempty"`
	Error   *ErrorInfo                 `json:"error,omitempty"`
}

type SessionStatusResponseData struct {
	SessionID     string                 `json:"session_id"`
	Action        string                 `json:"action,omitempty"`
	Status        string                 `json:"status"`
	Timestamp     time.Time              `json:"timestamp"`
	Name          string                 `json:"name,omitempty"`
	SessionStatus string                 `json:"session_status,omitempty"`
	DeviceJID     string                 `json:"device_jid,omitempty"`
	IsConnected   bool                   `json:"is_connected"`
	ClientStatus  string                 `json:"client_status,omitempty"`
	CreatedAt     time.Time              `json:"created_at,omitempty"`
	UpdatedAt     time.Time              `json:"updated_at,omitempty"`
	Connection    *SessionConnectionInfo `json:"connection,omitempty"`
}

type UpdateWebhookRequest struct {
	WebhookURL string   `json:"webhook_url" binding:"required" example:"https://example.com/webhook"`
	URL        string   `json:"url,omitempty" example:"https://example.com/webhook"`
	Events     []string `json:"events,omitempty" example:"Message,Connected"`
}

func (r UpdateWebhookRequest) Validate() error {
	if strings.TrimSpace(r.WebhookURL) == "" {
		return fmt.Errorf("webhook_url is required")
	}
	if !strings.HasPrefix(r.WebhookURL, "http://") && !strings.HasPrefix(r.WebhookURL, "https://") {
		return fmt.Errorf("webhook_url must be a valid HTTP or HTTPS URL")
	}
	return nil
}
