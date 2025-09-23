package chatwoot

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// Handler representa o handler HTTP para webhooks do Chatwoot
type Handler struct {
	service *Service
	logger  *slog.Logger
}

// NewHandler cria uma nova instância do handler
func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes registra as rotas do Chatwoot
func (h *Handler) RegisterRoutes(router fiber.Router) {
	chatwootGroup := router.Group("/chatwoot")
	chatwootGroup.Post("/webhook/:instanceName", h.HandleWebhook)
	chatwootGroup.Post("/config/:instanceName", h.CreateConfig)
	chatwootGroup.Get("/config/:instanceName", h.GetConfig)
	chatwootGroup.Put("/config/:instanceName", h.UpdateConfig)
	chatwootGroup.Delete("/config/:instanceName", h.DeleteConfig)
}

// HandleWebhook processa webhooks recebidos do Chatwoot
func (h *Handler) HandleWebhook(c *fiber.Ctx) error {
	instanceName := c.Params("instanceName")

	h.logger.Info("Received Chatwoot webhook", "instance", instanceName)

	var payload WebhookPayload
	if err := c.BodyParser(&payload); err != nil {
		h.logger.Error("Failed to parse webhook payload", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON payload",
		})
	}

	// Log do payload para debug
	conversationID := 0
	if payload.Conversation != nil {
		if id, ok := payload.Conversation["id"].(float64); ok {
			conversationID = int(id)
		}
	}

	messageType := ""
	if msgType, ok := payload.Message["message_type"].(string); ok {
		messageType = msgType
	}

	content := ""
	if contentVal, ok := payload.Message["content"].(string); ok {
		content = contentVal
	}

	h.logger.Debug("Webhook payload",
		"event", payload.Event,
		"messageType", messageType,
		"content", content,
		"conversationID", conversationID)

	// Processa o webhook
	if err := h.service.ProcessWebhook(c.Context(), &payload); err != nil {
		h.logger.Error("Failed to process webhook", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process webhook",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Webhook processed successfully",
	})
}

// CreateConfig cria uma nova configuração Chatwoot para uma instância
func (h *Handler) CreateConfig(c *fiber.Ctx) error {
	instanceName := c.Params("instanceName")

	var config ChatwootConfig
	if err := c.BodyParser(&config); err != nil {
		h.logger.Error("Failed to parse config", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON payload",
		})
	}

	// Validações básicas
	if config.IsActive {
		if config.URL == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "URL is required when Chatwoot is active",
			})
		}

		if config.AccountID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Account ID is required when Chatwoot is enabled",
			})
		}

		if config.Token == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Token is required when Chatwoot is enabled",
			})
		}

		// Valida URL
		if !isValidURL(config.URL) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid URL format",
			})
		}
	}

	// Define nome da inbox se não fornecido
	if config.NameInbox == "" {
		config.NameInbox = instanceName
	}

	// Aqui você salvaria a configuração no banco de dados
	// Por exemplo: h.configRepository.Save(instanceName, &config)

	h.logger.Info("Chatwoot config created", "instance", instanceName, "isActive", config.IsActive)

	// Retorna a configuração com URL do webhook
	response := fiber.Map{
		"isActive":                config.IsActive,
		"accountId":               config.AccountID,
		"url":                     config.URL,
		"nameInbox":               config.NameInbox,
		"signMsg":                 config.SignMsg,
		"signDelimiter":           config.SignDelimiter,
		"reopenConversation":      config.ReopenConversation,
		"conversationPending":     config.ConversationPending,
		"mergeBrazilContacts":     config.MergeBrazilContacts,
		"importContacts":          config.ImportContacts,
		"importMessages":          config.ImportMessages,
		"daysLimitImportMessages": config.DaysLimitImportMessages,
		"autoCreate":              config.AutoCreate,
		"organization":            config.Organization,
		"logo":                    config.Logo,
		"ignoreJids":              config.IgnoreJids,
		"webhook_url":             fmt.Sprintf("%s/chatwoot/webhook/%s", getBaseURL(c), instanceName),
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

// GetConfig retorna a configuração Chatwoot de uma instância
func (h *Handler) GetConfig(c *fiber.Ctx) error {
	instanceName := c.Params("instanceName")

	// Aqui você buscaria a configuração no banco de dados
	// Por exemplo: config, err := h.configRepository.FindByInstance(instanceName)

	// Por enquanto, retorna uma configuração padrão
	config := ChatwootConfig{
		IsActive: false,
		URL:      "",
	}

	response := fiber.Map{
		"isActive":                config.IsActive,
		"accountId":               config.AccountID,
		"url":                     config.URL,
		"nameInbox":               config.NameInbox,
		"signMsg":                 config.SignMsg,
		"signDelimiter":           config.SignDelimiter,
		"reopenConversation":      config.ReopenConversation,
		"conversationPending":     config.ConversationPending,
		"mergeBrazilContacts":     config.MergeBrazilContacts,
		"importContacts":          config.ImportContacts,
		"importMessages":          config.ImportMessages,
		"daysLimitImportMessages": config.DaysLimitImportMessages,
		"autoCreate":              config.AutoCreate,
		"organization":            config.Organization,
		"logo":                    config.Logo,
		"ignoreJids":              config.IgnoreJids,
		"webhook_url":             fmt.Sprintf("%s/chatwoot/webhook/%s", getBaseURL(c), instanceName),
	}

	return c.JSON(response)
}

// UpdateConfig atualiza a configuração Chatwoot de uma instância
func (h *Handler) UpdateConfig(c *fiber.Ctx) error {
	instanceName := c.Params("instanceName")

	var config ChatwootConfig
	if err := c.BodyParser(&config); err != nil {
		h.logger.Error("Failed to parse config", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON payload",
		})
	}

	// Validações básicas (mesmo que CreateConfig)
	if config.IsActive {
		if config.URL == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "URL is required when Chatwoot is active",
			})
		}

		if config.AccountID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Account ID is required when Chatwoot is enabled",
			})
		}

		if config.Token == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Token is required when Chatwoot is enabled",
			})
		}

		if !isValidURL(config.URL) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid URL format",
			})
		}
	}

	// Define nome da inbox se não fornecido
	if config.NameInbox == "" {
		config.NameInbox = instanceName
	}

	// Aqui você atualizaria a configuração no banco de dados
	// Por exemplo: h.configRepository.Update(instanceName, &config)

	h.logger.Info("Chatwoot config updated", "instance", instanceName, "isActive", config.IsActive)

	response := fiber.Map{
		"isActive":                config.IsActive,
		"accountId":               config.AccountID,
		"url":                     config.URL,
		"nameInbox":               config.NameInbox,
		"signMsg":                 config.SignMsg,
		"signDelimiter":           config.SignDelimiter,
		"reopenConversation":      config.ReopenConversation,
		"conversationPending":     config.ConversationPending,
		"mergeBrazilContacts":     config.MergeBrazilContacts,
		"importContacts":          config.ImportContacts,
		"importMessages":          config.ImportMessages,
		"daysLimitImportMessages": config.DaysLimitImportMessages,
		"autoCreate":              config.AutoCreate,
		"organization":            config.Organization,
		"logo":                    config.Logo,
		"ignoreJids":              config.IgnoreJids,
		"webhook_url":             fmt.Sprintf("%s/chatwoot/webhook/%s", getBaseURL(c), instanceName),
	}

	return c.JSON(response)
}

// DeleteConfig remove a configuração Chatwoot de uma instância
func (h *Handler) DeleteConfig(c *fiber.Ctx) error {
	instanceName := c.Params("instanceName")

	// Aqui você removeria a configuração do banco de dados
	// Por exemplo: h.configRepository.Delete(instanceName)

	h.logger.Info("Chatwoot config deleted", "instance", instanceName)

	return c.JSON(fiber.Map{
		"message": "Chatwoot configuration deleted successfully",
	})
}

// isValidURL valida se uma URL é válida
func isValidURL(url string) bool {
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

// getBaseURL extrai a URL base da requisição
func getBaseURL(c *fiber.Ctx) string {
	scheme := "http"
	if c.Protocol() == "https" {
		scheme = "https"
	}

	// Verifica headers de proxy
	if forwarded := c.Get("X-Forwarded-Proto"); forwarded != "" {
		scheme = forwarded
	}

	host := c.Hostname()
	if forwarded := c.Get("X-Forwarded-Host"); forwarded != "" {
		host = forwarded
	}

	return fmt.Sprintf("%s://%s", scheme, host)
}

// HealthCheck endpoint para verificar saúde da integração
func (h *Handler) HealthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "ok",
		"service": "chatwoot-integration",
	})
}
