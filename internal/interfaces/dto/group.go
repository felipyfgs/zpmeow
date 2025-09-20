package dto

import (
	"fmt"
	"time"
)

type CreateGroupRequest struct {
	Name         string   `json:"name" binding:"required" example:"My Group"`
	Participants []string `json:"participants" binding:"required" example:"[\"5511999999999\", \"5511888888888\"]"`
}

type GetGroupInfoRequest struct {
	GroupJID string `json:"group_jid" binding:"required" example:"120363025246125486@g.us"`
}

type JoinGroupRequest struct {
	GroupJID string `json:"group_jid" binding:"required" example:"120363025246125486@g.us"`
}

type JoinGroupWithInviteRequest struct {
	InviteCode string `json:"invite_code" binding:"required" example:"ABC123DEF456"`
}

type LeaveGroupRequest struct {
	GroupJID string `json:"group_jid" binding:"required" example:"120363025246125486@g.us"`
}

type GetInviteLinkRequest struct {
	GroupJID string `json:"group_jid" binding:"required" example:"120363025246125486@g.us"`
	Reset    bool   `json:"reset" example:"false"`
}

type GetInviteInfoRequest struct {
	InviteCode string `json:"invite_code" binding:"required" example:"ABC123DEF456"`
}

type GroupInviteInfoReq struct {
	InviteCode string `json:"invite_code" binding:"required" example:"ABC123DEF456"`
}

type UpdateParticipantsRequest struct {
	GroupJID     string   `json:"group_jid" binding:"required" example:"120363025246125486@g.us"`
	Action       string   `json:"action" binding:"required" example:"add"` // "add" or "remove"
	Participants []string `json:"participants" binding:"required" example:"[\"5511999999999\", \"5511888888888\"]"`
}

type SetGroupNameRequest struct {
	GroupJID string `json:"group_jid" binding:"required" example:"120363025246125486@g.us"`
	Name     string `json:"name" binding:"required" example:"New Group Name"`
}

type SetGroupTopicRequest struct {
	GroupJID string `json:"group_jid" binding:"required" example:"120363025246125486@g.us"`
	Topic    string `json:"topic" binding:"required" example:"Group topic description"`
}

type SetGroupPhotoRequest struct {
	GroupJID string `json:"group_jid" binding:"required" example:"120363025246125486@g.us"`
	Photo    string `json:"photo" binding:"required" example:"base64_encoded_image"`
}

type RemoveGroupPhotoRequest struct {
	GroupJID string `json:"group_jid" binding:"required" example:"120363025246125486@g.us"`
}

type SetGroupAnnounceRequest struct {
	GroupJID string `json:"group_jid" binding:"required" example:"120363025246125486@g.us"`
	Announce bool   `json:"announce" example:"true"`
}

type SetGroupLockedRequest struct {
	GroupJID string `json:"group_jid" binding:"required" example:"120363025246125486@g.us"`
	Locked   bool   `json:"locked" example:"true"`
}

type SetGroupEphemeralRequest struct {
	GroupJID  string `json:"group_jid" binding:"required" example:"120363025246125486@g.us"`
	Ephemeral bool   `json:"ephemeral" example:"true"`
}

type SetGroupJoinApprovalRequest struct {
	GroupJID     string `json:"group_jid" binding:"required" example:"120363025246125486@g.us"`
	JoinApproval bool   `json:"join_approval" example:"true"`
}

type SetGroupMemberAddModeRequest struct {
	GroupJID      string `json:"group_jid" binding:"required" example:"120363025246125486@g.us"`
	MemberAddMode string `json:"member_add_mode" binding:"required" example:"admin_only"`
}

type GetGroupRequestParticipantsRequest struct {
	GroupJID string `json:"group_jid" binding:"required" example:"120363025246125486@g.us"`
}

type UpdateGroupRequestParticipantsRequest struct {
	GroupJID     string   `json:"group_jid" binding:"required" example:"120363025246125486@g.us"`
	Action       string   `json:"action" binding:"required" example:"approve"` // "approve" or "reject"
	Participants []string `json:"participants" binding:"required" example:"[\"5511999999999\", \"5511888888888\"]"`
}

type GroupJoinApprovalReq = SetGroupJoinApprovalRequest
type GroupMemberModeReq = SetGroupMemberAddModeRequest
type GetGroupRequestsReq = GetGroupRequestParticipantsRequest
type UpdateGroupRequestsReq = UpdateGroupRequestParticipantsRequest

func (r *CreateGroupRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("group name is required")
	}
	if len(r.Participants) == 0 {
		return fmt.Errorf("at least one participant is required")
	}
	return nil
}

func (r *GetGroupInfoRequest) Validate() error {
	if r.GroupJID == "" {
		return fmt.Errorf("group JID is required")
	}
	return nil
}

func (r *JoinGroupRequest) Validate() error {
	if r.GroupJID == "" {
		return fmt.Errorf("group JID is required")
	}
	return nil
}

func (r *JoinGroupWithInviteRequest) Validate() error {
	if r.InviteCode == "" {
		return fmt.Errorf("invite code is required")
	}
	return nil
}

func (r *LeaveGroupRequest) Validate() error {
	if r.GroupJID == "" {
		return fmt.Errorf("group JID is required")
	}
	return nil
}

func (r *GetInviteLinkRequest) Validate() error {
	if r.GroupJID == "" {
		return fmt.Errorf("group JID is required")
	}
	return nil
}

func (r *GetInviteInfoRequest) Validate() error {
	if r.InviteCode == "" {
		return fmt.Errorf("invite code is required")
	}
	return nil
}

func (r *GroupInviteInfoReq) Validate() error {
	if r.InviteCode == "" {
		return fmt.Errorf("invite code is required")
	}
	return nil
}

func (r *UpdateParticipantsRequest) Validate() error {
	if r.GroupJID == "" {
		return fmt.Errorf("group JID is required")
	}
	if r.Action == "" {
		return fmt.Errorf("action is required")
	}
	if r.Action != "add" && r.Action != "remove" {
		return fmt.Errorf("action must be 'add' or 'remove'")
	}
	if len(r.Participants) == 0 {
		return fmt.Errorf("at least one participant is required")
	}
	return nil
}

func (r *SetGroupNameRequest) Validate() error {
	if r.GroupJID == "" {
		return fmt.Errorf("group JID is required")
	}
	if r.Name == "" {
		return fmt.Errorf("group name is required")
	}
	return nil
}

func (r *SetGroupTopicRequest) Validate() error {
	if r.GroupJID == "" {
		return fmt.Errorf("group JID is required")
	}
	return nil
}

func (r *SetGroupPhotoRequest) Validate() error {
	if r.GroupJID == "" {
		return fmt.Errorf("group JID is required")
	}
	if r.Photo == "" {
		return fmt.Errorf("photo is required")
	}
	return nil
}

func (r *RemoveGroupPhotoRequest) Validate() error {
	if r.GroupJID == "" {
		return fmt.Errorf("group JID is required")
	}
	return nil
}

func (r *SetGroupAnnounceRequest) Validate() error {
	if r.GroupJID == "" {
		return fmt.Errorf("group JID is required")
	}
	return nil
}

func (r *SetGroupLockedRequest) Validate() error {
	if r.GroupJID == "" {
		return fmt.Errorf("group JID is required")
	}
	return nil
}

func (r *SetGroupEphemeralRequest) Validate() error {
	if r.GroupJID == "" {
		return fmt.Errorf("group JID is required")
	}
	return nil
}

func (r *SetGroupJoinApprovalRequest) Validate() error {
	if r.GroupJID == "" {
		return fmt.Errorf("group JID is required")
	}
	return nil
}

func (r *SetGroupMemberAddModeRequest) Validate() error {
	if r.GroupJID == "" {
		return fmt.Errorf("group JID is required")
	}
	if r.MemberAddMode == "" {
		return fmt.Errorf("member add mode is required")
	}
	return nil
}

func (r *GetGroupRequestParticipantsRequest) Validate() error {
	if r.GroupJID == "" {
		return fmt.Errorf("group JID is required")
	}
	return nil
}

func (r *UpdateGroupRequestParticipantsRequest) Validate() error {
	if r.GroupJID == "" {
		return fmt.Errorf("group JID is required")
	}
	if r.Action == "" {
		return fmt.Errorf("action is required")
	}
	if r.Action != "approve" && r.Action != "reject" {
		return fmt.Errorf("action must be 'approve' or 'reject'")
	}
	if len(r.Participants) == 0 {
		return fmt.Errorf("at least one participant is required")
	}
	return nil
}

type GroupInfo struct {
	JID              string   `json:"jid"`
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	Topic            string   `json:"topic"`
	Participants     []string `json:"participants"`
	Admins           []string `json:"admins"`
	Owner            string   `json:"owner"`
	CreatedAt        int64    `json:"createdAt"`
	Size             int      `json:"size"`
	IsAnnounce       bool     `json:"isAnnounce"`
	IsLocked         bool     `json:"isLocked"`
	IsEphemeral      bool     `json:"isEphemeral"`
	Announce         bool     `json:"announce"`
	Locked           bool     `json:"locked"`
	Ephemeral        bool     `json:"ephemeral"`
	ParticipantCount int      `json:"participantCount"`
}

type GroupList struct {
	Groups []GroupInfo `json:"groups"`
	Total  int         `json:"total"`
}

type InviteInfo struct {
	GroupJID   string `json:"groupJid"`
	GroupName  string `json:"groupName"`
	InviteCode string `json:"inviteCode"`
	Inviter    string `json:"inviter"`
	ExpiresAt  int64  `json:"expiresAt"`
	IsValid    bool   `json:"isValid"`
}

type GroupResponse struct {
	Success bool                `json:"success"`
	Code    int                 `json:"code"`
	Data    GroupData           `json:"data"`
	Error   *GroupErrorResponse `json:"error,omitempty"`
}

type GroupData struct {
	SessionID  string      `json:"session_id" example:"default"`
	Action     string      `json:"action" example:"create"`
	Status     string      `json:"status" example:"success"`
	Timestamp  time.Time   `json:"timestamp" example:"2023-01-01T00:00:00Z"`
	Group      *GroupInfo  `json:"group,omitempty"`
	Groups     []GroupInfo `json:"groups,omitempty"`
	InviteLink string      `json:"invite_link,omitempty"`
	Message    string      `json:"message,omitempty"`
}

type GroupErrorResponse struct {
	Code    string `json:"code" example:"INVALID_GROUP_JID"`
	Message string `json:"message" example:"Invalid group JID format"`
	Details string `json:"details,omitempty" example:"Group JID must end with @g.us"`
}

func NewGroupSuccessResponse(sessionID, action string, group *GroupInfo) *GroupResponse {
	return &GroupResponse{
		Success: true,
		Code:    200,
		Data: GroupData{
			SessionID: sessionID,
			Action:    action,
			Status:    "success",
			Timestamp: time.Now(),
			Group:     group,
		},
	}
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

func NewGroupListResponse(sessionID string, groups []GroupInfo) *GroupResponse {
	return &GroupResponse{
		Success: true,
		Code:    200,
		Data: GroupData{
			SessionID: sessionID,
			Action:    "list_groups",
			Status:    "success",
			Timestamp: time.Now(),
			Groups:    groups,
		},
	}
}

func NewGroupOperationResponse(sessionID, action, message string) *GroupResponse {
	return &GroupResponse{
		Success: true,
		Code:    200,
		Data: GroupData{
			SessionID: sessionID,
			Action:    action,
			Status:    "success",
			Timestamp: time.Now(),
			Message:   message,
		},
	}
}

func NewInviteLinkResponse(sessionID, groupJID, inviteLink string) *GroupResponse {
	return &GroupResponse{
		Success: true,
		Code:    200,
		Data: GroupData{
			SessionID:  sessionID,
			Action:     "get_invite_link",
			Status:     "success",
			Timestamp:  time.Now(),
			InviteLink: inviteLink,
		},
	}
}
