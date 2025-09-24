package chatwoot

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"zpmeow/internal/infra/database/models"
	"zpmeow/internal/infra/database/repository"
)

// ConversationMapper gerencia o mapeamento inteligente entre chats do WhatsApp e conversas do Chatwoot
type ConversationMapper struct {
	client    *Client
	chatRepo  *repository.ChatRepository
	logger    *slog.Logger
	sessionID string
	inboxID   int
}

// NewConversationMapper cria um novo mapper de conversas
func NewConversationMapper(client *Client, chatRepo *repository.ChatRepository, logger *slog.Logger, sessionID string, inboxID int) *ConversationMapper {
	return &ConversationMapper{
		client:    client,
		chatRepo:  chatRepo,
		logger:    logger,
		sessionID: sessionID,
		inboxID:   inboxID,
	}
}

// ConversationMapping representa o mapeamento de uma conversa
type ConversationMapping struct {
	ChatID         string
	ContactID      int
	ConversationID int
	IsValid        bool
	NeedsUpdate    bool
}

// GetOrCreateConversationMapping obtém ou cria mapeamento inteligente para uma conversa
// Inspirado na Evolution API com melhorias para garantir separação de conversas
func (cm *ConversationMapper) GetOrCreateConversationMapping(ctx context.Context, chatJid, phoneNumber string) (*ConversationMapping, error) {
	cm.logger.Info("🔍 [MAPPER] Starting intelligent conversation mapping",
		"session_id", cm.sessionID,
		"chat_jid", chatJid,
		"phone_number", phoneNumber,
		"inbox_id", cm.inboxID)

	// 1. Busca chat existente no zpmeow
	chat, err := cm.chatRepo.GetChatWithChatwootMapping(ctx, cm.sessionID, chatJid)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat: %w", err)
	}

	// 2. Se chat não existe, retorna nil (será criado depois)
	if chat == nil {
		cm.logger.Info("📝 [MAPPER] Chat not found, will be created",
			"chat_jid", chatJid)
		return nil, nil
	}

	// 3. VALIDAÇÃO INTELIGENTE: Verifica se mapeamento salvo ainda é válido
	if chat.ChatwootContactId != nil && chat.ChatwootConversationId != nil {
		cm.logger.Info("🔍 [MAPPER] Found saved mapping, validating",
			"chat_id", chat.ID,
			"saved_contact_id", *chat.ChatwootContactId,
			"saved_conversation_id", *chat.ChatwootConversationId)

		// Valida se a conversa ainda existe E está ativa
		isValid, conversation, err := cm.validateAndGetConversation(ctx, int(*chat.ChatwootConversationId))
		if err != nil {
			cm.logger.Error("❌ [MAPPER] Error validating conversation",
				"conversation_id", *chat.ChatwootConversationId,
				"error", err)
		}

		if isValid && conversation != nil {
			// GARANTIA DE SEPARAÇÃO: Verifica se a conversa pertence à inbox correta
			if conversation.InboxID == cm.inboxID {
				cm.logger.Info("✅ [MAPPER] Saved mapping is valid and belongs to correct inbox",
					"conversation_id", *chat.ChatwootConversationId,
					"inbox_id", cm.inboxID)
				return &ConversationMapping{
					ChatID:         chat.ID,
					ContactID:      int(*chat.ChatwootContactId),
					ConversationID: int(*chat.ChatwootConversationId),
					IsValid:        true,
					NeedsUpdate:    false,
				}, nil
			} else {
				cm.logger.Warn("⚠️ [MAPPER] Conversation belongs to different inbox, will recreate",
					"conversation_id", *chat.ChatwootConversationId,
					"expected_inbox", cm.inboxID,
					"actual_inbox", conversation.InboxID)
			}
		}

		cm.logger.Warn("⚠️ [MAPPER] Saved mapping is invalid, will recreate",
			"old_conversation_id", *chat.ChatwootConversationId)
	}

	// 4. CRIAÇÃO/BUSCA INTELIGENTE (baseado na Evolution API)
	cm.logger.Info("🔄 [MAPPER] Creating new mapping",
		"chat_jid", chatJid,
		"phone_number", phoneNumber)

	// Busca ou cria contato
	contact, err := cm.findOrCreateContact(ctx, phoneNumber, chatJid)
	if err != nil {
		return nil, fmt.Errorf("failed to find/create contact: %w", err)
	}

	// ESTRATÉGIA EVOLUTION: Busca conversa ativa específica para esta inbox
	conversation, err := cm.findActiveConversationForInbox(ctx, contact.ID, cm.inboxID)
	if err != nil {
		return nil, fmt.Errorf("failed to find active conversation: %w", err)
	}

	// Se não encontrou conversa ativa para esta inbox, cria nova
	if conversation == nil {
		conversation, err = cm.createNewConversation(ctx, contact.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to create conversation: %w", err)
		}
		cm.logger.Info("✅ [MAPPER] Created new conversation",
			"conversation_id", conversation.ID,
			"contact_id", contact.ID,
			"inbox_id", cm.inboxID)
	} else {
		cm.logger.Info("✅ [MAPPER] Found existing active conversation",
			"conversation_id", conversation.ID,
			"contact_id", contact.ID,
			"inbox_id", cm.inboxID,
			"status", conversation.Status)
	}

	// 5. Atualiza mapeamento no banco
	err = cm.updateChatMapping(ctx, chat.ID, int64(contact.ID), int64(conversation.ID))
	if err != nil {
		cm.logger.Error("❌ [MAPPER] Failed to update chat mapping",
			"chat_id", chat.ID,
			"contact_id", contact.ID,
			"conversation_id", conversation.ID,
			"error", err)
	}

	cm.logger.Info("✅ [MAPPER] Successfully created/updated mapping",
		"chat_id", chat.ID,
		"contact_id", contact.ID,
		"conversation_id", conversation.ID,
		"inbox_id", cm.inboxID)

	return &ConversationMapping{
		ChatID:         chat.ID,
		ContactID:      contact.ID,
		ConversationID: conversation.ID,
		IsValid:        true,
		NeedsUpdate:    true,
	}, nil
}

// validateAndGetConversation verifica se uma conversa ainda existe E retorna seus dados
func (cm *ConversationMapper) validateAndGetConversation(ctx context.Context, conversationID int) (bool, *Conversation, error) {
	// Tenta buscar a conversa no Chatwoot
	conversation, err := cm.client.GetConversation(ctx, conversationID)
	if err != nil {
		// Se erro 404, conversa não existe mais
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return false, nil, nil
		}
		// Outros erros são problemas de conectividade
		return false, nil, err
	}

	// Verifica se a conversa não está resolvida (como Evolution API)
	if conversation.Status == "resolved" {
		cm.logger.Info("⚠️ [MAPPER] Conversation is resolved, will create new one",
			"conversation_id", conversationID,
			"status", conversation.Status)
		return false, conversation, nil
	}

	return true, conversation, nil
}



// findOrCreateContact busca ou cria contato no Chatwoot
func (cm *ConversationMapper) findOrCreateContact(ctx context.Context, phoneNumber, chatJid string) (*Contact, error) {
	// Busca contato existente
	contacts, err := cm.client.SearchContacts(ctx, phoneNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to search contacts: %w", err)
	}

	// Se encontrou contato, retorna o primeiro
	if len(contacts) > 0 {
		cm.logger.Info("✅ [MAPPER] Found existing contact",
			"contact_id", contacts[0].ID,
			"phone_number", phoneNumber)
		return &contacts[0], nil
	}

	// Cria novo contato
	isGroup := strings.Contains(chatJid, "@g.us")
	contactReq := &ContactCreateRequest{
		Name:        phoneNumber, // Usar phone como nome inicial
		PhoneNumber: fmt.Sprintf("+%s", phoneNumber),
		Identifier:  chatJid,
		InboxID:     cm.inboxID,
	}

	contact, err := cm.client.CreateContact(ctx, *contactReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create contact: %w", err)
	}

	cm.logger.Info("✅ [MAPPER] Created new contact",
		"contact_id", contact.ID,
		"phone_number", phoneNumber,
		"is_group", isGroup)

	return contact, nil
}

// findActiveConversationForInbox busca conversa ativa para um contato ESPECÍFICA para esta inbox
// Baseado na Evolution API - garante separação absoluta entre inboxes
func (cm *ConversationMapper) findActiveConversationForInbox(ctx context.Context, contactID, inboxID int) (*Conversation, error) {
	conversations, err := cm.client.ListContactConversations(ctx, contactID)
	if err != nil {
		return nil, fmt.Errorf("failed to list conversations: %w", err)
	}

	cm.logger.Info("🔍 [MAPPER] Searching for active conversation in specific inbox",
		"contact_id", contactID,
		"inbox_id", inboxID,
		"total_conversations", len(conversations))

	// ESTRATÉGIA EVOLUTION: Busca conversa ativa ESPECÍFICA para esta inbox
	for _, conv := range conversations {
		cm.logger.Info("🔍 [MAPPER] Checking conversation",
			"conversation_id", conv.ID,
			"inbox_id", conv.InboxID,
			"status", conv.Status,
			"target_inbox", inboxID)

		// GARANTIA DE SEPARAÇÃO: Deve ser da inbox correta E não resolvida
		if conv.InboxID == inboxID && conv.Status != "resolved" {
			cm.logger.Info("✅ [MAPPER] Found active conversation for inbox",
				"conversation_id", conv.ID,
				"contact_id", contactID,
				"inbox_id", inboxID,
				"status", conv.Status)
			return &conv, nil
		}
	}

	cm.logger.Info("📝 [MAPPER] No active conversation found for inbox",
		"contact_id", contactID,
		"inbox_id", inboxID)
	return nil, nil
}



// createNewConversation cria nova conversa no Chatwoot
func (cm *ConversationMapper) createNewConversation(ctx context.Context, contactID int) (*Conversation, error) {
	convReq := ConversationCreateRequest{
		ContactID: contactID,
		InboxID:   cm.inboxID,
		Status:    "pending",
	}

	conversation, err := cm.client.CreateConversation(ctx, convReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	cm.logger.Info("✅ [MAPPER] Created new conversation",
		"conversation_id", conversation.ID,
		"contact_id", contactID,
		"inbox_id", cm.inboxID)

	return conversation, nil
}

// updateChatMapping atualiza o mapeamento no banco de dados
func (cm *ConversationMapper) updateChatMapping(ctx context.Context, chatID string, contactID, conversationID int64) error {
	return cm.chatRepo.UpdateChatwootMapping(ctx, chatID, contactID, conversationID)
}

// SaveConversationMapping salva mapeamento de conversa (usado pela estratégia híbrida)
func (cm *ConversationMapper) SaveConversationMapping(ctx context.Context, chatJid, phoneNumber string, contactID, conversationID int) error {
	cm.logger.Info("💾 [MAPPER] Saving conversation mapping",
		"session_id", cm.sessionID,
		"chat_jid", chatJid,
		"phone_number", phoneNumber,
		"contact_id", contactID,
		"conversation_id", conversationID)

	// Extrai chatID do JID
	chatID := strings.Split(chatJid, "@")[0]

	// Busca ou cria chat no zpmeow
	chat, err := cm.chatRepo.GetChatBySessionAndJID(ctx, cm.sessionID, chatJid)
	if err != nil {
		cm.logger.Error("❌ [MAPPER] Failed to get chat by JID", "error", err, "chat_jid", chatJid)
		return fmt.Errorf("failed to get chat by JID: %w", err)
	}

	if chat == nil {
		// Cria novo chat se não existir
		cm.logger.Info("📝 [MAPPER] Creating new chat for mapping",
			"chat_jid", chatJid,
			"chat_id", chatID)

		newChat := &models.ChatModel{
			SessionId:              cm.sessionID,
			ChatJid:                chatJid,
			ChatName:               &phoneNumber, // Usa número como nome inicial
			PhoneNumber:            &phoneNumber,
			IsGroup:                strings.Contains(chatJid, "@g.us"),
			ChatwootConversationId: &[]int64{int64(conversationID)}[0],
			ChatwootContactId:      &[]int64{int64(contactID)}[0],
			UnreadCount:            0,
			IsArchived:             false,
		}

		err = cm.chatRepo.CreateChat(ctx, newChat)
		if err != nil {
			cm.logger.Error("❌ [MAPPER] Failed to create chat", "error", err)
			return fmt.Errorf("failed to create chat: %w", err)
		}

		cm.logger.Info("✅ [MAPPER] Created new chat with mapping",
			"chat_id", newChat.ID,
			"chat_jid", chatJid,
			"conversation_id", conversationID,
			"contact_id", contactID)
	} else {
		// Atualiza mapeamento do chat existente
		cm.logger.Info("🔄 [MAPPER] Updating existing chat mapping",
			"chat_id", chat.ID,
			"old_conversation_id", chat.ChatwootConversationId,
			"new_conversation_id", conversationID,
			"old_contact_id", chat.ChatwootContactId,
			"new_contact_id", contactID)

		err = cm.updateChatMapping(ctx, chat.ID, int64(contactID), int64(conversationID))
		if err != nil {
			cm.logger.Error("❌ [MAPPER] Failed to update chat mapping", "error", err)
			return fmt.Errorf("failed to update chat mapping: %w", err)
		}

		cm.logger.Info("✅ [MAPPER] Updated chat mapping successfully",
			"chat_id", chat.ID,
			"conversation_id", conversationID,
			"contact_id", contactID)
	}

	return nil
}
