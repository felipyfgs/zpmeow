package handlers

import (
	"encoding/base64"
	"fmt"
	"time"

	"zpmeow/internal/application"
	"zpmeow/internal/application/ports"
	"zpmeow/internal/infra/http/dto"
	"zpmeow/internal/infra/wmeow"

	"github.com/gofiber/fiber/v2"
)

type GroupHandler struct {
	groupService    *application.GroupApp
	wmeowService    wmeow.WameowService
	operationHelper *GroupOperationHelper
}

func NewGroupHandler(groupService *application.GroupApp, wmeowService wmeow.WameowService) *GroupHandler {
	return &GroupHandler{
		groupService:    groupService,
		wmeowService:    wmeowService,
		operationHelper: NewGroupOperationHelper(),
	}
}

func (h *GroupHandler) resolveSessionID(_ *fiber.Ctx, sessionIDOrName string) (string, error) {
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
func (h *GroupHandler) CreateGroup(c *fiber.Ctx) error {
	var req dto.CreateGroupRequest
	sessionID, err := h.operationHelper.ValidateSessionAndParseRequest(c, &req, h.resolveSessionID)
	if err != nil {
		return err // Error already handled by helper
	}

	ctx := c.Context()
	groupInfo, err := h.wmeowService.CreateGroup(ctx, sessionID, req.Name, req.Participants)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewGroupErrorResponse(
			fiber.StatusInternalServerError,
			"CREATE_GROUP_FAILED",
			"Failed to create group",
			err.Error(),
		))
	}

	dtoGroupInfo := convertWmeowGroupInfoToDTO(groupInfo)

	response := dto.NewGroupSuccessResponse(sessionID, "create", "success", dtoGroupInfo)
	response.Code = fiber.StatusCreated
	return c.Status(fiber.StatusCreated).JSON(response)
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
func (h *GroupHandler) GetGroupInfo(c *fiber.Ctx) error {
	var req dto.GetGroupInfoRequest
	sessionID, err := h.operationHelper.ValidateSessionAndParseRequest(c, &req, h.resolveSessionID)
	if err != nil {
		return err // Error already handled by helper
	}

	ctx := c.Context()
	groupInfo, err := h.wmeowService.GetGroupInfo(ctx, sessionID, req.GroupJID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewGroupErrorResponse(
			fiber.StatusInternalServerError,
			"GET_GROUP_INFO_FAILED",
			"Failed to get group information",
			err.Error(),
		))
	}

	dtoGroupInfo := convertWmeowGroupInfoToDTO(groupInfo)

	response := dto.NewGroupSuccessResponse(sessionID, "info", "success", dtoGroupInfo)
	return c.Status(fiber.StatusOK).JSON(response)
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
func (h *GroupHandler) ListGroups(c *fiber.Ctx) error {
	sessionIDOrName := c.Params("sessionId")
	if sessionIDOrName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.NewGroupErrorResponse(
			fiber.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
	}

	ctx := c.Context()

	req := application.ListGroupsRequest{
		SessionID: sessionID,
	}

	result, err := h.groupService.ListGroups(ctx, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewGroupErrorResponse(
			fiber.StatusInternalServerError,
			"LIST_GROUPS_FAILED",
			"Failed to list groups",
			err.Error(),
		))
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
	return c.Status(fiber.StatusOK).JSON(response)
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
func (h *GroupHandler) JoinGroup(c *fiber.Ctx) error {
	sessionIDOrName := c.Params("sessionId")
	if sessionIDOrName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.NewGroupErrorResponse(
			fiber.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
	}

	var req dto.JoinGroupRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	ctx := c.Context()
	_, err = h.wmeowService.JoinGroup(ctx, sessionID, req.GroupJID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewGroupErrorResponse(
			fiber.StatusInternalServerError,
			"JOIN_GROUP_FAILED",
			"Failed to join group",
			err.Error(),
		))
	}

	groupInfo, err := h.wmeowService.GetGroupInfo(ctx, sessionID, req.GroupJID)
	if err != nil {
		response := dto.NewGroupSuccessResponse(sessionID, "join", "success", nil)
		response.Data.Message = "Group joined successfully"
		return c.Status(fiber.StatusOK).JSON(response)
	}

	dtoGroupInfo := convertWmeowGroupInfoToDTO(groupInfo)
	response := dto.NewGroupSuccessResponse(sessionID, "join", "success", dtoGroupInfo)
	return c.Status(fiber.StatusOK).JSON(response)
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
func (h *GroupHandler) JoinGroupWithInvite(c *fiber.Ctx) error {
	sessionIDOrName := c.Params("sessionId")
	if sessionIDOrName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.NewGroupErrorResponse(
			fiber.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
	}

	var req dto.JoinGroupWithInviteRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	ctx := c.Context()
	groupInfo, err := h.wmeowService.JoinGroupWithInvite(ctx, sessionID, "", "", req.InviteCode, 0)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewGroupErrorResponse(
			fiber.StatusInternalServerError,
			"JOIN_GROUP_WITH_INVITE_FAILED",
			"Failed to join group with invite",
			err.Error(),
		))
	}

	dtoGroupInfo := convertWmeowGroupInfoToDTO(groupInfo)

	response := dto.NewGroupSuccessResponse(sessionID, "join_with_invite", "success", dtoGroupInfo)
	return c.Status(fiber.StatusOK).JSON(response)
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
func (h *GroupHandler) LeaveGroup(c *fiber.Ctx) error {
	sessionIDOrName := c.Params("sessionId")
	if sessionIDOrName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.NewGroupErrorResponse(
			fiber.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
	}

	var req dto.LeaveGroupRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	ctx := c.Context()
	err = h.wmeowService.LeaveGroup(ctx, sessionID, req.GroupJID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewGroupErrorResponse(
			fiber.StatusInternalServerError,
			"LEAVE_GROUP_FAILED",
			"Failed to leave group",
			err.Error(),
		))
	}

	response := dto.NewGroupOperationResponse(sessionID, "leave", "Successfully left the group")
	return c.Status(fiber.StatusOK).JSON(response)
}

// GetInviteLink godoc
// @Summary Get group invite link
// @Description Retrieves the invite link for a WhatsApp group
// @Tags Groups
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.GetInviteLinkRequest true "Get invite link request"
// @Success 200 {object} dto.GroupResponse "Invite link retrieved successfully"
// @Failure 400 {object} dto.GroupResponse "Invalid request data"
// @Failure 401 {object} dto.GroupResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.GroupResponse "Session not found"
// @Failure 500 {object} dto.GroupResponse "Failed to get invite link"
// @Router /session/{sessionId}/group/invitelink [post]
func (h *GroupHandler) GetInviteLink(c *fiber.Ctx) error {
	sessionIDOrName := c.Params("sessionId")
	if sessionIDOrName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.NewGroupErrorResponse(
			fiber.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
	}

	var req dto.GetInviteLinkRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	ctx := c.Context()
	inviteLink, err := h.wmeowService.GetInviteLink(ctx, sessionID, req.GroupJID, req.Reset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewGroupErrorResponse(
			fiber.StatusInternalServerError,
			"GET_INVITE_LINK_FAILED",
			"Failed to get invite link",
			err.Error(),
		))
	}

	response := dto.NewInviteLinkResponse(req.GroupJID, inviteLink, "")
	return c.Status(fiber.StatusOK).JSON(response)
}

func convertWmeowGroupInfoToDTO(groupInfo *ports.GroupInfo) *dto.GroupInfo {
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

// GetInviteInfo godoc
// @Summary Get group invite information
// @Description Retrieves information about a WhatsApp group from an invite link
// @Tags Groups
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.GetInviteInfoRequest true "Get invite info request"
// @Success 200 {object} dto.GroupResponse "Invite info retrieved successfully"
// @Failure 400 {object} dto.GroupResponse "Invalid request data"
// @Failure 401 {object} dto.GroupResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.GroupResponse "Session not found"
// @Failure 500 {object} dto.GroupResponse "Failed to get invite info"
// @Router /session/{sessionId}/group/inviteinfo [post]
func (h *GroupHandler) GetInviteInfo(c *fiber.Ctx) error {
	sessionIDOrName := c.Params("sessionId")
	if sessionIDOrName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.NewGroupErrorResponse(
			fiber.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
	}

	var req dto.GetInviteInfoRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	ctx := c.Context()
	inviteInfo, err := h.wmeowService.GetInviteInfo(ctx, sessionID, req.InviteCode)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewGroupErrorResponse(
			fiber.StatusInternalServerError,
			"GET_INVITE_INFO_FAILED",
			"Failed to get invite info",
			err.Error(),
		))
	}

	response := dto.NewGroupOperationResponse(sessionID, "invite_info", "Invite info retrieved successfully")
	response.Data.Message = fmt.Sprintf("Group: %s, Created by: %s", inviteInfo.Name, inviteInfo.CreatedBy)
	return c.Status(fiber.StatusOK).JSON(response)
}

// GetGroupInfoFromInvite godoc
// @Summary Get detailed group info from invite
// @Description Retrieves detailed information about a WhatsApp group from an invite link
// @Tags Groups
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.GetInviteInfoRequest true "Get group info from invite request"
// @Success 200 {object} dto.GroupResponse "Group info retrieved successfully"
// @Failure 400 {object} dto.GroupResponse "Invalid request data"
// @Failure 401 {object} dto.GroupResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.GroupResponse "Session not found"
// @Failure 500 {object} dto.GroupResponse "Failed to get group info"
// @Router /session/{sessionId}/group/inviteinfo-specific [post]
func (h *GroupHandler) GetGroupInfoFromInvite(c *fiber.Ctx) error {
	sessionIDOrName := c.Params("sessionId")
	if sessionIDOrName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.NewGroupErrorResponse(
			fiber.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
	}

	var req dto.GetInviteInfoRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	ctx := c.Context()
	groupInfo, err := h.wmeowService.GetGroupInfoFromInvite(ctx, sessionID, "", "", req.InviteCode, 0)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewGroupErrorResponse(
			fiber.StatusInternalServerError,
			"GET_GROUP_INFO_FROM_INVITE_FAILED",
			"Failed to get group info from invite",
			err.Error(),
		))
	}

	dtoGroupInfo := convertWmeowGroupInfoToDTO(groupInfo)

	response := dto.NewGroupSuccessResponse(sessionID, "invite_info_specific", "success", dtoGroupInfo)
	return c.Status(fiber.StatusOK).JSON(response)
}

// UpdateParticipants godoc
// @Summary Update group participants
// @Description Adds, removes, promotes or demotes participants in a WhatsApp group
// @Tags Groups
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.UpdateParticipantsRequest true "Update participants request"
// @Success 200 {object} dto.GroupResponse "Participants updated successfully"
// @Failure 400 {object} dto.GroupResponse "Invalid request data"
// @Failure 401 {object} dto.GroupResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.GroupResponse "Session not found"
// @Failure 500 {object} dto.GroupResponse "Failed to update participants"
// @Router /session/{sessionId}/group/participants/update [post]
func (h *GroupHandler) UpdateParticipants(c *fiber.Ctx) error {
	sessionIDOrName := c.Params("sessionId")
	if sessionIDOrName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.NewGroupErrorResponse(
			fiber.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
	}

	var req dto.UpdateParticipantsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"VALIDATION_ERROR",
			"Request validation failed",
			err.Error(),
		))
	}

	ctx := c.Context()
	err = h.wmeowService.UpdateParticipants(ctx, sessionID, req.GroupJID, req.Action, req.Participants)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewGroupErrorResponse(
			fiber.StatusInternalServerError,
			"UPDATE_PARTICIPANTS_FAILED",
			"Failed to update participants",
			err.Error(),
		))
	}

	message := "Successfully " + req.Action + "ed participants"
	response := dto.NewGroupOperationResponse(sessionID, "update_participants", message)
	return c.Status(fiber.StatusOK).JSON(response)
}

// SetName godoc
// @Summary Set group name
// @Description Sets the name of a WhatsApp group
// @Tags Groups
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.SetGroupNameRequest true "Set group name request"
// @Success 200 {object} dto.GroupResponse "Group name updated successfully"
// @Failure 400 {object} dto.GroupResponse "Invalid request data"
// @Failure 401 {object} dto.GroupResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.GroupResponse "Session not found"
// @Failure 500 {object} dto.GroupResponse "Failed to set group name"
// @Router /session/{sessionId}/group/name/set [post]
func (h *GroupHandler) SetName(c *fiber.Ctx) error {
	sessionIDOrName := c.Params("sessionId")
	if sessionIDOrName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.NewGroupErrorResponse(
			fiber.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
	}

	var req dto.SetGroupNameRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	ctx := c.Context()
	err = h.wmeowService.SetGroupName(ctx, sessionID, req.GroupJID, req.Name)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewGroupErrorResponse(
			fiber.StatusInternalServerError,
			"SET_GROUP_NAME_FAILED",
			"Failed to set group name",
			err.Error(),
		))
	}

	response := dto.NewGroupOperationResponse(sessionID, "set_name", "Group name updated successfully")
	return c.Status(fiber.StatusOK).JSON(response)
}

// SetTopic godoc
// @Summary Set group topic/description
// @Description Sets the topic/description of a WhatsApp group
// @Tags Groups
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.SetGroupTopicRequest true "Set group topic request"
// @Success 200 {object} dto.GroupResponse "Group topic updated successfully"
// @Failure 400 {object} dto.GroupResponse "Invalid request data"
// @Failure 401 {object} dto.GroupResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.GroupResponse "Session not found"
// @Failure 500 {object} dto.GroupResponse "Failed to set group topic"
// @Router /session/{sessionId}/group/topic/set [post]
func (h *GroupHandler) SetTopic(c *fiber.Ctx) error {
	sessionIDOrName := c.Params("sessionId")
	if sessionIDOrName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.NewGroupErrorResponse(
			fiber.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
	}

	var req dto.SetGroupTopicRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	ctx := c.Context()
	err = h.wmeowService.SetGroupTopic(ctx, sessionID, req.GroupJID, req.Topic)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewGroupErrorResponse(
			fiber.StatusInternalServerError,
			"SET_GROUP_TOPIC_FAILED",
			"Failed to set group topic",
			err.Error(),
		))
	}

	response := dto.NewGroupOperationResponse(sessionID, "set_topic", "Group topic updated successfully")
	return c.Status(fiber.StatusOK).JSON(response)
}

// SetPhoto godoc
// @Summary Set group photo
// @Description Sets the profile photo of a WhatsApp group
// @Tags Groups
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.SetGroupPhotoRequest true "Set group photo request"
// @Success 200 {object} dto.GroupResponse "Group photo updated successfully"
// @Failure 400 {object} dto.GroupResponse "Invalid request data"
// @Failure 401 {object} dto.GroupResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.GroupResponse "Session not found"
// @Failure 500 {object} dto.GroupResponse "Failed to set group photo"
// @Router /session/{sessionId}/group/photo/set [post]
func (h *GroupHandler) SetPhoto(c *fiber.Ctx) error {
	sessionIDOrName := c.Params("sessionId")
	if sessionIDOrName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.NewGroupErrorResponse(
			fiber.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
	}

	var req dto.SetGroupPhotoRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	photoData, err := base64.StdEncoding.DecodeString(req.Photo)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_PHOTO_DATA",
			"Invalid base64 photo data",
			err.Error(),
		))
	}

	ctx := c.Context()
	err = h.wmeowService.SetGroupPhoto(ctx, sessionID, req.GroupJID, photoData)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewGroupErrorResponse(
			fiber.StatusInternalServerError,
			"SET_GROUP_PHOTO_FAILED",
			"Failed to set group photo",
			err.Error(),
		))
	}

	response := dto.NewGroupOperationResponse(sessionID, "set_photo", "Group photo updated successfully")
	return c.Status(fiber.StatusOK).JSON(response)
}

// RemovePhoto godoc
// @Summary Remove group photo
// @Description Removes the profile photo of a WhatsApp group
// @Tags Groups
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.RemoveGroupPhotoRequest true "Remove group photo request"
// @Success 200 {object} dto.GroupResponse "Group photo removed successfully"
// @Failure 400 {object} dto.GroupResponse "Invalid request data"
// @Failure 401 {object} dto.GroupResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.GroupResponse "Session not found"
// @Failure 500 {object} dto.GroupResponse "Failed to remove group photo"
// @Router /session/{sessionId}/group/photo/remove [post]
func (h *GroupHandler) RemovePhoto(c *fiber.Ctx) error {
	sessionIDOrName := c.Params("sessionId")
	if sessionIDOrName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.NewGroupErrorResponse(
			fiber.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
	}

	var req dto.RemoveGroupPhotoRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	ctx := c.Context()
	err = h.wmeowService.RemoveGroupPhoto(ctx, sessionID, req.GroupJID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewGroupErrorResponse(
			fiber.StatusInternalServerError,
			"REMOVE_GROUP_PHOTO_FAILED",
			"Failed to remove group photo",
			err.Error(),
		))
	}

	response := dto.NewGroupOperationResponse(sessionID, "remove_photo", "Group photo removed successfully")
	return c.Status(fiber.StatusOK).JSON(response)
}

// SetAnnounce godoc
// @Summary Set group announce mode
// @Description Sets whether only admins can send messages in a WhatsApp group
// @Tags Groups
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.SetGroupAnnounceRequest true "Set group announce request"
// @Success 200 {object} dto.GroupResponse "Group announce setting updated successfully"
// @Failure 400 {object} dto.GroupResponse "Invalid request data"
// @Failure 401 {object} dto.GroupResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.GroupResponse "Session not found"
// @Failure 500 {object} dto.GroupResponse "Failed to set announce mode"
// @Router /session/{sessionId}/group/announce/set [post]
func (h *GroupHandler) SetAnnounce(c *fiber.Ctx) error {
	sessionIDOrName := c.Params("sessionId")
	if sessionIDOrName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.NewGroupErrorResponse(
			fiber.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
	}

	var req dto.SetGroupAnnounceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	ctx := c.Context()
	err = h.wmeowService.SetGroupAnnounce(ctx, sessionID, req.GroupJID, req.AnnounceOnly)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewGroupErrorResponse(
			fiber.StatusInternalServerError,
			"SET_GROUP_ANNOUNCE_FAILED",
			"Failed to set group announce setting",
			err.Error(),
		))
	}

	response := dto.NewGroupOperationResponse(sessionID, "set_announce", "Group announce setting updated successfully")
	return c.Status(fiber.StatusOK).JSON(response)
}

// SetLocked godoc
// @Summary Set group locked mode
// @Description Sets whether only admins can edit group info in a WhatsApp group
// @Tags Groups
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.SetGroupLockedRequest true "Set group locked request"
// @Success 200 {object} dto.GroupResponse "Group locked setting updated successfully"
// @Failure 400 {object} dto.GroupResponse "Invalid request data"
// @Failure 401 {object} dto.GroupResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.GroupResponse "Session not found"
// @Failure 500 {object} dto.GroupResponse "Failed to set locked mode"
// @Router /session/{sessionId}/group/locked/set [post]
func (h *GroupHandler) SetLocked(c *fiber.Ctx) error {
	sessionIDOrName := c.Params("sessionId")
	if sessionIDOrName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.NewGroupErrorResponse(
			fiber.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
	}

	var req dto.SetGroupLockedRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	ctx := c.Context()
	err = h.wmeowService.SetGroupLocked(ctx, sessionID, req.GroupJID, req.Locked)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewGroupErrorResponse(
			fiber.StatusInternalServerError,
			"SET_GROUP_LOCKED_FAILED",
			"Failed to set group locked setting",
			err.Error(),
		))
	}

	response := dto.NewGroupOperationResponse(sessionID, "set_locked", "Group locked setting updated successfully")
	return c.Status(fiber.StatusOK).JSON(response)
}

// SetEphemeral godoc
// @Summary Set group ephemeral mode
// @Description Sets the disappearing messages timer for a WhatsApp group
// @Tags Groups
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.SetGroupEphemeralRequest true "Set group ephemeral request"
// @Success 200 {object} dto.GroupResponse "Group ephemeral setting updated successfully"
// @Failure 400 {object} dto.GroupResponse "Invalid request data"
// @Failure 401 {object} dto.GroupResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.GroupResponse "Session not found"
// @Failure 500 {object} dto.GroupResponse "Failed to set ephemeral mode"
// @Router /session/{sessionId}/group/ephemeral/set [post]
func (h *GroupHandler) SetEphemeral(c *fiber.Ctx) error {
	sessionIDOrName := c.Params("sessionId")
	if sessionIDOrName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.NewGroupErrorResponse(
			fiber.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
	}

	var req dto.SetGroupEphemeralRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	ctx := c.Context()
	duration := 0
	if req.Ephemeral {
		duration = 604800
	}
	err = h.wmeowService.SetGroupEphemeral(ctx, sessionID, req.GroupJID, req.Ephemeral, duration)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewGroupErrorResponse(
			fiber.StatusInternalServerError,
			"SET_GROUP_EPHEMERAL_FAILED",
			"Failed to set group ephemeral setting",
			err.Error(),
		))
	}

	response := dto.NewGroupOperationResponse(sessionID, "set_ephemeral", "Group ephemeral setting updated successfully")
	return c.Status(fiber.StatusOK).JSON(response)
}

// SetJoinApproval godoc
// @Summary Set group join approval mode
// @Description Sets whether admin approval is required for new members to join the group
// @Tags Groups
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.GroupJoinApprovalReq true "Set group join approval request"
// @Success 200 {object} dto.GroupResponse "Group join approval mode updated successfully"
// @Failure 400 {object} dto.GroupResponse "Invalid request data"
// @Failure 401 {object} dto.GroupResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.GroupResponse "Session not found"
// @Failure 500 {object} dto.GroupResponse "Failed to set join approval mode"
// @Router /session/{sessionId}/group/join-approval/set [post]
func (h *GroupHandler) SetJoinApproval(c *fiber.Ctx) error {
	sessionIDOrName := c.Params("sessionId")
	if sessionIDOrName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.NewGroupErrorResponse(
			fiber.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
	}

	var req dto.GroupJoinApprovalReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request body",
			err.Error(),
		))
	}

	ctx := c.Context()
	err = h.wmeowService.SetGroupJoinApproval(ctx, sessionID, req.GroupJID, req.RequireApproval)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewGroupErrorResponse(
			fiber.StatusInternalServerError,
			"SET_JOIN_APPROVAL_FAILED",
			"Failed to set group join approval mode",
			err.Error(),
		))
	}

	response := dto.NewGroupOperationResponse(sessionID, "set_join_approval", "Group join approval mode updated successfully")
	return c.Status(fiber.StatusOK).JSON(response)
}

// SetMemberAddMode godoc
// @Summary Set group member add mode
// @Description Sets who can add new members to the WhatsApp group (admins only or all participants)
// @Tags Groups
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.GroupMemberModeReq true "Set group member add mode request"
// @Success 200 {object} dto.GroupResponse "Group member add mode updated successfully"
// @Failure 400 {object} dto.GroupResponse "Invalid request data"
// @Failure 401 {object} dto.GroupResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.GroupResponse "Session not found"
// @Failure 500 {object} dto.GroupResponse "Failed to set member add mode"
// @Router /session/{sessionId}/group/member-add-mode/set [post]
func (h *GroupHandler) SetMemberAddMode(c *fiber.Ctx) error {
	sessionIDOrName := c.Params("sessionId")
	if sessionIDOrName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.NewGroupErrorResponse(
			fiber.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
	}

	var req dto.GroupMemberModeReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request body",
			err.Error(),
		))
	}

	ctx := c.Context()
	err = h.wmeowService.SetGroupMemberAddMode(ctx, sessionID, req.GroupJID, req.Mode)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewGroupErrorResponse(
			fiber.StatusInternalServerError,
			"SET_MEMBER_ADD_MODE_FAILED",
			"Failed to set group member add mode",
			err.Error(),
		))
	}

	response := dto.NewGroupOperationResponse(sessionID, "set_member_add_mode", "Group member add mode updated successfully")
	return c.Status(fiber.StatusOK).JSON(response)
}

// GetGroupRequestParticipants godoc
// @Summary Get group join requests
// @Description Retrieves the list of participants requesting to join a WhatsApp group
// @Tags Groups
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.GetGroupRequestsReq true "Get group request participants request"
// @Success 200 {object} dto.GroupResponse "Group request participants retrieved successfully"
// @Failure 400 {object} dto.GroupResponse "Invalid request data"
// @Failure 401 {object} dto.GroupResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.GroupResponse "Session not found"
// @Failure 500 {object} dto.GroupResponse "Failed to get request participants"
// @Router /session/{sessionId}/group/requests/list [post]
func (h *GroupHandler) GetGroupRequestParticipants(c *fiber.Ctx) error {
	sessionIDOrName := c.Params("sessionId")
	if sessionIDOrName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.NewGroupErrorResponse(
			fiber.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
	}

	var req dto.GetGroupRequestsReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request body",
			err.Error(),
		))
	}

	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"VALIDATION_ERROR",
			"Request validation failed",
			err.Error(),
		))
	}

	ctx := c.Context()
	participants, err := h.wmeowService.GetGroupRequestParticipants(ctx, sessionID, req.GroupJID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewGroupErrorResponse(
			fiber.StatusInternalServerError,
			"GET_GROUP_REQUEST_PARTICIPANTS_FAILED",
			"Failed to get group request participants",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"code":    200,
		"data": fiber.Map{
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

// UpdateGroupRequestParticipants godoc
// @Summary Update group join requests
// @Description Approves or rejects participants requesting to join a WhatsApp group
// @Tags Groups
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.UpdateGroupRequestsReq true "Update group request participants request"
// @Success 200 {object} dto.GroupResponse "Group request participants updated successfully"
// @Failure 400 {object} dto.GroupResponse "Invalid request data"
// @Failure 401 {object} dto.GroupResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.GroupResponse "Session not found"
// @Failure 500 {object} dto.GroupResponse "Failed to update request participants"
// @Router /session/{sessionId}/group/requests/update [post]
func (h *GroupHandler) UpdateGroupRequestParticipants(c *fiber.Ctx) error {
	sessionIDOrName := c.Params("sessionId")
	if sessionIDOrName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.NewGroupErrorResponse(
			fiber.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
	}

	var req dto.UpdateGroupRequestsReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request body",
			err.Error(),
		))
	}

	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewGroupErrorResponse(
			fiber.StatusBadRequest,
			"VALIDATION_ERROR",
			"Request validation failed",
			err.Error(),
		))
	}

	ctx := c.Context()
	err = h.wmeowService.UpdateGroupRequestParticipants(ctx, sessionID, req.GroupJID, req.Action, req.Participants)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewGroupErrorResponse(
			fiber.StatusInternalServerError,
			"UPDATE_GROUP_REQUEST_PARTICIPANTS_FAILED",
			"Failed to update group request participants",
			err.Error(),
		))
	}

	message := fmt.Sprintf("Successfully %sed %d participants", req.Action, len(req.Participants))
	response := dto.NewGroupOperationResponse(sessionID, "update_request_participants", message)
	return c.Status(fiber.StatusOK).JSON(response)
}
