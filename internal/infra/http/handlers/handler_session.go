package handlers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"

	"zpmeow/internal/application"
	"zpmeow/internal/domain/session"
	"zpmeow/internal/infra/http/dto"
	"zpmeow/internal/infra/wmeow"
)

type SessionHandler struct {
	*BaseHandler
	sessionService *application.SessionApp
	wmeowService   wmeow.WameowService
}

func NewSessionHandler(sessionService *application.SessionApp, wmeowService wmeow.WameowService) *SessionHandler {
	return &SessionHandler{
		BaseHandler:    NewBaseHandler("session-handler"),
		sessionService: sessionService,
		wmeowService:   wmeowService,
	}
}

func (h *SessionHandler) validateSessionID(c *fiber.Ctx) (string, bool) {
	sessionIDOrName := c.Params("sessionId")
	if sessionIDOrName == "" {
		if err := h.sendErrorResponse(c, fiber.StatusBadRequest, "SESSION_ID_REQUIRED", "Session ID or name is required", "Missing session ID or name in path"); err != nil {
			h.logger.Errorf("Failed to send error response: %v", err)
		}
		return "", false
	}
	return sessionIDOrName, true
}

func (h *SessionHandler) bindAndValidateRequest(c *fiber.Ctx, req interface{}) bool {
	if err := h.BindAndValidate(c, req); err != nil {
		h.logger.Errorf("Failed to bind or validate request: %v", err)
		if sendErr := h.SendValidationErrorResponse(c, err); sendErr != nil {
			h.logger.Errorf("Failed to send validation error response: %v", sendErr)
		}
		return false
	}
	return true
}

func (h *SessionHandler) sendSuccessResponse(c *fiber.Ctx, sessionID, action string, data interface{}) error {
	response := &dto.SessionResponse{
		Success: true,
		Code:    fiber.StatusOK,
		Data: dto.SessionData{
			SessionId: sessionID,
			Action:    action,
			Status:    "success",
			Timestamp: time.Now(),
		},
	}

	switch v := data.(type) {
	case *dto.SessionInfo:
		response.Data.Session = v
	case []dto.SessionInfo:
		response.Data.Sessions = v
	case string:
		response.Data.QRCode = v
	}

	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		h.logger.Errorf("Failed to marshal response: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to format response"})
	}

	c.Set("Content-Type", "application/json")
	return c.Status(fiber.StatusOK).Send(jsonBytes)
}

func (h *SessionHandler) sendErrorResponse(c *fiber.Ctx, status int, errorCode, message, _ string) error {
	response := &dto.SessionResponse{
		Success: false,
		Code:    status,
		Data: dto.SessionData{
			Status:    "error",
			Timestamp: time.Now(),
		},
		Error: &dto.ErrorInfo{
			Code:    errorCode,
			Message: message,
		},
	}

	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		h.logger.Errorf("Failed to marshal error response: %v", err)
		return c.Status(status).JSON(fiber.Map{"error": "Failed to format error response"})
	}

	c.Set("Content-Type", "application/json")
	return c.Status(status).Send(jsonBytes)
}

func (h *SessionHandler) convertToSessionInfo(session *session.Session) *dto.SessionInfo {
	sessionInfo := &dto.SessionInfo{
		ID:        session.SessionID().Value(),
		Name:      session.Name().Value(),
		Status:    string(session.Status()),
		DeviceJID: session.DeviceJID().Value(),
		ApiKey:    session.ApiKey().Value(),
		CreatedAt: session.CreatedAt().Value(),
		UpdatedAt: session.UpdatedAt().Value(),
	}

	return sessionInfo
}

func (h *SessionHandler) logOperation(operation, details string) {
	h.logger.Infof("%s: %s", operation, details)
}

func (h *SessionHandler) logSuccess(operation, details string) {
	h.logger.Infof("%s completed successfully: %s", operation, details)
}

func (h *SessionHandler) logError(operation string, err error) {
	h.logger.Errorf("Failed to %s: %v", operation, err)
}

// GetSessions godoc
// @Summary List all sessions
// @Description Retrieves a list of all WhatsApp sessions
// @Tags Sessions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} dto.SessionResponse{data=dto.SessionData{sessions=[]dto.SessionInfo}} "List of sessions"
// @Failure 401 {object} dto.SessionResponse "Unauthorized - Invalid API key"
// @Failure 500 {object} dto.SessionResponse "Failed to get sessions"
// @Router /sessions/list [get]
func (h *SessionHandler) GetSessions(c *fiber.Ctx) error {
	h.logOperation("Getting all sessions", "")

	sessions, err := h.sessionService.GetAllSessions(c.Context())
	if err != nil {
		h.logError("get all sessions", err)
		return h.sendErrorResponse(c, fiber.StatusInternalServerError, "GET_SESSIONS_FAILED", "Failed to get sessions", err.Error())
	}

	sessionInfos := make([]dto.SessionInfo, len(sessions))
	for i, session := range sessions {
		sessionInfos[i] = *h.convertToSessionInfo(session)
	}

	h.logSuccess("Get all sessions", fmt.Sprintf("retrieved %d sessions", len(sessions)))
	return h.sendSuccessResponse(c, "", "list", sessionInfos)
}

// GetSession godoc
// @Summary Get session information
// @Description Retrieves detailed information about a specific WhatsApp session
// @Tags Sessions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Success 200 {object} dto.SessionResponse{data=dto.SessionData{session=dto.SessionInfo}} "Session information"
// @Failure 400 {object} dto.SessionResponse "Invalid session ID"
// @Failure 401 {object} dto.SessionResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.SessionResponse "Session not found"
// @Router /sessions/{sessionId}/info [get]
func (h *SessionHandler) GetSession(c *fiber.Ctx) error {
	sessionID, ok := h.validateSessionID(c)
	if !ok {
		return nil // validateSessionId já enviou a resposta de erro
	}

	h.logOperation("Getting session", sessionID)

	session, err := h.sessionService.GetSession(c.Context(), sessionID)
	if err != nil {
		h.logError("get session "+sessionID, err)
		return h.sendErrorResponse(c, fiber.StatusNotFound, "SESSION_NOT_FOUND", "Session not found", err.Error())
	}

	sessionInfo := h.convertToSessionInfo(session)
	h.logSuccess("Get session", sessionID)
	return h.sendSuccessResponse(c, sessionID, "get", sessionInfo)
}

// CreateSession godoc
// @Summary Create a new WhatsApp session
// @Description Creates a new WhatsApp session with the specified name
// @Tags Sessions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.CreateSessionRequest true "Session creation request"
// @Success 201 {object} dto.CreateSessionResponse "Session created successfully"
// @Failure 400 {object} dto.SessionResponse "Invalid request data"
// @Failure 401 {object} dto.SessionResponse "Unauthorized - Invalid API key"
// @Failure 500 {object} dto.SessionResponse "Failed to create session"
// @Router /sessions/create [post]
func (h *SessionHandler) CreateSession(c *fiber.Ctx) error {
	var req dto.CreateSessionRequest
	if !h.bindAndValidateRequest(c, &req) {
		return nil // bindAndValidateRequest já enviou a resposta de erro
	}

	h.logOperation("Creating session", "name: "+req.Name)

	appReq := application.CreateSessionRequest{
		Name: req.Name,
	}

	session, err := h.sessionService.CreateSessionWithRequest(c.Context(), appReq)
	if err != nil {
		h.logError("create session", err)
		return h.sendErrorResponse(c, fiber.StatusInternalServerError, "CREATE_SESSION_FAILED", "Failed to create session", err.Error())
	}

	sessionInfo := h.convertToSessionInfo(session)

	response := &dto.CreateSessionResponse{
		Success: true,
		Code:    fiber.StatusCreated,
		Data: &dto.SessionCreateData{
			Action:    "create",
			Status:    "success",
			Timestamp: time.Now(),
			Session:   sessionInfo,
		},
	}

	h.logSuccess("Create session", session.SessionID().Value())
	return c.Status(fiber.StatusCreated).JSON(response)
}

// DeleteSession godoc
// @Summary Delete a WhatsApp session
// @Description Permanently deletes a WhatsApp session and stops its client
// @Tags Sessions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Success 200 {object} dto.SessionResponse "Session deleted successfully"
// @Failure 400 {object} dto.SessionResponse "Invalid session ID"
// @Failure 401 {object} dto.SessionResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.SessionResponse "Session not found"
// @Failure 500 {object} dto.SessionResponse "Failed to delete session"
// @Router /sessions/{sessionId}/delete [delete]
func (h *SessionHandler) DeleteSession(c *fiber.Ctx) error {
	sessionID, ok := h.validateSessionID(c)
	if !ok {
		return nil // validateSessionId já enviou a resposta de erro
	}

	h.logOperation("Deleting session", sessionID)

	session, err := h.sessionService.GetSession(c.Context(), sessionID)
	if err != nil {
		h.logError("get session "+sessionID+" for deletion", err)
		return h.sendErrorResponse(c, fiber.StatusNotFound, "SESSION_NOT_FOUND", "Session not found", err.Error())
	}

	if err := h.wmeowService.StopClient(session.SessionID().Value()); err != nil {
		h.logger.Warnf("Could not stop client for session %s (may already be stopped): %v", session.SessionID().Value(), err)
	}

	if err := h.sessionService.DeleteSession(c.Context(), session.SessionID().Value()); err != nil {
		h.logError("delete session "+session.SessionID().Value(), err)
		return h.sendErrorResponse(c, fiber.StatusInternalServerError, "DELETE_SESSION_FAILED", "Failed to delete session", err.Error())
	}

	h.logSuccess("Delete session", sessionID)
	return h.sendSuccessResponse(c, sessionID, "delete", nil)
}

// ConnectSession godoc
// @Summary Connect a WhatsApp session
// @Description Starts the WhatsApp client for a session and generates QR code if needed
// @Tags Sessions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Success 200 {object} dto.ConnectSessionResponse "Session connection initiated"
// @Failure 400 {object} dto.SessionResponse "Invalid session ID"
// @Failure 401 {object} dto.SessionResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.SessionResponse "Session not found"
// @Failure 409 {object} dto.SessionResponse "Device already in use"
// @Failure 500 {object} dto.SessionResponse "Failed to start client"
// @Router /sessions/{sessionId}/connect [post]
func (h *SessionHandler) ConnectSession(c *fiber.Ctx) error {
	sessionID, ok := h.validateSessionID(c)
	if !ok {
		return nil // validateSessionId já enviou a resposta de erro
	}

	h.logOperation("Connecting session", sessionID)

	session, err := h.sessionService.GetSession(c.Context(), sessionID)
	if err != nil {
		h.logError("get session "+sessionID+" for connection", err)
		return h.sendErrorResponse(c, fiber.StatusNotFound, "SESSION_NOT_FOUND", "Session not found", err.Error())
	}

	if !session.DeviceJID().IsEmpty() {
		existingSession, err := h.sessionService.GetSessionByDeviceJID(c.Context(), session.DeviceJID().Value())
		if err == nil && existingSession.SessionID() != session.SessionID() {
			return h.sendErrorResponse(c, fiber.StatusConflict, "DEVICE_ALREADY_IN_USE",
				fmt.Sprintf("Device %s is already in use by session %s (%s)", session.DeviceJID().Value(), existingSession.SessionID().Value(), existingSession.Name().Value()),
				"Each meow device can only be used by one session at a time")
		}
	}

	if err := h.wmeowService.StartClient(session.SessionID().Value()); err != nil {
		h.logError("start client for session "+session.SessionID().Value(), err)
		return h.sendErrorResponse(c, fiber.StatusInternalServerError, "START_CLIENT_FAILED", "Failed to start client", err.Error())
	}

	var qrCode string
	isConnected := h.wmeowService.IsClientConnected(session.SessionID().Value())
	if !isConnected {
		qrCode, err = h.wmeowService.GetQRCode(session.SessionID().Value())
		if err != nil {
			h.logger.Errorf("Failed to get QR code for session %s: %v", session.SessionID().Value(), err)
		}
	}

	connectionInfo := &dto.SessionConnectionInfo{
		QRCode:      qrCode,
		Connected:   isConnected,
		IsConnected: isConnected,
	}

	response := &dto.ConnectSessionResponse{
		Success: true,
		Code:    fiber.StatusOK,
		Data: &dto.SessionConnectData{
			SessionId:  session.SessionID().Value(),
			Action:     "connect",
			Status:     "success",
			Timestamp:  time.Now(),
			Session:    h.convertToSessionInfo(session),
			Connection: connectionInfo,
			QRCode:     qrCode,
		},
	}

	h.logSuccess("Connect session", sessionID)
	return c.Status(fiber.StatusOK).JSON(response)
}

// DisconnectSession godoc
// @Summary Disconnect a WhatsApp session
// @Description Stops the WhatsApp client for a session and disconnects it
// @Tags Sessions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Success 200 {object} dto.SessionResponse "Session disconnected successfully"
// @Failure 400 {object} dto.SessionResponse "Invalid session ID"
// @Failure 401 {object} dto.SessionResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.SessionResponse "Session not found"
// @Failure 500 {object} dto.SessionResponse "Failed to disconnect session"
// @Router /sessions/{sessionId}/disconnect [post]
func (h *SessionHandler) DisconnectSession(c *fiber.Ctx) error {
	sessionID, ok := h.validateSessionID(c)
	if !ok {
		h.logger.Debugf("DisconnectSession: validateSessionId failed for %s", sessionID)
		return nil // validateSessionId já enviou a resposta de erro
	}

	h.logOperation("Disconnecting session", sessionID)
	h.logger.Debugf("DisconnectSession: Starting disconnect for session %s", sessionID)

	session, err := h.sessionService.GetSession(c.Context(), sessionID)
	if err != nil {
		h.logger.Debugf("DisconnectSession: Failed to get session %s: %v", sessionID, err)
		h.logError("get session "+sessionID+" for disconnection", err)
		return h.sendErrorResponse(c, fiber.StatusNotFound, "SESSION_NOT_FOUND", "Session not found", err.Error())
	}

	h.logger.Debugf("DisconnectSession: Found session %s (ID: %s), calling StopClient", sessionID, session.SessionID().Value())

	if err := h.wmeowService.StopClient(session.SessionID().Value()); err != nil {
		h.logger.Debugf("DisconnectSession: StopClient failed for session %s: %v", session.SessionID().Value(), err)
		h.logError("stop client for session "+session.SessionID().Value(), err)
		return h.sendErrorResponse(c, fiber.StatusInternalServerError, "STOP_CLIENT_FAILED", "Failed to disconnect session", err.Error())
	}

	h.logger.Debugf("DisconnectSession: StopClient succeeded for session %s, sending success response", sessionID)
	h.logSuccess("Disconnect session", sessionID)
	h.logger.Debugf("DisconnectSession: Completed successfully for session %s", sessionID)
	return h.sendSuccessResponse(c, sessionID, "disconnect", nil)
}

// PairPhone godoc
// @Summary Pair phone with session
// @Description Pairs a phone number with a WhatsApp session using pairing code
// @Tags Sessions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.PairPhoneRequest true "Phone pairing request"
// @Success 200 {object} dto.PairPhoneResponse "Phone paired successfully"
// @Failure 400 {object} dto.SessionResponse "Invalid request data"
// @Failure 401 {object} dto.SessionResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.SessionResponse "Session not found"
// @Failure 500 {object} dto.SessionResponse "Failed to pair phone"
// @Router /sessions/{sessionId}/pair [post]
func (h *SessionHandler) PairPhone(c *fiber.Ctx) error {
	sessionID, ok := h.validateSessionID(c)
	if !ok {
		return nil // validateSessionId já enviou a resposta de erro
	}

	var req dto.PairPhoneRequest
	if !h.bindAndValidateRequest(c, &req) {
		return nil // bindAndValidateRequest já enviou a resposta de erro
	}

	h.logOperation("Pairing phone for session", fmt.Sprintf("session: %s, phone: %s", sessionID, req.Phone))

	session, err := h.sessionService.GetSession(c.Context(), sessionID)
	if err != nil {
		h.logError("get session "+sessionID+" for phone pairing", err)
		return h.sendErrorResponse(c, fiber.StatusNotFound, "SESSION_NOT_FOUND", "Session not found", err.Error())
	}

	pairCode, err := h.wmeowService.PairPhone(session.SessionID().Value(), req.Phone)
	if err != nil {
		h.logError("pair phone for session "+session.SessionID().Value(), err)
		return h.sendErrorResponse(c, fiber.StatusInternalServerError, "PHONE_PAIRING_FAILED", "Failed to pair phone", err.Error())
	}

	response := &dto.PairPhoneResponse{
		Success: true,
		Code:    fiber.StatusOK,
		Data: &dto.PairPhoneResponseData{
			SessionId: sessionID,
			Action:    "pair",
			Status:    "success",
			Timestamp: time.Now(),
			Phone:     req.Phone,
			Code:      pairCode,
		},
	}

	h.logSuccess("Pair phone", fmt.Sprintf("session: %s, phone: %s", sessionID, req.Phone))
	return c.Status(fiber.StatusOK).JSON(response)
}

// GetSessionStatus godoc
// @Summary Get session status
// @Description Retrieves the current status and connection state of a WhatsApp session
// @Tags Sessions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Success 200 {object} dto.SessionStatusResponse "Session status information"
// @Failure 400 {object} dto.SessionResponse "Invalid session ID"
// @Failure 401 {object} dto.SessionResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.SessionResponse "Session not found"
// @Router /sessions/{sessionId}/status [get]
func (h *SessionHandler) GetSessionStatus(c *fiber.Ctx) error {
	sessionID, ok := h.validateSessionID(c)
	if !ok {
		return nil // validateSessionId já enviou a resposta de erro
	}

	h.logOperation("Getting status for session", sessionID)

	session, err := h.sessionService.GetSession(c.Context(), sessionID)
	if err != nil {
		h.logError("get session "+sessionID+" for status", err)
		return h.sendErrorResponse(c, fiber.StatusNotFound, "SESSION_NOT_FOUND", "Session not found", err.Error())
	}

	isConnected := h.wmeowService.IsClientConnected(session.SessionID().Value())
	clientStatus := "disconnected"
	if isConnected {
		clientStatus = "connected"
	}

	response := &dto.SessionStatusResponse{
		Success: true,
		Code:    fiber.StatusOK,
		Data: &dto.SessionStatusResponseData{
			SessionId:     sessionID,
			Action:        "status",
			Status:        "success",
			Timestamp:     time.Now(),
			Name:          session.Name().Value(),
			SessionStatus: string(session.Status()),
			DeviceJID:     session.DeviceJID().Value(),
			IsConnected:   isConnected,
			ClientStatus:  string(clientStatus),
			CreatedAt:     session.CreatedAt().Value(),
			UpdatedAt:     session.UpdatedAt().Value(),
		},
	}

	h.logSuccess("Get session status", sessionID)
	return c.Status(fiber.StatusOK).JSON(response)
}

// UpdateSessionWebhook godoc
// @Summary Update session webhook
// @Description Updates the webhook URL and event subscriptions for a WhatsApp session
// @Tags Sessions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID"
// @Param request body dto.UpdateWebhookRequest true "Webhook update request"
// @Success 200 {object} dto.SessionResponse "Webhook updated successfully"
// @Failure 400 {object} dto.SessionResponse "Invalid request data"
// @Failure 401 {object} dto.SessionResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} dto.SessionResponse "Session not found"
// @Failure 500 {object} dto.SessionResponse "Failed to update webhook"
// @Router /sessions/{sessionId}/webhook [put]
func (h *SessionHandler) UpdateSessionWebhook(c *fiber.Ctx) error {
	sessionID, ok := h.validateSessionID(c)
	if !ok {
		return nil // validateSessionId já enviou a resposta de erro
	}

	var req dto.UpdateWebhookRequest
	if !h.bindAndValidateRequest(c, &req) {
		return nil // bindAndValidateRequest já enviou a resposta de erro
	}

	h.logOperation("Updating webhook for session", fmt.Sprintf("session: %s, url: %s", sessionID, req.URL))

	session, err := h.sessionService.GetSession(c.Context(), sessionID)
	if err != nil {
		h.logError("get session "+sessionID+" for webhook update", err)
		return h.sendErrorResponse(c, fiber.StatusNotFound, "SESSION_NOT_FOUND", "Session not found", err.Error())
	}

	if err := h.wmeowService.UpdateSessionWebhook(session.SessionID().Value(), req.URL); err != nil {
		h.logError("update webhook for session "+session.SessionID().Value(), err)
		return h.sendErrorResponse(c, fiber.StatusInternalServerError, "WEBHOOK_UPDATE_FAILED", "Failed to update webhook", err.Error())
	}

	if len(req.Events) > 0 {
		if err := h.wmeowService.UpdateSessionSubscriptions(session.SessionID().Value(), req.Events); err != nil {
			h.logError("update events subscription for session "+session.SessionID().Value(), err)
			return h.sendErrorResponse(c, fiber.StatusInternalServerError, "EVENTS_UPDATE_FAILED", "Failed to update events subscription", err.Error())
		}
	}

	h.logSuccess("Update session webhook", sessionID)
	return h.sendSuccessResponse(c, sessionID, "webhook_update", nil)
}
