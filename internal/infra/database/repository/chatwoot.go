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
		INSERT INTO chatwoot (
			session_id, enabled, account_id, token, url, name_inbox,
			sign_msg, sign_delimiter, number, reopen_conversation,
			conversation_pending, merge_brazil_contacts, import_contacts,
			import_messages, days_limit_import_messages, auto_create,
			organization, logo, ignore_jids, sync_status
		) VALUES (
			:session_id, :enabled, :account_id, :token, :url, :name_inbox,
			:sign_msg, :sign_delimiter, :number, :reopen_conversation,
			:conversation_pending, :merge_brazil_contacts, :import_contacts,
			:import_messages, :days_limit_import_messages, :auto_create,
			:organization, :logo, :ignore_jids, :sync_status
		) RETURNING id, created_at, updated_at`

	rows, err := r.db.NamedQueryContext(ctx, query, config)
	if err != nil {
		return fmt.Errorf("failed to create chatwoot config: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&config.ID, &config.CreatedAt, &config.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan created chatwoot config: %w", err)
		}
	}

	return nil
}

// GetBySessionID busca configuração Chatwoot por ID da sessão
func (r *ChatwootRepository) GetBySessionID(ctx context.Context, sessionID string) (*models.ChatwootModel, error) {
	var config models.ChatwootModel
	query := `
		SELECT id, session_id, enabled, account_id, token, url, name_inbox,
			   sign_msg, sign_delimiter, number, reopen_conversation,
			   conversation_pending, merge_brazil_contacts, import_contacts,
			   import_messages, days_limit_import_messages, auto_create,
			   organization, logo, ignore_jids, inbox_id, inbox_name,
			   last_sync, sync_status, error_message, messages_count,
			   contacts_count, conversations_count, created_at, updated_at
		FROM chatwoot 
		WHERE session_id = $1`

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
		SELECT id, session_id, enabled, account_id, token, url, name_inbox,
			   sign_msg, sign_delimiter, number, reopen_conversation,
			   conversation_pending, merge_brazil_contacts, import_contacts,
			   import_messages, days_limit_import_messages, auto_create,
			   organization, logo, ignore_jids, inbox_id, inbox_name,
			   last_sync, sync_status, error_message, messages_count,
			   contacts_count, conversations_count, created_at, updated_at
		FROM chatwoot 
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
		UPDATE chatwoot SET
			enabled = :enabled,
			account_id = :account_id,
			token = :token,
			url = :url,
			name_inbox = :name_inbox,
			sign_msg = :sign_msg,
			sign_delimiter = :sign_delimiter,
			number = :number,
			reopen_conversation = :reopen_conversation,
			conversation_pending = :conversation_pending,
			merge_brazil_contacts = :merge_brazil_contacts,
			import_contacts = :import_contacts,
			import_messages = :import_messages,
			days_limit_import_messages = :days_limit_import_messages,
			auto_create = :auto_create,
			organization = :organization,
			logo = :logo,
			ignore_jids = :ignore_jids,
			inbox_id = :inbox_id,
			inbox_name = :inbox_name,
			last_sync = :last_sync,
			sync_status = :sync_status,
			error_message = :error_message,
			messages_count = :messages_count,
			contacts_count = :contacts_count,
			conversations_count = :conversations_count,
			updated_at = NOW()
		WHERE session_id = :session_id
		RETURNING updated_at`

	rows, err := r.db.NamedQueryContext(ctx, query, config)
	if err != nil {
		return fmt.Errorf("failed to update chatwoot config: %w", err)
	}
	defer rows.Close()

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
	query := `DELETE FROM chatwoot WHERE session_id = $1`
	
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
func (r *ChatwootRepository) List(ctx context.Context, enabled *bool) ([]*models.ChatwootModel, error) {
	var configs []*models.ChatwootModel
	
	query := `
		SELECT c.id, c.session_id, c.enabled, c.account_id, c.token, c.url, c.name_inbox,
			   c.sign_msg, c.sign_delimiter, c.number, c.reopen_conversation,
			   c.conversation_pending, c.merge_brazil_contacts, c.import_contacts,
			   c.import_messages, c.days_limit_import_messages, c.auto_create,
			   c.organization, c.logo, c.ignore_jids, c.inbox_id, c.inbox_name,
			   c.last_sync, c.sync_status, c.error_message, c.messages_count,
			   c.contacts_count, c.conversations_count, c.created_at, c.updated_at
		FROM chatwoot c
		INNER JOIN sessions s ON c.session_id = s.id`

	args := []interface{}{}
	
	if enabled != nil {
		query += ` WHERE c.enabled = $1`
		args = append(args, *enabled)
	}
	
	query += ` ORDER BY c.created_at DESC`

	err := r.db.SelectContext(ctx, &configs, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list chatwoot configs: %w", err)
	}

	return configs, nil
}

// ListWithSessionInfo lista configurações com informações da sessão
func (r *ChatwootRepository) ListWithSessionInfo(ctx context.Context, enabled *bool) ([]*ChatwootWithSession, error) {
	var configs []*ChatwootWithSession
	
	query := `
		SELECT c.id, c.session_id, c.enabled, c.account_id, c.token, c.url, c.name_inbox,
			   c.sign_msg, c.sign_delimiter, c.number, c.reopen_conversation,
			   c.conversation_pending, c.merge_brazil_contacts, c.import_contacts,
			   c.import_messages, c.days_limit_import_messages, c.auto_create,
			   c.organization, c.logo, c.ignore_jids, c.inbox_id, c.inbox_name,
			   c.last_sync, c.sync_status, c.error_message, c.messages_count,
			   c.contacts_count, c.conversations_count, c.created_at, c.updated_at,
			   s.name as session_name, s.status as session_status, s.connected as session_connected
		FROM chatwoot c
		INNER JOIN sessions s ON c.session_id = s.id`

	args := []interface{}{}
	
	if enabled != nil {
		query += ` WHERE c.enabled = $1`
		args = append(args, *enabled)
	}
	
	query += ` ORDER BY c.created_at DESC`

	err := r.db.SelectContext(ctx, &configs, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list chatwoot configs with session info: %w", err)
	}

	return configs, nil
}

// UpdateSyncStatus atualiza apenas o status de sincronização
func (r *ChatwootRepository) UpdateSyncStatus(ctx context.Context, sessionID, status string, errorMsg *string) error {
	query := `
		UPDATE chatwoot SET
			sync_status = $2,
			error_message = $3,
			last_sync = NOW(),
			updated_at = NOW()
		WHERE session_id = $1`

	_, err := r.db.ExecContext(ctx, query, sessionID, status, errorMsg)
	if err != nil {
		return fmt.Errorf("failed to update sync status: %w", err)
	}

	return nil
}

// UpdateMetrics atualiza as métricas de uma configuração
func (r *ChatwootRepository) UpdateMetrics(ctx context.Context, sessionID string, messagesCount, contactsCount, conversationsCount int) error {
	query := `
		UPDATE chatwoot SET
			messages_count = $2,
			contacts_count = $3,
			conversations_count = $4,
			updated_at = NOW()
		WHERE session_id = $1`

	_, err := r.db.ExecContext(ctx, query, sessionID, messagesCount, contactsCount, conversationsCount)
	if err != nil {
		return fmt.Errorf("failed to update metrics: %w", err)
	}

	return nil
}

// UpdateInboxInfo atualiza informações da inbox após criação
func (r *ChatwootRepository) UpdateInboxInfo(ctx context.Context, sessionID string, inboxID int, inboxName string) error {
	query := `
		UPDATE chatwoot SET
			inbox_id = $2,
			inbox_name = $3,
			updated_at = NOW()
		WHERE session_id = $1`

	_, err := r.db.ExecContext(ctx, query, sessionID, inboxID, inboxName)
	if err != nil {
		return fmt.Errorf("failed to update inbox info: %w", err)
	}

	return nil
}

// GetEnabledConfigs retorna apenas configurações habilitadas
func (r *ChatwootRepository) GetEnabledConfigs(ctx context.Context) ([]*models.ChatwootModel, error) {
	enabled := true
	return r.List(ctx, &enabled)
}

// GetMetrics retorna métricas agregadas
func (r *ChatwootRepository) GetMetrics(ctx context.Context) (*ChatwootMetrics, error) {
	var metrics ChatwootMetrics
	
	query := `
		SELECT 
			COUNT(*) as total_configs,
			COUNT(CASE WHEN enabled = true THEN 1 END) as enabled_configs,
			COUNT(CASE WHEN enabled = true AND inbox_id IS NOT NULL THEN 1 END) as connected_configs,
			COALESCE(SUM(messages_count), 0) as total_messages,
			COALESCE(SUM(contacts_count), 0) as total_contacts,
			COALESCE(SUM(conversations_count), 0) as total_conversations
		FROM chatwoot`

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
	TotalConfigs     int `db:"total_configs" json:"total_configs"`
	EnabledConfigs   int `db:"enabled_configs" json:"enabled_configs"`
	ConnectedConfigs int `db:"connected_configs" json:"connected_configs"`
	TotalMessages    int `db:"total_messages" json:"total_messages"`
	TotalContacts    int `db:"total_contacts" json:"total_contacts"`
	TotalConversations int `db:"total_conversations" json:"total_conversations"`
}
