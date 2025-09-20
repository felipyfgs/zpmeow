package dto

import (
	"time"
)

type CheckContactRequest struct {
	Phones []string `json:"phones" binding:"required" example:"[\"5511999999999\", \"5511888888888\"]"`
}

type GetContactInfoRequest struct {
	Phones []string `json:"phones" binding:"required" example:"[\"5511999999999\", \"5511888888888\"]"`
}

type GetAvatarRequest struct {
	Phone string `json:"phone" binding:"required" example:"5511999999999"`
}

type SetContactPresenceRequest struct {
	State string `json:"state" binding:"required" example:"available"`
}

type ContactCheckResult struct {
	Query        string `json:"query" example:"5511999999999"`
	IsInmeow     bool   `json:"is_in_meow" example:"true"`
	JID          string `json:"jid" example:"5511999999999@s.meow.net"`
	VerifiedName string `json:"verified_name,omitempty" example:"João Silva"`
}

type ContactInfo struct {
	JID          string `json:"jid" example:"5511999999999@s.meow.net"`
	Name         string `json:"name,omitempty" example:"João Silva"`
	DisplayName  string `json:"display_name,omitempty" example:"João Silva"`
	VerifiedName string `json:"verified_name,omitempty" example:"João Silva Empresa"`
	Avatar       string `json:"avatar,omitempty" example:"https://..."`
	Status       string `json:"status,omitempty" example:"Disponível"`
	PictureID    string `json:"picture_id,omitempty" example:"pic_123"`
	DeviceCount  int    `json:"device_count,omitempty" example:"2"`
	Notify       string `json:"notify,omitempty" example:"João"`
	PushName     string `json:"push_name,omitempty" example:"João"`
	BusinessName string `json:"business_name,omitempty" example:"Empresa João"`
	Phone        string `json:"phone,omitempty" example:"5511999999999"`
	IsBlocked    bool   `json:"is_blocked" example:"false"`
	IsMuted      bool   `json:"is_muted" example:"false"`
}

type AvatarInfo struct {
	Phone     string    `json:"phone" example:"5511999999999"`
	JID       string    `json:"jid" example:"5511999999999@s.meow.net"`
	AvatarURL string    `json:"avatar_url,omitempty" example:"https://pps.meow.net/..."`
	PictureID string    `json:"picture_id,omitempty" example:"pic_123"`
	Timestamp time.Time `json:"timestamp" example:"2023-01-01T12:00:00Z"`
}

type ContactResponse struct {
	Success bool                  `json:"success"`
	Code    int                   `json:"code"`
	Data    ContactData           `json:"data"`
	Error   *ContactErrorResponse `json:"error,omitempty"`
}

type ContactData struct {
	Action       string               `json:"action" example:"check_contacts"`
	Status       string               `json:"status" example:"success"`
	Timestamp    time.Time            `json:"timestamp" example:"2023-01-01T00:00:00Z"`
	CheckResults []ContactCheckResult `json:"check_results,omitempty"`
	ContactInfos []ContactInfo        `json:"contact_infos,omitempty"`
	Avatar       *AvatarInfo          `json:"avatar,omitempty"`
}

type ContactErrorResponse struct {
	Code    string `json:"code" example:"INVALID_PHONE"`
	Message string `json:"message" example:"Invalid phone number format"`
	Details string `json:"details,omitempty" example:"Phone number must include country code"`
}

type ContactsResponse struct {
	Success bool                  `json:"success"`
	Code    int                   `json:"code"`
	Data    ContactsData          `json:"data"`
	Error   *ContactErrorResponse `json:"error,omitempty"`
}

type ContactsData struct {
	Action    string        `json:"action" example:"get_contacts"`
	Status    string        `json:"status" example:"success"`
	Timestamp time.Time     `json:"timestamp" example:"2023-01-01T00:00:00Z"`
	Contacts  []ContactInfo `json:"contacts"`
	Count     int           `json:"count" example:"10"`
}

func NewContactSuccessResponse(action string, checkResults []ContactCheckResult, contactInfos []ContactInfo) *ContactResponse {
	return &ContactResponse{
		Success: true,
		Code:    200,
		Data: ContactData{
			Action:       action,
			Status:       "success",
			Timestamp:    time.Now(),
			CheckResults: checkResults,
			ContactInfos: contactInfos,
		},
	}
}

func NewContactErrorResponse(code int, errorCode, message, details string) *ContactResponse {
	return &ContactResponse{
		Success: false,
		Code:    code,
		Data: ContactData{
			Status:    "error",
			Timestamp: time.Now(),
		},
		Error: &ContactErrorResponse{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}
}

func NewContactAvatarResponse(avatar *AvatarInfo) *ContactResponse {
	return &ContactResponse{
		Success: true,
		Code:    200,
		Data: ContactData{
			Action:    "get_avatar",
			Status:    "success",
			Timestamp: time.Now(),
			Avatar:    avatar,
		},
	}
}

func NewContactsResponse(contacts []ContactInfo) *ContactsResponse {
	return &ContactsResponse{
		Success: true,
		Code:    200,
		Data: ContactsData{
			Action:    "get_contacts",
			Status:    "success",
			Timestamp: time.Now(),
			Contacts:  contacts,
			Count:     len(contacts),
		},
	}
}
