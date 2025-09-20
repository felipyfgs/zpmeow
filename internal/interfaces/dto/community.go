package dto

import (
	"time"
)

type LinkGroupRequest struct {
	CommunityJID string `json:"community_jid" binding:"required" example:"120363025246125486@g.us"`
	GroupJID     string `json:"group_jid" binding:"required" example:"120363025246125487@g.us"`
}

type UnlinkGroupRequest struct {
	CommunityJID string `json:"community_jid" binding:"required" example:"120363025246125486@g.us"`
	GroupJID     string `json:"group_jid" binding:"required" example:"120363025246125487@g.us"`
}

type GetSubGroupsRequest struct {
	CommunityJID string `json:"community_jid" binding:"required" example:"120363025246125486@g.us"`
}

type GetLinkedGroupsParticipantsRequest struct {
	CommunityJID string `json:"community_jid" binding:"required" example:"120363025246125486@g.us"`
}

type CommunityResponse struct {
	Success bool                    `json:"success"`
	Code    int                     `json:"code"`
	Data    CommunityData           `json:"data"`
	Error   *CommunityErrorResponse `json:"error,omitempty"`
}

type CommunityData struct {
	SessionID    string    `json:"session_id" example:"default"`
	CommunityJID string    `json:"community_jid,omitempty" example:"120363025246125486@g.us"`
	GroupJID     string    `json:"group_jid,omitempty" example:"120363025246125487@g.us"`
	Action       string    `json:"action" example:"link_group"`
	Status       string    `json:"status" example:"success"`
	Timestamp    time.Time `json:"timestamp" example:"2023-01-01T00:00:00Z"`
}

type CommunityErrorResponse struct {
	Code    string `json:"code" example:"INVALID_COMMUNITY_JID"`
	Message string `json:"message" example:"Invalid community JID format"`
	Details string `json:"details,omitempty" example:"Community JID must be in format: number@g.us"`
}

type CommunitySubGroupsResponse struct {
	Success bool                    `json:"success"`
	Code    int                     `json:"code"`
	Data    CommunitySubGroupsData  `json:"data"`
	Error   *CommunityErrorResponse `json:"error,omitempty"`
}

type CommunitySubGroupsData struct {
	SessionID    string    `json:"session_id" example:"default"`
	CommunityJID string    `json:"community_jid" example:"120363025246125486@g.us"`
	Action       string    `json:"action" example:"get_subgroups"`
	Status       string    `json:"status" example:"success"`
	Timestamp    time.Time `json:"timestamp" example:"2023-01-01T00:00:00Z"`
	SubGroups    []string  `json:"subgroups" example:"[\"120363025246125487@g.us\", \"120363025246125488@g.us\"]"`
	Total        int       `json:"total" example:"2"`
}

type CommunityParticipantsResponse struct {
	Success bool                      `json:"success"`
	Code    int                       `json:"code"`
	Data    CommunityParticipantsData `json:"data"`
	Error   *CommunityErrorResponse   `json:"error,omitempty"`
}

type CommunityParticipantsData struct {
	SessionID    string    `json:"session_id" example:"default"`
	CommunityJID string    `json:"community_jid" example:"120363025246125486@g.us"`
	Action       string    `json:"action" example:"get_participants"`
	Status       string    `json:"status" example:"success"`
	Timestamp    time.Time `json:"timestamp" example:"2023-01-01T00:00:00Z"`
	Participants []string  `json:"participants" example:"[\"5511999999999\", \"5511888888888\"]"`
	Total        int       `json:"total" example:"2"`
}

type CommunityValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *CommunityValidationError) Error() string {
	return e.Message
}

func NewCommunitySuccessResponse(sessionID, communityJID, groupJID, action string) *CommunityResponse {
	return &CommunityResponse{
		Success: true,
		Code:    200,
		Data: CommunityData{
			SessionID:    sessionID,
			CommunityJID: communityJID,
			GroupJID:     groupJID,
			Action:       action,
			Status:       "success",
			Timestamp:    time.Now(),
		},
	}
}

func NewCommunityErrorResponse(code int, errorCode, message, details string) *CommunityResponse {
	return &CommunityResponse{
		Success: false,
		Code:    code,
		Error: &CommunityErrorResponse{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}
}

func NewCommunitySubGroupsResponse(sessionID, communityJID string, subGroups []string) *CommunitySubGroupsResponse {
	return &CommunitySubGroupsResponse{
		Success: true,
		Code:    200,
		Data: CommunitySubGroupsData{
			SessionID:    sessionID,
			CommunityJID: communityJID,
			Action:       "get_subgroups",
			Status:       "success",
			Timestamp:    time.Now(),
			SubGroups:    subGroups,
			Total:        len(subGroups),
		},
	}
}

func NewCommunityParticipantsResponse(sessionID, communityJID string, participants []string) *CommunityParticipantsResponse {
	return &CommunityParticipantsResponse{
		Success: true,
		Code:    200,
		Data: CommunityParticipantsData{
			SessionID:    sessionID,
			CommunityJID: communityJID,
			Action:       "get_participants",
			Status:       "success",
			Timestamp:    time.Now(),
			Participants: participants,
			Total:        len(participants),
		},
	}
}

func validateCommunityJID(jid string) bool {
	if jid == "" {
		return false
	}
	return len(jid) > 10 && (jid[len(jid)-5:] == "@g.us")
}

func (r *LinkGroupRequest) Validate() error {
	if r.CommunityJID == "" {
		return &CommunityValidationError{Field: "community_jid", Message: "Community JID is required"}
	}
	if !validateCommunityJID(r.CommunityJID) {
		return &CommunityValidationError{Field: "community_jid", Message: "Invalid community JID format"}
	}
	if r.GroupJID == "" {
		return &CommunityValidationError{Field: "group_jid", Message: "Group JID is required"}
	}
	if !validateCommunityJID(r.GroupJID) {
		return &CommunityValidationError{Field: "group_jid", Message: "Invalid group JID format"}
	}
	return nil
}

func (r *UnlinkGroupRequest) Validate() error {
	if r.CommunityJID == "" {
		return &CommunityValidationError{Field: "community_jid", Message: "Community JID is required"}
	}
	if !validateCommunityJID(r.CommunityJID) {
		return &CommunityValidationError{Field: "community_jid", Message: "Invalid community JID format"}
	}
	if r.GroupJID == "" {
		return &CommunityValidationError{Field: "group_jid", Message: "Group JID is required"}
	}
	if !validateCommunityJID(r.GroupJID) {
		return &CommunityValidationError{Field: "group_jid", Message: "Invalid group JID format"}
	}
	return nil
}

func (r *GetSubGroupsRequest) Validate() error {
	if r.CommunityJID == "" {
		return &CommunityValidationError{Field: "community_jid", Message: "Community JID is required"}
	}
	if !validateCommunityJID(r.CommunityJID) {
		return &CommunityValidationError{Field: "community_jid", Message: "Invalid community JID format"}
	}
	return nil
}

func (r *GetLinkedGroupsParticipantsRequest) Validate() error {
	if r.CommunityJID == "" {
		return &CommunityValidationError{Field: "community_jid", Message: "Community JID is required"}
	}
	if !validateCommunityJID(r.CommunityJID) {
		return &CommunityValidationError{Field: "community_jid", Message: "Invalid community JID format"}
	}
	return nil
}
