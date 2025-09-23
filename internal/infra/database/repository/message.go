package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"zpmeow/internal/infra/database/models"
)

type MessageRepository struct {
	db *sqlx.DB
}

func NewMessageRepository(db *sqlx.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// CreateMessage cria uma nova mensagem
func (r *MessageRepository) CreateMessage(ctx context.Context, message *models.MessageModel) error {
	// Se ID não foi fornecido, deixa o PostgreSQL gerar um UUID
	if message.ID == "" {
		query := `
			INSERT INTO messages (
				chat_id, session_id, whatsapp_message_id, message_type, content,
				media_url, media_mime_type, media_size, media_filename, thumbnail_url,
				sender_jid, sender_name, is_from_me, is_forwarded, is_broadcast,
				quoted_message_id, quoted_content, status, timestamp, reaction, metadata
			) VALUES (
				:chat_id, :session_id, :whatsapp_message_id, :message_type, :content,
				:media_url, :media_mime_type, :media_size, :media_filename, :thumbnail_url,
				:sender_jid, :sender_name, :is_from_me, :is_forwarded, :is_broadcast,
				:quoted_message_id, :quoted_content, :status, :timestamp, :reaction, :metadata
			) RETURNING id, created_at, updated_at`

		rows, err := r.db.NamedQueryContext(ctx, query, message)
		if err != nil {
			return fmt.Errorf("failed to create message: %w", err)
		}
		defer rows.Close()

		if rows.Next() {
			err = rows.Scan(&message.ID, &message.CreatedAt, &message.UpdatedAt)
			if err != nil {
				return fmt.Errorf("failed to scan created message: %w", err)
			}
		}
	} else {
		// Se ID foi fornecido, usa ele
		query := `
			INSERT INTO messages (
				id, chat_id, session_id, whatsapp_message_id, message_type, content,
				media_url, media_mime_type, media_size, media_filename, thumbnail_url,
				sender_jid, sender_name, is_from_me, is_forwarded, is_broadcast,
				quoted_message_id, quoted_content, status, timestamp, reaction, metadata
			) VALUES (
				:id, :chat_id, :session_id, :whatsapp_message_id, :message_type, :content,
				:media_url, :media_mime_type, :media_size, :media_filename, :thumbnail_url,
				:sender_jid, :sender_name, :is_from_me, :is_forwarded, :is_broadcast,
				:quoted_message_id, :quoted_content, :status, :timestamp, :reaction, :metadata
			) RETURNING created_at, updated_at`

		rows, err := r.db.NamedQueryContext(ctx, query, message)
		if err != nil {
			return fmt.Errorf("failed to create message: %w", err)
		}
		defer rows.Close()

		if rows.Next() {
			err = rows.Scan(&message.CreatedAt, &message.UpdatedAt)
			if err != nil {
				return fmt.Errorf("failed to scan created message: %w", err)
			}
		}
	}

	return nil
}

// GetMessageByID busca uma mensagem por ID
func (r *MessageRepository) GetMessageByID(ctx context.Context, id string) (*models.MessageModel, error) {
	var message models.MessageModel
	query := `
		SELECT id, chat_id, session_id, whatsapp_message_id, message_type, content,
			   media_url, media_mime_type, media_size, media_filename, thumbnail_url,
			   sender_jid, sender_name, is_from_me, is_forwarded, is_broadcast,
			   quoted_message_id, quoted_content, status, timestamp, edit_timestamp,
			   is_deleted, deleted_at, reaction, metadata, created_at, updated_at
		FROM messages 
		WHERE id = $1`

	err := r.db.GetContext(ctx, &message, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	return &message, nil
}

// GetMessageByWhatsAppID busca uma mensagem por WhatsApp message ID
func (r *MessageRepository) GetMessageByWhatsAppID(ctx context.Context, sessionID, whatsappMessageID string) (*models.MessageModel, error) {
	var message models.MessageModel
	query := `
		SELECT id, chat_id, session_id, whatsapp_message_id, message_type, content,
			   media_url, media_mime_type, media_size, media_filename, thumbnail_url,
			   sender_jid, sender_name, is_from_me, is_forwarded, is_broadcast,
			   quoted_message_id, quoted_content, status, timestamp, edit_timestamp,
			   is_deleted, deleted_at, reaction, metadata, created_at, updated_at
		FROM messages 
		WHERE session_id = $1 AND whatsapp_message_id = $2`

	err := r.db.GetContext(ctx, &message, query, sessionID, whatsappMessageID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get message by whatsapp id: %w", err)
	}

	return &message, nil
}

// GetMessagesByChatID busca mensagens de um chat
func (r *MessageRepository) GetMessagesByChatID(ctx context.Context, chatID string, limit, offset int) ([]*models.MessageModel, error) {
	var messages []*models.MessageModel
	query := `
		SELECT id, chat_id, session_id, whatsapp_message_id, message_type, content,
			   media_url, media_mime_type, media_size, media_filename, thumbnail_url,
			   sender_jid, sender_name, is_from_me, is_forwarded, is_broadcast,
			   quoted_message_id, quoted_content, status, timestamp, edit_timestamp,
			   is_deleted, deleted_at, reaction, metadata, created_at, updated_at
		FROM messages 
		WHERE chat_id = $1 AND is_deleted = FALSE
		ORDER BY timestamp DESC
		LIMIT $2 OFFSET $3`

	err := r.db.SelectContext(ctx, &messages, query, chatID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages by chat: %w", err)
	}

	return messages, nil
}

// UpdateMessage atualiza uma mensagem
func (r *MessageRepository) UpdateMessage(ctx context.Context, message *models.MessageModel) error {
	query := `
		UPDATE messages SET
			content = :content,
			media_url = :media_url,
			media_mime_type = :media_mime_type,
			media_size = :media_size,
			media_filename = :media_filename,
			thumbnail_url = :thumbnail_url,
			sender_name = :sender_name,
			status = :status,
			edit_timestamp = :edit_timestamp,
			reaction = :reaction,
			metadata = :metadata,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = :id
		RETURNING updated_at`

	rows, err := r.db.NamedQueryContext(ctx, query, message)
	if err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&message.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan updated message: %w", err)
		}
	}

	return nil
}

// UpdateMessageStatus atualiza o status de uma mensagem
func (r *MessageRepository) UpdateMessageStatus(ctx context.Context, id string, status string) error {
	query := `UPDATE messages SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update message status: %w", err)
	}
	return nil
}

// EditMessage edita o conteúdo de uma mensagem
func (r *MessageRepository) EditMessage(ctx context.Context, id string, newContent string) error {
	query := `
		UPDATE messages SET 
			content = $1, 
			edit_timestamp = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, newContent, id)
	if err != nil {
		return fmt.Errorf("failed to edit message: %w", err)
	}
	return nil
}

// DeleteMessage marca uma mensagem como deletada
func (r *MessageRepository) DeleteMessage(ctx context.Context, id string) error {
	query := `
		UPDATE messages SET 
			is_deleted = TRUE, 
			deleted_at = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP 
		WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}
	return nil
}

// AddReaction adiciona uma reação a uma mensagem
func (r *MessageRepository) AddReaction(ctx context.Context, id string, reaction string) error {
	query := `UPDATE messages SET reaction = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, reaction, id)
	if err != nil {
		return fmt.Errorf("failed to add reaction: %w", err)
	}
	return nil
}

// RemoveReaction remove uma reação de uma mensagem
func (r *MessageRepository) RemoveReaction(ctx context.Context, id string) error {
	query := `UPDATE messages SET reaction = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to remove reaction: %w", err)
	}
	return nil
}

// GetUnreadMessagesByChat busca mensagens não lidas de um chat
func (r *MessageRepository) GetUnreadMessagesByChat(ctx context.Context, chatID string) ([]*models.MessageModel, error) {
	var messages []*models.MessageModel
	query := `
		SELECT id, chat_id, session_id, whatsapp_message_id, message_type, content,
			   media_url, media_mime_type, media_size, media_filename, thumbnail_url,
			   sender_jid, sender_name, is_from_me, is_forwarded, is_broadcast,
			   quoted_message_id, quoted_content, status, timestamp, edit_timestamp,
			   is_deleted, deleted_at, reaction, metadata, created_at, updated_at
		FROM messages 
		WHERE chat_id = $1 AND is_from_me = FALSE AND status != 'read' AND is_deleted = FALSE
		ORDER BY timestamp ASC`

	err := r.db.SelectContext(ctx, &messages, query, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unread messages: %w", err)
	}

	return messages, nil
}

// MarkMessagesAsRead marca mensagens como lidas
func (r *MessageRepository) MarkMessagesAsRead(ctx context.Context, messageIDs []string) error {
	if len(messageIDs) == 0 {
		return nil
	}

	query, args, err := sqlx.In(`UPDATE messages SET status = 'read', updated_at = CURRENT_TIMESTAMP WHERE id IN (?)`, messageIDs)
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	query = r.db.Rebind(query)
	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to mark messages as read: %w", err)
	}

	return nil
}

// GetMessagesByTimeRange busca mensagens em um intervalo de tempo
func (r *MessageRepository) GetMessagesByTimeRange(ctx context.Context, chatID string, startTime, endTime time.Time, limit, offset int) ([]*models.MessageModel, error) {
	var messages []*models.MessageModel
	query := `
		SELECT id, chat_id, session_id, whatsapp_message_id, message_type, content,
			   media_url, media_mime_type, media_size, media_filename, thumbnail_url,
			   sender_jid, sender_name, is_from_me, is_forwarded, is_broadcast,
			   quoted_message_id, quoted_content, status, timestamp, edit_timestamp,
			   is_deleted, deleted_at, reaction, metadata, created_at, updated_at
		FROM messages 
		WHERE chat_id = $1 AND timestamp BETWEEN $2 AND $3 AND is_deleted = FALSE
		ORDER BY timestamp DESC
		LIMIT $4 OFFSET $5`

	err := r.db.SelectContext(ctx, &messages, query, chatID, startTime, endTime, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages by time range: %w", err)
	}

	return messages, nil
}
