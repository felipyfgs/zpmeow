package chatwoot

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"zpmeow/internal/application/ports"
)

// InboxService gerencia operações relacionadas a inboxes
type InboxService struct {
	client       *Client
	logger       *slog.Logger
	errorHandler ports.ChatwootErrorHandler
	validator    ports.ChatwootValidator
	sessionID    string
	adapter      *InboxAdapter
}

// NewInboxService cria um novo serviço de inbox
func NewInboxService(client *Client, logger *slog.Logger, sessionID string) *InboxService {
	return &InboxService{
		client:       client,
		logger:       logger,
		errorHandler: NewErrorHandler(),
		validator:    NewValidator(),
		sessionID:    sessionID,
		adapter:      NewInboxAdapter(),
	}
}

// InitializeInbox inicializa ou encontra a inbox configurada
func (is *InboxService) InitializeInbox(ctx context.Context, config *ports.ChatwootConfig) (*ports.InboxResponse, error) {
	// Valida a configuração
	if config == nil {
		return nil, fmt.Errorf("chatwoot config is nil")
	}

	if config.NameInbox == "" {
		return nil, fmt.Errorf("inbox name is required")
	}

	// Lista inboxes existentes
	inboxes, err := is.client.ListInboxes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list inboxes: %w", err)
	}

	is.logger.Info("Listed existing inboxes", "count", len(inboxes), "target_name", config.NameInbox)

	// Procura por uma inbox existente com o nome configurado
	for _, inbox := range inboxes {
		if inbox.Name == config.NameInbox {
			is.logger.Info("Found existing inbox", "name", inbox.Name, "id", inbox.ID)

			// Converte para tipo da interface
			inboxResponse := is.adapter.ToPortsInbox(&inbox)

			// Valida a inbox encontrada
			if err := is.validateInboxResponse(inboxResponse); err != nil {
				is.logger.Warn("Found inbox has validation issues", "error", err)
				continue
			}

			return inboxResponse, nil
		}
	}

	// Se não encontrou, cria uma nova inbox se autoCreate estiver habilitado
	if config.AutoCreate {
		return is.createInbox(ctx, config.NameInbox)
	}

	return nil, fmt.Errorf("inbox '%s' not found and auto-create is disabled", config.NameInbox)
}

// createInbox cria uma nova inbox
func (is *InboxService) createInbox(ctx context.Context, name string) (*ports.InboxResponse, error) {
	webhookURL := is.generateWebhookURL()

	req := InboxCreateRequest{
		Name: name,
		Channel: map[string]interface{}{
			"type":        "api",
			"webhook_url": webhookURL,
		},
	}

	is.logger.Info("Creating new inbox", "name", name, "webhook_url", webhookURL)
	inbox, err := is.client.CreateInbox(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create inbox: %w", err)
	}

	// Converte para tipo da interface
	inboxResponse := is.adapter.ToPortsInbox(inbox)

	// Valida a inbox criada
	if err := is.validateInboxResponse(inboxResponse); err != nil {
		return nil, fmt.Errorf("invalid inbox response: %w", err)
	}

	is.logger.Info("Successfully created inbox", "name", inbox.Name, "id", inbox.ID, "webhook", webhookURL)
	return inboxResponse, nil
}

// generateWebhookURL gera a URL do webhook para a inbox
func (is *InboxService) generateWebhookURL() string {
	// Tenta obter do ambiente primeiro
	if serverHost := os.Getenv("SERVER_HOST"); serverHost != "" {
		webhookURL := fmt.Sprintf("%s/chatwoot/webhook/%s", serverHost, is.sessionID)
		is.logger.Info("Generated webhook URL from SERVER_HOST", "webhook", webhookURL, "server_host", serverHost)
		return webhookURL
	}

	// Fallback para URL padrão
	webhookURL := fmt.Sprintf("http://localhost:8080/chatwoot/webhook/%s", is.sessionID)
	is.logger.Info("Generated default webhook URL", "webhook", webhookURL)
	return webhookURL
}

// validateInboxResponse valida se uma inbox está configurada corretamente
func (is *InboxService) validateInboxResponse(inbox *ports.InboxResponse) error {
	if inbox == nil {
		return fmt.Errorf("inbox is nil")
	}

	if inbox.ID == 0 {
		return fmt.Errorf("inbox ID is invalid")
	}

	if inbox.Name == "" {
		return fmt.Errorf("inbox name is empty")
	}

	return nil
}

// GetInboxByID busca uma inbox pelo ID
func (is *InboxService) GetInboxByID(ctx context.Context, inboxID int) (*ports.InboxResponse, error) {
	is.logger.Info("Getting inbox by ID", "inbox_id", inboxID)

	// Lista todas as inboxes e procura pelo ID
	inboxes, err := is.client.ListInboxes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list inboxes: %w", err)
	}

	for _, inbox := range inboxes {
		if inbox.ID == inboxID {
			// Converte para tipo da interface
			inboxResponse := is.adapter.ToPortsInbox(&inbox)

			// Valida a inbox
			if err := is.validateInboxResponse(inboxResponse); err != nil {
				return nil, fmt.Errorf("invalid inbox response: %w", err)
			}

			return inboxResponse, nil
		}
	}

	return nil, fmt.Errorf("inbox with ID %d not found", inboxID)
}

// FindInboxByName encontra uma inbox pelo nome
func (is *InboxService) FindInboxByName(ctx context.Context, name string) (*ports.InboxResponse, error) {
	is.logger.Info("Finding inbox by name", "name", name)

	inboxes, err := is.ListInboxes(ctx)
	if err != nil {
		return nil, err
	}

	for _, inbox := range inboxes {
		if inbox.Name == name {
			is.logger.Info("Found inbox by name", "name", name, "id", inbox.ID)
			return inbox, nil
		}
	}

	return nil, fmt.Errorf("inbox with name '%s' not found", name)
}

// ListInboxes lista todas as inboxes disponíveis
func (is *InboxService) ListInboxes(ctx context.Context) ([]*ports.InboxResponse, error) {
	inboxes, err := is.client.ListInboxes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list inboxes: %w", err)
	}

	// Converte para tipos da interface
	inboxResponses := is.adapter.ToPortsInboxList(inboxes)

	is.logger.Info("Listed inboxes", "count", len(inboxResponses))
	return inboxResponses, nil
}

// UpdateInboxWebhook atualiza a URL do webhook de uma inbox
func (is *InboxService) UpdateInboxWebhook(ctx context.Context, inboxID int) error {
	webhookURL := is.generateWebhookURL()

	is.logger.Info("Would update inbox webhook", "inbox_id", inboxID, "webhook_url", webhookURL)

	// Esta funcionalidade pode ser implementada futuramente
	// Por enquanto, apenas logamos a intenção
	return nil
}

// GetInboxStats retorna estatísticas da inbox
func (is *InboxService) GetInboxStats(ctx context.Context, inboxID int) (*InboxStats, error) {
	is.logger.Info("Getting inbox stats", "inbox_id", inboxID)

	// Esta funcionalidade pode ser implementada futuramente
	// Por enquanto, retorna estatísticas básicas
	stats := &InboxStats{
		InboxID:               inboxID,
		TotalConversations:    0,
		ActiveConversations:   0,
		PendingConversations:  0,
		ResolvedConversations: 0,
	}

	is.logger.Info("Retrieved inbox stats", "inbox_id", inboxID, "stats", stats)
	return stats, nil
}

// InboxStats representa estatísticas de uma inbox
type InboxStats struct {
	InboxID               int `json:"inbox_id"`
	TotalConversations    int `json:"total_conversations"`
	ActiveConversations   int `json:"active_conversations"`
	PendingConversations  int `json:"pending_conversations"`
	ResolvedConversations int `json:"resolved_conversations"`
}

// InboxHealth verifica a saúde de uma inbox
func (is *InboxService) InboxHealth(ctx context.Context, inbox *ports.InboxResponse) *InboxHealthStatus {
	status := &InboxHealthStatus{
		InboxID:   inbox.ID,
		Name:      inbox.Name,
		IsHealthy: true,
		Issues:    []string{},
	}

	// Verifica se a inbox existe
	if err := is.validateInboxResponse(inbox); err != nil {
		status.IsHealthy = false
		status.Issues = append(status.Issues, fmt.Sprintf("Inbox validation failed: %v", err))
	}

	// Verifica webhook URL
	if inbox.WebhookURL == "" {
		status.IsHealthy = false
		status.Issues = append(status.Issues, "Webhook URL is not configured")
	}

	// Verifica se a inbox ainda existe no Chatwoot
	_, err := is.GetInboxByID(context.Background(), inbox.ID)
	if err != nil {
		status.IsHealthy = false
		status.Issues = append(status.Issues, fmt.Sprintf("Inbox not found in Chatwoot: %v", err))
	}

	is.logger.Info("Inbox health check completed",
		"inbox_id", inbox.ID,
		"is_healthy", status.IsHealthy,
		"issues_count", len(status.Issues))

	return status
}

// InboxHealthStatus representa o status de saúde de uma inbox
type InboxHealthStatus struct {
	InboxID   int      `json:"inbox_id"`
	Name      string   `json:"name"`
	IsHealthy bool     `json:"is_healthy"`
	Issues    []string `json:"issues"`
}

// DeleteInbox deleta uma inbox (se suportado pela API)
func (is *InboxService) DeleteInbox(ctx context.Context, inboxID int) error {
	is.logger.Info("Would delete inbox", "inbox_id", inboxID)

	// Esta funcionalidade pode ser implementada futuramente
	// Por enquanto, apenas logamos a intenção
	return fmt.Errorf("inbox deletion not implemented")
}

// ValidateInboxConfiguration valida a configuração de uma inbox
func (is *InboxService) ValidateInboxConfiguration(ctx context.Context, config *ports.ChatwootConfig) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}

	if !config.IsActive {
		return fmt.Errorf("chatwoot integration is disabled")
	}

	if config.NameInbox == "" {
		return fmt.Errorf("inbox name is required")
	}

	if len(config.NameInbox) > 100 {
		return fmt.Errorf("inbox name is too long (max 100 characters)")
	}

	// Valida URL se fornecida
	if config.URL != "" {
		if err := is.validator.ValidateURL(config.URL); err != nil {
			return fmt.Errorf("invalid URL: %w", err)
		}
	}

	// Valida token se fornecido
	if config.Token != "" {
		if err := is.validator.ValidateToken(config.Token); err != nil {
			return fmt.Errorf("invalid token: %w", err)
		}
	}

	// Valida account ID se fornecido
	if config.AccountID > 0 {
		if err := is.validator.ValidateAccountID(config.AccountID); err != nil {
			return fmt.Errorf("invalid account ID: %w", err)
		}
	}

	is.logger.Info("Inbox configuration validation passed", "inbox_name", config.NameInbox)
	return nil
}
