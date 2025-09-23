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
		INSERT INTO "zpChats" (
			"sessionId", "chatJid", "chatName", "phoneNumber", "isGroup",
			"lastMsgAt", "unreadCount", "isArchived", metadata
		) VALUES (
			:sessionId, :chatJid, :chatName, :phoneNumber, :isGroup,
			:lastMsgAt, :unreadCount, :isArchived, :metadata
		) RETURNING id, "createdAt", "updatedAt"`

	rows, err := r.db.NamedQueryContext(ctx, query, chat)
	if err != nil {
		return fmt.Errorf("failed to create chat: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close rows in CreateChat: %v\n", closeErr)
		}
	}()

	if rows.Next() {
		err = rows.Scan(&chat.ID, &chat.CreatedAt, &chat.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan created chat: %w", err)
		}
	}

	return nil
}

// GetChatBySessionAndJID busca um chat por sessionId e chatJid (OTIMIZADA)
func (r *ChatRepository) GetChatBySessionAndJID(ctx context.Context, sessionID, chatJID string) (*models.ChatModel, error) {
	var chat models.ChatModel
	query := `
		SELECT id, "sessionId", "chatJid", "chatName", "phoneNumber", "isGroup",
			   "lastMsgAt", "unreadCount", "isArchived", metadata, "createdAt", "updatedAt"
		FROM "zpChats"
		WHERE "sessionId" = $1 AND "chatJid" = $2`

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
		SELECT id, "sessionId", "chatJid", "chatName", "phoneNumber", "isGroup",
			   "lastMsgAt", "unreadCount", "isArchived", metadata, "createdAt", "updatedAt"
		FROM "zpChats"
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

// GetChatByChatwootConversationID - MÉTODO REMOVIDO (campos Chatwoot removidos)
// Use relação separada para vincular chats com Chatwoot
func (r *ChatRepository) GetChatByChatwootConversationID(ctx context.Context, conversationID string) (*models.ChatModel, error) {
	// TODO: Implementar busca via tabela de relação zpCwMessages
	return nil, fmt.Errorf("método removido - usar relação separada para Chatwoot")
}

// UpdateChat atualiza um chat
func (r *ChatRepository) UpdateChat(ctx context.Context, chat *models.ChatModel) error {
	query := `
		UPDATE "zpChats" SET
			"chatName" = :chatName,
			"phoneNumber" = :phoneNumber,
			"isGroup" = :isGroup,
			"lastMsgAt" = :lastMsgAt,
			"unreadCount" = :unreadCount,
			"isArchived" = :isArchived,
			metadata = :metadata,
			"updatedAt" = CURRENT_TIMESTAMP
		WHERE id = :id
		RETURNING "updatedAt"`

	rows, err := r.db.NamedQueryContext(ctx, query, chat)
	if err != nil {
		return fmt.Errorf("failed to update chat: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close rows in UpdateChat: %v\n", closeErr)
		}
	}()

	if rows.Next() {
		err = rows.Scan(&chat.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan updated chat: %w", err)
		}
	}

	return nil
}

// UpdateLastMessageAt atualiza o timestamp da última mensagem (OTIMIZADA)
func (r *ChatRepository) UpdateLastMessageAt(ctx context.Context, chatID string, timestamp time.Time) error {
	query := `UPDATE "zpChats" SET "lastMsgAt" = $1, "updatedAt" = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, timestamp, chatID)
	if err != nil {
		return fmt.Errorf("failed to update last message at: %w", err)
	}
	return nil
}

// UpdateUnreadCount atualiza o contador de mensagens não lidas (OTIMIZADA)
func (r *ChatRepository) UpdateUnreadCount(ctx context.Context, chatID string, count int) error {
	query := `UPDATE "zpChats" SET "unreadCount" = $1, "updatedAt" = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, count, chatID)
	if err != nil {
		return fmt.Errorf("failed to update unread count: %w", err)
	}
	return nil
}

// GetChatsBySessionId busca todos os chats de uma sessão (OTIMIZADA)
func (r *ChatRepository) GetChatsBySessionID(ctx context.Context, sessionID string, limit, offset int) ([]*models.ChatModel, error) {
	var chats []*models.ChatModel
	query := `
		SELECT id, "sessionId", "chatJid", "chatName", "phoneNumber", "isGroup",
			   "lastMsgAt", "unreadCount", "isArchived", metadata, "createdAt", "updatedAt"
		FROM "zpChats"
		WHERE "sessionId" = $1 AND "isArchived" = FALSE
		ORDER BY "lastMsgAt" DESC NULLS LAST
		LIMIT $2 OFFSET $3`

	err := r.db.SelectContext(ctx, &chats, query, sessionID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get chats by session: %w", err)
	}

	return chats, nil
}

// DeleteChat deleta um chat
func (r *ChatRepository) DeleteChat(ctx context.Context, id string) error {
	query := `DELETE FROM "zpChats" WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete chat: %w", err)
	}
	return nil
}

// CreateOrUpdateChat cria ou atualiza um chat
func (r *ChatRepository) CreateOrUpdateChat(ctx context.Context, chat *models.ChatModel) error {
	existing, err := r.GetChatBySessionAndJID(ctx, chat.SessionId, chat.ChatJid)
	if err != nil {
		return err
	}

	if existing == nil {
		return r.CreateChat(ctx, chat)
	}

	// Atualiza o chat existente com novos dados (campos otimizados)
	existing.ChatName = chat.ChatName
	existing.PhoneNumber = chat.PhoneNumber
	existing.IsGroup = chat.IsGroup
	existing.Metadata = chat.Metadata
	// ChatType removido - redundante com IsGroup
	// GroupSubject, GroupDescription movidos para Metadata
	// Campos Chatwoot removidos - usar relação separada

	*chat = *existing
	return r.UpdateChat(ctx, chat)
}
