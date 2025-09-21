package dto

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

type UpdatePrivacySettingsRequest struct {
	LastSeen     string `json:"last_seen,omitempty" example:"everyone"`
	ProfilePhoto string `json:"profile_photo,omitempty" example:"contacts"`
	Status       string `json:"status,omitempty" example:"contacts"`
	About        string `json:"about,omitempty" example:"contacts"`
	ReadReceipts string `json:"read_receipts,omitempty" example:"everyone"`
	Groups       string `json:"groups,omitempty" example:"contacts"`
	CallAdd      string `json:"call_add,omitempty" example:"contacts"`
}

func (r UpdatePrivacySettingsRequest) Validate() error {
	validOptions := []string{"everyone", "contacts", "nobody"}

	if r.LastSeen != "" && !contains(validOptions, r.LastSeen) {
		return fmt.Errorf("invalid last_seen option, must be one of: %s", strings.Join(validOptions, ", "))
	}
	if r.ProfilePhoto != "" && !contains(validOptions, r.ProfilePhoto) {
		return fmt.Errorf("invalid profile_photo option, must be one of: %s", strings.Join(validOptions, ", "))
	}
	if r.Status != "" && !contains(validOptions, r.Status) {
		return fmt.Errorf("invalid status option, must be one of: %s", strings.Join(validOptions, ", "))
	}
	if r.About != "" && !contains(validOptions, r.About) {
		return fmt.Errorf("invalid about option, must be one of: %s", strings.Join(validOptions, ", "))
	}
	if r.ReadReceipts != "" && !contains(validOptions, r.ReadReceipts) {
		return fmt.Errorf("invalid read_receipts option, must be one of: %s", strings.Join(validOptions, ", "))
	}
	if r.Groups != "" && !contains(validOptions, r.Groups) {
		return fmt.Errorf("invalid groups option, must be one of: %s", strings.Join(validOptions, ", "))
	}
	if r.CallAdd != "" && !contains(validOptions, r.CallAdd) {
		return fmt.Errorf("invalid call_add option, must be one of: %s", strings.Join(validOptions, ", "))
	}

	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

type BlockContactsRequest struct {
	Contacts []string `json:"contacts" binding:"required" example:"[\"5511999999999\", \"5511888888888\"]"`
}

func (r BlockContactsRequest) Validate() error {
	if len(r.Contacts) == 0 {
		return fmt.Errorf("contacts list cannot be empty")
	}
	if len(r.Contacts) > 100 {
		return fmt.Errorf("maximum 100 contacts can be blocked at once")
	}
	return nil
}

type UnblockContactsRequest struct {
	Contacts []string `json:"contacts" binding:"required" example:"[\"5511999999999\", \"5511888888888\"]"`
}

func (r UnblockContactsRequest) Validate() error {
	if len(r.Contacts) == 0 {
		return fmt.Errorf("contacts list cannot be empty")
	}
	if len(r.Contacts) > 100 {
		return fmt.Errorf("maximum 100 contacts can be unblocked at once")
	}
	return nil
}

type PrivacyErrorResponse struct {
	Code    string `json:"code" example:"PRIVACY_UPDATE_FAILED"`
	Message string `json:"message" example:"Failed to update privacy settings"`
	Details string `json:"details" example:"Invalid privacy option provided"`
}

type PrivacySettings struct {
	LastSeen     string    `json:"last_seen" example:"everyone"`
	ProfilePhoto string    `json:"profile_photo" example:"contacts"`
	Status       string    `json:"status" example:"contacts"`
	About        string    `json:"about" example:"contacts"`
	ReadReceipts string    `json:"read_receipts" example:"everyone"`
	Groups       string    `json:"groups" example:"contacts"`
	CallAdd      string    `json:"call_add" example:"contacts"`
	UpdatedAt    time.Time `json:"updated_at" example:"2023-01-01T12:00:00Z"`
}

type PrivacyResponse struct {
	Success bool                  `json:"success"`
	Code    int                   `json:"code"`
	Data    *PrivacyResponseData  `json:"data,omitempty"`
	Error   *PrivacyErrorResponse `json:"error,omitempty"`
}

type PrivacyResponseData struct {
	SessionID string           `json:"session_id"`
	Action    string           `json:"action"`
	Status    string           `json:"status"`
	Message   string           `json:"message,omitempty"`
	Settings  *PrivacySettings `json:"settings,omitempty"`
	Timestamp time.Time        `json:"timestamp"`
}

type BlockedContactsResponse struct {
	Success bool                  `json:"success"`
	Code    int                   `json:"code"`
	Data    *BlockedContactsData  `json:"data,omitempty"`
	Error   *PrivacyErrorResponse `json:"error,omitempty"`
}

type BlockedContactsData struct {
	SessionID       string   `json:"session_id"`
	BlockedContacts []string `json:"blocked_contacts"`
	Count           int      `json:"count"`
}

type BlockActionResponse struct {
	Success bool                  `json:"success"`
	Code    int                   `json:"code"`
	Data    *BlockActionData      `json:"data,omitempty"`
	Error   *PrivacyErrorResponse `json:"error,omitempty"`
}

type BlockActionData struct {
	SessionID string    `json:"session_id"`
	Action    string    `json:"action"`
	Status    string    `json:"status"`
	Message   string    `json:"message,omitempty"`
	Contacts  []string  `json:"contacts"`
	Timestamp time.Time `json:"timestamp"`
}

func NewPrivacyErrorResponse(code int, errorCode, message, details string) *PrivacyResponse {
	return &PrivacyResponse{
		Success: false,
		Code:    code,
		Error: &PrivacyErrorResponse{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}
}

func NewPrivacySuccessResponse(sessionID, action, message string, settings *PrivacySettings) *PrivacyResponse {
	return &PrivacyResponse{
		Success: true,
		Code:    http.StatusOK,
		Data: &PrivacyResponseData{
			SessionID: sessionID,
			Action:    action,
			Status:    "success",
			Message:   message,
			Settings:  settings,
			Timestamp: time.Now(),
		},
	}
}

func NewBlockedContactsSuccessResponse(sessionID string, blockedContacts []string) *BlockedContactsResponse {
	return &BlockedContactsResponse{
		Success: true,
		Code:    http.StatusOK,
		Data: &BlockedContactsData{
			SessionID:       sessionID,
			BlockedContacts: blockedContacts,
			Count:           len(blockedContacts),
		},
	}
}

func NewBlockedContactsErrorResponse(code int, errorCode, message, details string) *BlockedContactsResponse {
	return &BlockedContactsResponse{
		Success: false,
		Code:    code,
		Error: &PrivacyErrorResponse{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}
}

func NewBlockActionSuccessResponse(sessionID, action, message string, contacts []string) *BlockActionResponse {
	return &BlockActionResponse{
		Success: true,
		Code:    http.StatusOK,
		Data: &BlockActionData{
			SessionID: sessionID,
			Action:    action,
			Status:    "success",
			Message:   message,
			Contacts:  contacts,
			Timestamp: time.Now(),
		},
	}
}

func NewBlockActionErrorResponse(code int, errorCode, message, details string) *BlockActionResponse {
	return &BlockActionResponse{
		Success: false,
		Code:    code,
		Error: &PrivacyErrorResponse{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}
}

type SetAllPrivacySettingsRequest struct {
	LastSeen          string `json:"last_seen" binding:"required" example:"contacts"`
	ProfilePhoto      string `json:"profile_photo" binding:"required" example:"contacts"`
	Status            string `json:"status" binding:"required" example:"contacts"`
	ReadReceipts      bool   `json:"read_receipts" example:"true"`
	GroupsAddMe       string `json:"groups_add_me" binding:"required" example:"contacts"`
	CallsAddMe        string `json:"calls_add_me" binding:"required" example:"contacts"`
	DisappearingChats bool   `json:"disappearing_chats" example:"false"`
}

func (r SetAllPrivacySettingsRequest) Validate() error {
	validValues := []string{"everyone", "contacts", "nobody"}

	if !contains(validValues, r.LastSeen) {
		return fmt.Errorf("last_seen must be one of: everyone, contacts, nobody")
	}
	if !contains(validValues, r.ProfilePhoto) {
		return fmt.Errorf("profile_photo must be one of: everyone, contacts, nobody")
	}
	if !contains(validValues, r.Status) {
		return fmt.Errorf("status must be one of: everyone, contacts, nobody")
	}
	if !contains(validValues, r.GroupsAddMe) {
		return fmt.Errorf("groups_add_me must be one of: everyone, contacts, nobody")
	}
	if !contains(validValues, r.CallsAddMe) {
		return fmt.Errorf("calls_add_me must be one of: everyone, contacts, nobody")
	}

	return nil
}

type PrivacySettingsResponse struct {
	Success bool                 `json:"success"`
	Code    int                  `json:"code"`
	Data    *PrivacySettingsData `json:"data,omitempty"`
	Error   *ErrorInfo           `json:"error,omitempty"`
}

type PrivacySettingsData struct {
	SessionID         string `json:"session_id"`
	LastSeen          string `json:"last_seen"`
	ProfilePhoto      string `json:"profile_photo"`
	Status            string `json:"status"`
	ReadReceipts      bool   `json:"read_receipts"`
	GroupsAddMe       string `json:"groups_add_me"`
	CallsAddMe        string `json:"calls_add_me"`
	DisappearingChats bool   `json:"disappearing_chats"`
	UpdatedAt         string `json:"updated_at"`
}

type BlocklistResponse struct {
	Success bool                 `json:"success"`
	Code    int                  `json:"code"`
	Data    *BlockedContactsData `json:"data,omitempty"`
	Error   *ErrorInfo           `json:"error,omitempty"`
}

type UpdateBlocklistRequest struct {
	Action   string   `json:"action" binding:"required" example:"block"`
	Contacts []string `json:"contacts" binding:"required" example:"[\"5511999999999\"]"`
}

func (r UpdateBlocklistRequest) Validate() error {
	if r.Action != "block" && r.Action != "unblock" {
		return fmt.Errorf("action must be either 'block' or 'unblock'")
	}
	if len(r.Contacts) == 0 {
		return fmt.Errorf("at least one contact must be provided")
	}
	return nil
}

type FindPrivacySettingsRequest struct {
	Settings []string `json:"settings" example:"[\"last_seen\", \"profile_photo\"]"`
}

func (r FindPrivacySettingsRequest) Validate() error {
	validSettings := []string{"last_seen", "profile_photo", "status", "read_receipts", "groups_add_me", "calls_add_me"}

	for _, setting := range r.Settings {
		found := false
		for _, valid := range validSettings {
			if setting == valid {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("invalid setting: %s", setting)
		}
	}
	return nil
}
