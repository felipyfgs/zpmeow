package handlers

import (
	"encoding/base64"
	"net/http"
	"time"

	"zpmeow/internal/application"
	"zpmeow/internal/infra/http/dto"
	"zpmeow/internal/infra/wmeow"

	"github.com/gin-gonic/gin"
)

type ContactHandler struct {
	contactService *application.ContactApp
	wmeowService   wmeow.WameowService
}

func NewContactHandler(contactService *application.ContactApp, wmeowService wmeow.WameowService) *ContactHandler {
	return &ContactHandler{
		contactService: contactService,
		wmeowService:   wmeowService,
	}
}

// CheckContact godoc
// @Summary Check if contacts are on WhatsApp
// @Description Verifies if phone numbers are registered on WhatsApp
// @Tags Contacts
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.CheckContactRequest true "Contact check request"
// @Success 200 {object} dto.ContactResponse "Contact check results"
// @Failure 400 {object} dto.ContactResponse "Invalid request data"
// @Failure 401 {object} dto.ContactResponse "Unauthorized - Invalid API key" "Invalid request data"
// @Failure 404 {object} dto.ContactResponse "Session not found"
// @Failure 500 {object} dto.ContactResponse "Failed to check contacts"
// @Router /session/{sessionId}/contacts/check [post]
func (h *ContactHandler) CheckContact(c *gin.Context) {
	sessionID := c.Param("sessionId")

	var req dto.CheckContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewContactErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
		return
	}

	if len(req.Phones) == 0 {
		c.JSON(http.StatusBadRequest, dto.NewContactErrorResponse(
			http.StatusBadRequest,
			"MISSING_PHONES",
			"At least one phone number is required",
			"",
		))
		return
	}

	ctx := c.Request.Context()

	appReq := application.CheckContactRequest{
		SessionID: sessionID,
		Phones:    req.Phones,
	}

	result, err := h.contactService.CheckContact(ctx, appReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewContactErrorResponse(
			http.StatusInternalServerError,
			"CHECK_CONTACT_FAILED",
			"Failed to check contacts",
			err.Error(),
		))
		return
	}

	var checkResults []dto.ContactCheckResult
	for _, checkResult := range result.Results {
		checkResults = append(checkResults, dto.ContactCheckResult{
			Query:        checkResult.Query,
			IsInmeow:     checkResult.IsInMeow,
			JID:          checkResult.JID,
			VerifiedName: checkResult.VerifiedName,
		})
	}

	response := dto.NewCheckContactSuccessResponse(sessionID, checkResults)
	c.JSON(http.StatusOK, response)
}

// GetContactInfo godoc
// @Summary Get contact information
// @Description Retrieves detailed information about WhatsApp contacts
// @Tags Contacts
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.GetContactInfoRequest true "Contact info request"
// @Success 200 {object} dto.ContactResponse "Contact information"
// @Failure 400 {object} dto.ContactResponse "Invalid request data"
// @Failure 401 {object} dto.ContactResponse "Unauthorized - Invalid API key" "Invalid request data"
// @Failure 404 {object} dto.ContactResponse "Session not found"
// @Failure 500 {object} dto.ContactResponse "Failed to get contact info"
// @Router /session/{sessionId}/contacts/info [post]
func (h *ContactHandler) GetContactInfo(c *gin.Context) {
	sessionID := c.Param("sessionId")

	var req dto.GetContactInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewContactErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
		return
	}

	if len(req.Phones) == 0 {
		c.JSON(http.StatusBadRequest, dto.NewContactErrorResponse(
			http.StatusBadRequest,
			"MISSING_PHONES",
			"At least one phone number is required",
			"",
		))
		return
	}

	ctx := c.Request.Context()
	results, err := h.wmeowService.GetUserInfo(ctx, sessionID, req.Phones)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewContactErrorResponse(
			http.StatusInternalServerError,
			"GET_CONTACT_INFO_FAILED",
			"Failed to get contact information",
			err.Error(),
		))
		return
	}

	var contactInfos []dto.ContactInfo
	for _, result := range results {
		contactInfos = append(contactInfos, dto.ContactInfo{
			JID:          result.JID,
			Name:         result.Name,
			DisplayName:  result.Name,
			VerifiedName: "",
			Notify:       result.Notify,
			PushName:     result.PushName,
			BusinessName: result.BusinessName,
			Phone:        "",
			IsBlocked:    result.IsBlocked,
			IsMuted:      result.IsMuted,
		})
	}

	response := dto.NewContactsResponse(sessionID, contactInfos)
	c.JSON(http.StatusOK, response)
}

// GetAvatar godoc
// @Summary Get contact avatar
// @Description Retrieves the avatar/profile picture of a WhatsApp contact
// @Tags Contacts
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.GetAvatarRequest true "Avatar request"
// @Success 200 {object} dto.ContactResponse "Contact avatar"
// @Failure 400 {object} dto.ContactResponse "Invalid request data"
// @Failure 401 {object} dto.ContactResponse "Unauthorized - Invalid API key" "Invalid request data"
// @Failure 404 {object} dto.ContactResponse "Session not found"
// @Failure 500 {object} dto.ContactResponse "Failed to get avatar"
// @Router /session/{sessionId}/contacts/avatar [post]
func (h *ContactHandler) GetAvatar(c *gin.Context) {
	sessionID := c.Param("sessionId")

	var req dto.GetAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewContactErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
		return
	}

	if req.Phone == "" {
		c.JSON(http.StatusBadRequest, dto.NewContactErrorResponse(
			http.StatusBadRequest,
			"MISSING_PHONE",
			"Phone number is required",
			"",
		))
		return
	}

	ctx := c.Request.Context()
	result, err := h.wmeowService.GetAvatar(ctx, sessionID, req.Phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewContactErrorResponse(
			http.StatusInternalServerError,
			"GET_AVATAR_FAILED",
			"Failed to get contact avatar",
			err.Error(),
		))
		return
	}

	avatarInfo := &dto.AvatarInfo{
		Phone:     result.Phone,
		JID:       result.JID,
		AvatarURL: result.AvatarURL,
		PictureID: result.PictureID,
		Timestamp: time.Now(),
	}

	response := dto.NewContactAvatarResponse(avatarInfo)
	c.JSON(http.StatusOK, response)
}

// SetPresence godoc
// @Summary Set presence status
// @Description Sets the presence status (online, offline, typing, etc.) for a session
// @Tags Contacts
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.SetContactPresenceRequest true "Presence request"
// @Success 200 {object} dto.ContactResponse "Presence set successfully"
// @Failure 400 {object} dto.ContactResponse "Invalid request data"
// @Failure 401 {object} dto.ContactResponse "Unauthorized - Invalid API key" "Invalid request data"
// @Failure 404 {object} dto.ContactResponse "Session not found"
// @Failure 500 {object} dto.ContactResponse "Failed to set presence"
// @Router /session/{sessionId}/presences/set [put]
func (h *ContactHandler) SetPresence(c *gin.Context) {
	sessionID := c.Param("sessionId")

	var req dto.SetContactPresenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewContactErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
		return
	}

	if req.State == "" {
		c.JSON(http.StatusBadRequest, dto.NewContactErrorResponse(
			http.StatusBadRequest,
			"MISSING_STATE",
			"State is required",
			"Valid states: available, unavailable",
		))
		return
	}

	ctx := c.Request.Context()
	err := h.wmeowService.SetUserPresence(ctx, sessionID, req.State)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewContactErrorResponse(
			http.StatusInternalServerError,
			"SET_PRESENCE_FAILED",
			"Failed to set contact presence",
			err.Error(),
		))
		return
	}

	response := dto.NewContactActionSuccessResponse(sessionID, req.Phone, "set_presence", "Presence updated successfully")
	c.JSON(http.StatusOK, response)
}

// GetContacts godoc
// @Summary Get contacts list
// @Description Retrieves the list of WhatsApp contacts for a session
// @Tags Contacts
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Success 200 {object} dto.ContactResponse "Contacts list"
// @Failure 400 {object} dto.ContactResponse "Invalid request data"
// @Failure 401 {object} dto.ContactResponse "Unauthorized - Invalid API key" "Invalid request data"
// @Failure 404 {object} dto.ContactResponse "Session not found"
// @Failure 500 {object} dto.ContactResponse "Failed to get contacts"
// @Router /session/{sessionId}/contacts/list [get]
// @Router /session/{sessionId}/contacts/sync [post]
func (h *ContactHandler) GetContacts(c *gin.Context) {
	sessionID := c.Param("sessionId")

	ctx := c.Request.Context()
	limit := 100
	offset := 0

	req := application.GetContactsRequest{
		SessionID: sessionID,
		Limit:     limit,
		Offset:    offset,
	}

	result, err := h.contactService.GetContacts(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewContactErrorResponse(
			http.StatusInternalServerError,
			"GET_CONTACTS_FAILED",
			"Failed to get contacts",
			err.Error(),
		))
		return
	}

	var contacts []dto.ContactInfo
	for _, contact := range result.Contacts {
		contacts = append(contacts, dto.ContactInfo{
			JID:          contact.JID,
			Name:         contact.Name,
			Notify:       contact.Notify,
			PushName:     contact.PushName,
			BusinessName: contact.BusinessName,
			IsBlocked:    contact.IsBlocked,
			IsMuted:      contact.IsMuted,
		})
	}

	response := dto.NewContactsResponse(sessionID, contacts)
	c.JSON(http.StatusOK, response)
}

func (h *ContactHandler) GetBlockedContacts(c *gin.Context) {
	sessionID := c.Param("sessionId")

	ctx := c.Request.Context()
	blocklist, err := h.wmeowService.GetBlocklist(ctx, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewContactErrorResponse(
			http.StatusInternalServerError,
			"GET_BLOCKLIST_FAILED",
			"Failed to get blocked contacts",
			err.Error(),
		))
		return
	}

	response := dto.NewContactsResponse(sessionID, []dto.ContactInfo{})
	for _, jid := range blocklist {
		response.Data.Contacts = append(response.Data.Contacts, dto.ContactInfo{
			JID:  jid,
			Name: jid,
		})
	}

	c.JSON(http.StatusOK, response)
}

func (h *ContactHandler) UpdateProfile(c *gin.Context) {
	sessionID := c.Param("sessionId")

	var req struct {
		Name  string `json:"name,omitempty"`
		About string `json:"about,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewContactErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
		return
	}

	if req.Name == "" && req.About == "" {
		c.JSON(http.StatusBadRequest, dto.NewContactErrorResponse(
			http.StatusBadRequest,
			"MISSING_DATA",
			"At least name or about must be provided",
			"",
		))
		return
	}

	ctx := c.Request.Context()
	err := h.wmeowService.UpdateProfile(ctx, sessionID, req.Name, req.About)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewContactErrorResponse(
			http.StatusInternalServerError,
			"UPDATE_PROFILE_FAILED",
			"Failed to update profile",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Profile updated successfully",
		"data": gin.H{
			"session_id": sessionID,
			"name":       req.Name,
			"about":      req.About,
		},
	})
}

func (h *ContactHandler) SetProfilePicture(c *gin.Context) {
	sessionID := c.Param("sessionId")

	var req struct {
		Image string `json:"image" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewContactErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
		return
	}

	imageData, err := base64.StdEncoding.DecodeString(req.Image)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewContactErrorResponse(
			http.StatusBadRequest,
			"INVALID_IMAGE",
			"Invalid base64 image data",
			err.Error(),
		))
		return
	}

	ctx := c.Request.Context()
	err = h.wmeowService.SetProfilePicture(ctx, sessionID, imageData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewContactErrorResponse(
			http.StatusInternalServerError,
			"SET_PROFILE_PICTURE_FAILED",
			"Failed to set profile picture",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Profile picture set successfully",
		"data": gin.H{
			"session_id": sessionID,
		},
	})
}

func (h *ContactHandler) RemoveProfilePicture(c *gin.Context) {
	sessionID := c.Param("sessionId")

	ctx := c.Request.Context()
	err := h.wmeowService.RemoveProfilePicture(ctx, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewContactErrorResponse(
			http.StatusInternalServerError,
			"REMOVE_PROFILE_PICTURE_FAILED",
			"Failed to remove profile picture",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Profile picture removed successfully",
		"data": gin.H{
			"session_id": sessionID,
		},
	})
}

func (h *ContactHandler) CheckUser(c *gin.Context)     { h.CheckContact(c) }
func (h *ContactHandler) GetUserInfo(c *gin.Context)   { h.GetContactInfo(c) }
func (h *ContactHandler) CheckUsers(c *gin.Context)    { h.CheckContact(c) }
func (h *ContactHandler) GetUserAvatar(c *gin.Context) { h.GetAvatar(c) }
func (h *ContactHandler) GetUserStatus(c *gin.Context) {
	sessionID := c.Param("sessionId")
	phone := c.Query("phone")

	if phone == "" {
		c.JSON(http.StatusBadRequest, dto.NewContactErrorResponse(
			http.StatusBadRequest,
			"MISSING_PHONE",
			"Phone number is required",
			"",
		))
		return
	}

	ctx := c.Request.Context()
	status, err := h.wmeowService.GetUserStatus(ctx, sessionID, phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewContactErrorResponse(
			http.StatusInternalServerError,
			"GET_USER_STATUS_FAILED",
			"Failed to get user status",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"session_id": sessionID,
			"phone":      phone,
			"status":     status,
		},
	})
}
func (h *ContactHandler) SetStatus(c *gin.Context) {
	sessionID := c.Param("sessionId")

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewContactErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
		return
	}

	ctx := c.Request.Context()
	err := h.wmeowService.SetStatus(ctx, sessionID, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewContactErrorResponse(
			http.StatusInternalServerError,
			"SET_STATUS_FAILED",
			"Failed to set status",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Status set successfully",
		"data": gin.H{
			"session_id": sessionID,
			"status":     req.Status,
		},
	})
}
func (h *ContactHandler) GetPrivacySettings(c *gin.Context) {
	sessionID := c.Param("sessionId")

	ctx := c.Request.Context()
	settings, err := h.wmeowService.GetPrivacySettings(ctx, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get privacy settings",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    settings,
	})
}
func (h *ContactHandler) UpdatePrivacySettings(c *gin.Context) {
	sessionID := c.Param("sessionId")

	var req struct {
		Setting string `json:"setting" binding:"required"`
		Value   string `json:"value" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	err := h.wmeowService.SetPrivacySetting(ctx, sessionID, req.Setting, req.Value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update privacy setting",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Privacy setting updated successfully",
	})
}
func (h *ContactHandler) SetUserPresence(c *gin.Context) { h.SetPresence(c) }
func (h *ContactHandler) BlockUser(c *gin.Context) {
	sessionID := c.Param("sessionId")

	var req struct {
		Phone string `json:"phone" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	err := h.wmeowService.UpdateBlocklist(ctx, sessionID, "block", []string{req.Phone})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to block contact",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Contact blocked successfully",
	})
}
func (h *ContactHandler) UnblockUser(c *gin.Context) {
	sessionID := c.Param("sessionId")

	var req struct {
		Phone string `json:"phone" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	err := h.wmeowService.UpdateBlocklist(ctx, sessionID, "unblock", []string{req.Phone})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to unblock contact",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Contact unblocked successfully",
	})
}
