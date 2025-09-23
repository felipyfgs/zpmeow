package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"zpmeow/internal/infra/database/models"
)

type ChatRepository struct {
	db *sqlx.DB
}

func NewChatRepository(db *sqlx.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

// CreateChat cria um novo chat
func (r *ChatRepository) CreateChat(ctx context.Context, chat *models.ChatModel) error {
	query := `
		INSERT INTO chats (
			session_id, chat_jid, chat_name, chat_type, phone_number, is_group,
			group_subject, group_description, chatwoot_conversation_id, chatwoot_contact_id,
			last_message_at, unread_count, is_archived, metadata
		) VALUES (
			:session_id, :chat_jid, :chat_name, :chat_type, :phone_number, :is_group,
			:group_subject, :group_description, :chatwoot_conversation_id, :chatwoot_contact_id,
			:last_message_at, :unread_count, :is_archived, :metadata
		) RETURNING id, created_at, updated_at`

	rows, err := r.db.NamedQueryContext(ctx, query, chat)
	if err != nil {
		return fmt.Errorf("failed to create chat: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&chat.ID, &chat.CreatedAt, &chat.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan created chat: %w", err)
		}
	}

	return nil
}

// GetChatBySessionAndJID busca um chat por session_id e chat_jid
func (r *ChatRepository) GetChatBySessionAndJID(ctx context.Context, sessionID, chatJID string) (*models.ChatModel, error) {
	var chat models.ChatModel
	query := `
		SELECT id, session_id, chat_jid, chat_name, chat_type, phone_number, is_group,
			   group_subject, group_description, chatwoot_conversation_id, chatwoot_contact_id,
			   last_message_at, unread_count, is_archived, metadata, created_at, updated_at
		FROM chats 
		WHERE session_id = $1 AND chat_jid = $2`

	err := r.db.GetContext(ctx, &chat, query, sessionID, chatJID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get chat: %w", err)
	}

	return &chat, nil
}

// GetChatByID busca um chat por ID
func (r *ChatRepository) GetChatByID(ctx context.Context, id string) (*models.ChatModel, error) {
	var chat models.ChatModel
	query := `
		SELECT id, session_id, chat_jid, chat_name, chat_type, phone_number, is_group,
			   group_subject, group_description, chatwoot_conversation_id, chatwoot_contact_id,
			   last_message_at, unread_count, is_archived, metadata, created_at, updated_at
		FROM chats 
		WHERE id = $1`

	err := r.db.GetContext(ctx, &chat, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get chat by id: %w", err)
	}

	return &chat, nil
}

// GetChatsByChatwootConversationID busca chats por chatwoot_conversation_id
func (r *ChatRepository) GetChatByChatwootConversationID(ctx context.Context, conversationID string) (*models.ChatModel, error) {
	var chat models.ChatModel
	query := `
		SELECT id, session_id, chat_jid, chat_name, chat_type, phone_number, is_group,
			   group_subject, group_description, chatwoot_conversation_id, chatwoot_contact_id,
			   last_message_at, unread_count, is_archived, metadata, created_at, updated_at
		FROM chats 
		WHERE chatwoot_conversation_id = $1`

	err := r.db.GetContext(ctx, &chat, query, conversationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get chat by chatwoot conversation id: %w", err)
	}

	return &chat, nil
}

// UpdateChat atualiza um chat
func (r *ChatRepository) UpdateChat(ctx context.Context, chat *models.ChatModel) error {
	query := `
		UPDATE chats SET
			chat_name = :chat_name,
			chat_type = :chat_type,
			phone_number = :phone_number,
			is_group = :is_group,
			group_subject = :group_subject,
			group_description = :group_description,
			chatwoot_conversation_id = :chatwoot_conversation_id,
			chatwoot_contact_id = :chatwoot_contact_id,
			last_message_at = :last_message_at,
			unread_count = :unread_count,
			is_archived = :is_archived,
			metadata = :metadata,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = :id
		RETURNING updated_at`

	rows, err := r.db.NamedQueryContext(ctx, query, chat)
	if err != nil {
		return fmt.Errorf("failed to update chat: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&chat.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan updated chat: %w", err)
		}
	}

	return nil
}

// UpdateLastMessageAt atualiza o timestamp da última mensagem
func (r *ChatRepository) UpdateLastMessageAt(ctx context.Context, chatID string, timestamp time.Time) error {
	query := `UPDATE chats SET last_message_at = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, timestamp, chatID)
	if err != nil {
		return fmt.Errorf("failed to update last message at: %w", err)
	}
	return nil
}

// UpdateUnreadCount atualiza o contador de mensagens não lidas
func (r *ChatRepository) UpdateUnreadCount(ctx context.Context, chatID string, count int) error {
	query := `UPDATE chats SET unread_count = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, count, chatID)
	if err != nil {
		return fmt.Errorf("failed to update unread count: %w", err)
	}
	return nil
}

// GetChatsBySessionID busca todos os chats de uma sessão
func (r *ChatRepository) GetChatsBySessionID(ctx context.Context, sessionID string, limit, offset int) ([]*models.ChatModel, error) {
	var chats []*models.ChatModel
	query := `
		SELECT id, session_id, chat_jid, chat_name, chat_type, phone_number, is_group,
			   group_subject, group_description, chatwoot_conversation_id, chatwoot_contact_id,
			   last_message_at, unread_count, is_archived, metadata, created_at, updated_at
		FROM chats 
		WHERE session_id = $1 AND is_archived = FALSE
		ORDER BY last_message_at DESC NULLS LAST
		LIMIT $2 OFFSET $3`

	err := r.db.SelectContext(ctx, &chats, query, sessionID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get chats by session: %w", err)
	}

	return chats, nil
}

// DeleteChat deleta um chat
func (r *ChatRepository) DeleteChat(ctx context.Context, id string) error {
	query := `DELETE FROM chats WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete chat: %w", err)
	}
	return nil
}

// CreateOrUpdateChat cria ou atualiza um chat
func (r *ChatRepository) CreateOrUpdateChat(ctx context.Context, chat *models.ChatModel) error {
	existing, err := r.GetChatBySessionAndJID(ctx, chat.SessionID, chat.ChatJID)
	if err != nil {
		return err
	}

	if existing == nil {
		return r.CreateChat(ctx, chat)
	}

	// Atualiza o chat existente com novos dados
	existing.ChatName = chat.ChatName
	existing.ChatType = chat.ChatType
	existing.PhoneNumber = chat.PhoneNumber
	existing.IsGroup = chat.IsGroup
	existing.GroupSubject = chat.GroupSubject
	existing.GroupDescription = chat.GroupDescription
	existing.ChatwootConversationID = chat.ChatwootConversationID
	existing.ChatwootContactID = chat.ChatwootContactID
	existing.Metadata = chat.Metadata

	*chat = *existing
	return r.UpdateChat(ctx, chat)
}
