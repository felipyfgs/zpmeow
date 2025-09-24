package repository

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"strings"
	"time"

	"zpmeow/internal/domain/session"
	"zpmeow/internal/infra/database/models"

	"github.com/jmoiron/sqlx"
)

type PostgresRepo struct {
	db *sqlx.DB
}

func NewPostgresRepo(db *sqlx.DB) session.Repository {
	return &PostgresRepo{
		db: db,
	}
}

func (r *PostgresRepo) CreateWithGeneratedID(ctx context.Context, sessionEntity *session.Session) (string, error) {
	// Webhook events agora são gerenciados pela tabela zpWebhooks separada

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
		INSERT INTO "zpSessions" (name, "deviceJid", status, "qrCode", "proxyUrl", connected, "apiKey", "createdAt", "updatedAt")
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	isConnected := sessionEntity.Status() == session.StatusConnected && sessionEntity.DeviceJID().Value() != ""

	apiKey := sessionEntity.ApiKey().Value()
	if apiKey == "" || apiKey == "temp-key" {
		var err error
		apiKey, err = r.generateAPIKey()
		if err != nil {
			return "", fmt.Errorf("failed to generate API key: %w", err)
		}
	}

	var generatedID string
	err := r.db.QueryRowContext(ctx, query,
		sessionEntity.Name().Value(),
		sessionEntity.DeviceJID().Value(),
		string(sessionEntity.Status()),
		sessionEntity.QRCode().Value(),
		sessionEntity.ProxyConfiguration().Value(),
		isConnected,
		apiKey,
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

func (r *PostgresRepo) Create(ctx context.Context, sessionEntity *session.Session) error {
	_, err := r.CreateWithGeneratedID(ctx, sessionEntity)
	return err
}

func (r *PostgresRepo) Exists(ctx context.Context, name string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM "zpSessions" WHERE name = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, name).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check session existence: %w", err)
	}
	return exists, nil
}

func (r *PostgresRepo) GetByID(ctx context.Context, id string) (*session.Session, error) {
	var model models.SessionModel
	query := `
		SELECT id, name, "deviceJid", status, "qrCode", "proxyUrl", connected, "apiKey", "createdAt", "updatedAt"
		FROM "zpSessions" WHERE id = $1
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

func (r *PostgresRepo) GetByName(ctx context.Context, name string) (*session.Session, error) {
	var model models.SessionModel
	query := `
		SELECT id, name, "deviceJid", status, "qrCode", "proxyUrl", connected, "apiKey", "createdAt", "updatedAt"
		FROM "zpSessions" WHERE name = $1
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

func (r *PostgresRepo) GetAll(ctx context.Context) ([]*session.Session, error) {
	var sessionModels []models.SessionModel
	query := `
		SELECT id, name, "deviceJid", status, "qrCode", "proxyUrl", connected, "apiKey", "createdAt", "updatedAt"
		FROM "zpSessions" ORDER BY "createdAt" DESC
	`

	err := r.db.SelectContext(ctx, &sessionModels, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all sessions: %w", err)
	}

	sessions := make([]*session.Session, len(sessionModels))
	for i, model := range sessionModels {
		session, err := r.modelToDomain(&model)
		if err != nil {
			return nil, err
		}
		sessions[i] = session
	}

	return sessions, nil
}

func (r *PostgresRepo) Update(ctx context.Context, sessionEntity *session.Session) error {
	// Webhook events agora são gerenciados pela tabela zpWebhooks separada
	updatedAt := time.Now()

	query := `
		UPDATE "zpSessions"
		SET name = $2, "deviceJid" = $3, status = $4, "qrCode" = $5, "proxyUrl" = $6,
		    connected = $7, "apiKey" = $8, "updatedAt" = $9
		WHERE id = $1
	`

	isConnected := sessionEntity.Status() == session.StatusConnected && sessionEntity.DeviceJID().Value() != ""

	result, err := r.db.ExecContext(ctx, query,
		sessionEntity.SessionID().Value(),
		sessionEntity.Name().Value(),
		sessionEntity.DeviceJID().Value(),
		string(sessionEntity.Status()),
		sessionEntity.QRCode().Value(),
		sessionEntity.ProxyConfiguration().Value(),
		isConnected,
		sessionEntity.ApiKey().Value(),
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
	query := `DELETE FROM "zpSessions" WHERE id = $1`

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

func (r *PostgresRepo) List(ctx context.Context, limit, offset int, status string) ([]*session.Session, int, error) {
	var sessionModels []models.SessionModel
	var totalCount int

	countQuery := `SELECT COUNT(*) FROM "zpSessions"`
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
		SELECT id, name, "deviceJid", status, "qrCode", "proxyUrl", connected, "apiKey", "createdAt", "updatedAt"
		FROM "zpSessions"
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

	sessions := make([]*session.Session, len(sessionModels))
	for i, model := range sessionModels {
		session, err := r.modelToDomain(&model)
		if err != nil {
			return nil, 0, err
		}
		sessions[i] = session
	}

	return sessions, totalCount, nil
}

func (r *PostgresRepo) GetActive(ctx context.Context) ([]*session.Session, error) {
	var sessionModels []models.SessionModel
	query := `
		SELECT id, name, "deviceJid", status, "qrCode", "proxyUrl", connected, "apiKey", "createdAt", "updatedAt"
		FROM "zpSessions" WHERE "deviceJid" IS NOT NULL AND "deviceJid" != '' ORDER BY "createdAt" DESC
	`

	err := r.db.SelectContext(ctx, &sessionModels, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions with credentials: %w", err)
	}

	sessions := make([]*session.Session, len(sessionModels))
	for i, model := range sessionModels {
		session, err := r.modelToDomain(&model)
		if err != nil {
			return nil, err
		}
		sessions[i] = session
	}

	return sessions, nil
}

func (r *PostgresRepo) GetInactive(ctx context.Context) ([]*session.Session, error) {
	var sessionModels []models.SessionModel
	query := `
		SELECT id, name, "deviceJid", status, "qrCode", "proxyUrl", connected, "apiKey", "createdAt", "updatedAt"
		FROM "zpSessions" WHERE status != $1 ORDER BY "createdAt" DESC
	`

	err := r.db.SelectContext(ctx, &sessionModels, query, string("connected"))
	if err != nil {
		return nil, fmt.Errorf("failed to get inactive sessions: %w", err)
	}

	sessions := make([]*session.Session, len(sessionModels))
	for i, model := range sessionModels {
		session, err := r.modelToDomain(&model)
		if err != nil {
			return nil, err
		}
		sessions[i] = session
	}

	return sessions, nil
}

func (r *PostgresRepo) GetByApiKey(ctx context.Context, apiKey string) (*session.Session, error) {
	var model models.SessionModel
	query := `
		SELECT id, name, "deviceJid", status, "qrCode", "proxyUrl", connected, "apiKey", "createdAt", "updatedAt"
		FROM "zpSessions" WHERE "apiKey" = $1
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

func (r *PostgresRepo) GetByDeviceJID(ctx context.Context, deviceJID string) (*session.Session, error) {
	if deviceJID == "" {
		return nil, fmt.Errorf("device JID cannot be empty")
	}

	var model models.SessionModel
	query := `
		SELECT id, name, "deviceJid", status, "qrCode", "proxyUrl", connected, "apiKey", "createdAt", "updatedAt"
		FROM "zpSessions" WHERE "deviceJid" = $1
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
		return nil
	}

	var count int
	query := `
		SELECT COUNT(*) FROM "zpSessions"
		WHERE "deviceJid" = $1 AND id != $2
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

// SessionBuilder ajuda na construção de entidades Session a partir de modelos
type SessionBuilder struct {
	model *models.SessionModel
}

// NewSessionBuilder cria um novo builder
func NewSessionBuilder(model *models.SessionModel) *SessionBuilder {
	return &SessionBuilder{model: model}
}

// buildValueObjects cria os value objects necessários
func (b *SessionBuilder) buildValueObjects() (session.SessionID, session.SessionName, error) {
	sessionID, err := session.NewSessionID(b.model.ID)
	if err != nil {
		return session.SessionID{}, session.SessionName{}, fmt.Errorf("failed to create session ID: %w", err)
	}

	sessionName, err := session.NewSessionName(b.model.Name)
	if err != nil {
		return session.SessionID{}, session.SessionName{}, fmt.Errorf("failed to create session name: %w", err)
	}

	return sessionID, sessionName, nil
}

// buildOptionalFields cria os campos opcionais
func (b *SessionBuilder) buildOptionalFields() (session.ProxyConfiguration, session.DeviceJID, session.QRCode, session.ApiKey, error) {
	var proxyURL session.ProxyConfiguration
	var deviceJID session.DeviceJID
	var qrCode session.QRCode
	var apiKey session.ApiKey

	// Proxy URL
	if b.model.ProxyUrl != "" {
		proxy, err := session.NewProxyConfiguration(b.model.ProxyUrl)
		if err != nil {
			return proxyURL, deviceJID, qrCode, apiKey, fmt.Errorf("failed to create proxy configuration: %w", err)
		}
		proxyURL = proxy
	}

	// Device JID
	if b.model.DeviceJid != "" {
		jid, err := session.NewDeviceJID(b.model.DeviceJid)
		if err != nil {
			return proxyURL, deviceJID, qrCode, apiKey, fmt.Errorf("failed to create device JID: %w", err)
		}
		deviceJID = jid
	}

	// QR Code
	if b.model.QrCode != "" {
		qr, err := session.NewQRCode(b.model.QrCode)
		if err != nil {
			return proxyURL, deviceJID, qrCode, apiKey, fmt.Errorf("failed to create QR code: %w", err)
		}
		qrCode = qr
	}

	// API Key
	if b.model.ApiKey != "" {
		key, err := session.NewApiKey(b.model.ApiKey)
		if err != nil {
			return proxyURL, deviceJID, qrCode, apiKey, fmt.Errorf("failed to create API key: %w", err)
		}
		apiKey = key
	}

	return proxyURL, deviceJID, qrCode, apiKey, nil
}

// configureStatus configura o status da sessão
func (b *SessionBuilder) configureStatus(sessionEntity *session.Session) error {
	if b.model.Status != string(session.StatusDisconnected) {
		switch session.Status(b.model.Status) {
		case session.StatusConnecting:
			if err := sessionEntity.Connect(); err != nil {
				return fmt.Errorf("failed to set connecting status: %w", err)
			}
		case session.StatusConnected:
			if err := sessionEntity.Connect(); err != nil {
				return fmt.Errorf("failed to set connecting status: %w", err)
			}
			if err := sessionEntity.SetConnected(); err != nil {
				return fmt.Errorf("failed to set connected status: %w", err)
			}
		case session.StatusError:
			sessionEntity.SetError("Restored from database")
		}
	}
	return nil
}

// configureOptionalProperties configura propriedades opcionais
func (b *SessionBuilder) configureOptionalProperties(sessionEntity *session.Session, proxyURL session.ProxyConfiguration, deviceJID session.DeviceJID, qrCode session.QRCode, apiKey session.ApiKey) error {
	if !proxyURL.IsEmpty() {
		if err := sessionEntity.SetProxyConfiguration(proxyURL.Value()); err != nil {
			return fmt.Errorf("failed to set proxy URL: %w", err)
		}
	}

	if !deviceJID.IsEmpty() {
		if err := sessionEntity.Authenticate(deviceJID.Value()); err != nil {
			return fmt.Errorf("failed to set device JID: %w", err)
		}
	}

	if !qrCode.IsEmpty() {
		if err := sessionEntity.SetQRCode(qrCode.Value()); err != nil {
			return fmt.Errorf("failed to set QR code: %w", err)
		}
	}

	if !apiKey.IsEmpty() {
		if err := sessionEntity.SetApiKey(apiKey.Value()); err != nil {
			return fmt.Errorf("failed to set API key: %w", err)
		}
	}

	return nil
}

func (r *PostgresRepo) modelToDomain(model *models.SessionModel) (*session.Session, error) {
	builder := NewSessionBuilder(model)

	// Cria value objects básicos
	sessionID, sessionName, err := builder.buildValueObjects()
	if err != nil {
		return nil, err
	}

	// Cria campos opcionais
	proxyURL, deviceJID, qrCode, apiKey, err := builder.buildOptionalFields()
	if err != nil {
		return nil, err
	}

	// Cria entidade Session
	sessionEntity, err := session.NewSession(sessionID.Value(), sessionName.Value())
	if err != nil {
		return nil, fmt.Errorf("failed to create session entity: %w", err)
	}

	// Configura status
	if err := builder.configureStatus(sessionEntity); err != nil {
		return nil, err
	}

	// Configura propriedades opcionais
	if err := builder.configureOptionalProperties(sessionEntity, proxyURL, deviceJID, qrCode, apiKey); err != nil {
		return nil, err
	}

	return sessionEntity, nil
}

func (r *PostgresRepo) generateAPIKey() (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const keyLength = 32

	b := make([]byte, keyLength)
	for i := range b {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		b[i] = charset[randomIndex.Int64()]
	}
	return string(b), nil
}
