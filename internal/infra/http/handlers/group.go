package handlers

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"zpmeow/internal/application"
	"zpmeow/internal/infra/http/dto"
	"zpmeow/internal/infra/wmeow"

	"github.com/gin-gonic/gin"
)

type GroupHandler struct {
	groupService *application.GroupApp
	wmeowService wmeow.WameowService
}

func NewGroupHandler(groupService *application.GroupApp, wmeowService wmeow.WameowService) *GroupHandler {
	return &GroupHandler{
		groupService: groupService,
		wmeowService: wmeowService,
	}
}

func (h *GroupHandler) resolveSessionID(_ *gin.Context, sessionIDOrName string) (string, error) {
	return sessionIDOrName, nil
}

// CreateGroup godoc
// @Summary Create a new WhatsApp group
// @Description Creates a new WhatsApp group with specified participants
// @Tags Groups
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.CreateGroupRequest true "Group creation request"
// @Success 201 {object} dto.GroupResponse "Group created successfully"
// @Failure 400 {object} dto.GroupResponse "Invalid request data"
// @Failure 401 {object} dto.GroupResponse "Unauthorized - Invalid API key" "Invalid request data"
// @Failure 404 {object} dto.GroupResponse "Session not found"
// @Failure 500 {object} dto.GroupResponse "Failed to create group"
// @Router /session/{sessionId}/group/create [post]
func (h *GroupHandler) CreateGroup(c *gin.Context) {
	sessionIDOrName := c.Param("sessionId")
	if sessionIDOrName == "" {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
		return
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewGroupErrorResponse(
			http.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
		return
	}

	var req dto.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"VALIDATION_ERROR",
			"Request validation failed",
			err.Error(),
		))
		return
	}

	ctx := c.Request.Context()
	groupInfo, err := h.wmeowService.CreateGroup(ctx, sessionID, req.Name, req.Participants)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewGroupErrorResponse(
			http.StatusInternalServerError,
			"CREATE_GROUP_FAILED",
			"Failed to create group",
			err.Error(),
		))
		return
	}

	dtoGroupInfo := convertWmeowGroupInfoToDTO(groupInfo)

	response := dto.NewGroupSuccessResponse(sessionID, "create", "success", dtoGroupInfo)
	response.Code = http.StatusCreated
	c.JSON(http.StatusCreated, response)
}

// GetGroupInfo godoc
// @Summary Get group information
// @Description Retrieves detailed information about a WhatsApp group
// @Tags Groups
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.GetGroupInfoRequest true "Group info request"
// @Success 200 {object} dto.GroupResponse "Group information"
// @Failure 400 {object} dto.GroupResponse "Invalid request data"
// @Failure 401 {object} dto.GroupResponse "Unauthorized - Invalid API key" "Invalid request data"
// @Failure 404 {object} dto.GroupResponse "Session not found"
// @Failure 500 {object} dto.GroupResponse "Failed to get group info"
// @Router /session/{sessionId}/group/info [post]
func (h *GroupHandler) GetGroupInfo(c *gin.Context) {
	sessionIDOrName := c.Param("sessionId")
	if sessionIDOrName == "" {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
		return
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewGroupErrorResponse(
			http.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
		return
	}

	var req dto.GetGroupInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
		return
	}

	ctx := c.Request.Context()
	groupInfo, err := h.wmeowService.GetGroupInfo(ctx, sessionID, req.GroupJID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewGroupErrorResponse(
			http.StatusInternalServerError,
			"GET_GROUP_INFO_FAILED",
			"Failed to get group information",
			err.Error(),
		))
		return
	}

	dtoGroupInfo := convertWmeowGroupInfoToDTO(groupInfo)

	response := dto.NewGroupSuccessResponse(sessionID, "info", "success", dtoGroupInfo)
	c.JSON(http.StatusOK, response)
}

// ListGroups godoc
// @Summary List WhatsApp groups
// @Description Retrieves a list of all WhatsApp groups for a session
// @Tags Groups
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Success 200 {object} dto.GroupResponse "Groups list"
// @Failure 400 {object} dto.GroupResponse "Invalid request data"
// @Failure 401 {object} dto.GroupResponse "Unauthorized - Invalid API key" "Invalid request data"
// @Failure 404 {object} dto.GroupResponse "Session not found"
// @Failure 500 {object} dto.GroupResponse "Failed to list groups"
// @Router /session/{sessionId}/group/list [get]
func (h *GroupHandler) ListGroups(c *gin.Context) {
	sessionIDOrName := c.Param("sessionId")
	if sessionIDOrName == "" {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
		return
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewGroupErrorResponse(
			http.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
		return
	}

	ctx := c.Request.Context()

	req := application.ListGroupsRequest{
		SessionID: sessionID,
	}

	result, err := h.groupService.ListGroups(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewGroupErrorResponse(
			http.StatusInternalServerError,
			"LIST_GROUPS_FAILED",
			"Failed to list groups",
			err.Error(),
		))
		return
	}

	var dtoGroups []dto.GroupInfo
	for _, group := range result.Groups {
		var participants []dto.GroupParticipant
		for _, participantJID := range group.Participants {
			participants = append(participants, dto.GroupParticipant{
				JID:          participantJID,
				Phone:        participantJID,
				IsAdmin:      false,
				IsSuperAdmin: false,
			})
		}

		dtoGroups = append(dtoGroups, dto.GroupInfo{
			JID:          group.JID,
			Name:         group.Name,
			Description:  group.Description,
			Participants: participants,
			Admins:       group.Admins,
			Owner:        group.Owner,
			IsAnnounce:   group.IsAnnounce,
			IsLocked:     group.IsLocked,
		})
	}

	response := dto.NewGroupListResponse(sessionID, dtoGroups)
	c.JSON(http.StatusOK, response)
}

// JoinGroup godoc
// @Summary Join a WhatsApp group
// @Description Joins a WhatsApp group using group JID
// @Tags Groups
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.JoinGroupRequest true "Join group request"
// @Success 200 {object} dto.GroupResponse "Joined group successfully"
// @Failure 400 {object} dto.GroupResponse "Invalid request data"
// @Failure 401 {object} dto.GroupResponse "Unauthorized - Invalid API key" "Invalid request data"
// @Failure 404 {object} dto.GroupResponse "Session not found"
// @Failure 500 {object} dto.GroupResponse "Failed to join group"
// @Router /session/{sessionId}/group/join [post]
func (h *GroupHandler) JoinGroup(c *gin.Context) {
	sessionIDOrName := c.Param("sessionId")
	if sessionIDOrName == "" {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
		return
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewGroupErrorResponse(
			http.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
		return
	}

	var req dto.JoinGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
		return
	}

	ctx := c.Request.Context()
	_, err = h.wmeowService.JoinGroup(ctx, sessionID, req.GroupJID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewGroupErrorResponse(
			http.StatusInternalServerError,
			"JOIN_GROUP_FAILED",
			"Failed to join group",
			err.Error(),
		))
		return
	}

	groupInfo, err := h.wmeowService.GetGroupInfo(ctx, sessionID, req.GroupJID)
	if err != nil {
		response := dto.NewGroupSuccessResponse(sessionID, "join", "success", nil)
		response.Data.Message = "Group joined successfully"
		c.JSON(http.StatusOK, response)
		return
	}

	dtoGroupInfo := convertWmeowGroupInfoToDTO(groupInfo)
	response := dto.NewGroupSuccessResponse(sessionID, "join", "success", dtoGroupInfo)
	c.JSON(http.StatusOK, response)
}

// JoinGroupWithInvite godoc
// @Summary Join group with invite link
// @Description Joins a WhatsApp group using an invite link
// @Tags Groups
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.JoinGroupWithInviteRequest true "Join group with invite request"
// @Success 200 {object} dto.GroupResponse "Joined group successfully"
// @Failure 400 {object} dto.GroupResponse "Invalid request data"
// @Failure 401 {object} dto.GroupResponse "Unauthorized - Invalid API key" "Invalid request data"
// @Failure 404 {object} dto.GroupResponse "Session not found"
// @Failure 500 {object} dto.GroupResponse "Failed to join group"
// @Router /session/{sessionId}/group/join-with-invite [post]
func (h *GroupHandler) JoinGroupWithInvite(c *gin.Context) {
	sessionIDOrName := c.Param("sessionId")
	if sessionIDOrName == "" {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
		return
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewGroupErrorResponse(
			http.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
		return
	}

	var req dto.JoinGroupWithInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
		return
	}

	ctx := c.Request.Context()
	groupInfo, err := h.wmeowService.JoinGroupWithInvite(ctx, sessionID, "", "", req.InviteCode, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewGroupErrorResponse(
			http.StatusInternalServerError,
			"JOIN_GROUP_WITH_INVITE_FAILED",
			"Failed to join group with invite",
			err.Error(),
		))
		return
	}

	dtoGroupInfo := convertWmeowGroupInfoToDTO(groupInfo)

	response := dto.NewGroupSuccessResponse(sessionID, "join_with_invite", "success", dtoGroupInfo)
	c.JSON(http.StatusOK, response)
}

// LeaveGroup godoc
// @Summary Leave a WhatsApp group
// @Description Leaves a WhatsApp group
// @Tags Groups
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.LeaveGroupRequest true "Leave group request"
// @Success 200 {object} dto.GroupResponse "Left group successfully"
// @Failure 400 {object} dto.GroupResponse "Invalid request data"
// @Failure 401 {object} dto.GroupResponse "Unauthorized - Invalid API key" "Invalid request data"
// @Failure 404 {object} dto.GroupResponse "Session not found"
// @Failure 500 {object} dto.GroupResponse "Failed to leave group"
// @Router /session/{sessionId}/group/leave [post]
func (h *GroupHandler) LeaveGroup(c *gin.Context) {
	sessionIDOrName := c.Param("sessionId")
	if sessionIDOrName == "" {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
		return
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewGroupErrorResponse(
			http.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
		return
	}

	var req dto.LeaveGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
		return
	}

	ctx := c.Request.Context()
	err = h.wmeowService.LeaveGroup(ctx, sessionID, req.GroupJID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewGroupErrorResponse(
			http.StatusInternalServerError,
			"LEAVE_GROUP_FAILED",
			"Failed to leave group",
			err.Error(),
		))
		return
	}

	response := dto.NewGroupOperationResponse(sessionID, "leave", "Successfully left the group")
	c.JSON(http.StatusOK, response)
}

func (h *GroupHandler) GetInviteLink(c *gin.Context) {
	sessionIDOrName := c.Param("sessionId")
	if sessionIDOrName == "" {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
		return
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewGroupErrorResponse(
			http.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
		return
	}

	var req dto.GetInviteLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
		return
	}

	ctx := c.Request.Context()
	inviteLink, err := h.wmeowService.GetInviteLink(ctx, sessionID, req.GroupJID, req.Reset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewGroupErrorResponse(
			http.StatusInternalServerError,
			"GET_INVITE_LINK_FAILED",
			"Failed to get invite link",
			err.Error(),
		))
		return
	}

	response := dto.NewInviteLinkResponse(req.GroupJID, inviteLink, "")
	c.JSON(http.StatusOK, response)
}

func convertWmeowGroupInfoToDTO(groupInfo *wmeow.GroupInfo) *dto.GroupInfo {
	if groupInfo == nil {
		return nil
	}

	var participants []dto.GroupParticipant
	for _, participantJID := range groupInfo.Participants {
		participants = append(participants, dto.GroupParticipant{
			JID:          participantJID,
			Phone:        participantJID,
			IsAdmin:      false,
			IsSuperAdmin: false,
		})
	}

	createdAt := time.Unix(groupInfo.CreatedAt, 0)

	return &dto.GroupInfo{
		JID:              groupInfo.JID,
		Name:             groupInfo.Name,
		Topic:            groupInfo.Topic,
		Participants:     participants,
		Admins:           []string{},
		Owner:            groupInfo.CreatedBy,
		CreatedAt:        createdAt,
		Size:             len(groupInfo.Participants),
		ParticipantCount: len(groupInfo.Participants),
		Announce:         groupInfo.IsAnnounce,
		IsAnnounce:       groupInfo.IsAnnounce,
		Locked:           groupInfo.IsLocked,
		IsLocked:         groupInfo.IsLocked,
		Ephemeral:        groupInfo.IsEphemeral,
		IsEphemeral:      groupInfo.IsEphemeral,
	}
}

func (h *GroupHandler) GetInviteInfo(c *gin.Context) {
	sessionIDOrName := c.Param("sessionId")
	if sessionIDOrName == "" {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
		return
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewGroupErrorResponse(
			http.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
		return
	}

	var req dto.GetInviteInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
		return
	}

	ctx := c.Request.Context()
	inviteInfo, err := h.wmeowService.GetInviteInfo(ctx, sessionID, req.InviteCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewGroupErrorResponse(
			http.StatusInternalServerError,
			"GET_INVITE_INFO_FAILED",
			"Failed to get invite info",
			err.Error(),
		))
		return
	}

	response := dto.NewGroupOperationResponse(sessionID, "invite_info", "Invite info retrieved successfully")
	response.Data.Message = fmt.Sprintf("Group: %s, Created by: %s", inviteInfo.Name, inviteInfo.CreatedBy)
	c.JSON(http.StatusOK, response)
}

func (h *GroupHandler) GetGroupInfoFromInvite(c *gin.Context) {
	sessionIDOrName := c.Param("sessionId")
	if sessionIDOrName == "" {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
		return
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewGroupErrorResponse(
			http.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
		return
	}

	var req dto.GetInviteInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
		return
	}

	ctx := c.Request.Context()
	groupInfo, err := h.wmeowService.GetGroupInfoFromInvite(ctx, sessionID, "", "", req.InviteCode, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewGroupErrorResponse(
			http.StatusInternalServerError,
			"GET_GROUP_INFO_FROM_INVITE_FAILED",
			"Failed to get group info from invite",
			err.Error(),
		))
		return
	}

	dtoGroupInfo := convertWmeowGroupInfoToDTO(groupInfo)

	response := dto.NewGroupSuccessResponse(sessionID, "invite_info_specific", "success", dtoGroupInfo)
	c.JSON(http.StatusOK, response)
}

func (h *GroupHandler) UpdateParticipants(c *gin.Context) {
	sessionIDOrName := c.Param("sessionId")
	if sessionIDOrName == "" {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
		return
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewGroupErrorResponse(
			http.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
		return
	}

	var req dto.UpdateParticipantsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"VALIDATION_ERROR",
			"Request validation failed",
			err.Error(),
		))
		return
	}

	ctx := c.Request.Context()
	err = h.wmeowService.UpdateParticipants(ctx, sessionID, req.GroupJID, req.Action, req.Participants)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewGroupErrorResponse(
			http.StatusInternalServerError,
			"UPDATE_PARTICIPANTS_FAILED",
			"Failed to update participants",
			err.Error(),
		))
		return
	}

	message := "Successfully " + req.Action + "ed participants"
	response := dto.NewGroupOperationResponse(sessionID, "update_participants", message)
	c.JSON(http.StatusOK, response)
}

func (h *GroupHandler) SetName(c *gin.Context) {
	sessionIDOrName := c.Param("sessionId")
	if sessionIDOrName == "" {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
		return
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewGroupErrorResponse(
			http.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
		return
	}

	var req dto.SetGroupNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
		return
	}

	ctx := c.Request.Context()
	err = h.wmeowService.SetGroupName(ctx, sessionID, req.GroupJID, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewGroupErrorResponse(
			http.StatusInternalServerError,
			"SET_GROUP_NAME_FAILED",
			"Failed to set group name",
			err.Error(),
		))
		return
	}

	response := dto.NewGroupOperationResponse(sessionID, "set_name", "Group name updated successfully")
	c.JSON(http.StatusOK, response)
}

func (h *GroupHandler) SetTopic(c *gin.Context) {
	sessionIDOrName := c.Param("sessionId")
	if sessionIDOrName == "" {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
		return
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewGroupErrorResponse(
			http.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
		return
	}

	var req dto.SetGroupTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
		return
	}

	ctx := c.Request.Context()
	err = h.wmeowService.SetGroupTopic(ctx, sessionID, req.GroupJID, req.Topic)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewGroupErrorResponse(
			http.StatusInternalServerError,
			"SET_GROUP_TOPIC_FAILED",
			"Failed to set group topic",
			err.Error(),
		))
		return
	}

	response := dto.NewGroupOperationResponse(sessionID, "set_topic", "Group topic updated successfully")
	c.JSON(http.StatusOK, response)
}

func (h *GroupHandler) SetPhoto(c *gin.Context) {
	sessionIDOrName := c.Param("sessionId")
	if sessionIDOrName == "" {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
		return
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewGroupErrorResponse(
			http.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
		return
	}

	var req dto.SetGroupPhotoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
		return
	}

	photoData, err := base64.StdEncoding.DecodeString(req.Photo)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"INVALID_PHOTO_DATA",
			"Invalid base64 photo data",
			err.Error(),
		))
		return
	}

	ctx := c.Request.Context()
	err = h.wmeowService.SetGroupPhoto(ctx, sessionID, req.GroupJID, photoData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewGroupErrorResponse(
			http.StatusInternalServerError,
			"SET_GROUP_PHOTO_FAILED",
			"Failed to set group photo",
			err.Error(),
		))
		return
	}

	response := dto.NewGroupOperationResponse(sessionID, "set_photo", "Group photo updated successfully")
	c.JSON(http.StatusOK, response)
}

func (h *GroupHandler) RemovePhoto(c *gin.Context) {
	sessionIDOrName := c.Param("sessionId")
	if sessionIDOrName == "" {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
		return
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewGroupErrorResponse(
			http.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
		return
	}

	var req dto.RemoveGroupPhotoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
		return
	}

	ctx := c.Request.Context()
	err = h.wmeowService.RemoveGroupPhoto(ctx, sessionID, req.GroupJID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewGroupErrorResponse(
			http.StatusInternalServerError,
			"REMOVE_GROUP_PHOTO_FAILED",
			"Failed to remove group photo",
			err.Error(),
		))
		return
	}

	response := dto.NewGroupOperationResponse(sessionID, "remove_photo", "Group photo removed successfully")
	c.JSON(http.StatusOK, response)
}

func (h *GroupHandler) SetAnnounce(c *gin.Context) {
	sessionIDOrName := c.Param("sessionId")
	if sessionIDOrName == "" {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
		return
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewGroupErrorResponse(
			http.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
		return
	}

	var req dto.SetGroupAnnounceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
		return
	}

	ctx := c.Request.Context()
	err = h.wmeowService.SetGroupAnnounce(ctx, sessionID, req.GroupJID, req.AnnounceOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewGroupErrorResponse(
			http.StatusInternalServerError,
			"SET_GROUP_ANNOUNCE_FAILED",
			"Failed to set group announce setting",
			err.Error(),
		))
		return
	}

	response := dto.NewGroupOperationResponse(sessionID, "set_announce", "Group announce setting updated successfully")
	c.JSON(http.StatusOK, response)
}

func (h *GroupHandler) SetLocked(c *gin.Context) {
	sessionIDOrName := c.Param("sessionId")
	if sessionIDOrName == "" {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
		return
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewGroupErrorResponse(
			http.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
		return
	}

	var req dto.SetGroupLockedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
		return
	}

	ctx := c.Request.Context()
	err = h.wmeowService.SetGroupLocked(ctx, sessionID, req.GroupJID, req.Locked)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewGroupErrorResponse(
			http.StatusInternalServerError,
			"SET_GROUP_LOCKED_FAILED",
			"Failed to set group locked setting",
			err.Error(),
		))
		return
	}

	response := dto.NewGroupOperationResponse(sessionID, "set_locked", "Group locked setting updated successfully")
	c.JSON(http.StatusOK, response)
}

func (h *GroupHandler) SetEphemeral(c *gin.Context) {
	sessionIDOrName := c.Param("sessionId")
	if sessionIDOrName == "" {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
		return
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewGroupErrorResponse(
			http.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
		return
	}

	var req dto.SetGroupEphemeralRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
		return
	}

	ctx := c.Request.Context()
	duration := 0
	if req.Ephemeral {
		duration = 604800
	}
	err = h.wmeowService.SetGroupEphemeral(ctx, sessionID, req.GroupJID, req.Ephemeral, duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewGroupErrorResponse(
			http.StatusInternalServerError,
			"SET_GROUP_EPHEMERAL_FAILED",
			"Failed to set group ephemeral setting",
			err.Error(),
		))
		return
	}

	response := dto.NewGroupOperationResponse(sessionID, "set_ephemeral", "Group ephemeral setting updated successfully")
	c.JSON(http.StatusOK, response)
}

func (h *GroupHandler) SetJoinApproval(c *gin.Context) {
	sessionIDOrName := c.Param("sessionId")
	if sessionIDOrName == "" {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
		return
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewGroupErrorResponse(
			http.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
		return
	}

	var req dto.GroupJoinApprovalReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request body",
			err.Error(),
		))
		return
	}

	ctx := c.Request.Context()
	err = h.wmeowService.SetGroupJoinApproval(ctx, sessionID, req.GroupJID, req.RequireApproval)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewGroupErrorResponse(
			http.StatusInternalServerError,
			"SET_JOIN_APPROVAL_FAILED",
			"Failed to set group join approval mode",
			err.Error(),
		))
		return
	}

	response := dto.NewGroupOperationResponse(sessionID, "set_join_approval", "Group join approval mode updated successfully")
	c.JSON(http.StatusOK, response)
}

func (h *GroupHandler) SetMemberAddMode(c *gin.Context) {
	sessionIDOrName := c.Param("sessionId")
	if sessionIDOrName == "" {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
		return
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewGroupErrorResponse(
			http.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
		return
	}

	var req dto.GroupMemberModeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request body",
			err.Error(),
		))
		return
	}

	ctx := c.Request.Context()
	err = h.wmeowService.SetGroupMemberAddMode(ctx, sessionID, req.GroupJID, req.Mode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewGroupErrorResponse(
			http.StatusInternalServerError,
			"SET_MEMBER_ADD_MODE_FAILED",
			"Failed to set group member add mode",
			err.Error(),
		))
		return
	}

	response := dto.NewGroupOperationResponse(sessionID, "set_member_add_mode", "Group member add mode updated successfully")
	c.JSON(http.StatusOK, response)
}

func (h *GroupHandler) GetGroupRequestParticipants(c *gin.Context) {
	sessionIDOrName := c.Param("sessionId")
	if sessionIDOrName == "" {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
		return
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewGroupErrorResponse(
			http.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
		return
	}

	var req dto.GetGroupRequestsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request body",
			err.Error(),
		))
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"VALIDATION_ERROR",
			"Request validation failed",
			err.Error(),
		))
		return
	}

	ctx := c.Request.Context()
	participants, err := h.wmeowService.GetGroupRequestParticipants(ctx, sessionID, req.GroupJID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewGroupErrorResponse(
			http.StatusInternalServerError,
			"GET_GROUP_REQUEST_PARTICIPANTS_FAILED",
			"Failed to get group request participants",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"data": gin.H{
			"session_id":   sessionID,
			"action":       "get_request_participants",
			"status":       "success",
			"timestamp":    time.Now(),
			"group_jid":    req.GroupJID,
			"participants": participants,
			"total":        len(participants),
		},
	})
}

func (h *GroupHandler) UpdateGroupRequestParticipants(c *gin.Context) {
	sessionIDOrName := c.Param("sessionId")
	if sessionIDOrName == "" {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
		return
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewGroupErrorResponse(
			http.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
		return
	}

	var req dto.UpdateGroupRequestsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request body",
			err.Error(),
		))
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewGroupErrorResponse(
			http.StatusBadRequest,
			"VALIDATION_ERROR",
			"Request validation failed",
			err.Error(),
		))
		return
	}

	ctx := c.Request.Context()
	err = h.wmeowService.UpdateGroupRequestParticipants(ctx, sessionID, req.GroupJID, req.Action, req.Participants)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewGroupErrorResponse(
			http.StatusInternalServerError,
			"UPDATE_GROUP_REQUEST_PARTICIPANTS_FAILED",
			"Failed to update group request participants",
			err.Error(),
		))
		return
	}

	message := fmt.Sprintf("Successfully %sed %d participants", req.Action, len(req.Participants))
	response := dto.NewGroupOperationResponse(sessionID, "update_request_participants", message)
	c.JSON(http.StatusOK, response)
}
