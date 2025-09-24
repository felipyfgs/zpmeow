package chatwoot

import (
	"context"
	"fmt"
	"log/slog"

	"zpmeow/internal/application/ports"
)

// ConversationService gerencia operações relacionadas a conversas
type ConversationService struct {
	client       *Client
	logger       *slog.Logger
	cacheManager ports.ChatwootCacheManager
	errorHandler ports.ChatwootErrorHandler
	validator    ports.ChatwootValidator

	adapter *ConversationAdapter
}

// NewConversationService cria um novo serviço de conversas
func NewConversationService(client *Client, logger *slog.Logger, cacheManager ports.ChatwootCacheManager) *ConversationService {
	return &ConversationService{
		client:       client,
		logger:       logger,
		cacheManager: cacheManager,
		errorHandler: NewErrorHandler(),
		validator:    NewValidator(),
		adapter:      NewConversationAdapter(),
	}
}

// GetOrCreateConversation obtém ou cria uma conversa
func (cs *ConversationService) GetOrCreateConversation(ctx context.Context, contactID int, inboxID int) (*ports.ConversationResponse, error) {
	// Verifica cache primeiro
	if conversation, found := cs.cacheManager.GetConversation(contactID); found {
		cs.logger.Info("Conversation found in cache", "contact_id", contactID, "conversation_id", conversation.ID)
		return conversation, nil
	}

	// Busca conversa ativa existente
	conversation, err := cs.findActiveConversation(ctx, contactID, inboxID)
	if err != nil {
		cs.logger.Error("Failed to find active conversation", "error", err, "contact_id", contactID)
	}

	if conversation != nil {
		// Salva no cache
		cs.cacheManager.SetConversation(contactID, conversation, ConversationCacheTTL)
		cs.logger.Info("Found existing active conversation", "contact_id", contactID, "conversation_id", conversation.ID)
		return conversation, nil
	}

	// Cria nova conversa
	conversation, err = cs.createNewConversation(ctx, contactID, inboxID)
	if err != nil {
		return nil, cs.errorHandler.HandleConversationError(err, 0)
	}

	// Salva no cache
	cs.cacheManager.SetConversation(contactID, conversation, ConversationCacheTTL)
	cs.logger.Info("Successfully created conversation", "contact_id", contactID, "conversation_id", conversation.ID)
	return conversation, nil
}

// findActiveConversation busca uma conversa ativa para o contato
func (cs *ConversationService) findActiveConversation(ctx context.Context, contactID int, inboxID int) (*ports.ConversationResponse, error) {
	cs.logger.Info("Searching for active conversation", "contact_id", contactID, "inbox_id", inboxID)

	// Lista conversas do contato
	conversations, err := cs.client.ListContactConversations(ctx, contactID)
	if err != nil {
		return nil, fmt.Errorf("failed to list contact conversations: %w", err)
	}

	cs.logger.Info("Found conversations for contact", "contact_id", contactID, "total_conversations", len(conversations))

	// Procura conversa ativa para esta inbox
	for _, conv := range conversations {
		if conv.InboxID == inboxID && cs.isActiveStatus(conv.Status) {
			cs.logger.Info("Found active conversation for inbox",
				"conversation_id", conv.ID,
				"contact_id", contactID,
				"inbox_id", inboxID,
				"status", conv.Status)
			return cs.adapter.ToPortsConversation(&conv), nil
		}
	}

	cs.logger.Info("No active conversation found", "contact_id", contactID, "inbox_id", inboxID)
	return nil, nil
}

// createNewConversation cria uma nova conversa
func (cs *ConversationService) createNewConversation(ctx context.Context, contactID int, inboxID int) (*ports.ConversationResponse, error) {
	req := ConversationCreateRequest{
		ContactID: contactID,
		InboxID:   inboxID,
		Status:    "pending",
	}

	cs.logger.Info("Creating new conversation", "contact_id", contactID, "inbox_id", inboxID)
	conversation, err := cs.client.CreateConversation(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	// Converte para tipo da interface
	conversationResponse := cs.adapter.ToPortsConversation(conversation)

	// Valida a conversa criada
	if err := cs.validateConversationResponse(conversationResponse); err != nil {
		return nil, fmt.Errorf("invalid conversation response: %w", err)
	}

	cs.logger.Info("Successfully created conversation",
		"conversation_id", conversation.ID,
		"contact_id", contactID,
		"inbox_id", inboxID,
		"status", conversation.Status)

	return conversationResponse, nil
}

// isActiveStatus verifica se o status da conversa é ativo
func (cs *ConversationService) isActiveStatus(status string) bool {
	activeStatuses := []string{"open", "pending"}
	for _, activeStatus := range activeStatuses {
		if status == activeStatus {
			return true
		}
	}
	return false
}

// validateConversationResponse valida uma response de conversa
func (cs *ConversationService) validateConversationResponse(conversation *ports.ConversationResponse) error {
	if conversation == nil {
		return fmt.Errorf("conversation response is nil")
	}

	if conversation.ID == 0 {
		return fmt.Errorf("conversation ID is invalid")
	}

	if conversation.ContactID == 0 {
		return fmt.Errorf("conversation contact ID is invalid")
	}

	if conversation.InboxID == 0 {
		return fmt.Errorf("conversation inbox ID is invalid")
	}

	return nil
}

// GetConversationByID busca uma conversa pelo ID
func (cs *ConversationService) GetConversationByID(ctx context.Context, conversationID int) (*ports.ConversationResponse, error) {
	cs.logger.Info("Getting conversation by ID", "conversation_id", conversationID)

	conversation, err := cs.client.GetConversation(ctx, conversationID)
	if err != nil {
		return nil, cs.errorHandler.HandleConversationError(err, conversationID)
	}

	// Converte para tipo da interface
	conversationResponse := cs.adapter.ToPortsConversation(conversation)

	// Valida a conversa
	if err := cs.validateConversationResponse(conversationResponse); err != nil {
		return nil, fmt.Errorf("invalid conversation response: %w", err)
	}

	return conversationResponse, nil
}

// ClearConversationCache limpa o cache de uma conversa específica
func (cs *ConversationService) ClearConversationCache(contactID int) {
	cs.cacheManager.DeleteConversation(contactID)
	cs.logger.Info("Cleared conversation cache", "contact_id", contactID)
}

// MapConversation cria mapeamento entre chat e conversa (deprecated - now handled by service)
func (cs *ConversationService) MapConversation(ctx context.Context, chatJID string, contactID int, conversationID int) error {
	cs.logger.Info("Conversation mapping handled by unified strategy",
		"chat_jid", chatJID,
		"contact_id", contactID,
		"conversation_id", conversationID)

	// Mapeamento agora é feito pela estratégia unificada no service
	return nil
}

// ListActiveConversations lista conversas ativas para uma inbox
func (cs *ConversationService) ListActiveConversations(ctx context.Context, inboxID int) ([]*ports.ConversationResponse, error) {
	cs.logger.Info("Listing active conversations", "inbox_id", inboxID)

	// Esta funcionalidade pode ser implementada futuramente
	// Por enquanto, retorna lista vazia
	cs.logger.Info("Active conversations listing not implemented yet")
	return []*ports.ConversationResponse{}, nil
}

// UpdateConversationStatus atualiza o status de uma conversa
func (cs *ConversationService) UpdateConversationStatus(ctx context.Context, conversationID int, status string) error {
	cs.logger.Info("Updating conversation status", "conversation_id", conversationID, "status", status)

	// Valida o status
	validStatuses := []string{"open", "pending", "resolved", "snoozed"}
	isValidStatus := false
	for _, validStatus := range validStatuses {
		if status == validStatus {
			isValidStatus = true
			break
		}
	}

	if !isValidStatus {
		return fmt.Errorf("invalid conversation status: %s", status)
	}

	// Esta funcionalidade pode ser implementada futuramente
	cs.logger.Info("Conversation status update not implemented yet")
	return nil
}

// GetConversationStats retorna estatísticas de conversas
func (cs *ConversationService) GetConversationStats(ctx context.Context, inboxID int) (*ConversationStats, error) {
	cs.logger.Info("Getting conversation stats", "inbox_id", inboxID)

	// Esta funcionalidade pode ser implementada futuramente
	stats := &ConversationStats{
		InboxID:               inboxID,
		TotalConversations:    0,
		ActiveConversations:   0,
		PendingConversations:  0,
		ResolvedConversations: 0,
	}

	cs.logger.Info("Retrieved conversation stats", "inbox_id", inboxID, "stats", stats)
	return stats, nil
}

// ConversationStats representa estatísticas de conversas
type ConversationStats struct {
	InboxID               int `json:"inbox_id"`
	TotalConversations    int `json:"total_conversations"`
	ActiveConversations   int `json:"active_conversations"`
	PendingConversations  int `json:"pending_conversations"`
	ResolvedConversations int `json:"resolved_conversations"`
}

// ConversationHealth verifica a saúde de uma conversa
func (cs *ConversationService) ConversationHealth(ctx context.Context, conversationID int) *ConversationHealthStatus {
	status := &ConversationHealthStatus{
		ConversationID: conversationID,
		IsHealthy:      true,
		Issues:         []string{},
	}

	// Verifica se a conversa existe
	conversation, err := cs.GetConversationByID(ctx, conversationID)
	if err != nil {
		status.IsHealthy = false
		status.Issues = append(status.Issues, fmt.Sprintf("Conversation not found: %v", err))
		return status
	}

	// Verifica se a conversa tem dados válidos
	if err := cs.validateConversationResponse(conversation); err != nil {
		status.IsHealthy = false
		status.Issues = append(status.Issues, fmt.Sprintf("Invalid conversation data: %v", err))
	}

	cs.logger.Info("Conversation health check completed",
		"conversation_id", conversationID,
		"is_healthy", status.IsHealthy,
		"issues_count", len(status.Issues))

	return status
}

// ConversationHealthStatus representa o status de saúde de uma conversa
type ConversationHealthStatus struct {
	ConversationID int      `json:"conversation_id"`
	IsHealthy      bool     `json:"is_healthy"`
	Issues         []string `json:"issues"`
}
