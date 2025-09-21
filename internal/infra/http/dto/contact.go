package dto

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Contact request DTOs

type CheckContactRequest struct {
	Phones []string `json:"phones" binding:"required" example:"[\"5511999999999\", \"5511888888888\"]"`
}

func (r CheckContactRequest) Validate() error {
	if len(r.Phones) == 0 {
		return fmt.Errorf("phones list cannot be empty")
	}
	if len(r.Phones) > 100 {
		return fmt.Errorf("maximum 100 phone numbers allowed")
	}
	for i, phone := range r.Phones {
		if strings.TrimSpace(phone) == "" {
			return fmt.Errorf("phone number at index %d cannot be empty", i)
		}
		if len(phone) < 10 {
			return fmt.Errorf("phone number at index %d must be at least 10 digits", i)
		}
	}
	return nil
}

type GetContactRequest struct {
	Phone string `json:"phone" binding:"required" example:"5511999999999"`
}

func (r GetContactRequest) Validate() error {
	if strings.TrimSpace(r.Phone) == "" {
		return fmt.Errorf("phone is required")
	}
	if len(r.Phone) < 10 {
		return fmt.Errorf("phone number must be at least 10 digits")
	}
	return nil
}

type GetContactsRequest struct {
	Limit  int    `json:"limit,omitempty" example:"50"`
	Offset int    `json:"offset,omitempty" example:"0"`
	Filter string `json:"filter,omitempty" example:"name"`
}

func (r GetContactsRequest) Validate() error {
	if r.Limit < 0 {
		return fmt.Errorf("limit cannot be negative")
	}
	if r.Limit > 1000 {
		return fmt.Errorf("limit cannot exceed 1000")
	}
	if r.Offset < 0 {
		return fmt.Errorf("offset cannot be negative")
	}
	return nil
}

type BlockContactRequest struct {
	Phone string `json:"phone" binding:"required" example:"5511999999999"`
}

func (r BlockContactRequest) Validate() error {
	if strings.TrimSpace(r.Phone) == "" {
		return fmt.Errorf("phone is required")
	}
	return nil
}

type UnblockContactRequest struct {
	Phone string `json:"phone" binding:"required" example:"5511999999999"`
}

func (r UnblockContactRequest) Validate() error {
	if strings.TrimSpace(r.Phone) == "" {
		return fmt.Errorf("phone is required")
	}
	return nil
}

type GetContactInfoRequest struct {
	Phone  string   `json:"phone,omitempty" example:"5511999999999"`
	Phones []string `json:"phones,omitempty" example:"[\"5511999999999\", \"5511888888888\"]"`
}

func (r GetContactInfoRequest) Validate() error {
	if r.Phone == "" && len(r.Phones) == 0 {
		return fmt.Errorf("either phone or phones is required")
	}
	if r.Phone != "" && len(r.Phones) > 0 {
		return fmt.Errorf("provide either phone or phones, not both")
	}
	if len(r.Phones) > 100 {
		return fmt.Errorf("maximum 100 phone numbers allowed")
	}
	return nil
}

type GetAvatarRequest struct {
	Phone string `json:"phone" binding:"required" example:"5511999999999"`
}

func (r GetAvatarRequest) Validate() error {
	if strings.TrimSpace(r.Phone) == "" {
		return fmt.Errorf("phone is required")
	}
	return nil
}

type SetContactPresenceRequest struct {
	Phone string `json:"phone" binding:"required" example:"5511999999999"`
	State string `json:"state" binding:"required" example:"available"`
}

func (r SetContactPresenceRequest) Validate() error {
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

// Contact response DTOs

type ContactErrorResponse struct {
	Code    string `json:"code" example:"CONTACT_NOT_FOUND"`
	Message string `json:"message" example:"Contact not found"`
	Details string `json:"details" example:"Contact with phone 5511999999999 not found"`
}

type ContactInfo struct {
	Phone        string    `json:"phone" example:"5511999999999"`
	Name         string    `json:"name,omitempty" example:"João Silva"`
	DisplayName  string    `json:"display_name,omitempty" example:"João Silva"`
	VerifiedName string    `json:"verified_name,omitempty" example:"João Silva"`
	PushName     string    `json:"push_name,omitempty" example:"João"`
	Notify       string    `json:"notify,omitempty" example:"João"`
	IsOnWhatsApp bool      `json:"is_on_whatsapp" example:"true"`
	IsBlocked    bool      `json:"is_blocked" example:"false"`
	IsMuted      bool      `json:"is_muted" example:"false"`
	BusinessName string    `json:"business_name,omitempty" example:"João's Business"`
	Avatar       string    `json:"avatar,omitempty" example:"https://example.com/avatar.jpg"`
	Status       string    `json:"status,omitempty" example:"Hey there! I am using WhatsApp."`
	LastSeen     time.Time `json:"last_seen,omitempty" example:"2023-01-01T12:00:00Z"`
	JID          string    `json:"jid,omitempty" example:"5511999999999@s.whatsapp.net"`
}

type ContactResponse struct {
	Success bool                  `json:"success"`
	Code    int                   `json:"code"`
	Data    *ContactResponseData  `json:"data,omitempty"`
	Error   *ContactErrorResponse `json:"error,omitempty"`
}

type ContactResponseData struct {
	SessionID string        `json:"session_id"`
	Action    string        `json:"action"`
	Contact   *ContactInfo  `json:"contact,omitempty"`
	Contacts  []ContactInfo `json:"contacts,omitempty"`
	Phone     string        `json:"phone,omitempty"`
	Status    string        `json:"status"`
	Message   string        `json:"message,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
}

type ContactListResponse struct {
	Success bool                  `json:"success"`
	Code    int                   `json:"code"`
	Data    *ContactListData      `json:"data,omitempty"`
	Error   *ContactErrorResponse `json:"error,omitempty"`
}

type ContactListData struct {
	SessionID string        `json:"session_id"`
	Contacts  []ContactInfo `json:"contacts"`
	Count     int           `json:"count"`
	Limit     int           `json:"limit"`
	Offset    int           `json:"offset"`
	Total     int           `json:"total"`
}

type CheckContactResponse struct {
	Success bool                  `json:"success"`
	Code    int                   `json:"code"`
	Data    *CheckContactData     `json:"data,omitempty"`
	Error   *ContactErrorResponse `json:"error,omitempty"`
}

type CheckContactData struct {
	SessionID string               `json:"session_id"`
	Results   []ContactCheckResult `json:"results"`
	Count     int                  `json:"count"`
	Timestamp time.Time            `json:"timestamp"`
}

type ContactCheckResult struct {
	Phone        string `json:"phone" example:"5511999999999"`
	Query        string `json:"query,omitempty" example:"5511999999999"`
	IsOnWhatsApp bool   `json:"is_on_whatsapp" example:"true"`
	IsInmeow     bool   `json:"is_in_meow" example:"true"`
	JID          string `json:"jid,omitempty" example:"5511999999999@s.whatsapp.net"`
	VerifiedName string `json:"verified_name,omitempty" example:"João Silva"`
}

type ContactActionResponse struct {
	Success bool                  `json:"success"`
	Code    int                   `json:"code"`
	Data    *ContactActionData    `json:"data,omitempty"`
	Error   *ContactErrorResponse `json:"error,omitempty"`
}

type ContactActionData struct {
	SessionID string    `json:"session_id"`
	Phone     string    `json:"phone"`
	Action    string    `json:"action"`
	Status    string    `json:"status"`
	Message   string    `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// Avatar related DTOs
type AvatarInfo struct {
	Phone     string    `json:"phone" example:"5511999999999"`
	JID       string    `json:"jid,omitempty" example:"5511999999999@s.whatsapp.net"`
	AvatarURL string    `json:"avatar_url,omitempty" example:"https://example.com/avatar.jpg"`
	HasAvatar bool      `json:"has_avatar" example:"true"`
	PictureID string    `json:"picture_id,omitempty" example:"1234567890"`
	Timestamp time.Time `json:"timestamp,omitempty" example:"2023-01-01T12:00:00Z"`
}

type ContactAvatarResponse struct {
	Success bool                  `json:"success"`
	Code    int                   `json:"code"`
	Data    *AvatarInfo           `json:"data,omitempty"`
	Error   *ContactErrorResponse `json:"error,omitempty"`
}

// Constructor functions

func NewContactErrorResponse(code int, errorCode, message, details string) *ContactResponse {
	return &ContactResponse{
		Success: false,
		Code:    code,
		Error: &ContactErrorResponse{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}
}

func NewContactSuccessResponse(sessionID, action string, contact *ContactInfo) *ContactResponse {
	return &ContactResponse{
		Success: true,
		Code:    http.StatusOK,
		Data: &ContactResponseData{
			SessionID: sessionID,
			Action:    action,
			Contact:   contact,
			Status:    "success",
			Timestamp: time.Now(),
		},
	}
}

func NewContactListSuccessResponse(sessionID string, contacts []ContactInfo, limit, offset, total int) *ContactListResponse {
	return &ContactListResponse{
		Success: true,
		Code:    http.StatusOK,
		Data: &ContactListData{
			SessionID: sessionID,
			Contacts:  contacts,
			Count:     len(contacts),
			Limit:     limit,
			Offset:    offset,
			Total:     total,
		},
	}
}

func NewContactListErrorResponse(code int, errorCode, message, details string) *ContactListResponse {
	return &ContactListResponse{
		Success: false,
		Code:    code,
		Error: &ContactErrorResponse{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}
}

func NewCheckContactSuccessResponse(sessionID string, results []ContactCheckResult) *CheckContactResponse {
	return &CheckContactResponse{
		Success: true,
		Code:    http.StatusOK,
		Data: &CheckContactData{
			SessionID: sessionID,
			Results:   results,
			Count:     len(results),
			Timestamp: time.Now(),
		},
	}
}

func NewCheckContactErrorResponse(code int, errorCode, message, details string) *CheckContactResponse {
	return &CheckContactResponse{
		Success: false,
		Code:    code,
		Error: &ContactErrorResponse{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}
}

func NewContactActionSuccessResponse(sessionID, phone, action, message string) *ContactActionResponse {
	return &ContactActionResponse{
		Success: true,
		Code:    http.StatusOK,
		Data: &ContactActionData{
			SessionID: sessionID,
			Phone:     phone,
			Action:    action,
			Status:    "success",
			Message:   message,
			Timestamp: time.Now(),
		},
	}
}

func NewContactActionErrorResponse(code int, errorCode, message, details string) *ContactActionResponse {
	return &ContactActionResponse{
		Success: false,
		Code:    code,
		Error: &ContactErrorResponse{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}
}

func NewContactAvatarResponse(avatarInfo *AvatarInfo) *ContactAvatarResponse {
	return &ContactAvatarResponse{
		Success: true,
		Code:    http.StatusOK,
		Data:    avatarInfo,
	}
}

func NewContactsResponse(sessionID string, contacts []ContactInfo) *ContactListResponse {
	return &ContactListResponse{
		Success: true,
		Code:    http.StatusOK,
		Data: &ContactListData{
			SessionID: sessionID,
			Contacts:  contacts,
			Count:     len(contacts),
			Total:     len(contacts),
		},
	}
}
