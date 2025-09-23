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
			INSERT INTO "zpMessages" (
				"chatId", "sessionId", "msgId", "msgType", content,
				"mediaInfo", "senderJid", "senderName", "isFromMe", "isForwarded", "isBroadcast",
				"quotedMsgId", "quotedContent", status, timestamp, "editTimestamp", "isDeleted", "deletedAt", reaction, metadata
			) VALUES (
				$1, $2, $3, $4, $5,
				$6, $7, $8, $9, $10, $11,
				$12, $13, $14, $15, $16, $17, $18, $19, $20
			) RETURNING id, "createdAt", "updatedAt"`

		err := r.db.QueryRowContext(ctx, query,
			message.ChatId, message.SessionId, message.MsgId, message.MsgType, message.Content,
			message.MediaInfo, message.SenderJid, message.SenderName, message.IsFromMe, message.IsForwarded, message.IsBroadcast,
			message.QuotedMsgId, message.QuotedContent, message.Status, message.Timestamp, message.EditTimestamp, message.IsDeleted, message.DeletedAt, message.Reaction, message.Metadata,
		).Scan(&message.ID, &message.CreatedAt, &message.UpdatedAt)

		if err != nil {
			return fmt.Errorf("failed to create message: %w", err)
		}
	} else {
		// Se ID foi fornecido, usa ele
		query := `
			INSERT INTO "zpMessages" (
				id, "chatId", "sessionId", "msgId", "msgType", content,
				"mediaInfo", "senderJid", "senderName", "isFromMe", "isForwarded", "isBroadcast",
				"quotedMsgId", "quotedContent", status, timestamp, "editTimestamp", "isDeleted", "deletedAt", reaction, metadata
			) VALUES (
				$1, $2, $3, $4, $5, $6,
				$7, $8, $9, $10, $11, $12,
				$13, $14, $15, $16, $17, $18, $19, $20, $21
			) RETURNING "createdAt", "updatedAt"`

		err := r.db.QueryRowContext(ctx, query,
			message.ID, message.ChatId, message.SessionId, message.MsgId, message.MsgType, message.Content,
			message.MediaInfo, message.SenderJid, message.SenderName, message.IsFromMe, message.IsForwarded, message.IsBroadcast,
			message.QuotedMsgId, message.QuotedContent, message.Status, message.Timestamp, message.EditTimestamp, message.IsDeleted, message.DeletedAt, message.Reaction, message.Metadata,
		).Scan(&message.CreatedAt, &message.UpdatedAt)

		if err != nil {
			return fmt.Errorf("failed to create message: %w", err)
		}
	}

	return nil
}

// GetMessageByID busca uma mensagem por ID
func (r *MessageRepository) GetMessageByID(ctx context.Context, id string) (*models.MessageModel, error) {
	var message models.MessageModel
	query := `
		SELECT id, "chatId", "sessionId", "msgId", "msgType", content,
			   "mediaInfo", "senderJid", "senderName", "isFromMe", "isForwarded", "isBroadcast",
			   "quotedMsgId", "quotedContent", status, timestamp, "editTimestamp",
			   "isDeleted", "deletedAt", reaction, metadata, "createdAt", "updatedAt"
		FROM "zpMessages"
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
		SELECT id, "chatId", "sessionId", "msgId", "msgType", content,
			   "mediaInfo", "senderJid", "senderName", "isFromMe", "isForwarded", "isBroadcast",
			   "quotedMsgId", "quotedContent", status, timestamp, "editTimestamp",
			   "isDeleted", "deletedAt", reaction, metadata, "createdAt", "updatedAt"
		FROM "zpMessages"
		WHERE "sessionId" = $1 AND "msgId" = $2`

	err := r.db.GetContext(ctx, &message, query, sessionID, whatsappMessageID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get message by whatsapp id: %w", err)
	}

	return &message, nil
}

// GetMessagesByChatId busca mensagens de um chat
func (r *MessageRepository) GetMessagesByChatId(ctx context.Context, chatID string, limit, offset int) ([]*models.MessageModel, error) {
	var messages []*models.MessageModel
	query := `
		SELECT id, "chatId", "sessionId", msgId, "msgType", content,
			   media_url, media_mime_type, media_size, media_filename, thumbnail_url,
			   "senderJid", "senderName", "isFromMe", "isForwarded", "isBroadcast",
			   "quotedMsgId", "quotedContent", status, timestamp, "editTimestamp",
			   "isDeleted", "deletedAt", reaction, metadata, "createdAt", "updatedAt"
		FROM "zpMessages" 
		WHERE "chatId" = $1 AND "isDeleted" = FALSE
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
		UPDATE "zpMessages" SET
			content = :content,
			media_url = :media_url,
			media_mime_type = :media_mime_type,
			media_size = :media_size,
			media_filename = :media_filename,
			thumbnail_url = :thumbnail_url,
			"senderName" = :"senderName",
			status = :status,
			"editTimestamp" = :"editTimestamp",
			reaction = :reaction,
			metadata = :metadata,
			"updatedAt" = CURRENT_TIMESTAMP
		WHERE id = :id
		RETURNING "updatedAt"`

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
	query := `UPDATE "zpMessages" SET status = $1, "updatedAt" = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update message status: %w", err)
	}
	return nil
}

// EditMessage edita o conteúdo de uma mensagem
func (r *MessageRepository) EditMessage(ctx context.Context, id string, newContent string) error {
	query := `
		UPDATE "zpMessages" SET
			content = $1,
			"editTimestamp" = CURRENT_TIMESTAMP,
			"updatedAt" = CURRENT_TIMESTAMP
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
		UPDATE "zpMessages" SET
			"isDeleted" = TRUE,
			"deletedAt" = CURRENT_TIMESTAMP,
			"updatedAt" = CURRENT_TIMESTAMP
		WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}
	return nil
}

// AddReaction adiciona uma reação a uma mensagem
func (r *MessageRepository) AddReaction(ctx context.Context, id string, reaction string) error {
	query := `UPDATE "zpMessages" SET reaction = $1, "updatedAt" = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, reaction, id)
	if err != nil {
		return fmt.Errorf("failed to add reaction: %w", err)
	}
	return nil
}

// RemoveReaction remove uma reação de uma mensagem
func (r *MessageRepository) RemoveReaction(ctx context.Context, id string) error {
	query := `UPDATE "zpMessages" SET reaction = NULL, "updatedAt" = CURRENT_TIMESTAMP WHERE id = $1`
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
		SELECT id, "chatId", "sessionId", msgId, "msgType", content,
			   media_url, media_mime_type, media_size, media_filename, thumbnail_url,
			   "senderJid", "senderName", "isFromMe", "isForwarded", "isBroadcast",
			   "quotedMsgId", "quotedContent", status, timestamp, "editTimestamp",
			   "isDeleted", "deletedAt", reaction, metadata, "createdAt", "updatedAt"
		FROM "zpMessages" 
		WHERE "chatId" = $1 AND "isFromMe" = FALSE AND status != 'read' AND "isDeleted" = FALSE
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

	query, args, err := sqlx.In(`UPDATE "zpMessages" SET status = 'read', "updatedAt" = CURRENT_TIMESTAMP WHERE id IN (?)`, messageIDs)
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
		SELECT id, "chatId", "sessionId", msgId, "msgType", content,
			   media_url, media_mime_type, media_size, media_filename, thumbnail_url,
			   "senderJid", "senderName", "isFromMe", "isForwarded", "isBroadcast",
			   "quotedMsgId", "quotedContent", status, timestamp, "editTimestamp",
			   "isDeleted", "deletedAt", reaction, metadata, "createdAt", "updatedAt"
		FROM "zpMessages" 
		WHERE "chatId" = $1 AND timestamp BETWEEN $2 AND $3 AND "isDeleted" = FALSE
		ORDER BY timestamp DESC
		LIMIT $4 OFFSET $5`

	err := r.db.SelectContext(ctx, &messages, query, chatID, startTime, endTime, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages by time range: %w", err)
	}

	return messages, nil
}
