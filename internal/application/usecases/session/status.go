package session

import (
	"context"
	"fmt"
	"strings"

	"zpmeow/internal/application/common"
	"zpmeow/internal/application/ports"
	"zpmeow/internal/domain/session"
)

type GetSessionStatusQuery struct {
	SessionID string
}

func (q GetSessionStatusQuery) Validate() error {
	if strings.TrimSpace(q.SessionID) == "" {
		return common.NewValidationError("sessionID", q.SessionID, "session ID is required")
	}

	return nil
}

type SessionStatusView struct {
	SessionID       string
	Status          string
	IsConnected     bool
	IsAuthenticated bool
	DeviceJID       string
	QRCode          string
	LastSeen        string
	ConnectionInfo  map[string]interface{}
}

type GetSessionStatusUseCase struct {
	sessionRepo     session.Repository
	whatsappService ports.WhatsAppService
	logger          ports.Logger
}

func NewGetSessionStatusUseCase(
	sessionRepo session.Repository,
	whatsappService ports.WhatsAppService,
	logger ports.Logger,
) *GetSessionStatusUseCase {
	return &GetSessionStatusUseCase{
		sessionRepo:     sessionRepo,
		whatsappService: whatsappService,
		logger:          logger,
	}
}

func (uc *GetSessionStatusUseCase) Handle(ctx context.Context, query GetSessionStatusQuery) (*SessionStatusView, error) {
	if err := query.Validate(); err != nil {
		uc.logger.Warn(ctx, "Invalid get session status query", "error", err)
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	sessionEntity, err := uc.sessionRepo.GetByID(ctx, query.SessionID)
	if err != nil {
		uc.logger.Error(ctx, "Failed to get session", "sessionID", query.SessionID, "error", err)
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	isConnected := uc.whatsappService.IsClientConnected(query.SessionID)
	var whatsappStatus string
	if isConnected {
		whatsappStatus = "connected"
	} else {
		whatsappStatus = "disconnected"
	}

	qrCode := ""
	if !sessionEntity.IsAuthenticated() && (sessionEntity.IsConnecting() || sessionEntity.IsDisconnected()) {
		qrCode, err = uc.whatsappService.GetQRCode(query.SessionID)
		if err != nil {
			uc.logger.Warn(ctx, "Failed to get QR code", "sessionID", query.SessionID, "error", err)
		}
	}

	connectionInfo := map[string]interface{}{
		"domain_status":   sessionEntity.Status().String(),
		"whatsapp_status": whatsappStatus,
		"has_proxy":       sessionEntity.HasProxy(),
		"has_webhook":     sessionEntity.HasWebhook(),
		"created_at":      sessionEntity.CreatedAt().Value().Format("2006-01-02T15:04:05Z07:00"),
		"updated_at":      sessionEntity.UpdatedAt().Value().Format("2006-01-02T15:04:05Z07:00"),
	}

	statusView := &SessionStatusView{
		SessionID:       sessionEntity.SessionID().Value(),
		Status:          whatsappStatus,
		IsConnected:     sessionEntity.IsConnected(),
		IsAuthenticated: sessionEntity.IsAuthenticated(),
		DeviceJID:       sessionEntity.WaJID().Value(),
		QRCode:          qrCode,
		LastSeen:        sessionEntity.UpdatedAt().Value().Format("2006-01-02T15:04:05Z07:00"),
		ConnectionInfo:  connectionInfo,
	}

	uc.logger.Debug(ctx, "Session status retrieved successfully",
		"sessionID", query.SessionID,
		"status", whatsappStatus,
		"isConnected", sessionEntity.IsConnected())

	return statusView, nil
}
