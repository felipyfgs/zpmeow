package handlers

import (
	"time"

	"zpmeow/internal/application"
	"zpmeow/internal/infra/http/dto"
	"zpmeow/internal/infra/wmeow"

	"github.com/gofiber/fiber/v2"
)

type CommunityHandler struct {
	sessionService *application.SessionApp
	wmeowService   wmeow.WameowService
}

func NewCommunityHandler(sessionService *application.SessionApp, wmeowService wmeow.WameowService) *CommunityHandler {
	return &CommunityHandler{
		sessionService: sessionService,
		wmeowService:   wmeowService,
	}
}

func (h *CommunityHandler) resolveSessionID(_ *fiber.Ctx, sessionIDOrName string) (string, error) {
	return sessionIDOrName, nil
}

// LinkGroup godoc
// @Summary Link group to community
// @Description Links a WhatsApp group to a community
// @Tags Communities
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.LinkGroupRequest true "Link group request"
// @Success 200 {object} dto.CommunityResponse "Group linked successfully"
// @Failure 400 {object} dto.CommunityResponse "Invalid request data"
// @Failure 401 {object} dto.CommunityResponse "Unauthorized - Invalid API key" "Invalid request data"
// @Failure 404 {object} dto.CommunityResponse "Session not found"
// @Failure 500 {object} dto.CommunityResponse "Failed to link group"
// @Router /session/{sessionId}/community/link [post]
func (h *CommunityHandler) LinkGroup(c *fiber.Ctx) error {
	sessionIDOrName := c.Params("sessionId")
	if sessionIDOrName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewCommunityErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.NewCommunityErrorResponse(
			fiber.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
	}

	var req dto.LinkGroupRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewCommunityErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request body",
			err.Error(),
		))
	}

	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewCommunityErrorResponse(
			fiber.StatusBadRequest,
			"VALIDATION_ERROR",
			"Request validation failed",
			err.Error(),
		))
	}

	ctx := c.Context()
	err = h.wmeowService.LinkGroup(ctx, sessionID, req.CommunityJID, req.GroupJID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewCommunityErrorResponse(
			fiber.StatusInternalServerError,
			"LINK_GROUP_FAILED",
			"Failed to link group to community",
			err.Error(),
		))
	}

	response := dto.NewCommunitySuccessResponse(sessionID, "link_group", "Group linked to community successfully", nil)
	return c.Status(fiber.StatusOK).JSON(response)
}

// UnlinkGroup godoc
// @Summary Unlink group from community
// @Description Unlinks a WhatsApp group from a community
// @Tags Communities
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.UnlinkGroupRequest true "Unlink group request"
// @Success 200 {object} dto.CommunityResponse "Group unlinked successfully"
// @Failure 400 {object} dto.CommunityResponse "Invalid request data"
// @Failure 401 {object} dto.CommunityResponse "Unauthorized - Invalid API key" "Invalid request data"
// @Failure 404 {object} dto.CommunityResponse "Session not found"
// @Failure 500 {object} dto.CommunityResponse "Failed to unlink group"
// @Router /session/{sessionId}/community/unlink [post]
func (h *CommunityHandler) UnlinkGroup(c *fiber.Ctx) error {
	sessionIDOrName := c.Params("sessionId")
	if sessionIDOrName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewCommunityErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.NewCommunityErrorResponse(
			fiber.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
	}

	var req dto.UnlinkGroupRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewCommunityErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request body",
			err.Error(),
		))
	}

	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewCommunityErrorResponse(
			fiber.StatusBadRequest,
			"VALIDATION_ERROR",
			"Request validation failed",
			err.Error(),
		))
	}

	ctx := c.Context()
	err = h.wmeowService.UnlinkGroup(ctx, sessionID, req.CommunityJID, req.GroupJID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewCommunityErrorResponse(
			fiber.StatusInternalServerError,
			"UNLINK_GROUP_FAILED",
			"Failed to unlink group from community",
			err.Error(),
		))
	}

	response := dto.NewCommunitySuccessResponse(sessionID, "unlink_group", "Group unlinked from community successfully", nil)
	return c.Status(fiber.StatusOK).JSON(response)
}

// GetSubGroups godoc
// @Summary Get community sub-groups
// @Description Retrieves the list of sub-groups in a WhatsApp community
// @Tags Communities
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.GetSubGroupsRequest true "Get sub-groups request"
// @Success 200 {object} dto.CommunityResponse "Sub-groups retrieved successfully"
// @Failure 400 {object} dto.CommunityResponse "Invalid request data"
// @Failure 401 {object} dto.CommunityResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.CommunityResponse "Session not found"
// @Failure 500 {object} dto.CommunityResponse "Failed to get sub-groups"
// @Router /session/{sessionId}/community/subgroups [post]
func (h *CommunityHandler) GetSubGroups(c *fiber.Ctx) error {
	sessionIDOrName := c.Params("sessionId")
	if sessionIDOrName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewCommunityErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.NewCommunityErrorResponse(
			fiber.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
	}

	var req dto.GetSubGroupsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewCommunityErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request body",
			err.Error(),
		))
	}

	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewCommunityErrorResponse(
			fiber.StatusBadRequest,
			"VALIDATION_ERROR",
			"Request validation failed",
			err.Error(),
		))
	}

	ctx := c.Context()
	subGroups, err := h.wmeowService.GetSubGroups(ctx, sessionID, req.CommunityJID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewCommunityErrorResponse(
			fiber.StatusInternalServerError,
			"GET_SUBGROUPS_FAILED",
			"Failed to get subgroups",
			err.Error(),
		))
	}

	var communityInfos []dto.CommunityInfo
	for _, groupJID := range subGroups {
		communityInfos = append(communityInfos, dto.CommunityInfo{
			JID:         groupJID,
			Name:        "",
			Description: "",
			CreatedAt:   time.Now(),
			MemberCount: 0,
		})
	}

	response := dto.NewCommunitySubGroupsResponse(sessionID, req.CommunityJID, communityInfos)
	return c.Status(fiber.StatusOK).JSON(response)
}

// GetLinkedGroupsParticipants godoc
// @Summary Get linked groups participants
// @Description Retrieves participants from all groups linked to a WhatsApp community
// @Tags Communities
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.GetLinkedGroupsParticipantsRequest true "Get linked groups participants request"
// @Success 200 {object} dto.CommunityResponse "Linked groups participants retrieved successfully"
// @Failure 400 {object} dto.CommunityResponse "Invalid request data"
// @Failure 401 {object} dto.CommunityResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.CommunityResponse "Session not found"
// @Failure 500 {object} dto.CommunityResponse "Failed to get linked groups participants"
// @Router /session/{sessionId}/community/participants [post]
func (h *CommunityHandler) GetLinkedGroupsParticipants(c *fiber.Ctx) error {
	sessionIDOrName := c.Params("sessionId")
	if sessionIDOrName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewCommunityErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.NewCommunityErrorResponse(
			fiber.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
	}

	var req dto.GetLinkedGroupsParticipantsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewCommunityErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request body",
			err.Error(),
		))
	}

	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewCommunityErrorResponse(
			fiber.StatusBadRequest,
			"VALIDATION_ERROR",
			"Request validation failed",
			err.Error(),
		))
	}

	ctx := c.Context()
	participants, err := h.wmeowService.GetLinkedGroupsParticipants(ctx, sessionID, req.CommunityJID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewCommunityErrorResponse(
			fiber.StatusInternalServerError,
			"GET_LINKED_GROUPS_PARTICIPANTS_FAILED",
			"Failed to get linked groups participants",
			err.Error(),
		))
	}

	var groupParticipants []dto.GroupParticipant
	for _, participantJID := range participants {
		groupParticipants = append(groupParticipants, dto.GroupParticipant{
			JID:          participantJID,
			Phone:        participantJID,
			Name:         "",
			IsAdmin:      false,
			IsSuperAdmin: false,
		})
	}

	response := dto.NewCommunityParticipantsResponse(sessionID, req.CommunityJID, groupParticipants)
	return c.Status(fiber.StatusOK).JSON(response)
}
