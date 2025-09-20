package session

import (
	"context"
	"fmt"
	"strings"

	"zpmeow/internal/application/common"
	"zpmeow/internal/application/ports"
	"zpmeow/internal/domain/session"
)

type ConnectSessionCommand struct {
	SessionID string
}

func (c ConnectSessionCommand) Validate() error {
	if strings.TrimSpace(c.SessionID) == "" {
		return common.NewValidationError("sessionID", c.SessionID, "session ID is required")
	}

	return nil
}

type ConnectSessionResult struct {
	SessionID string
	Status    string
	QRCode    string
}

type ConnectSessionUseCase struct {
	sessionRepo     session.Repository
	whatsappService ports.WhatsAppService
	eventPublisher  ports.EventPublisher
	logger          ports.Logger
}

func NewConnectSessionUseCase(
	sessionRepo session.Repository,
	whatsappService ports.WhatsAppService,
	eventPublisher ports.EventPublisher,
	logger ports.Logger,
) *ConnectSessionUseCase {
	return &ConnectSessionUseCase{
		sessionRepo:     sessionRepo,
		whatsappService: whatsappService,
		eventPublisher:  eventPublisher,
		logger:          logger,
	}
}

func (uc *ConnectSessionUseCase) Handle(ctx context.Context, cmd ConnectSessionCommand) (*ConnectSessionResult, error) {
	if err := cmd.Validate(); err != nil {
		uc.logger.Warn(ctx, "Invalid connect session command", "error", err)
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	sessionEntity, err := uc.sessionRepo.GetByID(ctx, cmd.SessionID)
	if err != nil {
		uc.logger.Error(ctx, "Failed to get session", "sessionID", cmd.SessionID, "error", err)
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if !sessionEntity.CanConnect() {
		return nil, common.NewBusinessRuleError(
			"session_connection_not_allowed",
			fmt.Sprintf("session cannot be connected from current status: %s", sessionEntity.Status()),
		)
	}

	qrCodeFromConnect, err := uc.whatsappService.ConnectSession(ctx, cmd.SessionID)
	if err != nil {
		uc.logger.Error(ctx, "Failed to connect session via WhatsApp service", "sessionID", cmd.SessionID, "error", err)

		sessionEntity.SetError("connection failed: " + err.Error())

		if updateErr := uc.sessionRepo.Update(ctx, sessionEntity); updateErr != nil {
			uc.logger.Error(ctx, "Failed to update session after connection error", "sessionID", cmd.SessionID, "error", updateErr)
		}

		return nil, fmt.Errorf("failed to connect session: %w", err)
	}

	if err := sessionEntity.Connect(); err != nil {
		uc.logger.Error(ctx, "Failed to set session to connecting state", "sessionID", cmd.SessionID, "error", err)
		return nil, fmt.Errorf("failed to update session state: %w", err)
	}

	qrCode := qrCodeFromConnect
	if !sessionEntity.IsAuthenticated() && qrCode == "" {
		qrCode, err = uc.whatsappService.GetQRCode(cmd.SessionID)
		if err != nil {
			uc.logger.Warn(ctx, "Failed to get QR code", "sessionID", cmd.SessionID, "error", err)
		}
	}

	if qrCode != "" {
		if err := sessionEntity.SetQRCode(qrCode); err != nil {
			uc.logger.Warn(ctx, "Failed to set QR code in session", "sessionID", cmd.SessionID, "error", err)
		}
	}

	if err := uc.sessionRepo.Update(ctx, sessionEntity); err != nil {
		uc.logger.Error(ctx, "Failed to update session", "sessionID", cmd.SessionID, "error", err)
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	events := sessionEntity.GetEvents()
	if len(events) > 0 {
		if err := uc.eventPublisher.PublishBatch(ctx, events); err != nil {
			uc.logger.Warn(ctx, "Failed to publish domain events", "sessionID", cmd.SessionID, "error", err)
		}
		sessionEntity.ClearEvents()
	}

	uc.logger.Info(ctx, "Session connection initiated", "sessionID", cmd.SessionID, "status", sessionEntity.Status())

	return &ConnectSessionResult{
		SessionID: sessionEntity.SessionID().Value(),
		Status:    sessionEntity.Status().String(),
		QRCode:    qrCode,
	}, nil
}
