package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"zpmeow/internal/infra/database/models"
)

type WebhookRepository struct {
	db *sqlx.DB
}

func NewWebhookRepository(db *sqlx.DB) *WebhookRepository {
	return &WebhookRepository{db: db}
}

// Create cria uma nova configuração de webhook
func (r *WebhookRepository) Create(ctx context.Context, webhook *models.WebhookModel) error {
	query := `
		INSERT INTO "zpWebhooks" (
			"sessionId", url, events, "isActive"
		) VALUES (
			$1, $2, $3, $4
		) RETURNING id, "createdAt", "updatedAt"`

	rows, err := r.db.QueryContext(ctx, query,
		webhook.SessionId, webhook.URL, pq.Array(webhook.Events), webhook.IsActive)
	if err != nil {
		return fmt.Errorf("failed to create webhook config: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close rows in CreateWebhook: %v\n", closeErr)
		}
	}()

	if rows.Next() {
		err = rows.Scan(&webhook.ID, &webhook.CreatedAt, &webhook.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan created webhook config: %w", err)
		}
	}

	return nil
}

// GetBySessionID busca configuração de webhook por sessionID
func (r *WebhookRepository) GetBySessionID(ctx context.Context, sessionID string) (*models.WebhookModel, error) {
	var webhook models.WebhookModel
	query := `
		SELECT id, "sessionId", url, events, "isActive", "createdAt", "updatedAt"
		FROM "zpWebhooks"
		WHERE "sessionId" = $1`

	err := r.db.GetContext(ctx, &webhook, query, sessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Não encontrado
		}
		return nil, fmt.Errorf("failed to get webhook by sessionID: %w", err)
	}

	return &webhook, nil
}

// Update atualiza uma configuração de webhook existente
func (r *WebhookRepository) Update(ctx context.Context, webhook *models.WebhookModel) error {
	query := `
		UPDATE "zpWebhooks"
		SET url = $2, events = $3, "isActive" = $4, "updatedAt" = CURRENT_TIMESTAMP
		WHERE "sessionId" = $1
		RETURNING "updatedAt"`

	rows, err := r.db.QueryContext(ctx, query,
		webhook.SessionId, webhook.URL, pq.Array(webhook.Events), webhook.IsActive)
	if err != nil {
		return fmt.Errorf("failed to update webhook config: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close rows in UpdateWebhook: %v\n", closeErr)
		}
	}()

	if rows.Next() {
		err = rows.Scan(&webhook.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan updated webhook config: %w", err)
		}
	}

	return nil
}

// Upsert cria ou atualiza uma configuração de webhook
func (r *WebhookRepository) Upsert(ctx context.Context, webhook *models.WebhookModel) error {
	query := `
		INSERT INTO "zpWebhooks" (
			"sessionId", url, events, "isActive"
		) VALUES (
			$1, $2, $3, $4
		) ON CONFLICT ("sessionId") DO UPDATE SET
			url = EXCLUDED.url,
			events = EXCLUDED.events,
			"isActive" = EXCLUDED."isActive",
			"updatedAt" = CURRENT_TIMESTAMP
		RETURNING id, "createdAt", "updatedAt"`

	rows, err := r.db.QueryContext(ctx, query,
		webhook.SessionId, webhook.URL, pq.Array(webhook.Events), webhook.IsActive)
	if err != nil {
		return fmt.Errorf("failed to upsert webhook config: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close rows in UpsertWebhook: %v\n", closeErr)
		}
	}()

	if rows.Next() {
		err = rows.Scan(&webhook.ID, &webhook.CreatedAt, &webhook.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan upserted webhook config: %w", err)
		}
	}

	return nil
}

// Delete remove uma configuração de webhook
func (r *WebhookRepository) Delete(ctx context.Context, sessionID string) error {
	query := `DELETE FROM "zpWebhooks" WHERE "sessionId" = $1`

	result, err := r.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete webhook config: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("webhook config not found for sessionID: %s", sessionID)
	}

	return nil
}

// List lista todas as configurações de webhook
func (r *WebhookRepository) List(ctx context.Context, isActive *bool) ([]*models.WebhookModel, error) {
	var webhooks []*models.WebhookModel

	query := `
		SELECT w.id, w."sessionId", w.url, w.events, w."isActive", w."createdAt", w."updatedAt"
		FROM "zpWebhooks" w
		INNER JOIN "zpSessions" s ON w."sessionId" = s.id`

	args := []interface{}{}

	if isActive != nil {
		query += ` WHERE w."isActive" = $1`
		args = append(args, *isActive)
	}

	query += ` ORDER BY w."createdAt" DESC`

	err := r.db.SelectContext(ctx, &webhooks, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list webhook configs: %w", err)
	}

	return webhooks, nil
}

// GetActiveWebhooks busca todos os webhooks ativos
func (r *WebhookRepository) GetActiveWebhooks(ctx context.Context) ([]*models.WebhookModel, error) {
	isActive := true
	return r.List(ctx, &isActive)
}

// SetWebhookForSession define webhook para uma sessão (método de conveniência)
func (r *WebhookRepository) SetWebhookForSession(ctx context.Context, sessionID, url string, events []string) error {
	webhook := &models.WebhookModel{
		SessionId: sessionID,
		URL:       url,
		Events:    events,
		IsActive:  true,
	}

	return r.Upsert(ctx, webhook)
}

// GetWebhookForSession busca webhook para uma sessão (método de conveniência)
func (r *WebhookRepository) GetWebhookForSession(ctx context.Context, sessionID string) (string, []string, error) {
	webhook, err := r.GetBySessionID(ctx, sessionID)
	if err != nil {
		return "", nil, err
	}

	if webhook == nil || !webhook.IsActive {
		return "", nil, nil
	}

	return webhook.URL, webhook.Events, nil
}
