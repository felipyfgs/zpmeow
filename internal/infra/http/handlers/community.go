package handlers

import (
	"net/http"
	"time"

	"zpmeow/internal/application"
	"zpmeow/internal/infra/http/dto"
	"zpmeow/internal/infra/wmeow"

	"github.com/gin-gonic/gin"
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

func (h *CommunityHandler) resolveSessionID(_ *gin.Context, sessionIDOrName string) (string, error) {
	return sessionIDOrName, nil
}

func (h *CommunityHandler) LinkGroup(c *gin.Context) {
	sessionIDOrName := c.Param("sessionId")
	if sessionIDOrName == "" {
		c.JSON(http.StatusBadRequest, dto.NewCommunityErrorResponse(
			http.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
		return
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewCommunityErrorResponse(
			http.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
		return
	}

	var req dto.LinkGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewCommunityErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request body",
			err.Error(),
		))
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewCommunityErrorResponse(
			http.StatusBadRequest,
			"VALIDATION_ERROR",
			"Request validation failed",
			err.Error(),
		))
		return
	}

	ctx := c.Request.Context()
	err = h.wmeowService.LinkGroup(ctx, sessionID, req.CommunityJID, req.GroupJID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewCommunityErrorResponse(
			http.StatusInternalServerError,
			"LINK_GROUP_FAILED",
			"Failed to link group to community",
			err.Error(),
		))
		return
	}

	response := dto.NewCommunitySuccessResponse(sessionID, "link_group", "Group linked to community successfully", nil)
	c.JSON(http.StatusOK, response)
}

func (h *CommunityHandler) UnlinkGroup(c *gin.Context) {
	sessionIDOrName := c.Param("sessionId")
	if sessionIDOrName == "" {
		c.JSON(http.StatusBadRequest, dto.NewCommunityErrorResponse(
			http.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
		return
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewCommunityErrorResponse(
			http.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
		return
	}

	var req dto.UnlinkGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewCommunityErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request body",
			err.Error(),
		))
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewCommunityErrorResponse(
			http.StatusBadRequest,
			"VALIDATION_ERROR",
			"Request validation failed",
			err.Error(),
		))
		return
	}

	ctx := c.Request.Context()
	err = h.wmeowService.UnlinkGroup(ctx, sessionID, req.CommunityJID, req.GroupJID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewCommunityErrorResponse(
			http.StatusInternalServerError,
			"UNLINK_GROUP_FAILED",
			"Failed to unlink group from community",
			err.Error(),
		))
		return
	}

	response := dto.NewCommunitySuccessResponse(sessionID, "unlink_group", "Group unlinked from community successfully", nil)
	c.JSON(http.StatusOK, response)
}

func (h *CommunityHandler) GetSubGroups(c *gin.Context) {
	sessionIDOrName := c.Param("sessionId")
	if sessionIDOrName == "" {
		c.JSON(http.StatusBadRequest, dto.NewCommunityErrorResponse(
			http.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
		return
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewCommunityErrorResponse(
			http.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
		return
	}

	var req dto.GetSubGroupsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewCommunityErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request body",
			err.Error(),
		))
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewCommunityErrorResponse(
			http.StatusBadRequest,
			"VALIDATION_ERROR",
			"Request validation failed",
			err.Error(),
		))
		return
	}

	ctx := c.Request.Context()
	subGroups, err := h.wmeowService.GetSubGroups(ctx, sessionID, req.CommunityJID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewCommunityErrorResponse(
			http.StatusInternalServerError,
			"GET_SUBGROUPS_FAILED",
			"Failed to get subgroups",
			err.Error(),
		))
		return
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
	c.JSON(http.StatusOK, response)
}

func (h *CommunityHandler) GetLinkedGroupsParticipants(c *gin.Context) {
	sessionIDOrName := c.Param("sessionId")
	if sessionIDOrName == "" {
		c.JSON(http.StatusBadRequest, dto.NewCommunityErrorResponse(
			http.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
		return
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewCommunityErrorResponse(
			http.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
		return
	}

	var req dto.GetLinkedGroupsParticipantsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewCommunityErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request body",
			err.Error(),
		))
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewCommunityErrorResponse(
			http.StatusBadRequest,
			"VALIDATION_ERROR",
			"Request validation failed",
			err.Error(),
		))
		return
	}

	ctx := c.Request.Context()
	participants, err := h.wmeowService.GetLinkedGroupsParticipants(ctx, sessionID, req.CommunityJID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewCommunityErrorResponse(
			http.StatusInternalServerError,
			"GET_LINKED_GROUPS_PARTICIPANTS_FAILED",
			"Failed to get linked groups participants",
			err.Error(),
		))
		return
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
	c.JSON(http.StatusOK, response)
}
