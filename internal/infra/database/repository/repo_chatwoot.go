package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"zpmeow/internal/infra/database/models"
)

type ChatwootRepository struct {
	db *sqlx.DB
}

func NewChatwootRepository(db *sqlx.DB) *ChatwootRepository {
	return &ChatwootRepository{db: db}
}

// Create cria uma nova configuração Chatwoot
func (r *ChatwootRepository) Create(ctx context.Context, config *models.ChatwootModel) error {
	query := `
		INSERT INTO "zpChatwoot" (
			"sessionId", "isActive", "accountId", token, url, "nameInbox",
			"inboxId", "syncStatus", config
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		) RETURNING id, "createdAt", "updatedAt"`

	rows, err := r.db.QueryContext(ctx, query,
		config.SessionId, config.IsActive, config.AccountId, config.Token,
		config.URL, config.NameInbox, config.InboxId, config.SyncStatus, config.Config)
	if err != nil {
		return fmt.Errorf("failed to create chatwoot config: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close rows in CreateChatwootConfig: %v\n", closeErr)
		}
	}()

	if rows.Next() {
		err = rows.Scan(&config.ID, &config.CreatedAt, &config.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan created chatwoot config: %w", err)
		}
	}

	return nil
}

// GetBySessionId busca configuração Chatwoot por ID da sessão
func (r *ChatwootRepository) GetBySessionID(ctx context.Context, sessionID string) (*models.ChatwootModel, error) {
	var config models.ChatwootModel
	query := `
		SELECT id, "sessionId", "isActive", "accountId", token, url, "nameInbox",
			   "inboxId", "lastSync", "syncStatus", config, "createdAt", "updatedAt"
		FROM "zpChatwoot"
		WHERE "sessionId" = $1`

	err := r.db.GetContext(ctx, &config, query, sessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Não encontrado
		}
		return nil, fmt.Errorf("failed to get chatwoot config: %w", err)
	}

	return &config, nil
}

// GetByID busca configuração Chatwoot por ID
func (r *ChatwootRepository) GetByID(ctx context.Context, id string) (*models.ChatwootModel, error) {
	var config models.ChatwootModel
	query := `
		SELECT id, "sessionId", "isActive", "accountId", token, url, "nameInbox",
			   "inboxId", "lastSync", "syncStatus", config, "createdAt", "updatedAt"
		FROM "zpChatwoot"
		WHERE id = $1`

	err := r.db.GetContext(ctx, &config, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get chatwoot config by id: %w", err)
	}

	return &config, nil
}

// Update atualiza uma configuração Chatwoot existente
func (r *ChatwootRepository) Update(ctx context.Context, config *models.ChatwootModel) error {
	query := `
		UPDATE "zpChatwoot" SET
			"isActive" = $1,
			"accountId" = $2,
			token = $3,
			url = $4,
			"nameInbox" = $5,
			"inboxId" = $6,
			"lastSync" = $7,
			"syncStatus" = $8,
			config = $9,
			"updatedAt" = NOW()
		WHERE "sessionId" = $10
		RETURNING "updatedAt"`

	rows, err := r.db.QueryContext(ctx, query,
		config.IsActive, config.AccountId, config.Token, config.URL,
		config.NameInbox, config.InboxId, config.LastSync, config.SyncStatus,
		config.Config, config.SessionId)
	if err != nil {
		return fmt.Errorf("failed to update chatwoot config: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close rows in UpdateChatwootConfig: %v\n", closeErr)
		}
	}()

	if rows.Next() {
		err = rows.Scan(&config.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan updated timestamp: %w", err)
		}
	}

	return nil
}

// Delete remove uma configuração Chatwoot
func (r *ChatwootRepository) Delete(ctx context.Context, sessionID string) error {
	query := `DELETE FROM "zpChatwoot" WHERE "sessionId" = $1`

	result, err := r.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete chatwoot config: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("chatwoot config not found for session: %s", sessionID)
	}

	return nil
}

// List lista todas as configurações Chatwoot
func (r *ChatwootRepository) List(ctx context.Context, isActive *bool) ([]*models.ChatwootModel, error) {
	var configs []*models.ChatwootModel

	query := `
		SELECT c.id, c."sessionId", c."isActive", c."accountId", c.token, c.url, c."nameInbox",
			   c."inboxId", c."lastSync", c."syncStatus", c.config, c."createdAt", c."updatedAt"
		FROM "zpChatwoot" c
		INNER JOIN "zpSessions" s ON c."sessionId" = s.id`

	args := []interface{}{}

	if isActive != nil {
		query += ` WHERE c."isActive" = $1`
		args = append(args, *isActive)
	}

	query += ` ORDER BY c."createdAt" DESC`

	err := r.db.SelectContext(ctx, &configs, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list chatwoot configs: %w", err)
	}

	return configs, nil
}

// ListWithSessionInfo lista configurações com informações da sessão
func (r *ChatwootRepository) ListWithSessionInfo(ctx context.Context, isActive *bool) ([]*ChatwootWithSession, error) {
	var configs []*ChatwootWithSession

	query := `
		SELECT c.id, c."sessionId", c."isActive", c."accountId", c.token, c.url, c."nameInbox",
			   c.sign_msg, c.sign_delimiter, c.number, c.reopen_conversation,
			   c.conversation_pending, c.merge_brazil_contacts, c.import_contacts,
			   c.import_messages, c.days_limit_import_messages, c.auto_create,
			   c.organization, c.logo, c.ignore_jids, c."inboxId", c.inbox_name,
			   c."lastSync", c."syncStatus", c.error_message, c.messages_count,
			   c.contacts_count, c.conversations_count, c."createdAt", c."updatedAt",
			   s.name as session_name, s.status as session_status, s.connected as session_connected
		FROM "zpChatwoot" c
		INNER JOIN sessions s ON c."sessionId" = s.id`

	args := []interface{}{}

	if isActive != nil {
		query += ` WHERE c."isActive" = $1`
		args = append(args, *isActive)
	}

	query += ` ORDER BY c."createdAt" DESC`

	err := r.db.SelectContext(ctx, &configs, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list chatwoot configs with session info: %w", err)
	}

	return configs, nil
}

// UpdateSyncStatus atualiza apenas o status de sincronização
func (r *ChatwootRepository) UpdateSyncStatus(ctx context.Context, sessionID, status string, errorMsg *string) error {
	query := `
		UPDATE "zpChatwoot" SET
			"syncStatus" = $2,
			error_message = $3,
			"lastSync" = NOW(),
			"updatedAt" = NOW()
		WHERE "sessionId" = $1`

	_, err := r.db.ExecContext(ctx, query, sessionID, status, errorMsg)
	if err != nil {
		return fmt.Errorf("failed to update sync status: %w", err)
	}

	return nil
}

// UpdateMetrics atualiza as métricas de uma configuração
func (r *ChatwootRepository) UpdateMetrics(ctx context.Context, sessionID string, messagesCount, contactsCount, conversationsCount int) error {
	query := `
		UPDATE "zpChatwoot" SET
			messages_count = $2,
			contacts_count = $3,
			conversations_count = $4,
			"updatedAt" = NOW()
		WHERE "sessionId" = $1`

	_, err := r.db.ExecContext(ctx, query, sessionID, messagesCount, contactsCount, conversationsCount)
	if err != nil {
		return fmt.Errorf("failed to update metrics: %w", err)
	}

	return nil
}

// UpdateInboxInfo atualiza informações da inbox após criação
func (r *ChatwootRepository) UpdateInboxInfo(ctx context.Context, sessionID string, inboxID int, inboxName string) error {
	query := `
		UPDATE "zpChatwoot" SET
			"inboxId" = $2,
			inbox_name = $3,
			"updatedAt" = NOW()
		WHERE "sessionId" = $1`

	_, err := r.db.ExecContext(ctx, query, sessionID, inboxID, inboxName)
	if err != nil {
		return fmt.Errorf("failed to update inbox info: %w", err)
	}

	return nil
}

// GetActiveConfigs retorna apenas configurações ativas
func (r *ChatwootRepository) GetActiveConfigs(ctx context.Context) ([]*models.ChatwootModel, error) {
	isActive := true
	return r.List(ctx, &isActive)
}

// GetMetrics retorna métricas agregadas
func (r *ChatwootRepository) GetMetrics(ctx context.Context) (*ChatwootMetrics, error) {
	var metrics ChatwootMetrics

	query := `
		SELECT
			COUNT(*) as total_configs,
			COUNT(CASE WHEN "isActive" = true THEN 1 END) as active_configs,
			COUNT(CASE WHEN "isActive" = true AND "inboxId" IS NOT NULL THEN 1 END) as connected_configs,
			COALESCE(SUM(messages_count), 0) as total_messages,
			COALESCE(SUM(contacts_count), 0) as total_contacts,
			COALESCE(SUM(conversations_count), 0) as total_conversations
		FROM "zpChatwoot"`

	err := r.db.GetContext(ctx, &metrics, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get chatwoot metrics: %w", err)
	}

	return &metrics, nil
}

// Estruturas auxiliares

// ChatwootWithSession representa uma configuração Chatwoot com informações da sessão
type ChatwootWithSession struct {
	models.ChatwootModel
	SessionName      string `db:"session_name" json:"session_name"`
	SessionStatus    string `db:"session_status" json:"session_status"`
	SessionConnected bool   `db:"session_connected" json:"session_connected"`
}

// ChatwootMetrics representa métricas agregadas
type ChatwootMetrics struct {
	TotalConfigs       int `db:"total_configs" json:"total_configs"`
	ActiveConfigs      int `db:"active_configs" json:"active_configs"`
	ConnectedConfigs   int `db:"connected_configs" json:"connected_configs"`
	TotalMessages      int `db:"total_messages" json:"total_messages"`
	TotalContacts      int `db:"total_contacts" json:"total_contacts"`
	TotalConversations int `db:"total_conversations" json:"total_conversations"`
}
