package handlers

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"zpmeow/internal/application"
	"zpmeow/internal/infra/wmeow"
	"zpmeow/internal/interfaces/dto"

	"github.com/gin-gonic/gin"
)

type GroupHandler struct {
	sessionService *application.SessionApp
	wmeowService   wmeow.WameowService
}

func NewGroupHandler(sessionService *application.SessionApp, wmeowService wmeow.WameowService) *GroupHandler {
	return &GroupHandler{
		sessionService: sessionService,
		wmeowService:   wmeowService,
	}
}

func (h *GroupHandler) resolveSessionID(c *gin.Context, sessionIDOrName string) (string, error) {
	if h.sessionService == nil {
		return sessionIDOrName, nil
	}

	ctx := c.Request.Context()
	session, err := h.sessionService.GetSession(ctx, sessionIDOrName)
	if err != nil {
		return "", err
	}

	return session.ID().String(), nil
}

// @Summary		Create a new group
// @Description	Create a new meow group with specified name and participants
// @Tags			Groups
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string					true	"Session ID"
// @Param			request		body		dto.CreateGroupRequest	true	"Create group request"
// @Success		201			{object}	dto.GroupResponse
// @Failure		400			{object}	dto.GroupResponse
// @Failure		404			{object}	dto.GroupResponse
// @Failure		500			{object}	dto.GroupResponse
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/group/create [post]
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

	response := dto.NewGroupSuccessResponse(sessionID, "create", dtoGroupInfo)
	response.Code = http.StatusCreated
	c.JSON(http.StatusCreated, response)
}

// @Summary		Get group information
// @Description	Get detailed information about a specific group
// @Tags			Groups
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string					true	"Session ID"
// @Param			request		body		dto.GetGroupInfoRequest	true	"Get group info request"
// @Success		200			{object}	dto.GroupResponse
// @Failure		400			{object}	dto.GroupResponse
// @Failure		404			{object}	dto.GroupResponse
// @Failure		500			{object}	dto.GroupResponse
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/group/info [post]
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

	response := dto.NewGroupSuccessResponse(sessionID, "info", dtoGroupInfo)
	c.JSON(http.StatusOK, response)
}

// @Summary		List all groups
// @Description	Get a list of all groups the user is a member of
// @Tags			Groups
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string	true	"Session ID"
// @Success		200			{object}	dto.GroupResponse
// @Failure		400			{object}	dto.GroupResponse
// @Failure		404			{object}	dto.GroupResponse
// @Failure		500			{object}	dto.GroupResponse
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/group/list [get]
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
	groups, err := h.wmeowService.ListGroups(ctx, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewGroupErrorResponse(
			http.StatusInternalServerError,
			"LIST_GROUPS_FAILED",
			"Failed to list groups",
			err.Error(),
		))
		return
	}

	dtoGroups := convertWmeowGroupInfoSliceToDTO(groups)

	response := dto.NewGroupListResponse(sessionID, dtoGroups)
	c.JSON(http.StatusOK, response)
}

// @Summary		Join group via invite link
// @Description	Join a meow group using an invite link
// @Tags			Groups
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string					true	"Session ID"
// @Param			request		body		dto.JoinGroupRequest	true	"Join group request"
// @Success		200			{object}	dto.GroupResponse
// @Failure		400			{object}	dto.GroupResponse
// @Failure		404			{object}	dto.GroupResponse
// @Failure		500			{object}	dto.GroupResponse
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/group/join [post]
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
		response := dto.NewGroupSuccessResponse(sessionID, "join", nil)
		response.Data.Message = "Group joined successfully"
		c.JSON(http.StatusOK, response)
		return
	}

	dtoGroupInfo := convertWmeowGroupInfoToDTO(groupInfo)
	response := dto.NewGroupSuccessResponse(sessionID, "join", dtoGroupInfo)
	c.JSON(http.StatusOK, response)
}

// @Summary		Join group with specific invite
// @Description	Join a meow group using specific invite details
// @Tags			Groups
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string							true	"Session ID"
// @Param			request		body		dto.JoinGroupWithInviteRequest	true	"Join group with invite request"
// @Success		200			{object}	dto.GroupResponse
// @Failure		400			{object}	dto.GroupResponse
// @Failure		404			{object}	dto.GroupResponse
// @Failure		500			{object}	dto.GroupResponse
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/group/join-with-invite [post]
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

	response := dto.NewGroupSuccessResponse(sessionID, "join_with_invite", dtoGroupInfo)
	c.JSON(http.StatusOK, response)
}

// @Summary		Leave group
// @Description	Leave a meow group
// @Tags			Groups
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string					true	"Session ID"
// @Param			request		body		dto.LeaveGroupRequest	true	"Leave group request"
// @Success		200			{object}	dto.GroupResponse
// @Failure		400			{object}	dto.GroupResponse
// @Failure		404			{object}	dto.GroupResponse
// @Failure		500			{object}	dto.GroupResponse
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/group/leave [post]
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

// @Summary		Get group invite link
// @Description	Get or reset the invite link for a group
// @Tags			Groups
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string						true	"Session ID"
// @Param			request		body		dto.GetInviteLinkRequest	true	"Get invite link request"
// @Success		200			{object}	dto.GroupResponse
// @Failure		400			{object}	dto.GroupResponse
// @Failure		404			{object}	dto.GroupResponse
// @Failure		500			{object}	dto.GroupResponse
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/group/invitelink [post]
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

	response := dto.NewInviteLinkResponse(sessionID, req.GroupJID, inviteLink)
	c.JSON(http.StatusOK, response)
}

func convertWmeowGroupInfoToDTO(groupInfo *wmeow.GroupInfo) *dto.GroupInfo {
	if groupInfo == nil {
		return nil
	}

	return &dto.GroupInfo{
		JID:          groupInfo.JID,
		Name:         groupInfo.Name,
		Topic:        groupInfo.Topic,
		Participants: groupInfo.Participants,
		Admins:       []string{}, // Campo removido da estrutura simplificada
		Owner:        groupInfo.CreatedBy,
		CreatedAt:    groupInfo.CreatedAt,
		Size:         len(groupInfo.Participants),
		Announce:     groupInfo.IsAnnounce,
		Locked:       groupInfo.IsLocked,
		Ephemeral:    groupInfo.IsEphemeral,
	}
}

func convertWmeowGroupInfoSliceToDTO(groups []wmeow.GroupInfo) []dto.GroupInfo {
	var dtoGroups []dto.GroupInfo
	for _, group := range groups {
		dtoGroups = append(dtoGroups, dto.GroupInfo{
			JID:          group.JID,
			Name:         group.Name,
			Topic:        group.Topic,
			Participants: group.Participants,
			Admins:       []string{}, // Campo removido da estrutura simplificada
			Owner:        group.CreatedBy,
			CreatedAt:    group.CreatedAt,
			Size:         len(group.Participants),
			Announce:     group.IsAnnounce,
			Locked:       group.IsLocked,
			Ephemeral:    group.IsEphemeral,
		})
	}
	return dtoGroups
}

// @Summary		Get group info from invite link
// @Description	Get group information from an invite link without joining
// @Tags			Groups
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string						true	"Session ID"
// @Param			request		body		dto.GetInviteInfoRequest	true	"Get invite info request"
// @Success		200			{object}	dto.GroupResponse
// @Failure		400			{object}	dto.GroupResponse
// @Failure		404			{object}	dto.GroupResponse
// @Failure		500			{object}	dto.GroupResponse
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/group/inviteinfo [post]
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
	response.Data.InviteLink = "" // Campo removido da estrutura simplificada
	response.Data.Message = fmt.Sprintf("Group: %s, Created by: %s", inviteInfo.Name, inviteInfo.CreatedBy)
	c.JSON(http.StatusOK, response)
}

// @Summary		Get group info from specific invite
// @Description	Get group information from specific invite details
// @Tags			Groups
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string								true	"Session ID"
// @Param			request		body		dto.GroupInviteInfoReq	true	"Get group info from invite request"
// @Success		200			{object}	dto.GroupResponse
// @Failure		400			{object}	dto.GroupResponse
// @Failure		404			{object}	dto.GroupResponse
// @Failure		500			{object}	dto.GroupResponse
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/group/inviteinfo-specific [post]
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

	var req dto.GroupInviteInfoReq
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

	response := dto.NewGroupSuccessResponse(sessionID, "invite_info_specific", dtoGroupInfo)
	c.JSON(http.StatusOK, response)
}

// @Summary		Update group participants
// @Description	Add or remove participants from a group
// @Tags			Groups
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string							true	"Session ID"
// @Param			request		body		dto.UpdateParticipantsRequest	true	"Update participants request"
// @Success		200			{object}	dto.GroupResponse
// @Failure		400			{object}	dto.GroupResponse
// @Failure		404			{object}	dto.GroupResponse
// @Failure		500			{object}	dto.GroupResponse
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/group/participants/update [post]
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

// @Summary		Set group name
// @Description	Update the name of a group
// @Tags			Groups
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string					true	"Session ID"
// @Param			request		body		dto.SetGroupNameRequest	true	"Set group name request"
// @Success		200			{object}	dto.GroupResponse
// @Failure		400			{object}	dto.GroupResponse
// @Failure		404			{object}	dto.GroupResponse
// @Failure		500			{object}	dto.GroupResponse
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/group/settings/name [post]
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

// @Summary		Set group topic
// @Description	Update the topic/description of a group
// @Tags			Groups
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string						true	"Session ID"
// @Param			request		body		dto.SetGroupTopicRequest	true	"Set group topic request"
// @Success		200			{object}	dto.GroupResponse
// @Failure		400			{object}	dto.GroupResponse
// @Failure		404			{object}	dto.GroupResponse
// @Failure		500			{object}	dto.GroupResponse
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/group/settings/topic [post]
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

// @Summary		Set group photo
// @Description	Update the photo/avatar of a group
// @Tags			Groups
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string						true	"Session ID"
// @Param			request		body		dto.SetGroupPhotoRequest	true	"Set group photo request"
// @Success		200			{object}	dto.GroupResponse
// @Failure		400			{object}	dto.GroupResponse
// @Failure		404			{object}	dto.GroupResponse
// @Failure		500			{object}	dto.GroupResponse
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/group/settings/photo/set [post]
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

// @Summary		Remove group photo
// @Description	Remove the photo/avatar of a group
// @Tags			Groups
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string						true	"Session ID"
// @Param			request		body		dto.RemoveGroupPhotoRequest	true	"Remove group photo request"
// @Success		200			{object}	dto.GroupResponse
// @Failure		400			{object}	dto.GroupResponse
// @Failure		404			{object}	dto.GroupResponse
// @Failure		500			{object}	dto.GroupResponse
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/group/settings/photo/remove [post]
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

// @Summary		Set group announce mode
// @Description	Set whether only admins can send messages to the group
// @Tags			Groups
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string						true	"Session ID"
// @Param			request		body		dto.SetGroupAnnounceRequest	true	"Set group announce request"
// @Success		200			{object}	dto.GroupResponse
// @Failure		400			{object}	dto.GroupResponse
// @Failure		404			{object}	dto.GroupResponse
// @Failure		500			{object}	dto.GroupResponse
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/group/settings/announce [post]
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
	err = h.wmeowService.SetGroupAnnounce(ctx, sessionID, req.GroupJID, req.Announce)
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

// @Summary		Set group locked mode
// @Description	Set whether only admins can edit group info
// @Tags			Groups
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string						true	"Session ID"
// @Param			request		body		dto.SetGroupLockedRequest	true	"Set group locked request"
// @Success		200			{object}	dto.GroupResponse
// @Failure		400			{object}	dto.GroupResponse
// @Failure		404			{object}	dto.GroupResponse
// @Failure		500			{object}	dto.GroupResponse
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/group/settings/locked [post]
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

// @Summary		Set group ephemeral mode
// @Description	Set disappearing messages for the group
// @Tags			Groups
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string							true	"Session ID"
// @Param			request		body		dto.SetGroupEphemeralRequest	true	"Set group ephemeral request"
// @Success		200			{object}	dto.GroupResponse
// @Failure		400			{object}	dto.GroupResponse
// @Failure		404			{object}	dto.GroupResponse
// @Failure		500			{object}	dto.GroupResponse
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/group/settings/ephemeral [post]
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
		duration = 604800 // 7 days in seconds
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

// @Summary		Set group join approval mode
// @Description	Set whether admin approval is required to join the group
// @Tags			Groups
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string							true	"Session ID"
// @Param			request		body		dto.GroupJoinApprovalReq	true	"Set group join approval request"
// @Success		200			{object}	dto.GroupResponse
// @Failure		400			{object}	dto.GroupResponse
// @Failure		404			{object}	dto.GroupResponse
// @Failure		500			{object}	dto.GroupResponse
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/group/settings/join-approval [post]
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
	err = h.wmeowService.SetGroupJoinApproval(ctx, sessionID, req.GroupJID, req.JoinApproval)
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

// @Summary		Set group member add mode
// @Description	Set who can add members to the group (all or admin only)
// @Tags			Groups
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string								true	"Session ID"
// @Param			request		body		dto.GroupMemberModeReq	true	"Set group member add mode request"
// @Success		200			{object}	dto.GroupResponse
// @Failure		400			{object}	dto.GroupResponse
// @Failure		404			{object}	dto.GroupResponse
// @Failure		500			{object}	dto.GroupResponse
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/group/settings/member-add-mode [post]
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
	err = h.wmeowService.SetGroupMemberAddMode(ctx, sessionID, req.GroupJID, req.MemberAddMode)
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

// @Summary		Get group request participants
// @Description	Get list of users requesting to join the group
// @Tags			Groups
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string									true	"Session ID"
// @Param			request		body		dto.GetGroupRequestsReq	true	"Get group request participants request"
// @Success		200			{object}	dto.GroupResponse
// @Failure		400			{object}	dto.GroupResponse
// @Failure		404			{object}	dto.GroupResponse
// @Failure		500			{object}	dto.GroupResponse
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/group/requests/list [post]
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

// @Summary		Update group request participants
// @Description	Approve or reject users requesting to join the group
// @Tags			Groups
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string										true	"Session ID"
// @Param			request		body		dto.UpdateGroupRequestsReq	true	"Update group request participants request"
// @Success		200			{object}	dto.GroupResponse
// @Failure		400			{object}	dto.GroupResponse
// @Failure		404			{object}	dto.GroupResponse
// @Failure		500			{object}	dto.GroupResponse
// @Security		ApiKeyAuth
// @Router			/session/{sessionId}/group/requests/update [post]
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
