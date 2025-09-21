package dto

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)


type CreateCommunityRequest struct {
	Name        string `json:"name" binding:"required" example:"My Community"`
	Description string `json:"description,omitempty" example:"This is my community"`
}

func (r CreateCommunityRequest) Validate() error {
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

type UpdateCommunityRequest struct {
	Name        string `json:"name,omitempty" example:"Updated Community"`
	Description string `json:"description,omitempty" example:"Updated description"`
}

func (r UpdateCommunityRequest) Validate() error {
	if r.Name != "" && len(r.Name) > 100 {
		return fmt.Errorf("name must not exceed 100 characters")
	}
	if len(r.Description) > 500 {
		return fmt.Errorf("description must not exceed 500 characters")
	}
	return nil
}

type JoinCommunityRequest struct {
	CommunityJID string `json:"community_jid" binding:"required" example:"120363025246125486@g.us"`
}

func (r JoinCommunityRequest) Validate() error {
	if strings.TrimSpace(r.CommunityJID) == "" {
		return fmt.Errorf("community_jid is required")
	}
	return nil
}

type LeaveCommunityRequest struct {
	CommunityJID string `json:"community_jid" binding:"required" example:"120363025246125486@g.us"`
}

func (r LeaveCommunityRequest) Validate() error {
	if strings.TrimSpace(r.CommunityJID) == "" {
		return fmt.Errorf("community_jid is required")
	}
	return nil
}

type LinkGroupRequest struct {
	CommunityJID string `json:"community_jid" binding:"required" example:"120363025246125486@g.us"`
	GroupJID     string `json:"group_jid" binding:"required" example:"120363025246125486@g.us"`
}

func (r LinkGroupRequest) Validate() error {
	if strings.TrimSpace(r.CommunityJID) == "" {
		return fmt.Errorf("community_jid is required")
	}
	if strings.TrimSpace(r.GroupJID) == "" {
		return fmt.Errorf("group_jid is required")
	}
	return nil
}

type UnlinkGroupRequest struct {
	CommunityJID string `json:"community_jid" binding:"required" example:"120363025246125486@g.us"`
	GroupJID     string `json:"group_jid" binding:"required" example:"120363025246125486@g.us"`
}

func (r UnlinkGroupRequest) Validate() error {
	if strings.TrimSpace(r.CommunityJID) == "" {
		return fmt.Errorf("community_jid is required")
	}
	if strings.TrimSpace(r.GroupJID) == "" {
		return fmt.Errorf("group_jid is required")
	}
	return nil
}

type GetSubGroupsRequest struct {
	CommunityJID string `json:"community_jid" binding:"required" example:"120363025246125486@g.us"`
}

func (r GetSubGroupsRequest) Validate() error {
	if strings.TrimSpace(r.CommunityJID) == "" {
		return fmt.Errorf("community_jid is required")
	}
	return nil
}

type GetLinkedGroupsParticipantsRequest struct {
	CommunityJID string `json:"community_jid" binding:"required" example:"120363025246125486@g.us"`
}

func (r GetLinkedGroupsParticipantsRequest) Validate() error {
	if strings.TrimSpace(r.CommunityJID) == "" {
		return fmt.Errorf("community_jid is required")
	}
	return nil
}


type CommunityErrorResponse struct {
	Code    string `json:"code" example:"COMMUNITY_NOT_FOUND"`
	Message string `json:"message" example:"Community not found"`
	Details string `json:"details" example:"Community with JID '120363025246125486@g.us' not found"`
}

type CommunityInfo struct {
	JID         string    `json:"jid" example:"120363025246125486@g.us"`
	Name        string    `json:"name" example:"My Community"`
	Description string    `json:"description,omitempty" example:"This is my community"`
	Owner       string    `json:"owner,omitempty" example:"5511999999999@s.whatsapp.net"`
	CreatedAt   time.Time `json:"created_at,omitempty" example:"2023-01-01T12:00:00Z"`
	MemberCount int       `json:"member_count" example:"50"`
	IsMember    bool      `json:"is_member" example:"true"`
	IsAdmin     bool      `json:"is_admin" example:"false"`
}

type CommunityResponse struct {
	Success bool                    `json:"success"`
	Code    int                     `json:"code"`
	Data    *CommunityResponseData  `json:"data,omitempty"`
	Error   *CommunityErrorResponse `json:"error,omitempty"`
}

type CommunityResponseData struct {
	SessionID string         `json:"session_id"`
	Action    string         `json:"action"`
	Status    string         `json:"status"`
	Message   string         `json:"message,omitempty"`
	Community *CommunityInfo `json:"community,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

type CommunityListResponse struct {
	Success bool                    `json:"success"`
	Code    int                     `json:"code"`
	Data    *CommunityListData      `json:"data,omitempty"`
	Error   *CommunityErrorResponse `json:"error,omitempty"`
}

type CommunityListData struct {
	SessionID   string          `json:"session_id"`
	Communities []CommunityInfo `json:"communities"`
	Count       int             `json:"count"`
	Total       int             `json:"total"`
}

type CommunityActionResponse struct {
	Success bool                    `json:"success"`
	Code    int                     `json:"code"`
	Data    *CommunityActionData    `json:"data,omitempty"`
	Error   *CommunityErrorResponse `json:"error,omitempty"`
}

type CommunityActionData struct {
	SessionID    string    `json:"session_id"`
	CommunityJID string    `json:"community_jid"`
	Action       string    `json:"action"`
	Status       string    `json:"status"`
	Message      string    `json:"message,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
}

type CommunitySubGroupsResponse struct {
	Success bool                    `json:"success"`
	Code    int                     `json:"code"`
	Data    *CommunitySubGroupsData `json:"data,omitempty"`
	Error   *CommunityErrorResponse `json:"error,omitempty"`
}

type CommunitySubGroupsData struct {
	SessionID    string          `json:"session_id"`
	CommunityJID string          `json:"community_jid"`
	SubGroups    []CommunityInfo `json:"sub_groups"`
	Count        int             `json:"count"`
}

type CommunityParticipantsResponse struct {
	Success bool                       `json:"success"`
	Code    int                        `json:"code"`
	Data    *CommunityParticipantsData `json:"data,omitempty"`
	Error   *CommunityErrorResponse    `json:"error,omitempty"`
}

type CommunityParticipantsData struct {
	SessionID    string             `json:"session_id"`
	CommunityJID string             `json:"community_jid"`
	Participants []GroupParticipant `json:"participants"`
	Count        int                `json:"count"`
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

func NewCommunitySuccessResponse(sessionID, action, message string, community *CommunityInfo) *CommunityResponse {
	return &CommunityResponse{
		Success: true,
		Code:    http.StatusOK,
		Data: &CommunityResponseData{
			SessionID: sessionID,
			Action:    action,
			Status:    "success",
			Message:   message,
			Community: community,
			Timestamp: time.Now(),
		},
	}
}

func NewCommunityListSuccessResponse(sessionID string, communities []CommunityInfo) *CommunityListResponse {
	return &CommunityListResponse{
		Success: true,
		Code:    http.StatusOK,
		Data: &CommunityListData{
			SessionID:   sessionID,
			Communities: communities,
			Count:       len(communities),
			Total:       len(communities),
		},
	}
}

func NewCommunityListErrorResponse(code int, errorCode, message, details string) *CommunityListResponse {
	return &CommunityListResponse{
		Success: false,
		Code:    code,
		Error: &CommunityErrorResponse{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}
}

func NewCommunityActionSuccessResponse(sessionID, communityJID, action, message string) *CommunityActionResponse {
	return &CommunityActionResponse{
		Success: true,
		Code:    http.StatusOK,
		Data: &CommunityActionData{
			SessionID:    sessionID,
			CommunityJID: communityJID,
			Action:       action,
			Status:       "success",
			Message:      message,
			Timestamp:    time.Now(),
		},
	}
}

func NewCommunityActionErrorResponse(code int, errorCode, message, details string) *CommunityActionResponse {
	return &CommunityActionResponse{
		Success: false,
		Code:    code,
		Error: &CommunityErrorResponse{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}
}

func NewCommunitySubGroupsResponse(sessionID, communityJID string, subGroups []CommunityInfo) *CommunitySubGroupsResponse {
	return &CommunitySubGroupsResponse{
		Success: true,
		Code:    http.StatusOK,
		Data: &CommunitySubGroupsData{
			SessionID:    sessionID,
			CommunityJID: communityJID,
			SubGroups:    subGroups,
			Count:        len(subGroups),
		},
	}
}

func NewCommunityParticipantsResponse(sessionID, communityJID string, participants []GroupParticipant) *CommunityParticipantsResponse {
	return &CommunityParticipantsResponse{
		Success: true,
		Code:    http.StatusOK,
		Data: &CommunityParticipantsData{
			SessionID:    sessionID,
			CommunityJID: communityJID,
			Participants: participants,
			Count:        len(participants),
		},
	}
}
