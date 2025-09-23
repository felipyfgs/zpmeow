package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"zpmeow/internal/infra/database/models"
)

type ZpCwMessageRepository struct {
	db *sqlx.DB
}

func NewZpCwMessageRepository(db *sqlx.DB) *ZpCwMessageRepository {
	return &ZpCwMessageRepository{db: db}
}

// CreateRelation cria uma nova relação entre mensagem zpmeow e Chatwoot
func (r *ZpCwMessageRepository) CreateRelation(ctx context.Context, relation *models.ZpCwMessageModel) error {
	query := `
		INSERT INTO zp_cw_messages (
			session_id, zpmeow_message_id, chatwoot_message_id, chatwoot_conversation_id,
			chatwoot_account_id, direction, sync_status, sync_error, last_sync_at,
			chatwoot_source_id, chatwoot_echo_id, metadata
		) VALUES (
			:session_id, :zpmeow_message_id, :chatwoot_message_id, :chatwoot_conversation_id,
			:chatwoot_account_id, :direction, :sync_status, :sync_error, :last_sync_at,
			:chatwoot_source_id, :chatwoot_echo_id, :metadata
		) RETURNING id, created_at, updated_at`

	rows, err := r.db.NamedQueryContext(ctx, query, relation)
	if err != nil {
		return fmt.Errorf("failed to create chatwoot message relation: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&relation.ID, &relation.CreatedAt, &relation.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan created relation: %w", err)
		}
	}

	return nil
}

// GetRelationByZpmeowMessageID busca relação por zpmeow message ID
func (r *ZpCwMessageRepository) GetRelationByZpmeowMessageID(ctx context.Context, zpmeowMessageID string) (*models.ZpCwMessageModel, error) {
	var relation models.ZpCwMessageModel
	query := `
		SELECT id, session_id, zpmeow_message_id, chatwoot_message_id, chatwoot_conversation_id,
			   chatwoot_account_id, direction, sync_status, sync_error, last_sync_at,
			   chatwoot_source_id, chatwoot_echo_id, metadata, created_at, updated_at
		FROM zp_cw_messages 
		WHERE zpmeow_message_id = $1`

	err := r.db.GetContext(ctx, &relation, query, zpmeowMessageID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get relation by zpmeow message id: %w", err)
	}

	return &relation, nil
}

// GetRelationByChatwootMessageID busca relação por Chatwoot message ID
func (r *ZpCwMessageRepository) GetRelationByChatwootMessageID(ctx context.Context, sessionID string, chatwootMessageID string) (*models.ZpCwMessageModel, error) {
	var relation models.ZpCwMessageModel
	query := `
		SELECT id, session_id, zpmeow_message_id, chatwoot_message_id, chatwoot_conversation_id,
			   chatwoot_account_id, direction, sync_status, sync_error, last_sync_at,
			   chatwoot_source_id, chatwoot_echo_id, metadata, created_at, updated_at
		FROM zp_cw_messages 
		WHERE session_id = $1 AND chatwoot_message_id = $2`

	err := r.db.GetContext(ctx, &relation, query, sessionID, chatwootMessageID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get relation by chatwoot message id: %w", err)
	}

	return &relation, nil
}

// GetRelationByChatwootSourceID busca relação por Chatwoot source ID (WAID:message_id)
func (r *ZpCwMessageRepository) GetRelationByChatwootSourceID(ctx context.Context, sessionID, sourceID string) (*models.ZpCwMessageModel, error) {
	var relation models.ZpCwMessageModel
	query := `
		SELECT id, session_id, zpmeow_message_id, chatwoot_message_id, chatwoot_conversation_id,
			   chatwoot_account_id, direction, sync_status, sync_error, last_sync_at,
			   chatwoot_source_id, chatwoot_echo_id, metadata, created_at, updated_at
		FROM zp_cw_messages 
		WHERE session_id = $1 AND chatwoot_source_id = $2`

	err := r.db.GetContext(ctx, &relation, query, sessionID, sourceID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get relation by chatwoot source id: %w", err)
	}

	return &relation, nil
}

// GetRelationByChatwootEchoID busca relação por Chatwoot echo ID
func (r *ZpCwMessageRepository) GetRelationByChatwootEchoID(ctx context.Context, sessionID, echoID string) (*models.ZpCwMessageModel, error) {
	var relation models.ZpCwMessageModel
	query := `
		SELECT id, session_id, zpmeow_message_id, chatwoot_message_id, chatwoot_conversation_id,
			   chatwoot_account_id, direction, sync_status, sync_error, last_sync_at,
			   chatwoot_source_id, chatwoot_echo_id, metadata, created_at, updated_at
		FROM zp_cw_messages 
		WHERE session_id = $1 AND chatwoot_echo_id = $2`

	err := r.db.GetContext(ctx, &relation, query, sessionID, echoID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get relation by chatwoot echo id: %w", err)
	}

	return &relation, nil
}

// UpdateRelation atualiza uma relação
func (r *ZpCwMessageRepository) UpdateRelation(ctx context.Context, relation *models.ZpCwMessageModel) error {
	query := `
		UPDATE zp_cw_messages SET
			chatwoot_message_id = :chatwoot_message_id,
			chatwoot_conversation_id = :chatwoot_conversation_id,
			chatwoot_account_id = :chatwoot_account_id,
			direction = :direction,
			sync_status = :sync_status,
			sync_error = :sync_error,
			last_sync_at = :last_sync_at,
			chatwoot_source_id = :chatwoot_source_id,
			chatwoot_echo_id = :chatwoot_echo_id,
			metadata = :metadata,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = :id
		RETURNING updated_at`

	rows, err := r.db.NamedQueryContext(ctx, query, relation)
	if err != nil {
		return fmt.Errorf("failed to update chatwoot message relation: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&relation.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan updated relation: %w", err)
		}
	}

	return nil
}

// UpdateSyncStatus atualiza o status de sincronização
func (r *ZpCwMessageRepository) UpdateSyncStatus(ctx context.Context, id string, status string, syncError *string) error {
	query := `
		UPDATE zp_cw_messages SET 
			sync_status = $1, 
			sync_error = $2,
			last_sync_at = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP 
		WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, status, syncError, id)
	if err != nil {
		return fmt.Errorf("failed to update sync status: %w", err)
	}
	return nil
}

// GetRelationsByConversationID busca relações por conversation ID
func (r *ZpCwMessageRepository) GetRelationsByConversationID(ctx context.Context, conversationID string, limit, offset int) ([]*models.ZpCwMessageModel, error) {
	var relations []*models.ZpCwMessageModel
	query := `
		SELECT id, session_id, zpmeow_message_id, chatwoot_message_id, chatwoot_conversation_id,
			   chatwoot_account_id, direction, sync_status, sync_error, last_sync_at,
			   chatwoot_source_id, chatwoot_echo_id, metadata, created_at, updated_at
		FROM zp_cw_messages 
		WHERE chatwoot_conversation_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	err := r.db.SelectContext(ctx, &relations, query, conversationID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get relations by conversation id: %w", err)
	}

	return relations, nil
}

// GetFailedSyncRelations busca relações com falha na sincronização
func (r *ZpCwMessageRepository) GetFailedSyncRelations(ctx context.Context, sessionID string, limit int) ([]*models.ZpCwMessageModel, error) {
	var relations []*models.ZpCwMessageModel
	query := `
		SELECT id, session_id, zpmeow_message_id, chatwoot_message_id, chatwoot_conversation_id,
			   chatwoot_account_id, direction, sync_status, sync_error, last_sync_at,
			   chatwoot_source_id, chatwoot_echo_id, metadata, created_at, updated_at
		FROM zp_cw_messages 
		WHERE session_id = $1 AND sync_status = 'failed'
		ORDER BY created_at ASC
		LIMIT $2`

	err := r.db.SelectContext(ctx, &relations, query, sessionID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get failed sync relations: %w", err)
	}

	return relations, nil
}

// DeleteRelation deleta uma relação
func (r *ZpCwMessageRepository) DeleteRelation(ctx context.Context, id string) error {
	query := `DELETE FROM zp_cw_messages WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete chatwoot message relation: %w", err)
	}
	return nil
}

// GetRelationStats busca estatísticas de sincronização
func (r *ZpCwMessageRepository) GetRelationStats(ctx context.Context, sessionID string) (map[string]int, error) {
	query := `
		SELECT sync_status, COUNT(*) as count
		FROM zp_cw_messages 
		WHERE session_id = $1
		GROUP BY sync_status`

	rows, err := r.db.QueryContext(ctx, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get relation stats: %w", err)
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("failed to scan stats: %w", err)
		}
		stats[status] = count
	}

	return stats, nil
}

// CreateOrUpdateRelation cria ou atualiza uma relação
func (r *ZpCwMessageRepository) CreateOrUpdateRelation(ctx context.Context, relation *models.ZpCwMessageModel) error {
	// Tenta buscar relação existente por zpmeow_message_id
	existing, err := r.GetRelationByZpmeowMessageID(ctx, relation.ZpmeowMessageID)
	if err != nil {
		return err
	}

	if existing == nil {
		return r.CreateRelation(ctx, relation)
	}

	// Atualiza a relação existente
	existing.ChatwootMessageID = relation.ChatwootMessageID
	existing.ChatwootConversationID = relation.ChatwootConversationID
	existing.ChatwootAccountID = relation.ChatwootAccountID
	existing.Direction = relation.Direction
	existing.SyncStatus = relation.SyncStatus
	existing.SyncError = relation.SyncError
	existing.LastSyncAt = relation.LastSyncAt
	existing.ChatwootSourceID = relation.ChatwootSourceID
	existing.ChatwootEchoID = relation.ChatwootEchoID
	existing.Metadata = relation.Metadata

	*relation = *existing
	return r.UpdateRelation(ctx, relation)
}
