package handlers

import (
	"encoding/base64"
	"time"

	"zpmeow/internal/application"
	"zpmeow/internal/infra/http/dto"
	"zpmeow/internal/infra/wmeow"

	"github.com/gofiber/fiber/v2"
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
func (h *ContactHandler) CheckContact(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	var req dto.CheckContactRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewContactErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	if len(req.Phones) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewContactErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_PHONES",
			"At least one phone number is required",
			"",
		))
	}

	ctx := c.Context()

	appReq := application.CheckContactRequest{
		SessionID: sessionID,
		Phones:    req.Phones,
	}

	result, err := h.contactService.CheckContact(ctx, appReq)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewContactErrorResponse(
			fiber.StatusInternalServerError,
			"CHECK_CONTACT_FAILED",
			"Failed to check contacts",
			err.Error(),
		))
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
	return c.Status(fiber.StatusOK).JSON(response)
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
func (h *ContactHandler) GetContactInfo(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	var req dto.GetContactInfoRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewContactErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	if len(req.Phones) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewContactErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_PHONES",
			"At least one phone number is required",
			"",
		))
	}

	ctx := c.Context()
	results, err := h.wmeowService.GetUserInfo(ctx, sessionID, req.Phones)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewContactErrorResponse(
			fiber.StatusInternalServerError,
			"GET_CONTACT_INFO_FAILED",
			"Failed to get contact information",
			err.Error(),
		))
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
	return c.Status(fiber.StatusOK).JSON(response)
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
func (h *ContactHandler) GetAvatar(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	var req dto.GetAvatarRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewContactErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	if req.Phone == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewContactErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_PHONE",
			"Phone number is required",
			"",
		))
	}

	ctx := c.Context()
	result, err := h.wmeowService.GetAvatar(ctx, sessionID, req.Phone)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewContactErrorResponse(
			fiber.StatusInternalServerError,
			"GET_AVATAR_FAILED",
			"Failed to get contact avatar",
			err.Error(),
		))
	}

	avatarInfo := &dto.AvatarInfo{
		Phone:     result.Phone,
		JID:       result.JID,
		AvatarURL: result.AvatarURL,
		PictureID: result.PictureID,
		Timestamp: time.Now(),
	}

	response := dto.NewContactAvatarResponse(avatarInfo)
	return c.Status(fiber.StatusOK).JSON(response)
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
func (h *ContactHandler) SetPresence(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	var req dto.SetContactPresenceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewContactErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	if req.State == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewContactErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_STATE",
			"State is required",
			"Valid states: available, unavailable",
		))
	}

	ctx := c.Context()
	err := h.wmeowService.SetUserPresence(ctx, sessionID, req.State)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewContactErrorResponse(
			fiber.StatusInternalServerError,
			"SET_PRESENCE_FAILED",
			"Failed to set contact presence",
			err.Error(),
		))
	}

	response := dto.NewContactActionSuccessResponse(sessionID, req.Phone, "set_presence", "Presence updated successfully")
	return c.Status(fiber.StatusOK).JSON(response)
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
func (h *ContactHandler) GetContacts(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	ctx := c.Context()
	limit := 100
	offset := 0

	req := application.GetContactsRequest{
		SessionID: sessionID,
		Limit:     limit,
		Offset:    offset,
	}

	result, err := h.contactService.GetContacts(ctx, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewContactErrorResponse(
			fiber.StatusInternalServerError,
			"GET_CONTACTS_FAILED",
			"Failed to get contacts",
			err.Error(),
		))
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
	return c.Status(fiber.StatusOK).JSON(response)
}

func (h *ContactHandler) GetBlockedContacts(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	ctx := c.Context()
	blocklist, err := h.wmeowService.GetBlocklist(ctx, sessionID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewContactErrorResponse(
			fiber.StatusInternalServerError,
			"GET_BLOCKLIST_FAILED",
			"Failed to get blocked contacts",
			err.Error(),
		))
	}

	response := dto.NewContactsResponse(sessionID, []dto.ContactInfo{})
	for _, jid := range blocklist {
		response.Data.Contacts = append(response.Data.Contacts, dto.ContactInfo{
			JID:  jid,
			Name: jid,
		})
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func (h *ContactHandler) UpdateProfile(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	var req struct {
		Name  string `json:"name,omitempty"`
		About string `json:"about,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewContactErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	if req.Name == "" && req.About == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewContactErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_DATA",
			"At least name or about must be provided",
			"",
		))
	}

	ctx := c.Context()
	err := h.wmeowService.UpdateProfile(ctx, sessionID, req.Name, req.About)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewContactErrorResponse(
			fiber.StatusInternalServerError,
			"UPDATE_PROFILE_FAILED",
			"Failed to update profile",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Profile updated successfully",
		"data": fiber.Map{
			"session_id": sessionID,
			"name":       req.Name,
			"about":      req.About,
		},
	})
}

func (h *ContactHandler) SetProfilePicture(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	var req struct {
		Image string `json:"image" binding:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewContactErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	imageData, err := base64.StdEncoding.DecodeString(req.Image)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewContactErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_IMAGE",
			"Invalid base64 image data",
			err.Error(),
		))
	}

	ctx := c.Context()
	err = h.wmeowService.SetProfilePicture(ctx, sessionID, imageData)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewContactErrorResponse(
			fiber.StatusInternalServerError,
			"SET_PROFILE_PICTURE_FAILED",
			"Failed to set profile picture",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Profile picture set successfully",
		"data": fiber.Map{
			"session_id": sessionID,
		},
	})
}

func (h *ContactHandler) RemoveProfilePicture(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	ctx := c.Context()
	err := h.wmeowService.RemoveProfilePicture(ctx, sessionID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewContactErrorResponse(
			fiber.StatusInternalServerError,
			"REMOVE_PROFILE_PICTURE_FAILED",
			"Failed to remove profile picture",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Profile picture removed successfully",
		"data": fiber.Map{
			"session_id": sessionID,
		},
	})
}

func (h *ContactHandler) CheckUser(c *fiber.Ctx) error     { return h.CheckContact(c) }
func (h *ContactHandler) GetUserInfo(c *fiber.Ctx) error   { return h.GetContactInfo(c) }
func (h *ContactHandler) CheckUsers(c *fiber.Ctx) error    { return h.CheckContact(c) }
func (h *ContactHandler) GetUserAvatar(c *fiber.Ctx) error { return h.GetAvatar(c) }
func (h *ContactHandler) GetUserStatus(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")
	phone := c.Query("phone")

	if phone == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewContactErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_PHONE",
			"Phone number is required",
			"",
		))
	}

	ctx := c.Context()
	status, err := h.wmeowService.GetUserStatus(ctx, sessionID, phone)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewContactErrorResponse(
			fiber.StatusInternalServerError,
			"GET_USER_STATUS_FAILED",
			"Failed to get user status",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"session_id": sessionID,
			"phone":      phone,
			"status":     status,
		},
	})
}
func (h *ContactHandler) SetStatus(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewContactErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	ctx := c.Context()
	err := h.wmeowService.SetStatus(ctx, sessionID, req.Status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewContactErrorResponse(
			fiber.StatusInternalServerError,
			"SET_STATUS_FAILED",
			"Failed to set status",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Status set successfully",
		"data": fiber.Map{
			"session_id": sessionID,
			"status":     req.Status,
		},
	})
}
func (h *ContactHandler) GetPrivacySettings(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	ctx := c.Context()
	settings, err := h.wmeowService.GetPrivacySettings(ctx, sessionID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get privacy settings",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    settings,
	})
}
func (h *ContactHandler) UpdatePrivacySettings(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	var req struct {
		Setting string `json:"setting" binding:"required"`
		Value   string `json:"value" binding:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request format",
			"details": err.Error(),
		})
	}

	ctx := c.Context()
	err := h.wmeowService.SetPrivacySetting(ctx, sessionID, req.Setting, req.Value)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to update privacy setting",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Privacy setting updated successfully",
	})
}
func (h *ContactHandler) SetUserPresence(c *fiber.Ctx) error { return h.SetPresence(c) }
func (h *ContactHandler) BlockUser(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	var req struct {
		Phone string `json:"phone" binding:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request format",
			"details": err.Error(),
		})
	}

	ctx := c.Context()
	err := h.wmeowService.UpdateBlocklist(ctx, sessionID, "block", []string{req.Phone})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to block contact",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Contact blocked successfully",
	})
}
func (h *ContactHandler) UnblockUser(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	var req struct {
		Phone string `json:"phone" binding:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request format",
			"details": err.Error(),
		})
	}

	ctx := c.Context()
	err := h.wmeowService.UpdateBlocklist(ctx, sessionID, "unblock", []string{req.Phone})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to unblock contact",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Contact unblocked successfully",
	})
}
