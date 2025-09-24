package handlers

import (
	"encoding/json"
	"fmt"
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

// validateSessionId valida e retorna o sessionID do parâmetro
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

// resolveSessionId resolve o sessionID ou nome para o ID real da sessão
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
// @Description Configure Chatwoot integration for a WhatsApp session. This endpoint allows you to set up the connection between your WhatsApp session and Chatwoot. Required fields when isActive=true: accountId, token, url. Optional fields include nameInbox, signMsg, signDelimiter, number, reopenConversation, conversationPending, mergeBrazilContacts, importContacts, importMessages, daysLimitImportMessages, autoCreate, organization, logo, ignoreJids.
// @Tags Chatwoot
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID or name" example("my-session")
// @Param request body dto.ChatwootConfigRequest true "Chatwoot configuration request. Required: isActive (boolean). When isActive=true, also required: accountId (string), token (string), url (string). Optional: nameInbox, signMsg, signDelimiter, number, reopenConversation, conversationPending, mergeBrazilContacts, importContacts, importMessages, daysLimitImportMessages, autoCreate, organization, logo, ignoreJids."
// @Success 200 {object} dto.ChatwootConfigResponse "Successfully configured Chatwoot integration"
// @Failure 400 {object} dto.StandardErrorResponse "Bad request - validation errors or missing required fields"
// @Failure 401 {object} dto.StandardErrorResponse "Unauthorized - API key required (use global or session-specific key)"
// @Failure 404 {object} dto.StandardErrorResponse "Session not found"
// @Failure 500 {object} dto.StandardErrorResponse "Internal server error"
// @Router /session/{sessionId}/chatwoot/set [post]
func (h *ChatwootHandler) SetChatwootConfig(c *fiber.Ctx) error {
	sessionIDOrName, valid := h.validateSessionID(c)
	if !valid {
		return nil
	}

	// Resolve sessionID ou nome para o UUID real
	sessionID, err := h.resolveSessionID(c, sessionIDOrName)
	if err != nil {
		h.logger.Errorf("Failed to resolve session ID %s: %v", sessionIDOrName, err)
		return h.SendErrorResponse(c, fiber.StatusNotFound, "SESSION_NOT_FOUND", "Session not found", err)
	}

	var req dto.ChatwootConfigRequest
	if !h.bindAndValidateRequest(c, &req) {
		return nil
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
// @Description Get current Chatwoot configuration for a session. Returns the real configuration from database if configured, or a generic "not configured" response if no configuration exists. Accepts both global API key and session-specific API key for authentication.
// @Tags Chatwoot
// @Produce json
// @Security ApiKeyAuth
// @Param sessionId path string true "Session ID or name" example("my-session")
// @Success 200 {object} dto.ChatwootConfigResponse "Configuration found (if configured=true) or not configured message (if configured=false)"
// @Failure 401 {object} dto.StandardErrorResponse "Unauthorized - API key required (use global or session-specific key)"
// @Failure 404 {object} dto.StandardErrorResponse "Session not found"
// @Failure 500 {object} dto.StandardErrorResponse "Internal server error"
// @Router /session/{sessionId}/chatwoot/find [get]
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

	if dbConfig == nil {
		// Retorna resposta genérica quando não há configuração
		return h.SendSuccessResponse(c, fiber.StatusOK, map[string]interface{}{
			"configured": false,
			"message":    "Chatwoot integration not configured for this session",
		})
	}

	// Converte do modelo do banco para configuração
	config := h.dbModelToConfig(dbConfig)
	response := h.configToDTO(config, sessionIDOrName, c)

	// Adiciona flag indicando que está configurado
	responseMap := map[string]interface{}{
		"configured": true,
		"config":     response,
	}

	return h.SendSuccessResponse(c, fiber.StatusOK, responseMap)
}

// UpdateChatwootConfig atualiza a configuração Chatwoot de uma sessão
// FUNÇÃO NÃO UTILIZADA - SEM ROTA CORRESPONDENTE
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
// FUNÇÃO NÃO UTILIZADA - SEM ROTA CORRESPONDENTE
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
// FUNÇÃO NÃO UTILIZADA - SEM ROTA CORRESPONDENTE
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
		Enabled:       dbConfig.IsActive,
		Connected:     serviceExists && dbConfig.IsActive && dbConfig.InboxId != nil,
		InboxName:     stringPtrToString(dbConfig.NameInbox),
		MessagesCount: 0, /* TODO: calcular dinamicamente */
		ContactsCount: 0, /* TODO: calcular dinamicamente */
	}

	if dbConfig.InboxId != nil {
		response.InboxID = dbConfig.InboxId
	}

	if dbConfig.LastSync != nil {
		response.LastSync = dbConfig.LastSync.Format(time.RFC3339)
	}

	// TODO: implementar ErrorMessage usando metadata
	// if dbConfig.ErrorMessage != nil {
	//     response.ErrorMessage = *dbConfig.ErrorMessage
	// }

	return h.SendSuccessResponse(c, fiber.StatusOK, response)
}

// ReceiveChatwootWebhook recebe webhooks do Chatwoot (interno, não documentado no swagger)
func (h *ChatwootHandler) ReceiveChatwootWebhook(c *fiber.Ctx) error {
	sessionIDOrName := c.Params("sessionId")
	if sessionIDOrName == "" {
		h.logger.Error("Session ID is required")
		return h.SendErrorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Session ID is required", nil)
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

	h.logger.Infof("Received Chatwoot webhook for session %s: event=%s", sessionID, payload.Event)

	// Converte DTO para payload interno
	internalPayload := h.webhookDTOToInternal(&payload)

	// Processa webhook
	if err := h.chatwootIntegration.ProcessWebhook(c.Context(), sessionID, internalPayload); err != nil {
		h.logger.Errorf("Failed to process Chatwoot webhook: %v", err)
		return h.SendErrorResponse(c, fiber.StatusInternalServerError, "WEBHOOK_ERROR", "Failed to process webhook", err)
	}

	return h.SendSuccessResponse(c, fiber.StatusOK, map[string]string{"message": "Webhook processed successfully"})
}

// webhookDTOToInternal converte DTO do webhook para payload interno
func (h *ChatwootHandler) webhookDTOToInternal(dtoPayload *dto.ChatwootWebhookPayload) *chatwoot.WebhookPayload {
	payload := &chatwoot.WebhookPayload{
		Event:   dtoPayload.Event,
		Message: make(map[string]interface{}),
	}

	// Popula o map Message com os dados do DTO
	if dtoPayload.ID != 0 {
		payload.Message["id"] = dtoPayload.ID
	}
	if dtoPayload.Content != "" {
		payload.Message["content"] = dtoPayload.Content
	}
	payload.Message["private"] = dtoPayload.Private
	if dtoPayload.SourceID != "" {
		payload.Message["source_id"] = dtoPayload.SourceID
	}
	if dtoPayload.ContentType != "" {
		payload.Message["content_type"] = dtoPayload.ContentType
	}
	if dtoPayload.ContentAttributes != nil {
		payload.Message["content_attributes"] = dtoPayload.ContentAttributes
	}

	// Converte MsgType (pode ser string ou int)
	if dtoPayload.MsgType != nil {
		switch v := dtoPayload.MsgType.(type) {
		case string:
			payload.Message["message_type"] = v
		case float64:
			switch v {
			case 0:
				payload.Message["message_type"] = "incoming"
			case 1:
				payload.Message["message_type"] = "outgoing"
			}
		case int:
			switch v {
			case 0:
				payload.Message["message_type"] = "incoming"
			case 1:
				payload.Message["message_type"] = "outgoing"
			}
		}
	}

	// Converte CreatedAt (pode ser string ou timestamp)
	if dtoPayload.CreatedAt != nil {
		switch v := dtoPayload.CreatedAt.(type) {
		case string:
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				payload.Message["created_at"] = t
			}
		case float64:
			payload.Message["created_at"] = time.Unix(int64(v), 0)
		case int64:
			payload.Message["created_at"] = time.Unix(v, 0)
		case int:
			payload.Message["created_at"] = time.Unix(int64(v), 0)
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
		payload.Contact = make(map[string]interface{})
		payload.Contact["id"] = contact.ID
		payload.Contact["name"] = contact.Name
		payload.Contact["phone_number"] = contact.PhoneNumber
		payload.Contact["email"] = contact.Email
		payload.Contact["identifier"] = contact.Identifier
	}

	if dtoPayload.Conversation != nil {
		payload.Conversation = make(map[string]interface{})
		payload.Conversation["id"] = dtoPayload.Conversation.ID
		payload.Conversation["inbox_id"] = dtoPayload.Conversation.InboxID
		payload.Conversation["status"] = dtoPayload.Conversation.Status
	}

	// Converte anexos para o map Message
	if len(dtoPayload.Attachments) > 0 {
		attachments := make([]interface{}, 0, len(dtoPayload.Attachments))
		for _, att := range dtoPayload.Attachments {
			attachment := map[string]interface{}{
				"id":        att.ID,
				"file_type": att.FileType,
				"data_url":  att.DataURL,
				"file_size": int(att.FileSize), // Convert int64 to int
				"fallback":  att.Fallback,
			}
			attachments = append(attachments, attachment)
		}
		payload.Message["attachments"] = attachments
	}

	return payload
}

// TestChatwootConnection testa a conexão com Chatwoot
// FUNÇÃO NÃO UTILIZADA - SEM ROTA CORRESPONDENTE
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
	client := chatwoot.NewClient(req.URL, req.Token, req.AccountID, nil)

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
		IsActive:  req.IsActive != nil && *req.IsActive,
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
		Enabled:                 config.IsActive,
		AccountID:               config.AccountID,
		URL:                     config.URL,
		NameInbox:               config.NameInbox,
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
		IgnoreJids:              config.IgnoreJids,
	}
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
		SessionId:  sessionID,
		IsActive:   config.IsActive,
		SyncStatus: "pending",
	}

	// Campos opcionais (apenas se habilitado)
	if config.IsActive {
		model.AccountId = &config.AccountID
		model.Token = &config.Token
		model.URL = &config.URL
		model.NameInbox = &config.NameInbox
	}

	// Criar configurações JSONB vazio por enquanto para testar
	model.Config = models.JSONB{}

	return model
}

// dbModelToConfig converte modelo do banco para configuração
func (h *ChatwootHandler) dbModelToConfig(model *models.ChatwootModel) *chatwoot.ChatwootConfig {
	config := &chatwoot.ChatwootConfig{
		IsActive: model.IsActive,
	}

	// Campos opcionais
	if model.AccountId != nil {
		config.AccountID = *model.AccountId
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
	// Extrair configurações do campo Config JSONB
	if len(model.Config) > 0 {
		var configData map[string]interface{}
		configBytes, err := json.Marshal(model.Config)
		if err == nil {
			if err := json.Unmarshal(configBytes, &configData); err == nil {
				// Mapear campos do JSONB para a configuração
				if signMsg, ok := configData["signMsg"].(bool); ok {
					config.SignMsg = signMsg
				}
				if signDelimiter, ok := configData["signDelimiter"].(string); ok {
					config.SignDelimiter = signDelimiter
				}
				if reopenConversation, ok := configData["reopenConversation"].(bool); ok {
					config.ReopenConversation = reopenConversation
				}
				if conversationPending, ok := configData["conversationPending"].(bool); ok {
					config.ConversationPending = conversationPending
				}
				if mergeBrazilContacts, ok := configData["mergeBrazilContacts"].(bool); ok {
					config.MergeBrazilContacts = mergeBrazilContacts
				}
				if importContacts, ok := configData["importContacts"].(bool); ok {
					config.ImportContacts = importContacts
				}
				if importMessages, ok := configData["importMessages"].(bool); ok {
					config.ImportMessages = importMessages
				}
				if daysLimit, ok := configData["daysLimitImportMessages"].(float64); ok {
					config.DaysLimitImportMessages = int(daysLimit)
				}
				if autoCreate, ok := configData["autoCreate"].(bool); ok {
					config.AutoCreate = autoCreate
				}
				if organization, ok := configData["organization"].(string); ok {
					config.Organization = organization
				}
				if logo, ok := configData["logo"].(string); ok {
					config.Logo = logo
				}
				if ignoreJids, ok := configData["ignoreJids"].([]interface{}); ok {
					config.IgnoreJids = make([]string, len(ignoreJids))
					for i, jid := range ignoreJids {
						if jidStr, ok := jid.(string); ok {
							config.IgnoreJids[i] = jidStr
						}
					}
				}
			}
		}
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
