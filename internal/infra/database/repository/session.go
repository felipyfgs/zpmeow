package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	sessiondomain "zpmeow/internal/domain/session"
	"zpmeow/internal/infra/database/models"

	"github.com/jmoiron/sqlx"
)

type PostgresRepo struct {
	db *sqlx.DB
}

func NewPostgresRepo(db *sqlx.DB) sessiondomain.Repository {
	return &PostgresRepo{
		db: db,
	}
}

func (r *PostgresRepo) CreateWithGeneratedID(ctx context.Context, sessionEntity *sessiondomain.Session) (string, error) {
	// Serialize webhook events to JSON
	eventsJSON := []byte("[]")
	if len(sessionEntity.GetWebhookEvents()) > 0 {
		if jsonBytes, err := json.Marshal(sessionEntity.GetWebhookEvents()); err == nil {
			eventsJSON = jsonBytes
		}
	}

	now := time.Now()
	createdAt := now
	updatedAt := now

	if !sessionEntity.CreatedAt().Value().IsZero() {
		createdAt = sessionEntity.CreatedAt().Value()
	}
	if !sessionEntity.UpdatedAt().Value().IsZero() {
		updatedAt = sessionEntity.UpdatedAt().Value()
	}

	query := `
		INSERT INTO sessions (name, device_jid, status, qr_code, proxy_url, webhook_url, webhook_events, connected, apikey, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`

	isConnected := sessionEntity.Status() == sessiondomain.StatusConnected && sessionEntity.GetDeviceJIDString() != ""

	apiKey := sessionEntity.ApiKey().Value()
	if apiKey == "" || apiKey == "temp-key" {
		apiKey = r.generateAPIKey()
	}

	var generatedID string
	err := r.db.QueryRowContext(ctx, query,
		sessionEntity.Name().Value(),
		sessionEntity.GetDeviceJIDString(),
		string(sessionEntity.Status()),
		sessionEntity.QRCode().Value(),
		sessionEntity.ProxyConfiguration().Value(), // ✅ CORREÇÃO: usar método correto
		sessionEntity.WebhookEndpoint().Value(),    // ✅ CORREÇÃO: usar método correto
		string(eventsJSON),
		isConnected,
		apiKey, // Use the generated or existing API key
		createdAt,
		updatedAt,
	).Scan(&generatedID)

	if err != nil {
		if strings.Contains(err.Error(), "unique_session_name") {
			return "", fmt.Errorf("session already exists")
		}
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return generatedID, nil
}

func (r *PostgresRepo) Create(ctx context.Context, sessionEntity *sessiondomain.Session) error {
	_, err := r.CreateWithGeneratedID(ctx, sessionEntity)
	return err
}

func (r *PostgresRepo) Exists(ctx context.Context, name string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM sessions WHERE name = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, name).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check session existence: %w", err)
	}
	return exists, nil
}

func (r *PostgresRepo) GetByID(ctx context.Context, id string) (*sessiondomain.Session, error) {
	var model models.SessionModel
	query := `
		SELECT id, name, device_jid, status, qr_code, proxy_url, webhook_url, webhook_events, connected, apikey, created_at, updated_at
		FROM sessions WHERE id = $1
	`

	err := r.db.GetContext(ctx, &model, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session by ID: %w", err)
	}

	return r.modelToDomain(&model)
}

func (r *PostgresRepo) GetByName(ctx context.Context, name string) (*sessiondomain.Session, error) {
	var model models.SessionModel
	query := `
		SELECT id, name, device_jid, status, qr_code, proxy_url, webhook_url, webhook_events, connected, apikey, created_at, updated_at
		FROM sessions WHERE name = $1
	`

	err := r.db.GetContext(ctx, &model, query, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session by name: %w", err)
	}

	return r.modelToDomain(&model)
}

func (r *PostgresRepo) GetAll(ctx context.Context) ([]*sessiondomain.Session, error) {
	var sessionModels []models.SessionModel
	query := `
		SELECT id, name, device_jid, status, qr_code, proxy_url, webhook_url, webhook_events, connected, apikey, created_at, updated_at
		FROM sessions ORDER BY created_at DESC
	`

	err := r.db.SelectContext(ctx, &sessionModels, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all sessions: %w", err)
	}

	sessions := make([]*sessiondomain.Session, len(sessionModels))
	for i, model := range sessionModels {
		session, err := r.modelToDomain(&model)
		if err != nil {
			return nil, err
		}
		sessions[i] = session
	}

	return sessions, nil
}

func (r *PostgresRepo) Update(ctx context.Context, session *sessiondomain.Session) error {
	// Serialize webhook events to JSON
	eventsJSON := []byte("[]")
	if len(session.GetWebhookEvents()) > 0 {
		if jsonBytes, err := json.Marshal(session.GetWebhookEvents()); err == nil {
			eventsJSON = jsonBytes
		}
	}
	updatedAt := time.Now()

	query := `
		UPDATE sessions
		SET name = $2, device_jid = $3, status = $4, qr_code = $5, proxy_url = $6,
		    webhook_url = $7, webhook_events = $8, connected = $9, apikey = $10, updated_at = $11
		WHERE id = $1
	`

	isConnected := session.Status() == sessiondomain.StatusConnected && session.GetDeviceJIDString() != ""

	result, err := r.db.ExecContext(ctx, query,
		session.SessionID().Value(),
		session.Name().Value(),
		session.GetDeviceJIDString(),
		string(session.Status()),
		session.QRCode().Value(),
		session.ProxyConfiguration().Value(),
		session.WebhookEndpoint().Value(),
		string(eventsJSON),
		isConnected,
		session.ApiKey().Value(),
		updatedAt,
	)

	if err != nil {
		if strings.Contains(err.Error(), "unique_session_name") {
			return fmt.Errorf("session already exists")
		}
		return fmt.Errorf("failed to update session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}

func (r *PostgresRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM sessions WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}

func (r *PostgresRepo) List(ctx context.Context, limit, offset int, status string) ([]*sessiondomain.Session, int, error) {
	var sessionModels []models.SessionModel
	var totalCount int

	countQuery := `SELECT COUNT(*) FROM sessions`
	args := []interface{}{}
	argIndex := 1

	if status != "" {
		countQuery += fmt.Sprintf(" WHERE status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	err := r.db.GetContext(ctx, &totalCount, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count sessions: %w", err)
	}

	query := `
		SELECT id, name, device_jid, status, qr_code, proxy_url, webhook_url, webhook_events, connected, apikey, created_at, updated_at
		FROM sessions
	`

	if status != "" {
		query += fmt.Sprintf(" WHERE status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	err = r.db.SelectContext(ctx, &sessionModels, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list sessions: %w", err)
	}

	sessions := make([]*sessiondomain.Session, len(sessionModels))
	for i, model := range sessionModels {
		session, err := r.modelToDomain(&model)
		if err != nil {
			return nil, 0, err
		}
		sessions[i] = session
	}

	return sessions, totalCount, nil
}

func (r *PostgresRepo) GetActive(ctx context.Context) ([]*sessiondomain.Session, error) {
	var sessionModels []models.SessionModel
	query := `
		SELECT id, name, device_jid, status, qr_code, proxy_url, webhook_url, webhook_events, connected, apikey, created_at, updated_at
		FROM sessions WHERE device_jid IS NOT NULL AND device_jid != '' ORDER BY created_at DESC
	`

	err := r.db.SelectContext(ctx, &sessionModels, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions with credentials: %w", err)
	}

	sessions := make([]*sessiondomain.Session, len(sessionModels))
	for i, model := range sessionModels {
		session, err := r.modelToDomain(&model)
		if err != nil {
			return nil, err
		}
		sessions[i] = session
	}

	return sessions, nil
}

func (r *PostgresRepo) GetInactive(ctx context.Context) ([]*sessiondomain.Session, error) {
	var sessionModels []models.SessionModel
	query := `
		SELECT id, name, device_jid, status, qr_code, proxy_url, webhook_url, webhook_events, connected, apikey, created_at, updated_at
		FROM sessions WHERE status != $1 ORDER BY created_at DESC
	`

	err := r.db.SelectContext(ctx, &sessionModels, query, string("connected"))
	if err != nil {
		return nil, fmt.Errorf("failed to get inactive sessions: %w", err)
	}

	sessions := make([]*sessiondomain.Session, len(sessionModels))
	for i, model := range sessionModels {
		session, err := r.modelToDomain(&model)
		if err != nil {
			return nil, err
		}
		sessions[i] = session
	}

	return sessions, nil
}

func (r *PostgresRepo) GetByApiKey(ctx context.Context, apiKey string) (*sessiondomain.Session, error) {
	var model models.SessionModel
	query := `
		SELECT id, name, device_jid, status, qr_code, proxy_url, webhook_url, webhook_events, connected, apikey, created_at, updated_at
		FROM sessions WHERE apikey = $1
	`

	err := r.db.GetContext(ctx, &model, query, apiKey)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session by API key: %w", err)
	}

	return r.modelToDomain(&model)
}

func (r *PostgresRepo) GetByDeviceJID(ctx context.Context, deviceJID string) (*sessiondomain.Session, error) {
	if deviceJID == "" {
		return nil, fmt.Errorf("device JID cannot be empty")
	}

	var model models.SessionModel
	query := `
		SELECT id, name, device_jid, status, qr_code, proxy_url, webhook_url, webhook_events, connected, apikey, created_at, updated_at
		FROM sessions WHERE device_jid = $1
	`

	err := r.db.GetContext(ctx, &model, query, deviceJID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session by device JID: %w", err)
	}

	return r.modelToDomain(&model)
}

func (r *PostgresRepo) ValidateDeviceUniqueness(ctx context.Context, sessionID, deviceJID string) error {
	if deviceJID == "" {
		return nil // Empty device JID is allowed (not connected yet)
	}

	var count int
	query := `
		SELECT COUNT(*) FROM sessions
		WHERE device_jid = $1 AND id != $2
	`

	err := r.db.GetContext(ctx, &count, query, deviceJID, sessionID)
	if err != nil {
		return fmt.Errorf("failed to validate device uniqueness: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("device JID %s is already in use by another session", deviceJID)
	}

	return nil
}

func (r *PostgresRepo) modelToDomain(model *models.SessionModel) (*sessiondomain.Session, error) {

	sessionID, err := sessiondomain.NewSessionID(model.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create session ID: %w", err)
	}

	sessionName, err := sessiondomain.NewSessionName(model.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to create session name: %w", err)
	}

	var proxyURL sessiondomain.ProxyConfiguration
	if model.ProxyURL != "" {
		proxy, err := sessiondomain.NewProxyConfiguration(model.ProxyURL)
		if err != nil {
			return nil, fmt.Errorf("failed to create proxy configuration: %w", err)
		}
		proxyURL = proxy
	}

	var DeviceJID sessiondomain.DeviceJID
	if model.DeviceJID != "" {
		jid, err := sessiondomain.NewDeviceJID(model.DeviceJID)
		if err != nil {
			return nil, fmt.Errorf("failed to create device JID: %w", err)
		}
		DeviceJID = jid
	}

	var qrCode sessiondomain.QRCode
	if model.QRCode != "" {
		qr, err := sessiondomain.NewQRCode(model.QRCode)
		if err != nil {
			return nil, fmt.Errorf("failed to create QR code: %w", err)
		}
		qrCode = qr
	}

	var apiKey sessiondomain.ApiKey
	if model.ApiKey != "" {
		key, err := sessiondomain.NewApiKey(model.ApiKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create API key: %w", err)
		}
		apiKey = key
	}

	var events []string
	if model.Events != "" && model.Events != "[]" {
		if err := json.Unmarshal([]byte(model.Events), &events); err != nil {
			return nil, fmt.Errorf("failed to unmarshal webhook events: %w", err)
		}
	}

	sessionEntity, err := sessiondomain.NewSession(sessionID.Value(), sessionName.Value())
	if err != nil {
		return nil, fmt.Errorf("failed to create session entity: %w", err)
	}

	if model.Status != string(sessiondomain.StatusDisconnected) {
		switch sessiondomain.Status(model.Status) {
		case sessiondomain.StatusConnecting:
			if err := sessionEntity.Connect(); err != nil {
				return nil, fmt.Errorf("failed to set connecting status: %w", err)
			}
		case sessiondomain.StatusConnected:
			if err := sessionEntity.Connect(); err != nil {
				return nil, fmt.Errorf("failed to set connecting status: %w", err)
			}
			if err := sessionEntity.SetConnected(); err != nil {
				return nil, fmt.Errorf("failed to set connected status: %w", err)
			}
		case sessiondomain.StatusError:
			sessionEntity.SetError("Restored from database")
		}
	}

	if !proxyURL.IsEmpty() {
		if err := sessionEntity.SetProxyConfiguration(proxyURL.Value()); err != nil {
			return nil, fmt.Errorf("failed to set proxy URL: %w", err)
		}
	}

	if !DeviceJID.IsEmpty() {
		if err := sessionEntity.Authenticate(DeviceJID.Value()); err != nil {
			return nil, fmt.Errorf("failed to set device JID: %w", err)
		}
	}

	if !qrCode.IsEmpty() {
		if err := sessionEntity.SetQRCode(qrCode.Value()); err != nil {
			return nil, fmt.Errorf("failed to set QR code: %w", err)
		}
	}

	if !apiKey.IsEmpty() {
		if err := sessionEntity.SetApiKey(apiKey.Value()); err != nil {
			return nil, fmt.Errorf("failed to set API key: %w", err)
		}
	}

	if model.WebhookURL != "" {
		if err := sessionEntity.SetWebhookEndpoint(model.WebhookURL); err != nil {
			return nil, fmt.Errorf("failed to set webhook URL: %w", err)
		}
	}

	// Set webhook events if they exist
	if len(events) > 0 {
		sessionEntity.SetWebhookEvents(events)
	}

	return sessionEntity, nil
}

func (r *PostgresRepo) generateAPIKey() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const keyLength = 32

	b := make([]byte, keyLength)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
