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
		INSERT INTO "zpCwMessages" (
			"sessionId", "msgId", "chatwootMsgId", "chatwootConvId",
			direction, "syncStatus", "sourceId", metadata
		) VALUES (
			:sessionId, :msgId, :chatwootMsgId, :chatwootConvId,
			:direction, :syncStatus, :sourceId, :metadata
		) RETURNING id, "createdAt", "updatedAt"`

	rows, err := r.db.NamedQueryContext(ctx, query, relation)
	if err != nil {
		return fmt.Errorf("failed to create chatwoot message relation: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close rows in CreateRelation: %v\n", closeErr)
		}
	}()

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
		SELECT id, "sessionId", "msgId", "chatwootMsgId", "chatwootConvId",
			   direction, "syncStatus", "sourceId", metadata, "createdAt", "updatedAt"
		FROM "zpCwMessages"
		WHERE "msgId" = $1`

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
		SELECT id, "sessionId", "msgId", "chatwootMsgId", "chatwootConvId",
			   direction, "syncStatus", "sourceId", metadata, "createdAt", "updatedAt"
		FROM "zpCwMessages"
		WHERE "sessionId" = $1 AND "chatwootMsgId" = $2`

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
		SELECT id, "sessionId", "msgId", "chatwootMsgId", "chatwootConvId",
			   direction, "syncStatus", "sourceId", metadata, "createdAt", "updatedAt"
		FROM "zpCwMessages"
		WHERE "sessionId" = $1 AND "sourceId" = $2`

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
		SELECT id, "sessionId", "msgId", "chatwootMsgId", "chatwootConvId",
			   direction, "syncStatus", "sourceId", metadata, "createdAt", "updatedAt"
		FROM "zpCwMessages"
		WHERE "sessionId" = $1 AND metadata->>'echoId' = $2`

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
		UPDATE "zpCwMessages" SET
			"chatwootMsgId" = :chatwootMsgId,
			"chatwootConvId" = :chatwootConvId,
			direction = :direction,
			"syncStatus" = :syncStatus,
			"sourceId" = :sourceId,
			metadata = :metadata,
			"updatedAt" = CURRENT_TIMESTAMP
		WHERE id = :id
		RETURNING "updatedAt"`

	rows, err := r.db.NamedQueryContext(ctx, query, relation)
	if err != nil {
		return fmt.Errorf("failed to update chatwoot message relation: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close rows in UpdateRelation: %v\n", closeErr)
		}
	}()

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
		UPDATE "zpCwMessages" SET
			"syncStatus" = $1,
			metadata = COALESCE(metadata, '{}'::jsonb) || jsonb_build_object('syncError', $2),
			"updatedAt" = CURRENT_TIMESTAMP
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
		SELECT id, "sessionId", "msgId", "chatwootMsgId", "chatwootConvId",
			   direction, "syncStatus", "sourceId", metadata, "createdAt", "updatedAt"
		FROM "zpCwMessages"
		WHERE "chatwootConvId" = $1
		ORDER BY "createdAt" DESC
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
		SELECT id, "sessionId", "msgId", "chatwootMsgId", "chatwootConvId",
			   direction, "syncStatus", "sourceId", metadata, "createdAt", "updatedAt"
		FROM "zpCwMessages"
		WHERE "sessionId" = $1 AND "syncStatus" = 'failed'
		ORDER BY "createdAt" ASC
		LIMIT $2`

	err := r.db.SelectContext(ctx, &relations, query, sessionID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get failed sync relations: %w", err)
	}

	return relations, nil
}

// DeleteRelation deleta uma relação
func (r *ZpCwMessageRepository) DeleteRelation(ctx context.Context, id string) error {
	query := `DELETE FROM "zpCwMessages" WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete chatwoot message relation: %w", err)
	}
	return nil
}

// GetRelationStats busca estatísticas de sincronização
func (r *ZpCwMessageRepository) GetRelationStats(ctx context.Context, sessionID string) (map[string]int, error) {
	query := `
		SELECT "syncStatus", COUNT(*) as count
		FROM "zpCwMessages"
		WHERE "sessionId" = $1
		GROUP BY "syncStatus"`

	rows, err := r.db.QueryContext(ctx, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get relation stats: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close rows in GetRelationStats: %v\n", closeErr)
		}
	}()

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
	existing, err := r.GetRelationByZpmeowMessageID(ctx, relation.MsgId)
	if err != nil {
		return err
	}

	if existing == nil {
		return r.CreateRelation(ctx, relation)
	}

	// Atualiza a relação existente
	existing.ChatwootMsgId = relation.ChatwootMsgId
	existing.ChatwootConvId = relation.ChatwootConvId
	existing.Direction = relation.Direction
	existing.SyncStatus = relation.SyncStatus
	existing.SourceId = relation.SourceId
	// SyncError, LastSyncAt, ChatwootEchoID movidos para Metadata
	existing.Metadata = relation.Metadata

	*relation = *existing
	return r.UpdateRelation(ctx, relation)
}
