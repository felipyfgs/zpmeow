package session

import (
	"context"
	"fmt"
	"strings"

	"zpmeow/internal/application/common"
	"zpmeow/internal/application/ports"
	"zpmeow/internal/domain/session"
)

type GetSessionQuery struct {
	SessionID string
	Name      string
	ApiKey    string
}

func (q GetSessionQuery) Validate() error {
	if strings.TrimSpace(q.SessionID) == "" &&
		strings.TrimSpace(q.Name) == "" &&
		strings.TrimSpace(q.ApiKey) == "" {
		return common.NewValidationError("identifier", "", "at least one identifier (sessionID, name, or apiKey) is required")
	}

	return nil
}

type SessionView struct {
	SessionID          string
	Name               string
	Status             string
	DeviceJID          string
	QRCode             string
	ProxyConfiguration string
	WebhookEndpoint    string
	HasProxy           bool
	HasWebhook         bool
	IsConnected        bool
	IsAuthenticated    bool
	CreatedAt          string
	UpdatedAt          string
}

type GetSessionUseCase struct {
	sessionRepo session.Repository
	logger      ports.Logger
}

func NewGetSessionUseCase(
	sessionRepo session.Repository,
	logger ports.Logger,
) *GetSessionUseCase {
	return &GetSessionUseCase{
		sessionRepo: sessionRepo,
		logger:      logger,
	}
}

func (uc *GetSessionUseCase) Handle(ctx context.Context, query GetSessionQuery) (*SessionView, error) {
	if err := query.Validate(); err != nil {
		uc.logger.Warn(ctx, "Invalid get session query", "error", err)
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	var sessionEntity *session.Session
	var err error

	switch {
	case query.SessionID != "":
		sessionEntity, err = uc.sessionRepo.GetByID(ctx, query.SessionID)
		if err != nil {
			uc.logger.Error(ctx, "Failed to get session by ID", "sessionID", query.SessionID, "error", err)
			return nil, fmt.Errorf("failed to get session by ID: %w", err)
		}

	case query.Name != "":
		sessionEntity, err = uc.sessionRepo.GetByName(ctx, query.Name)
		if err != nil {
			uc.logger.Error(ctx, "Failed to get session by name", "name", query.Name, "error", err)
			return nil, fmt.Errorf("failed to get session by name: %w", err)
		}

	case query.ApiKey != "":
		sessionEntity, err = uc.sessionRepo.GetByApiKey(ctx, query.ApiKey)
		if err != nil {
			uc.logger.Error(ctx, "Failed to get session by API key", "error", err)
			return nil, fmt.Errorf("failed to get session by API key: %w", err)
		}
	}

	view := &SessionView{
		SessionID:          sessionEntity.SessionID().Value(),
		Name:               sessionEntity.Name().Value(),
		Status:             sessionEntity.Status().String(),
		DeviceJID:          sessionEntity.DeviceJID().Value(),
		QRCode:             sessionEntity.QRCode().Value(),
		ProxyConfiguration: sessionEntity.ProxyConfiguration().Value(),
		WebhookEndpoint:    sessionEntity.WebhookEndpoint().Value(),
		HasProxy:           sessionEntity.HasProxy(),
		HasWebhook:         sessionEntity.HasWebhook(),
		IsConnected:        sessionEntity.IsConnected(),
		IsAuthenticated:    sessionEntity.IsAuthenticated(),
		CreatedAt:          sessionEntity.CreatedAt().Value().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:          sessionEntity.UpdatedAt().Value().Format("2006-01-02T15:04:05Z07:00"),
	}

	uc.logger.Debug(ctx, "Session retrieved successfully", "sessionID", view.SessionID, "name", view.Name)

	return view, nil
}

type GetAllSessionsQuery struct {
}

func (q GetAllSessionsQuery) Validate() error {
	return nil
}

type GetAllSessionsUseCase struct {
	sessionRepo session.Repository
	logger      ports.Logger
}

func NewGetAllSessionsUseCase(
	sessionRepo session.Repository,
	logger ports.Logger,
) *GetAllSessionsUseCase {
	return &GetAllSessionsUseCase{
		sessionRepo: sessionRepo,
		logger:      logger,
	}
}

func (uc *GetAllSessionsUseCase) Handle(ctx context.Context, query GetAllSessionsQuery) ([]*SessionView, error) {
	if err := query.Validate(); err != nil {
		uc.logger.Warn(ctx, "Invalid get all sessions query", "error", err)
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	sessions, err := uc.sessionRepo.GetAll(ctx)
	if err != nil {
		uc.logger.Error(ctx, "Failed to get all sessions", "error", err)
		return nil, fmt.Errorf("failed to get all sessions: %w", err)
	}

	views := make([]*SessionView, len(sessions))
	for i, sessionEntity := range sessions {
		views[i] = &SessionView{
			SessionID:          sessionEntity.SessionID().Value(),
			Name:               sessionEntity.Name().Value(),
			Status:             sessionEntity.Status().String(),
			DeviceJID:          sessionEntity.DeviceJID().Value(),
			QRCode:             sessionEntity.QRCode().Value(),
			ProxyConfiguration: sessionEntity.ProxyConfiguration().Value(),
			WebhookEndpoint:    sessionEntity.WebhookEndpoint().Value(),
			HasProxy:           sessionEntity.HasProxy(),
			HasWebhook:         sessionEntity.HasWebhook(),
			IsConnected:        sessionEntity.IsConnected(),
			IsAuthenticated:    sessionEntity.IsAuthenticated(),
			CreatedAt:          sessionEntity.CreatedAt().Value().Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:          sessionEntity.UpdatedAt().Value().Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	uc.logger.Debug(ctx, "All sessions retrieved successfully", "count", len(views))

	return views, nil
}
