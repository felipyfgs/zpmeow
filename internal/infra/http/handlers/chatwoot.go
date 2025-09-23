package handlers

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"zpmeow/internal/application"
	"zpmeow/internal/application/ports"
	"zpmeow/internal/infra/chatwoot"
	"zpmeow/internal/infra/database/models"
	"zpmeow/internal/infra/database/repository"
	"zpmeow/internal/infra/http/dto"
)

type ChatwootHandler struct {
	*BaseHandler
	sessionService      *application.SessionApp
	chatwootIntegration *chatwoot.Integration
	chatwootRepo        *repository.ChatwootRepository
}

func NewChatwootHandler(sessionService *application.SessionApp, chatwootIntegration *chatwoot.Integration, chatwootRepo *repository.ChatwootRepository, whatsappService ports.WhatsAppService) *ChatwootHandler {
	// Configura o serviço WhatsApp na integração Chatwoot
	chatwootIntegration.SetWhatsAppService(whatsappService)

	return &ChatwootHandler{
		BaseHandler:         NewBaseHandler("chatwoot-handler"),
		sessionService:      sessionService,
		chatwootIntegration: chatwootIntegration,
		chatwootRepo:        chatwootRepo,
	}
}

// validateSessionID valida e retorna o sessionID do parâmetro
func (h *ChatwootHandler) validateSessionID(c *fiber.Ctx) (string, bool) {
	sessionID := c.Params("sessionId")
	if sessionID == "" {
		if err := h.SendErrorResponse(c, fiber.StatusBadRequest, "SESSION_ID_REQUIRED", "Session ID is required", fmt.Errorf("missing session ID in path")); err != nil {
			h.logger.Errorf("Failed to send error response: %v", err)
		}
		return "", false
	}
	return sessionID, true
}

// resolveSessionID resolve o sessionID ou nome para o ID real da sessão
func (h *ChatwootHandler) resolveSessionID(c *fiber.Ctx, sessionIDOrName string) (string, error) {
	if h.sessionService == nil {
		return sessionIDOrName, nil
	}

	ctx := c.Context()
	session, err := h.sessionService.GetSession(ctx, sessionIDOrName)
	if err != nil {
		return "", err
	}

	return session.SessionID().Value(), nil
}

// SetChatwootConfig configura a integração Chatwoot para uma sessão
// @Summary Configure Chatwoot integration
// @Description Configure Chatwoot integration for a WhatsApp session
// @Tags Chatwoot
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body dto.ChatwootConfigRequest true "Chatwoot configuration"
// @Success 200 {object} dto.ChatwootConfigResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /session/{sessionId}/chatwoot/config [post]
func (h *ChatwootHandler) SetChatwootConfig(c *fiber.Ctx) error {
	sessionID, valid := h.validateSessionID(c)
	if !valid {
		return nil
	}

	var req dto.ChatwootConfigRequest
	if !h.bindAndValidateRequest(c, &req) {
		return nil
	}

	// Verifica se a sessão existe
	_, err := h.sessionService.GetSession(c.Context(), sessionID)
	if err != nil {
		return h.SendErrorResponse(c, fiber.StatusInternalServerError, "DATABASE_ERROR", "Database error", err)
	}

	// Converte DTO para configuração interna
	config := h.dtoToConfig(&req, sessionID)

	// Converte para modelo do banco
	dbModel := h.configToDBModel(config, sessionID)

	// Verifica se já existe configuração para esta sessão
	existingConfig, err := h.chatwootRepo.GetBySessionID(c.Context(), sessionID)
	if err != nil {
		h.logger.Errorf("Failed to check existing config: %v", err)
		return h.SendErrorResponse(c, fiber.StatusInternalServerError, "DATABASE_ERROR", "Database error", err)
	}

	// Salva ou atualiza no banco de dados
	if existingConfig == nil {
		// Cria nova configuração
		if err := h.chatwootRepo.Create(c.Context(), dbModel); err != nil {
			h.logger.Errorf("Failed to save Chatwoot config: %v", err)
			return h.SendErrorResponse(c, fiber.StatusInternalServerError, "DATABASE_ERROR", "Database error", err)
		}
	} else {
		// Atualiza configuração existente
		dbModel.ID = existingConfig.ID
		if err := h.chatwootRepo.Update(c.Context(), dbModel); err != nil {
			h.logger.Errorf("Failed to update Chatwoot config: %v", err)
			return h.SendErrorResponse(c, fiber.StatusInternalServerError, "DATABASE_ERROR", "Database error", err)
		}
	}

	// Registra a configuração na integração
	if err := h.chatwootIntegration.RegisterInstance(sessionID, config); err != nil {
		h.logger.Errorf("Failed to register Chatwoot instance: %v", err)
		return h.SendErrorResponse(c, fiber.StatusInternalServerError, "DATABASE_ERROR", "Database error", err)
	}

	// Retorna a configuração
	response := h.configToDTO(config, sessionID, c)
	return h.SendSuccessResponse(c, fiber.StatusOK, response)
}

// GetChatwootConfig retorna a configuração Chatwoot de uma sessão
// @Summary Get Chatwoot configuration
// @Description Get current Chatwoot configuration for a session
// @Tags Chatwoot
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} dto.ChatwootConfigResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /session/{sessionId}/chatwoot/config [get]
func (h *ChatwootHandler) GetChatwootConfig(c *fiber.Ctx) error {
	sessionIDOrName, valid := h.validateSessionID(c)
	if !valid {
		return nil
	}

	// Resolve sessionID ou nome para o ID real da sessão
	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		return h.SendErrorResponse(c, fiber.StatusNotFound, "SESSION_NOT_FOUND", "Session not found", err)
	}

	// Busca configuração do banco de dados
	dbConfig, err := h.chatwootRepo.GetBySessionID(c.Context(), sessionID)
	if err != nil {
		h.logger.Errorf("Failed to get Chatwoot config: %v", err)
		return h.SendErrorResponse(c, fiber.StatusInternalServerError, "DATABASE_ERROR", "Database error", err)
	}

	var config *chatwoot.ChatwootConfig
	if dbConfig == nil {
		// Retorna configuração padrão se não existir
		config = &chatwoot.ChatwootConfig{Enabled: false}
	} else {
		// Converte do modelo do banco para configuração
		config = h.dbModelToConfig(dbConfig)
	}

	response := h.configToDTO(config, sessionIDOrName, c)
	return h.SendSuccessResponse(c, fiber.StatusOK, response)
}

// UpdateChatwootConfig atualiza a configuração Chatwoot de uma sessão
// @Summary Update Chatwoot configuration
// @Description Update Chatwoot configuration for a session
// @Tags Chatwoot
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body dto.ChatwootConfigRequest true "Chatwoot configuration"
// @Success 200 {object} dto.ChatwootConfigResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /session/{sessionId}/chatwoot/config [put]
func (h *ChatwootHandler) UpdateChatwootConfig(c *fiber.Ctx) error {
	sessionID, valid := h.validateSessionID(c)
	if !valid {
		return nil
	}

	var req dto.ChatwootConfigRequest
	if !h.bindAndValidateRequest(c, &req) {
		return nil
	}

	// Verifica se a sessão existe
	_, err := h.sessionService.GetSession(c.Context(), sessionID)
	if err != nil {
		return h.SendErrorResponse(c, fiber.StatusInternalServerError, "DATABASE_ERROR", "Database error", err)
	}

	// Converte para configuração interna
	config := h.dtoToConfig(&req, sessionID)

	// Converte para modelo do banco
	dbModel := h.configToDBModel(config, sessionID)

	// Atualiza no banco de dados
	if err := h.chatwootRepo.Update(c.Context(), dbModel); err != nil {
		h.logger.Errorf("Failed to update Chatwoot config: %v", err)
		return h.SendErrorResponse(c, fiber.StatusInternalServerError, "DATABASE_ERROR", "Database error", err)
	}

	// Remove configuração anterior da integração
	h.chatwootIntegration.UnregisterInstance(sessionID)

	// Registra nova configuração na integração
	if err := h.chatwootIntegration.RegisterInstance(sessionID, config); err != nil {
		h.logger.Errorf("Failed to update Chatwoot instance: %v", err)
		return h.SendErrorResponse(c, fiber.StatusInternalServerError, "DATABASE_ERROR", "Database error", err)
	}

	response := h.configToDTO(config, sessionID, c)
	return h.SendSuccessResponse(c, fiber.StatusOK, response)
}

// DeleteChatwootConfig remove a configuração Chatwoot de uma sessão
// @Summary Delete Chatwoot configuration
// @Description Remove Chatwoot configuration for a session
// @Tags Chatwoot
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} dto.SuccessResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /session/{sessionId}/chatwoot/config [delete]
func (h *ChatwootHandler) DeleteChatwootConfig(c *fiber.Ctx) error {
	sessionID, valid := h.validateSessionID(c)
	if !valid {
		return nil
	}

	// Verifica se a sessão existe
	_, err := h.sessionService.GetSession(c.Context(), sessionID)
	if err != nil {
		return h.SendErrorResponse(c, fiber.StatusInternalServerError, "DATABASE_ERROR", "Database error", err)
	}

	// Remove do banco de dados
	if err := h.chatwootRepo.Delete(c.Context(), sessionID); err != nil {
		h.logger.Errorf("Failed to delete Chatwoot config: %v", err)
		return h.SendErrorResponse(c, fiber.StatusInternalServerError, "DATABASE_ERROR", "Database error", err)
	}

	// Remove da integração
	h.chatwootIntegration.UnregisterInstance(sessionID)

	return h.SendSuccessResponse(c, fiber.StatusOK, nil)
}

// GetChatwootStatus retorna o status da integração Chatwoot
// @Summary Get Chatwoot status
// @Description Get current status of Chatwoot integration
// @Tags Chatwoot
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} dto.ChatwootStatusResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /session/{sessionId}/chatwoot/status [get]
func (h *ChatwootHandler) GetChatwootStatus(c *fiber.Ctx) error {
	sessionID, valid := h.validateSessionID(c)
	if !valid {
		return nil
	}

	// Verifica se a sessão existe
	_, err := h.sessionService.GetSession(c.Context(), sessionID)
	if err != nil {
		return h.SendErrorResponse(c, fiber.StatusInternalServerError, "DATABASE_ERROR", "Database error", err)
	}

	// Busca configuração do banco de dados
	dbConfig, err := h.chatwootRepo.GetBySessionID(c.Context(), sessionID)
	if err != nil {
		h.logger.Errorf("Failed to get Chatwoot config: %v", err)
		return h.SendErrorResponse(c, fiber.StatusInternalServerError, "DATABASE_ERROR", "Database error", err)
	}

	if dbConfig == nil {
		response := &dto.ChatwootStatusResponse{
			Enabled:   false,
			Connected: false,
		}
		return h.SendSuccessResponse(c, fiber.StatusOK, response)
	}

	_, serviceExists := h.chatwootIntegration.GetService(sessionID)

	response := &dto.ChatwootStatusResponse{
		Enabled:       dbConfig.Enabled,
		Connected:     serviceExists && dbConfig.Enabled && dbConfig.InboxID != nil,
		InboxName:     stringPtrToString(dbConfig.InboxName),
		MessagesCount: dbConfig.MessagesCount,
		ContactsCount: dbConfig.ContactsCount,
	}

	if dbConfig.InboxID != nil {
		response.InboxID = dbConfig.InboxID
	}

	if dbConfig.LastSync != nil {
		response.LastSync = dbConfig.LastSync.Format(time.RFC3339)
	}

	if dbConfig.ErrorMessage != nil {
		response.ErrorMessage = *dbConfig.ErrorMessage
	}

	return h.SendSuccessResponse(c, fiber.StatusOK, response)
}

// ReceiveChatwootWebhook recebe webhooks do Chatwoot
// @Summary Receive Chatwoot webhook
// @Description Receive and process webhooks from Chatwoot
// @Tags Chatwoot
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param payload body dto.ChatwootWebhookPayload true "Webhook payload"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /chatwoot/webhook/{sessionId} [post]
func (h *ChatwootHandler) ReceiveChatwootWebhook(c *fiber.Ctx) error {
	sessionIDOrName := c.Params("sessionId")
	if sessionIDOrName == "" {
		h.logger.Error("Session ID is required")
		return h.SendErrorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Session ID is required", nil)
	}

	// Decodifica o sessionID se necessário
	if decodedSessionID, err := url.QueryUnescape(sessionIDOrName); err == nil {
		sessionIDOrName = decodedSessionID
	}

	// Resolve sessionID ou nome para o UUID real
	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		h.logger.Errorf("Failed to resolve session ID %s: %v", sessionIDOrName, err)
		return h.SendErrorResponse(c, fiber.StatusNotFound, "SESSION_NOT_FOUND", "Session not found", err)
	}

	// Log do payload bruto para debug
	rawBody := c.Body()
	h.logger.Infof("Raw webhook payload: %s", string(rawBody))

	var payload dto.ChatwootWebhookPayload
	if err := c.BodyParser(&payload); err != nil {
		h.logger.Errorf("Failed to parse webhook payload: %v", err)
		return h.SendErrorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON payload", err)
	}

	h.logger.Infof("Received Chatwoot webhook for session %s: event=%s, messageType=%s, content=%s",
		sessionID, payload.Event, payload.MessageType, payload.Content)

	// Log detalhado dos campos importantes
	h.logger.Infof("Webhook details - Contact: %v, Conversation: %v",
		payload.Contact != nil, payload.Conversation != nil)

	if payload.Conversation != nil {
		h.logger.Infof("Conversation details - Contact: %v, Meta: %v",
			payload.Conversation.Contact != nil, payload.Conversation.Meta != nil)
		if payload.Conversation.Meta != nil {
			h.logger.Infof("Meta details - Sender: %v",
				payload.Conversation.Meta.Sender != nil)
			if payload.Conversation.Meta.Sender != nil {
				h.logger.Infof("Sender details - ID: %d, Phone: %s, Identifier: %s",
					payload.Conversation.Meta.Sender.ID,
					payload.Conversation.Meta.Sender.PhoneNumber,
					payload.Conversation.Meta.Sender.Identifier)
			}
		}
	}

	// Converte DTO para payload interno
	internalPayload := h.webhookDTOToInternal(&payload)

	// Processa webhook
	if err := h.chatwootIntegration.ProcessWebhook(c.Context(), sessionID, internalPayload); err != nil {
		h.logger.Errorf("Failed to process Chatwoot webhook: %v", err)
		return h.SendErrorResponse(c, fiber.StatusInternalServerError, "WEBHOOK_ERROR", "Failed to process webhook", err)
	}

	return h.SendSuccessResponse(c, fiber.StatusOK, map[string]string{"message": "Webhook processed successfully"})
}

// TestChatwootConnection testa a conexão com Chatwoot
// @Summary Test Chatwoot connection
// @Description Test connection to Chatwoot API
// @Tags Chatwoot
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body dto.ChatwootTestConnectionRequest true "Connection test parameters"
// @Success 200 {object} dto.ChatwootTestConnectionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /session/{sessionId}/chatwoot/test [post]
func (h *ChatwootHandler) TestChatwootConnection(c *fiber.Ctx) error {
	sessionID, valid := h.validateSessionID(c)
	if !valid {
		return nil
	}

	var req dto.ChatwootTestConnectionRequest
	if !h.bindAndValidateRequest(c, &req) {
		return nil
	}

	// Verifica se a sessão existe
	_, err := h.sessionService.GetSession(c.Context(), sessionID)
	if err != nil {
		return h.SendErrorResponse(c, fiber.StatusInternalServerError, "DATABASE_ERROR", "Database error", err)
	}

	// Cria cliente temporário para teste
	client := chatwoot.NewClient(req.URL, req.Token, req.AccountID)

	// Testa conexão listando inboxes
	inboxes, err := client.ListInboxes(c.Context())
	if err != nil {
		response := &dto.ChatwootTestConnectionResponse{
			Success: false,
			Message: "Failed to connect to Chatwoot",
			ErrorDetails: map[string]interface{}{
				"error": err.Error(),
			},
		}
		return h.SendSuccessResponse(c, fiber.StatusOK, response)
	}

	response := &dto.ChatwootTestConnectionResponse{
		Success: true,
		Message: "Successfully connected to Chatwoot",
		AccountInfo: &dto.ChatwootAccountInfo{
			ID:   mustParseInt(req.AccountID),
			Name: "Test Account", // Aqui você poderia buscar info real da conta
		},
		InboxesCount: len(inboxes),
	}

	return h.SendSuccessResponse(c, fiber.StatusOK, response)
}

// Métodos auxiliares

func (h *ChatwootHandler) bindAndValidateRequest(c *fiber.Ctx, req interface{}) bool {
	if err := h.BindAndValidate(c, req); err != nil {
		h.logger.Errorf("Failed to bind or validate request: %v", err)
		if sendErr := h.SendValidationErrorResponse(c, err); sendErr != nil {
			h.logger.Errorf("Failed to send validation error response: %v", sendErr)
		}
		return false
	}
	return true
}

func (h *ChatwootHandler) dtoToConfig(req *dto.ChatwootConfigRequest, sessionID string) *chatwoot.ChatwootConfig {
	config := &chatwoot.ChatwootConfig{
		Enabled:   req.Enabled != nil && *req.Enabled,
		AccountID: req.AccountID,
		Token:     req.Token,
		URL:       req.URL,
		NameInbox: req.NameInbox,
	}

	if config.NameInbox == "" {
		config.NameInbox = sessionID
	}

	if req.SignMsg != nil {
		config.SignMsg = *req.SignMsg
	}
	config.SignDelimiter = req.SignDelimiter
	config.Number = req.Number

	if req.ReopenConversation != nil {
		config.ReopenConversation = *req.ReopenConversation
	}
	if req.ConversationPending != nil {
		config.ConversationPending = *req.ConversationPending
	}
	if req.MergeBrazilContacts != nil {
		config.MergeBrazilContacts = *req.MergeBrazilContacts
	}
	if req.ImportContacts != nil {
		config.ImportContacts = *req.ImportContacts
	}
	if req.ImportMessages != nil {
		config.ImportMessages = *req.ImportMessages
	}
	if req.DaysLimitImportMessages != nil {
		config.DaysLimitImportMessages = *req.DaysLimitImportMessages
	}
	if req.AutoCreate != nil {
		config.AutoCreate = *req.AutoCreate
	}

	config.Organization = req.Organization
	config.Logo = req.Logo
	config.IgnoreJids = req.IgnoreJids

	return config
}

func (h *ChatwootHandler) configToDTO(config *chatwoot.ChatwootConfig, sessionID string, c *fiber.Ctx) *dto.ChatwootConfigResponse {
	baseURL := h.getBaseURL(c)
	h.logger.Infof("DEBUG: Base URL obtained: %s", baseURL)

	// Buscar o nome da sessão para usar no webhook (como faz a Evolution API)
	session, err := h.sessionService.GetSession(c.Context(), sessionID)
	sessionIdentifier := sessionID // fallback para o UUID se não conseguir buscar o nome
	if err == nil && session != nil {
		sessionIdentifier = session.Name().Value()
		h.logger.Infof("DEBUG: Using session name as identifier: %s", sessionIdentifier)
	} else {
		h.logger.Infof("DEBUG: Using session UUID as identifier: %s", sessionIdentifier)
	}

	return &dto.ChatwootConfigResponse{
		Enabled:                     config.Enabled,
		AccountID:                   config.AccountID,
		URL:                         config.URL,
		NameInbox:                   config.NameInbox,
		SignMsg:                     config.SignMsg,
		SignDelimiter:               config.SignDelimiter,
		Number:                      config.Number,
		ReopenConversation:          config.ReopenConversation,
		ConversationPending:         config.ConversationPending,
		MergeBrazilContacts:         config.MergeBrazilContacts,
		ImportContacts:              config.ImportContacts,
		ImportMessages:              config.ImportMessages,
		DaysLimitImportMessages:     config.DaysLimitImportMessages,
		AutoCreate:                  config.AutoCreate,
		Organization:                config.Organization,
		Logo:                        config.Logo,
		IgnoreJids:                  config.IgnoreJids,
		WebhookURL:                  func() string {
			webhookURL := fmt.Sprintf("%s/chatwoot/webhook/%s", baseURL, url.QueryEscape(sessionIdentifier))
			h.logger.Infof("DEBUG: Final webhook URL generated: %s", webhookURL)
			return webhookURL
		}(),
	}
}

func (h *ChatwootHandler) webhookDTOToInternal(dtoPayload *dto.ChatwootWebhookPayload) *chatwoot.WebhookPayload {
	payload := &chatwoot.WebhookPayload{
		Event:             dtoPayload.Event,
		ID:                dtoPayload.ID,
		Content:           dtoPayload.Content,
		Private:           dtoPayload.Private,
		SourceID:          dtoPayload.SourceID,
		ContentType:       dtoPayload.ContentType,
		ContentAttributes: dtoPayload.ContentAttributes,
	}

	// Converte MessageType (pode ser string ou int)
	if dtoPayload.MessageType != nil {
		switch v := dtoPayload.MessageType.(type) {
		case string:
			payload.MessageType = v
		case float64:
			if v == 0 {
				payload.MessageType = "incoming"
			} else if v == 1 {
				payload.MessageType = "outgoing"
			}
		case int:
			if v == 0 {
				payload.MessageType = "incoming"
			} else if v == 1 {
				payload.MessageType = "outgoing"
			}
		}
	}

	// Converte CreatedAt (pode ser string ou timestamp)
	if dtoPayload.CreatedAt != nil {
		switch v := dtoPayload.CreatedAt.(type) {
		case string:
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				payload.CreatedAt = t
			}
		case float64:
			payload.CreatedAt = time.Unix(int64(v), 0)
		case int64:
			payload.CreatedAt = time.Unix(v, 0)
		case int:
			payload.CreatedAt = time.Unix(int64(v), 0)
		}
	}

	// Extrai contato do campo direto, conversation.contact ou conversation.meta.sender
	var contact *dto.ChatwootContact
	if dtoPayload.Contact != nil {
		contact = dtoPayload.Contact
	} else if dtoPayload.Conversation != nil {
		if dtoPayload.Conversation.Contact != nil {
			contact = dtoPayload.Conversation.Contact
		} else if dtoPayload.Conversation.Meta != nil && dtoPayload.Conversation.Meta.Sender != nil {
			contact = dtoPayload.Conversation.Meta.Sender
		}
	}

	if contact != nil {
		payload.Contact = &chatwoot.Contact{
			ID:           contact.ID,
			Name:         contact.Name,
			Avatar:       contact.Avatar,
			AvatarURL:    contact.AvatarURL,
			PhoneNumber:  contact.PhoneNumber,
			Email:        contact.Email,
			Identifier:   contact.Identifier,
			Thumbnail:    contact.Thumbnail,
			CustomAttributes: contact.CustomAttributes,
		}
	}

	if dtoPayload.Conversation != nil {
		payload.Conversation = &chatwoot.Conversation{
			ID:                   dtoPayload.Conversation.ID,
			AccountID:            dtoPayload.Conversation.AccountID,
			InboxID:              dtoPayload.Conversation.InboxID,
			Status:               dtoPayload.Conversation.Status,
			Timestamp:            dtoPayload.Conversation.Timestamp,
			UnreadCount:          dtoPayload.Conversation.UnreadCount,
			AdditionalAttributes: dtoPayload.Conversation.AdditionalAttributes,
			CustomAttributes:     dtoPayload.Conversation.CustomAttributes,
		}
	}

	// Converte anexos
	for _, att := range dtoPayload.Attachments {
		payload.Attachments = append(payload.Attachments, chatwoot.Attachment{
			ID:        att.ID,
			MessageID: att.MessageID,
			FileType:  att.FileType,
			AccountID: att.AccountID,
			Extension: att.Extension,
			DataURL:   att.DataURL,
			ThumbURL:  att.ThumbURL,
			FileSize:  att.FileSize,
			Fallback:  att.Fallback,
		})
	}

	return payload
}

func (h *ChatwootHandler) getBaseURL(c *fiber.Ctx) string {
	// Use SERVER_HOST environment variable if set, otherwise fallback to request host
	if serverHost := os.Getenv("SERVER_HOST"); serverHost != "" {
		h.logger.Infof("DEBUG: SERVER_HOST found: %s", serverHost)
		scheme := "http"
		if strings.Contains(serverHost, "https://") {
			h.logger.Infof("DEBUG: Using HTTPS scheme from SERVER_HOST")
			return serverHost
		}
		if strings.Contains(serverHost, "://") {
			h.logger.Infof("DEBUG: Using scheme from SERVER_HOST: %s", serverHost)
			return serverHost
		}
		finalURL := fmt.Sprintf("%s://%s", scheme, serverHost)
		h.logger.Infof("DEBUG: Generated base URL: %s", finalURL)
		return finalURL
	}

	// Fallback to request host
	h.logger.Warnf("DEBUG: SERVER_HOST not found, using request host fallback")
	scheme := "http"
	if c.Protocol() == "https" {
		scheme = "https"
	}

	if forwarded := c.Get("X-Forwarded-Proto"); forwarded != "" {
		scheme = forwarded
	}

	host := c.Hostname()
	if forwarded := c.Get("X-Forwarded-Host"); forwarded != "" {
		host = forwarded
	}

	fallbackURL := fmt.Sprintf("%s://%s", scheme, host)
	h.logger.Infof("DEBUG: Fallback URL generated: %s", fallbackURL)
	return fallbackURL
}

func mustParseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}

// configToDBModel converte configuração para modelo do banco
func (h *ChatwootHandler) configToDBModel(config *chatwoot.ChatwootConfig, sessionID string) *models.ChatwootModel {
	model := &models.ChatwootModel{
		SessionID:               sessionID,
		Enabled:                 config.Enabled,
		SignMsg:                 config.SignMsg,
		SignDelimiter:           config.SignDelimiter,
		Number:                  config.Number,
		ReopenConversation:      config.ReopenConversation,
		ConversationPending:     config.ConversationPending,
		MergeBrazilContacts:     config.MergeBrazilContacts,
		ImportContacts:          config.ImportContacts,
		ImportMessages:          config.ImportMessages,
		DaysLimitImportMessages: config.DaysLimitImportMessages,
		AutoCreate:              config.AutoCreate,
		Organization:            config.Organization,
		Logo:                    config.Logo,
		IgnoreJids:              models.StringArray(config.IgnoreJids),
		SyncStatus:              "pending",
	}

	// Campos opcionais (apenas se habilitado)
	if config.Enabled {
		model.AccountID = &config.AccountID
		model.Token = &config.Token
		model.URL = &config.URL
		model.NameInbox = &config.NameInbox
	}

	return model
}

// dbModelToConfig converte modelo do banco para configuração
func (h *ChatwootHandler) dbModelToConfig(model *models.ChatwootModel) *chatwoot.ChatwootConfig {
	config := &chatwoot.ChatwootConfig{
		Enabled:                 model.Enabled,
		SignMsg:                 model.SignMsg,
		SignDelimiter:           model.SignDelimiter,
		Number:                  model.Number,
		ReopenConversation:      model.ReopenConversation,
		ConversationPending:     model.ConversationPending,
		MergeBrazilContacts:     model.MergeBrazilContacts,
		ImportContacts:          model.ImportContacts,
		ImportMessages:          model.ImportMessages,
		DaysLimitImportMessages: model.DaysLimitImportMessages,
		AutoCreate:              model.AutoCreate,
		Organization:            model.Organization,
		Logo:                    model.Logo,
		IgnoreJids:              []string(model.IgnoreJids),
	}

	// Campos opcionais
	if model.AccountID != nil {
		config.AccountID = *model.AccountID
	}
	if model.Token != nil {
		config.Token = *model.Token
	}
	if model.URL != nil {
		config.URL = *model.URL
	}
	if model.NameInbox != nil {
		config.NameInbox = *model.NameInbox
	}

	return config
}

// stringPtrToString converte *string para string
func stringPtrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
