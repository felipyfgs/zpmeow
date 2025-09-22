package handlers

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	"zpmeow/internal/application"
	"zpmeow/internal/infra/http/dto"
	"zpmeow/internal/infra/wmeow"

	"github.com/gofiber/fiber/v2"
	"go.mau.fi/whatsmeow"
)

type MessageHandler struct {
	sessionService *application.SessionApp
	wmeowService   wmeow.WameowService
}

func NewMessageHandler(sessionService *application.SessionApp, wmeowService wmeow.WameowService) *MessageHandler {
	return &MessageHandler{
		sessionService: sessionService,
		wmeowService:   wmeowService,
	}
}

func (h *MessageHandler) resolveSessionID(c *fiber.Ctx, sessionIDOrName string) (string, error) {
	if h.sessionService == nil {
		return sessionIDOrName, nil
	}

	ctx := c.Context()
	session, err := h.sessionService.GetSession(ctx, sessionIDOrName)
	if err != nil {
		return "", err
	}

	return session.ID().String(), nil
}

func (h *MessageHandler) decodeMediaData(dataURL string) ([]byte, error) {
	if strings.HasPrefix(dataURL, "http://") || strings.HasPrefix(dataURL, "https://") {
		resp, err := http.Get(dataURL)
		if err != nil {
			return nil, fmt.Errorf("failed to download from URL: %w", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				fmt.Printf("Warning: failed to close response body: %v\n", err)
			}
		}()

		if resp.StatusCode != fiber.StatusOK {
			return nil, fmt.Errorf("failed to download from URL: status %d", resp.StatusCode)
		}

		return io.ReadAll(resp.Body)
	}

	if strings.HasPrefix(dataURL, "data:") {
		commaIndex := strings.Index(dataURL, ",")
		if commaIndex == -1 {
			return nil, fmt.Errorf("invalid data URL format")
		}

		base64Data := dataURL[commaIndex+1:]

		data, err := base64.StdEncoding.DecodeString(base64Data)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 data: %w", err)
		}

		return data, nil
	}

	data, err := base64.StdEncoding.DecodeString(dataURL)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 data: %w", err)
	}

	return data, nil
}

// SendText godoc
// @Summary Send text message
// @Description Sends a text message to a WhatsApp contact or group
// @Tags Messages
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.SendTextRequest true "Text message request"
// @Success 200 {object} dto.MessageResponse "Message sent successfully"
// @Failure 400 {object} dto.MessageResponse "Invalid request data"
// @Failure 401 {object} dto.MessageResponse "Unauthorized - Invalid API key" "Invalid request data"
// @Failure 401 {object} dto.MessageResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.MessageResponse "Session not found"
// @Failure 500 {object} dto.MessageResponse "Failed to send message"
// @Router /session/{sessionId}/message/send/text [post]
func (h *MessageHandler) SendText(c *fiber.Ctx) error {
	sessionIDOrName := c.Params("sessionId")
	if sessionIDOrName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.NewMessageErrorResponse(
			fiber.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
	}

	var req dto.SendTextRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request body",
			err.Error(),
		))
	}

	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"VALIDATION_ERROR",
			"Request validation failed",
			err.Error(),
		))
	}

	ctx := c.Context()
	sendResp, err := h.wmeowService.SendTextMessage(ctx, sessionID, req.Phone, req.Body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewMessageErrorResponse(
			fiber.StatusInternalServerError,
			"SEND_TEXT_FAILED",
			"Failed to send text message",
			err.Error(),
		))
	}

	messageID := string(sendResp.ID)
	response := dto.NewTextResponse(true, fiber.StatusOK, req.Phone, messageID, req.Body, true)
	return c.Status(fiber.StatusOK).JSON(response)
}

// SendMedia godoc
// @Summary Send media message
// @Description Sends a media message (image, video, audio, document) to a WhatsApp contact or group
// @Tags Messages
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.SendMediaRequest true "Media message request"
// @Success 200 {object} dto.MessageResponse "Media sent successfully"
// @Failure 400 {object} dto.MessageResponse "Invalid request data"
// @Failure 401 {object} dto.MessageResponse "Unauthorized - Invalid API key" "Invalid request data"
// @Failure 401 {object} dto.MessageResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.MessageResponse "Session not found"
// @Failure 500 {object} dto.MessageResponse "Failed to send media"
// @Router /session/{sessionId}/message/send/media [post]
func (h *MessageHandler) SendMedia(c *fiber.Ctx) error {
	sessionIDOrName := c.Params("sessionId")
	if sessionIDOrName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.NewMessageErrorResponse(
			fiber.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
	}

	var req dto.SendMediaRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request body",
			err.Error(),
		))
	}

	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"VALIDATION_ERROR",
			"Request validation failed",
			err.Error(),
		))
	}

	mediaData, err := h.decodeMediaData(req.MediaURL)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_MEDIA_DATA",
			"Failed to decode media data",
			err.Error(),
		))
	}

	ctx := c.Context()
	var sendResp *whatsmeow.SendResponse

	switch req.MediaType {
	case "image":
		sendResp, err = h.wmeowService.SendImageMessage(ctx, sessionID, req.Phone, mediaData, req.Caption, "image/jpeg")
	case "audio":
		sendResp, err = h.wmeowService.SendAudioMessageWithPTT(ctx, sessionID, req.Phone, mediaData, "audio/mpeg", req.PTT)
	case "video":
		sendResp, err = h.wmeowService.SendVideoMessage(ctx, sessionID, req.Phone, mediaData, req.Caption, "video/mp4")
	case "document":
		filename := "document"
		sendResp, err = h.wmeowService.SendDocumentMessage(ctx, sessionID, req.Phone, mediaData, filename, req.Caption, "application/octet-stream")
	case "sticker":
		sendResp, err = h.wmeowService.SendStickerMessage(ctx, sessionID, req.Phone, mediaData, "image/webp")
	default:
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_MEDIA_TYPE",
			"Invalid media type",
			"Supported types: image, audio, video, document, sticker",
		))
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewMessageErrorResponse(
			fiber.StatusInternalServerError,
			"SEND_MEDIA_FAILED",
			"Failed to send media message",
			err.Error(),
		))
	}

	messageID := string(sendResp.ID)

	var response *dto.MessageResponse
	switch req.MediaType {
	case "image":
		response = dto.NewImageResponse(true, fiber.StatusOK, req.Phone, messageID, "", req.Caption, true)
	case "audio":
		response = dto.NewAudioResponse(true, fiber.StatusOK, req.Phone, messageID, "", req.PTT, true)
	case "video":
		response = dto.NewVideoResponse(true, fiber.StatusOK, req.Phone, messageID, "", req.Caption, false, true)
	case "document":
		response = dto.NewDocumentResponse(true, fiber.StatusOK, req.Phone, messageID, "", "document", "application/octet-stream", true)
	case "sticker":
		response = dto.NewStickerResponse(true, fiber.StatusOK, req.Phone, messageID, "", true)
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// MarkAsRead godoc
// @Summary Mark message as read
// @Description Marks a message as read in a WhatsApp chat
// @Tags Messages
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.MarkAsReadRequest true "Mark as read request"
// @Success 200 {object} dto.MessageResponse "Message marked as read"
// @Failure 400 {object} dto.MessageResponse "Invalid request data"
// @Failure 401 {object} dto.MessageResponse "Unauthorized - Invalid API key" "Invalid request data"
// @Failure 401 {object} dto.MessageResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.MessageResponse "Session not found"
// @Failure 500 {object} dto.MessageResponse "Failed to mark as read"
// @Router /session/{sessionId}/message/markread [post]
func (h *MessageHandler) MarkAsRead(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	var req dto.MarkAsReadRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageActionErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	if req.Phone == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageActionErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_PHONE",
			"Phone number is required",
			"",
		))
	}

	if len(req.MessageIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageActionErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_MESSAGE_IDS",
			"At least one message ID is required",
			"",
		))
	}

	ctx := c.Context()
	if err := h.wmeowService.MarkAsRead(ctx, sessionID, req.Phone, req.MessageIDs); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewMessageActionErrorResponse(
			fiber.StatusInternalServerError,
			"MARK_READ_FAILED",
			"Failed to mark messages as read",
			err.Error(),
		))
	}

	response := dto.NewMessageActionSuccessResponse(req.Phone, "", "mark_read")
	return c.Status(fiber.StatusOK).JSON(response)
}

// ReactToMessage godoc
// @Summary React to message
// @Description Adds an emoji reaction to a WhatsApp message
// @Tags Messages
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.ReactToMessageRequest true "React to message request"
// @Success 200 {object} dto.MessageResponse "Reaction added successfully"
// @Failure 400 {object} dto.MessageResponse "Invalid request data"
// @Failure 401 {object} dto.MessageResponse "Unauthorized - Invalid API key" "Invalid request data"
// @Failure 404 {object} dto.MessageResponse "Session not found"
// @Failure 500 {object} dto.MessageResponse "Failed to add reaction"
// @Router /session/{sessionId}/message/react [post]
func (h *MessageHandler) ReactToMessage(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	var req dto.ReactToMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageActionErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	if req.Phone == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageActionErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_PHONE",
			"Phone number is required",
			"",
		))
	}

	if req.MessageID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageActionErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_MESSAGE_ID",
			"Message ID is required",
			"",
		))
	}

	if req.Emoji == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageActionErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_EMOJI",
			"Emoji is required (use 'remove' to remove reaction)",
			"",
		))
	}

	ctx := c.Context()
	err := h.wmeowService.ReactToMessage(ctx, sessionID, req.Phone, req.MessageID, req.Emoji)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewMessageActionErrorResponse(
			fiber.StatusInternalServerError,
			"REACT_FAILED",
			"Failed to react to message",
			err.Error(),
		))
	}

	response := dto.NewMessageActionSuccessResponse(req.Phone, req.MessageID, "react")
	return c.Status(fiber.StatusOK).JSON(response)
}

// DeleteMessage godoc
// @Summary Delete message
// @Description Deletes a previously sent WhatsApp message
// @Tags Messages
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.DeleteMessageRequest true "Delete message request"
// @Success 200 {object} dto.MessageResponse "Message deleted successfully"
// @Failure 400 {object} dto.MessageResponse "Invalid request data"
// @Failure 401 {object} dto.MessageResponse "Unauthorized - Invalid API key" "Invalid request data"
// @Failure 404 {object} dto.MessageResponse "Session not found"
// @Failure 500 {object} dto.MessageResponse "Failed to delete message"
// @Router /session/{sessionId}/message/delete [post]
func (h *MessageHandler) DeleteMessage(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	var req dto.DeleteMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageActionErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	if req.Phone == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageActionErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_PHONE",
			"Phone number is required",
			"",
		))
	}

	if req.MessageID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageActionErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_MESSAGE_ID",
			"Message ID is required",
			"",
		))
	}

	ctx := c.Context()
	err := h.wmeowService.DeleteMessage(ctx, sessionID, req.Phone, req.MessageID, true)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewMessageActionErrorResponse(
			fiber.StatusInternalServerError,
			"DELETE_FAILED",
			"Failed to delete message",
			err.Error(),
		))
	}

	response := dto.NewMessageActionSuccessResponse(req.Phone, req.MessageID, "delete")
	return c.Status(fiber.StatusOK).JSON(response)
}

// EditMessage godoc
// @Summary Edit message
// @Description Edits a previously sent WhatsApp message
// @Tags Messages
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.EditMessageRequest true "Edit message request"
// @Success 200 {object} dto.MessageResponse "Message edited successfully"
// @Failure 400 {object} dto.MessageResponse "Invalid request data"
// @Failure 401 {object} dto.MessageResponse "Unauthorized - Invalid API key" "Invalid request data"
// @Failure 404 {object} dto.MessageResponse "Session not found"
// @Failure 500 {object} dto.MessageResponse "Failed to edit message"
// @Router /session/{sessionId}/message/edit [post]
func (h *MessageHandler) EditMessage(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	var req dto.EditMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageActionErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	if req.Phone == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageActionErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_PHONE",
			"Phone number is required",
			"",
		))
	}

	if req.MessageID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageActionErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_MESSAGE_ID",
			"Message ID is required",
			"",
		))
	}

	if req.NewText == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageActionErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_NEW_TEXT",
			"New text is required",
			"",
		))
	}

	ctx := c.Context()
	sendResp, err := h.wmeowService.EditMessage(ctx, sessionID, req.Phone, req.MessageID, req.NewText)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewMessageActionErrorResponse(
			fiber.StatusInternalServerError,
			"EDIT_FAILED",
			"Failed to edit message",
			err.Error(),
		))
	}

	messageID := string(sendResp.ID)
	response := dto.NewMessageActionSuccessResponse(req.Phone, messageID, "edit")
	return c.Status(fiber.StatusOK).JSON(response)
}

// SendLocation godoc
// @Summary Send location message
// @Description Sends a location message to a WhatsApp contact or group
// @Tags Messages
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.SendLocationRequest true "Location message request"
// @Success 200 {object} dto.MessageResponse "Location sent successfully"
// @Failure 400 {object} dto.MessageResponse "Invalid request data"
// @Failure 401 {object} dto.MessageResponse "Unauthorized - Invalid API key" "Invalid request data"
// @Failure 404 {object} dto.MessageResponse "Session not found"
// @Failure 500 {object} dto.MessageResponse "Failed to send location"
// @Router /session/{sessionId}/message/send/location [post]
func (h *MessageHandler) SendLocation(c *fiber.Ctx) error {
	sessionIDOrName := c.Params("sessionId")
	if sessionIDOrName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.NewMessageErrorResponse(
			fiber.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
	}

	var req dto.SendLocationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request body",
			err.Error(),
		))
	}

	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"VALIDATION_ERROR",
			"Request validation failed",
			err.Error(),
		))
	}

	ctx := c.Context()
	sendResp, err := h.wmeowService.SendLocationMessage(ctx, sessionID, req.Phone, req.Latitude, req.Longitude, req.Name, req.Address)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewMessageErrorResponse(
			fiber.StatusInternalServerError,
			"SEND_LOCATION_FAILED",
			"Failed to send location message",
			err.Error(),
		))
	}

	messageID := string(sendResp.ID)
	messageResponse := dto.NewLocationResponse(true, fiber.StatusOK, req.Phone, messageID, req.Latitude, req.Longitude, req.Name, "", true)
	return c.Status(fiber.StatusOK).JSON(messageResponse)
}

// SendContact godoc
// @Summary Send contact message
// @Description Sends a contact message to a WhatsApp contact or group
// @Tags Messages
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.SendContactRequest true "Contact message request"
// @Success 200 {object} dto.MessageResponse "Contact sent successfully"
// @Failure 400 {object} dto.MessageResponse "Invalid request data"
// @Failure 401 {object} dto.MessageResponse "Unauthorized - Invalid API key" "Invalid request data"
// @Failure 404 {object} dto.MessageResponse "Session not found"
// @Failure 500 {object} dto.MessageResponse "Failed to send contact"
// @Router /session/{sessionId}/message/send/contact [post]
func (h *MessageHandler) SendContact(c *fiber.Ctx) error {
	sessionIDOrName := c.Params("sessionId")
	if sessionIDOrName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.NewMessageErrorResponse(
			fiber.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
	}

	var req dto.SendContactRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request body",
			err.Error(),
		))
	}

	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"VALIDATION_ERROR",
			"Request validation failed",
			err.Error(),
		))
	}

	ctx := c.Context()

	if req.IsSingleContact() {
		contacts := []wmeow.ContactData{{
			Name:  req.ContactName,
			Phone: req.ContactPhone,
		}}

		sendResp, err := h.wmeowService.SendContactsMessage(ctx, sessionID, req.Phone, contacts)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(dto.NewMessageErrorResponse(
				fiber.StatusInternalServerError,
				"SEND_CONTACT_FAILED",
				"Failed to send contact message",
				err.Error(),
			))
		}

		vcard := "BEGIN:VCARD\nVERSION:3.0\nFN:" + req.ContactName + "\nTEL:" + req.ContactPhone + "\nEND:VCARD"
		messageID := string(sendResp.ID)
		messageResponse := dto.NewContactResponse(true, fiber.StatusOK, req.Phone, messageID, req.ContactName, vcard, true)
		return c.Status(fiber.StatusOK).JSON(messageResponse)
	}

	if req.IsMultipleContacts() {
		var contacts []wmeow.ContactData
		for _, contact := range req.Contacts {
			contacts = append(contacts, wmeow.ContactData{
				Name:  contact.Name,
				Phone: contact.Phone,
			})
		}

		sendResp, err := h.wmeowService.SendContactsMessage(ctx, sessionID, req.Phone, contacts)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(dto.NewMessageErrorResponse(
				fiber.StatusInternalServerError,
				"SEND_CONTACTS_FAILED",
				"Failed to send contacts message",
				err.Error(),
			))
		}

		var vcards []string
		for _, contact := range req.Contacts {
			vcard := fmt.Sprintf("BEGIN:VCARD\nVERSION:3.0\nFN:%s\nTEL:%s\nEND:VCARD", contact.Name, contact.Phone)
			vcards = append(vcards, vcard)
		}

		messageID := string(sendResp.ID)
		messageResponse := dto.NewContactsMessageResponse(true, fiber.StatusOK, req.Phone, messageID, vcards, true)
		return c.Status(fiber.StatusOK).JSON(messageResponse)
	}

	return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
		fiber.StatusBadRequest,
		"INVALID_REQUEST_FORMAT",
		"Invalid request format",
		"Must provide either single contact or multiple contacts",
	))
}

// SendImage godoc
// @Summary Send image message
// @Description Sends an image message to a WhatsApp contact or group
// @Tags Messages
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.SendImageRequest true "Image message request"
// @Success 200 {object} dto.MessageResponse "Image sent successfully"
// @Failure 400 {object} dto.MessageResponse "Invalid request data"
// @Failure 401 {object} dto.MessageResponse "Unauthorized - Invalid API key" "Invalid request data"
// @Failure 404 {object} dto.MessageResponse "Session not found"
// @Failure 500 {object} dto.MessageResponse "Failed to send image"
// @Router /session/{sessionId}/message/send/image [post]
func (h *MessageHandler) SendImage(c *fiber.Ctx) error {
	sessionIDOrName := c.Params("sessionId")
	if sessionIDOrName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.NewMessageErrorResponse(
			fiber.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
	}

	var req dto.SendImageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request body",
			err.Error(),
		))
	}

	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"VALIDATION_ERROR",
			"Request validation failed",
			err.Error(),
		))
	}

	imageData, err := h.decodeMediaData(req.Image)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_IMAGE_DATA",
			"Failed to decode image data",
			err.Error(),
		))
	}

	ctx := c.Context()
	sendResp, err := h.wmeowService.SendImageMessage(ctx, sessionID, req.Phone, imageData, req.Caption, "image/jpeg")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewMessageErrorResponse(
			fiber.StatusInternalServerError,
			"SEND_IMAGE_FAILED",
			"Failed to send image message",
			err.Error(),
		))
	}

	messageID := string(sendResp.ID)
	messageResponse := dto.NewImageResponse(true, fiber.StatusOK, req.Phone, messageID, req.Image, req.Caption, true)
	return c.Status(fiber.StatusOK).JSON(messageResponse)
}

// SendAudio godoc
// @Summary Send audio message
// @Description Sends an audio message to a WhatsApp contact or group
// @Tags Messages
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.SendAudioRequest true "Audio message request"
// @Success 200 {object} dto.MessageResponse "Audio sent successfully"
// @Failure 400 {object} dto.MessageResponse "Invalid request data"
// @Failure 401 {object} dto.MessageResponse "Unauthorized - Invalid API key" "Invalid request data"
// @Failure 404 {object} dto.MessageResponse "Session not found"
// @Failure 500 {object} dto.MessageResponse "Failed to send audio"
// @Router /session/{sessionId}/message/send/audio [post]
func (h *MessageHandler) SendAudio(c *fiber.Ctx) error {
	sessionIDOrName := c.Params("sessionId")
	if sessionIDOrName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.NewMessageErrorResponse(
			fiber.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
	}

	var req dto.SendAudioRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request body",
			err.Error(),
		))
	}

	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"VALIDATION_ERROR",
			"Request validation failed",
			err.Error(),
		))
	}

	audioData, err := h.decodeMediaData(req.Audio)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_AUDIO_DATA",
			"Failed to decode audio data",
			err.Error(),
		))
	}

	ctx := c.Context()
	sendResp, err := h.wmeowService.SendAudioMessageWithPTT(ctx, sessionID, req.Phone, audioData, "audio/mpeg", req.PTT)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewMessageErrorResponse(
			fiber.StatusInternalServerError,
			"SEND_AUDIO_FAILED",
			"Failed to send audio message",
			err.Error(),
		))
	}

	messageID := string(sendResp.ID)
	messageResponse := dto.NewAudioResponse(true, fiber.StatusOK, req.Phone, messageID, req.Audio, req.PTT, true)
	return c.Status(fiber.StatusOK).JSON(messageResponse)
}

// SendDocument godoc
// @Summary Send document message
// @Description Sends a document message to a WhatsApp contact or group
// @Tags Messages
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.SendDocumentRequest true "Document message request"
// @Success 200 {object} dto.MessageResponse "Document sent successfully"
// @Failure 400 {object} dto.MessageResponse "Invalid request data"
// @Failure 401 {object} dto.MessageResponse "Unauthorized - Invalid API key" "Invalid request data"
// @Failure 404 {object} dto.MessageResponse "Session not found"
// @Failure 500 {object} dto.MessageResponse "Failed to send document"
// @Router /session/{sessionId}/message/send/document [post]
func (h *MessageHandler) SendDocument(c *fiber.Ctx) error {
	sessionIDOrName := c.Params("sessionId")
	if sessionIDOrName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.NewMessageErrorResponse(
			fiber.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
	}

	var req dto.SendDocumentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request body",
			err.Error(),
		))
	}

	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"VALIDATION_ERROR",
			"Request validation failed",
			err.Error(),
		))
	}

	documentData, err := h.decodeMediaData(req.Document)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_DOCUMENT_DATA",
			"Failed to decode document data",
			err.Error(),
		))
	}

	filename := req.FileName
	if filename == "" {
		filename = "document"
	}
	mimeType := req.MimeType
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	ctx := c.Context()
	var sendResp *whatsmeow.SendResponse
	sendResp, err = h.wmeowService.SendDocumentMessage(ctx, sessionID, req.Phone, documentData, filename, "", mimeType)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewMessageErrorResponse(
			fiber.StatusInternalServerError,
			"SEND_DOCUMENT_FAILED",
			"Failed to send document message",
			err.Error(),
		))
	}

	messageID := string(sendResp.ID)
	messageResponse := dto.NewDocumentResponse(true, fiber.StatusOK, req.Phone, messageID, req.Document, filename, mimeType, true)
	return c.Status(fiber.StatusOK).JSON(messageResponse)
}

// SendVideo godoc
// @Summary Send video message
// @Description Sends a video message to a WhatsApp contact or group
// @Tags Messages
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.SendVideoRequest true "Video message request"
// @Success 200 {object} dto.MessageResponse "Video sent successfully"
// @Failure 400 {object} dto.MessageResponse "Invalid request data"
// @Failure 401 {object} dto.MessageResponse "Unauthorized - Invalid API key" "Invalid request data"
// @Failure 404 {object} dto.MessageResponse "Session not found"
// @Failure 500 {object} dto.MessageResponse "Failed to send video"
// @Router /session/{sessionId}/message/send/video [post]
func (h *MessageHandler) SendVideo(c *fiber.Ctx) error {
	sessionIDOrName := c.Params("sessionId")
	if sessionIDOrName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.NewMessageErrorResponse(
			fiber.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
	}

	var req dto.SendVideoRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request body",
			err.Error(),
		))
	}

	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"VALIDATION_ERROR",
			"Request validation failed",
			err.Error(),
		))
	}

	videoData, err := h.decodeMediaData(req.Video)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_VIDEO_DATA",
			"Failed to decode video data",
			err.Error(),
		))
	}

	ctx := c.Context()
	sendResp, err := h.wmeowService.SendVideoMessage(ctx, sessionID, req.Phone, videoData, req.Caption, "video/mp4")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewMessageErrorResponse(
			fiber.StatusInternalServerError,
			"SEND_VIDEO_FAILED",
			"Failed to send video message",
			err.Error(),
		))
	}

	messageID := string(sendResp.ID)
	messageResponse := dto.NewVideoResponse(true, fiber.StatusOK, req.Phone, messageID, req.Video, req.Caption, req.GifPlayback, true)
	return c.Status(fiber.StatusOK).JSON(messageResponse)
}

// SendSticker godoc
// @Summary Send sticker message
// @Description Sends a sticker message to a WhatsApp contact or group
// @Tags Messages
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.SendStickerRequest true "Sticker message request"
// @Success 200 {object} dto.MessageResponse "Sticker sent successfully"
// @Failure 400 {object} dto.MessageResponse "Invalid request data"
// @Failure 401 {object} dto.MessageResponse "Unauthorized - Invalid API key" "Invalid request data"
// @Failure 404 {object} dto.MessageResponse "Session not found"
// @Failure 500 {object} dto.MessageResponse "Failed to send sticker"
// @Router /session/{sessionId}/message/send/sticker [post]
func (h *MessageHandler) SendSticker(c *fiber.Ctx) error {
	sessionIDOrName := c.Params("sessionId")
	if sessionIDOrName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_SESSION_ID",
			"Session ID is required",
			"Session ID must be provided in the URL path",
		))
	}

	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.NewMessageErrorResponse(
			fiber.StatusNotFound,
			"SESSION_NOT_FOUND",
			"Session not found",
			err.Error(),
		))
	}

	var req dto.SendStickerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request body",
			err.Error(),
		))
	}

	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"VALIDATION_ERROR",
			"Request validation failed",
			err.Error(),
		))
	}

	stickerData, err := h.decodeMediaData(req.Sticker)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_STICKER_DATA",
			"Failed to decode sticker data",
			err.Error(),
		))
	}

	ctx := c.Context()
	sendResp, err := h.wmeowService.SendStickerMessage(ctx, sessionID, req.Phone, stickerData, "image/webp")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewMessageErrorResponse(
			fiber.StatusInternalServerError,
			"SEND_STICKER_FAILED",
			"Failed to send sticker message",
			err.Error(),
		))
	}

	messageID := string(sendResp.ID)
	messageResponse := dto.NewStickerResponse(true, fiber.StatusOK, req.Phone, messageID, req.Sticker, true)
	return c.Status(fiber.StatusOK).JSON(messageResponse)
}

// SendButton godoc
// @Summary Send button message
// @Description Sends an interactive button message to a WhatsApp contact or group
// @Tags Messages
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.SendButtonMessageRequest true "Button message request"
// @Success 200 {object} dto.MessageResponse "Button message sent successfully"
// @Failure 400 {object} dto.MessageResponse "Invalid request data"
// @Failure 401 {object} dto.MessageResponse "Unauthorized - Invalid API key" "Invalid request data"
// @Failure 404 {object} dto.MessageResponse "Session not found"
// @Failure 500 {object} dto.MessageResponse "Failed to send button message"
// @Router /session/{sessionId}/message/send/buttons [post]
func (h *MessageHandler) SendButton(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	var req dto.SendButtonMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"VALIDATION_ERROR",
			"Request validation failed",
			err.Error(),
		))
	}

	var buttons []wmeow.ButtonData
	for _, btn := range req.Buttons {
		buttons = append(buttons, wmeow.ButtonData{
			ID:   btn.ID,
			Text: btn.Text,
			Type: btn.Type,
		})
	}

	ctx := c.Context()
	resp, err := h.wmeowService.SendButtonMessage(ctx, sessionID, req.Phone, req.Title, buttons)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewMessageErrorResponse(
			fiber.StatusInternalServerError,
			"SEND_BUTTON_MESSAGE_FAILED",
			"Failed to send button message",
			err.Error(),
		))
	}

	response := dto.NewMessageSuccessResponse(sessionID, req.Phone, "button_message_sent", resp.ID, resp.Timestamp.Unix())
	return c.Status(fiber.StatusOK).JSON(response)
}

// SendList godoc
// @Summary Send list message
// @Description Sends an interactive list message to a WhatsApp contact or group
// @Tags Messages
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.SendListMessageRequest true "List message request"
// @Success 200 {object} dto.MessageResponse "List message sent successfully"
// @Failure 400 {object} dto.MessageResponse "Invalid request data"
// @Failure 401 {object} dto.MessageResponse "Unauthorized - Invalid API key" "Invalid request data"
// @Failure 404 {object} dto.MessageResponse "Session not found"
// @Failure 500 {object} dto.MessageResponse "Failed to send list message"
// @Router /session/{sessionId}/message/send/list [post]
func (h *MessageHandler) SendList(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	var req dto.SendListMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"VALIDATION_ERROR",
			"Request validation failed",
			err.Error(),
		))
	}

	var sections []wmeow.ListSection
	for _, section := range req.Sections {
		var rows []wmeow.ListRow
		for _, row := range section.Rows {
			rows = append(rows, wmeow.ListRow{
				ID:          row.ID,
				Title:       row.Title,
				Description: row.Description,
			})
		}
		sections = append(sections, wmeow.ListSection{
			Title: section.Title,
			Rows:  rows,
		})
	}

	ctx := c.Context()
	resp, err := h.wmeowService.SendListMessage(ctx, sessionID, req.Phone, req.Title, req.Description, req.ButtonText, req.FooterText, sections)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewMessageErrorResponse(
			fiber.StatusInternalServerError,
			"SEND_LIST_MESSAGE_FAILED",
			"Failed to send list message",
			err.Error(),
		))
	}

	response := dto.NewMessageSuccessResponse(sessionID, req.Phone, "list_message_sent", resp.ID, resp.Timestamp.Unix())
	return c.Status(fiber.StatusOK).JSON(response)
}

// SendPoll godoc
// @Summary Send poll message
// @Description Sends a poll message to a WhatsApp contact or group
// @Tags Messages
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.SendPollMessageRequest true "Poll message request"
// @Success 200 {object} dto.MessageResponse "Poll message sent successfully"
// @Failure 400 {object} dto.MessageResponse "Invalid request data"
// @Failure 401 {object} dto.MessageResponse "Unauthorized - Invalid API key" "Invalid request data"
// @Failure 404 {object} dto.MessageResponse "Session not found"
// @Failure 500 {object} dto.MessageResponse "Failed to send poll message"
// @Router /session/{sessionId}/message/send/poll [post]
func (h *MessageHandler) SendPoll(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	var req dto.SendPollMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMessageErrorResponse(
			fiber.StatusBadRequest,
			"VALIDATION_ERROR",
			"Request validation failed",
			err.Error(),
		))
	}

	ctx := c.Context()
	resp, err := h.wmeowService.SendPollMessage(ctx, sessionID, req.Phone, req.Name, req.Options, req.SelectableCount)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewMessageErrorResponse(
			fiber.StatusInternalServerError,
			"SEND_POLL_MESSAGE_FAILED",
			"Failed to send poll message",
			err.Error(),
		))
	}

	response := dto.NewMessageSuccessResponse(sessionID, req.Phone, "poll_message_sent", resp.ID, resp.Timestamp.Unix())
	return c.Status(fiber.StatusOK).JSON(response)
}
