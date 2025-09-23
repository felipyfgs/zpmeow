package dto

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

type CreateGroupRequest struct {
	Name         string   `json:"name" binding:"required" example:"My Group"`
	Description  string   `json:"description,omitempty" example:"This is my group"`
	Participants []string `json:"participants,omitempty" example:"5511999999999,5511888888888"`
}

func (r CreateGroupRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if len(r.Name) > 100 {
		return fmt.Errorf("name must not exceed 100 characters")
	}
	if len(r.Description) > 500 {
		return fmt.Errorf("description must not exceed 500 characters")
	}
	if len(r.Participants) > 256 {
		return fmt.Errorf("maximum 256 participants allowed")
	}
	return nil
}

type UpdateGroupRequest struct {
	Name        string `json:"name,omitempty" example:"Updated Group Name"`
	Description string `json:"description,omitempty" example:"Updated description"`
}

func (r UpdateGroupRequest) Validate() error {
	if r.Name != "" && len(r.Name) > 100 {
		return fmt.Errorf("name must not exceed 100 characters")
	}
	if len(r.Description) > 500 {
		return fmt.Errorf("description must not exceed 500 characters")
	}
	return nil
}

type AddParticipantsRequest struct {
	Participants []string `json:"participants" binding:"required" example:"5511999999999,5511888888888"`
}

func (r AddParticipantsRequest) Validate() error {
	if len(r.Participants) == 0 {
		return fmt.Errorf("participants list cannot be empty")
	}
	if len(r.Participants) > 50 {
		return fmt.Errorf("maximum 50 participants can be added at once")
	}
	return nil
}

type RemoveParticipantsRequest struct {
	Participants []string `json:"participants" binding:"required" example:"5511999999999,5511888888888"`
}

func (r RemoveParticipantsRequest) Validate() error {
	if len(r.Participants) == 0 {
		return fmt.Errorf("participants list cannot be empty")
	}
	return nil
}

type PromoteParticipantsRequest struct {
	Participants []string `json:"participants" binding:"required" example:"5511999999999,5511888888888"`
}

func (r PromoteParticipantsRequest) Validate() error {
	if len(r.Participants) == 0 {
		return fmt.Errorf("participants list cannot be empty")
	}
	return nil
}

type DemoteParticipantsRequest struct {
	Participants []string `json:"participants" binding:"required" example:"5511999999999,5511888888888"`
}

func (r DemoteParticipantsRequest) Validate() error {
	if len(r.Participants) == 0 {
		return fmt.Errorf("participants list cannot be empty")
	}
	return nil
}

type GetGroupInfoRequest struct {
	GroupJID string `json:"group_jid" binding:"required" example:"120363025246125486@g.us"`
}

func (r GetGroupInfoRequest) Validate() error {
	if strings.TrimSpace(r.GroupJID) == "" {
		return fmt.Errorf("group_jid is required")
	}
	return nil
}

type JoinGroupRequest struct {
	GroupJID string `json:"group_jid" binding:"required" example:"120363025246125486@g.us"`
}

func (r JoinGroupRequest) Validate() error {
	if strings.TrimSpace(r.GroupJID) == "" {
		return fmt.Errorf("group_jid is required")
	}
	return nil
}

type JoinGroupWithInviteRequest struct {
	InviteCode string `json:"invite_code" binding:"required" example:"abc123def456"`
}

func (r JoinGroupWithInviteRequest) Validate() error {
	if strings.TrimSpace(r.InviteCode) == "" {
		return fmt.Errorf("invite_code is required")
	}
	return nil
}

type LeaveGroupRequest struct {
	GroupJID string `json:"group_jid" binding:"required" example:"120363025246125486@g.us"`
}

func (r LeaveGroupRequest) Validate() error {
	if strings.TrimSpace(r.GroupJID) == "" {
		return fmt.Errorf("group_jid is required")
	}
	return nil
}

type GetInviteLinkRequest struct {
	GroupJID string `json:"group_jid" binding:"required" example:"120363025246125486@g.us"`
	Reset    bool   `json:"reset,omitempty" example:"false"`
}

func (r GetInviteLinkRequest) Validate() error {
	if strings.TrimSpace(r.GroupJID) == "" {
		return fmt.Errorf("group_jid is required")
	}
	return nil
}

type GetInviteInfoRequest struct {
	InviteCode string `json:"invite_code" binding:"required" example:"abc123def456"`
	Reset      bool   `json:"reset,omitempty" example:"false"`
}

func (r GetInviteInfoRequest) Validate() error {
	if strings.TrimSpace(r.InviteCode) == "" {
		return fmt.Errorf("invite_code is required")
	}
	return nil
}

type GroupErrorResponse struct {
	Code    string `json:"code" example:"GROUP_NOT_FOUND"`
	Message string `json:"message" example:"Group not found"`
	Details string `json:"details" example:"Group with ID 'group123' not found"`
}

type GroupParticipant struct {
	JID          string `json:"jid" example:"5511999999999@s.whatsapp.net"`
	Phone        string `json:"phone" example:"5511999999999"`
	Name         string `json:"name,omitempty" example:"João Silva"`
	IsAdmin      bool   `json:"is_admin" example:"false"`
	IsSuperAdmin bool   `json:"is_super_admin" example:"false"`
}

type GroupInfo struct {
	JID              string             `json:"jid" example:"120363025246125486@g.us"`
	Name             string             `json:"name" example:"My Group"`
	Topic            string             `json:"topic,omitempty" example:"Group topic"`
	Description      string             `json:"description,omitempty" example:"This is my group"`
	Owner            string             `json:"owner,omitempty" example:"5511999999999@s.whatsapp.net"`
	CreatedAt        time.Time          `json:"created_at,omitempty" example:"2023-01-01T12:00:00Z"`
	Participants     []GroupParticipant `json:"participants,omitempty"`
	Admins           []string           `json:"admins,omitempty"`
	ParticipantCount int                `json:"participant_count" example:"5"`
	Size             int                `json:"size" example:"5"`
	IsAnnounce       bool               `json:"is_announce" example:"false"`
	Announce         bool               `json:"announce" example:"false"`
	IsLocked         bool               `json:"is_locked" example:"false"`
	Locked           bool               `json:"locked" example:"false"`
	IsEphemeral      bool               `json:"is_ephemeral" example:"false"`
	Ephemeral        bool               `json:"ephemeral" example:"false"`
}

type GroupResponse struct {
	Success bool                `json:"success"`
	Code    int                 `json:"code"`
	Data    *GroupResponseData  `json:"data,omitempty"`
	Error   *GroupErrorResponse `json:"error,omitempty"`
}

type GroupResponseData struct {
	SessionId string     `json:"session_id"`
	Action    string     `json:"action"`
	Status    string     `json:"status"`
	Message   string     `json:"message,omitempty"`
	Group     *GroupInfo `json:"group,omitempty"`
	Timestamp time.Time  `json:"timestamp"`
}

type GroupListResponse struct {
	Success bool                `json:"success"`
	Code    int                 `json:"code"`
	Data    *GroupListData      `json:"data,omitempty"`
	Error   *GroupErrorResponse `json:"error,omitempty"`
}

type GroupListData struct {
	SessionId string      `json:"session_id"`
	Groups    []GroupInfo `json:"groups"`
	Count     int         `json:"count"`
	Total     int         `json:"total"`
}

type GroupActionResponse struct {
	Success bool                `json:"success"`
	Code    int                 `json:"code"`
	Data    *GroupActionData    `json:"data,omitempty"`
	Error   *GroupErrorResponse `json:"error,omitempty"`
}

type GroupActionData struct {
	SessionId    string    `json:"session_id"`
	GroupJID     string    `json:"group_jid"`
	Action       string    `json:"action"`
	Status       string    `json:"status"`
	Message      string    `json:"message,omitempty"`
	Participants []string  `json:"participants,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
}

func NewGroupErrorResponse(code int, errorCode, message, details string) *GroupResponse {
	return &GroupResponse{
		Success: false,
		Code:    code,
		Error: &GroupErrorResponse{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}
}

func NewGroupSuccessResponse(sessionID, action, message string, group *GroupInfo) *GroupResponse {
	return &GroupResponse{
		Success: true,
		Code:    http.StatusOK,
		Data: &GroupResponseData{
			SessionId: sessionID,
			Action:    action,
			Status:    "success",
			Message:   message,
			Group:     group,
			Timestamp: time.Now(),
		},
	}
}

func NewGroupListSuccessResponse(sessionID string, groups []GroupInfo) *GroupListResponse {
	return &GroupListResponse{
		Success: true,
		Code:    http.StatusOK,
		Data: &GroupListData{
			SessionId: sessionID,
			Groups:    groups,
			Count:     len(groups),
			Total:     len(groups),
		},
	}
}

func NewGroupListErrorResponse(code int, errorCode, message, details string) *GroupListResponse {
	return &GroupListResponse{
		Success: false,
		Code:    code,
		Error: &GroupErrorResponse{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}
}

func NewGroupActionSuccessResponse(sessionID, groupJID, action, message string, participants []string) *GroupActionResponse {
	return &GroupActionResponse{
		Success: true,
		Code:    http.StatusOK,
		Data: &GroupActionData{
			SessionId:    sessionID,
			GroupJID:     groupJID,
			Action:       action,
			Status:       "success",
			Message:      message,
			Participants: participants,
			Timestamp:    time.Now(),
		},
	}
}

func NewGroupActionErrorResponse(code int, errorCode, message, details string) *GroupActionResponse {
	return &GroupActionResponse{
		Success: false,
		Code:    code,
		Error: &GroupErrorResponse{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}
}

func NewGroupListResponse(sessionID string, groups []GroupInfo) *GroupListResponse {
	return NewGroupListSuccessResponse(sessionID, groups)
}

func NewGroupOperationResponse(sessionID, groupJID, action string) *GroupActionResponse {
	return NewGroupActionSuccessResponse(sessionID, groupJID, action, "Operation completed successfully", nil)
}

type InviteLinkResponse struct {
	Success bool                `json:"success"`
	Code    int                 `json:"code"`
	Data    *InviteLinkData     `json:"data,omitempty"`
	Error   *GroupErrorResponse `json:"error,omitempty"`
}

type InviteLinkData struct {
	GroupJID   string `json:"group_jid"`
	InviteLink string `json:"invite_link"`
	InviteCode string `json:"invite_code"`
}

func NewInviteLinkResponse(groupJID, inviteLink, inviteCode string) *InviteLinkResponse {
	return &InviteLinkResponse{
		Success: true,
		Code:    http.StatusOK,
		Data: &InviteLinkData{
			GroupJID:   groupJID,
			InviteLink: inviteLink,
			InviteCode: inviteCode,
		},
	}
}

type UpdateParticipantsRequest struct {
	GroupJID     string   `json:"group_jid" binding:"required" example:"120363025246125486@g.us"`
	Participants []string `json:"participants" binding:"required" example:"5511999999999@s.whatsapp.net"`
	Action       string   `json:"action" binding:"required" example:"add"`
}

func (r UpdateParticipantsRequest) Validate() error {
	if strings.TrimSpace(r.GroupJID) == "" {
		return fmt.Errorf("group_jid is required")
	}
	if len(r.Participants) == 0 {
		return fmt.Errorf("participants list cannot be empty")
	}
	if r.Action != "add" && r.Action != "remove" && r.Action != "promote" && r.Action != "demote" {
		return fmt.Errorf("action must be one of: add, remove, promote, demote")
	}
	return nil
}

type SetGroupNameRequest struct {
	GroupJID string `json:"group_jid" binding:"required" example:"120363025246125486@g.us"`
	Name     string `json:"name" binding:"required" example:"New Group Name"`
}

func (r SetGroupNameRequest) Validate() error {
	if strings.TrimSpace(r.GroupJID) == "" {
		return fmt.Errorf("group_jid is required")
	}
	if strings.TrimSpace(r.Name) == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}

type SetGroupTopicRequest struct {
	GroupJID string `json:"group_jid" binding:"required" example:"120363025246125486@g.us"`
	Topic    string `json:"topic" binding:"required" example:"New group topic"`
}

func (r SetGroupTopicRequest) Validate() error {
	if strings.TrimSpace(r.GroupJID) == "" {
		return fmt.Errorf("group_jid is required")
	}
	if strings.TrimSpace(r.Topic) == "" {
		return fmt.Errorf("topic is required")
	}
	return nil
}

type SetGroupPhotoRequest struct {
	GroupJID string `json:"group_jid" binding:"required" example:"120363025246125486@g.us"`
	Photo    string `json:"photo" binding:"required" example:"base64_encoded_image_data"`
}

func (r SetGroupPhotoRequest) Validate() error {
	if strings.TrimSpace(r.GroupJID) == "" {
		return fmt.Errorf("group_jid is required")
	}
	if strings.TrimSpace(r.Photo) == "" {
		return fmt.Errorf("photo is required")
	}
	return nil
}

type RemoveGroupPhotoRequest struct {
	GroupJID string `json:"group_jid" binding:"required" example:"120363025246125486@g.us"`
}

func (r RemoveGroupPhotoRequest) Validate() error {
	if strings.TrimSpace(r.GroupJID) == "" {
		return fmt.Errorf("group_jid is required")
	}
	return nil
}

type SetGroupAnnounceRequest struct {
	GroupJID     string `json:"group_jid" binding:"required" example:"120363025246125486@g.us"`
	AnnounceOnly bool   `json:"announce_only" example:"true"`
}

func (r SetGroupAnnounceRequest) Validate() error {
	if strings.TrimSpace(r.GroupJID) == "" {
		return fmt.Errorf("group_jid is required")
	}
	return nil
}

type SetGroupLockedRequest struct {
	GroupJID string `json:"group_jid" binding:"required" example:"120363025246125486@g.us"`
	Locked   bool   `json:"locked" example:"true"`
}

func (r SetGroupLockedRequest) Validate() error {
	if strings.TrimSpace(r.GroupJID) == "" {
		return fmt.Errorf("group_jid is required")
	}
	return nil
}

type SetGroupEphemeralRequest struct {
	GroupJID  string `json:"group_jid" binding:"required" example:"120363025246125486@g.us"`
	Ephemeral bool   `json:"ephemeral" example:"true"`
	Duration  int    `json:"duration,omitempty" example:"86400"`
}

func (r SetGroupEphemeralRequest) Validate() error {
	if strings.TrimSpace(r.GroupJID) == "" {
		return fmt.Errorf("group_jid is required")
	}
	return nil
}

type GroupJoinApprovalReq struct {
	GroupJID        string `json:"group_jid" binding:"required" example:"120363025246125486@g.us"`
	RequireApproval bool   `json:"require_approval" example:"true"`
}

func (r GroupJoinApprovalReq) Validate() error {
	if strings.TrimSpace(r.GroupJID) == "" {
		return fmt.Errorf("group_jid is required")
	}
	return nil
}

type GroupMemberModeReq struct {
	GroupJID string `json:"group_jid" binding:"required" example:"120363025246125486@g.us"`
	Mode     string `json:"mode" binding:"required" example:"admin_add"`
}

func (r GroupMemberModeReq) Validate() error {
	if strings.TrimSpace(r.GroupJID) == "" {
		return fmt.Errorf("group_jid is required")
	}
	if r.Mode != "admin_add" && r.Mode != "all_member_add" {
		return fmt.Errorf("mode must be either 'admin_add' or 'all_member_add'")
	}
	return nil
}

type GetGroupRequestsReq struct {
	GroupJID string `json:"group_jid" binding:"required" example:"120363025246125486@g.us"`
}

func (r GetGroupRequestsReq) Validate() error {
	if strings.TrimSpace(r.GroupJID) == "" {
		return fmt.Errorf("group_jid is required")
	}
	return nil
}

type UpdateGroupRequestsReq struct {
	GroupJID     string   `json:"group_jid" binding:"required" example:"120363025246125486@g.us"`
	Participants []string `json:"participants" binding:"required" example:"5511999999999@s.whatsapp.net"`
	Action       string   `json:"action" binding:"required" example:"approve"`
}

func (r UpdateGroupRequestsReq) Validate() error {
	if strings.TrimSpace(r.GroupJID) == "" {
		return fmt.Errorf("group_jid is required")
	}
	if len(r.Participants) == 0 {
		return fmt.Errorf("participants list cannot be empty")
	}
	if r.Action != "approve" && r.Action != "reject" {
		return fmt.Errorf("action must be either 'approve' or 'reject'")
	}
	return nil
}
