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
}

func NewConnectSessionUseCase(
	sessionRepo session.Repository,
	whatsappService ports.WhatsAppService,
	eventPublisher ports.EventPublisher,
) *ConnectSessionUseCase {
	return &ConnectSessionUseCase{
		sessionRepo:     sessionRepo,
		whatsappService: whatsappService,
		eventPublisher:  eventPublisher,
	}
}

func (uc *ConnectSessionUseCase) Handle(ctx context.Context, cmd ConnectSessionCommand) (*ConnectSessionResult, error) {
	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	sessionEntity, err := uc.sessionRepo.GetByID(ctx, cmd.SessionID)
	if err != nil {
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
		sessionEntity.SetError("connection failed: " + err.Error())

		if updateErr := uc.sessionRepo.Update(ctx, sessionEntity); updateErr != nil {
			// Log this error through domain events or return it
		}

		return nil, fmt.Errorf("failed to connect session: %w", err)
	}

	if err := sessionEntity.Connect(); err != nil {
		return nil, fmt.Errorf("failed to update session state: %w", err)
	}

	qrCode := qrCodeFromConnect
	if !sessionEntity.IsAuthenticated() && qrCode == "" {
		qrCode, err = uc.whatsappService.GetQRCode(cmd.SessionID)
		if err != nil {
			// QR code failure is not critical, continue without it
		}
	}

	if qrCode != "" {
		if err := sessionEntity.SetQRCode(qrCode); err != nil {
			// QR code setting failure is not critical
		}
	}

	if err := uc.sessionRepo.Update(ctx, sessionEntity); err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	events := sessionEntity.GetEvents()
	if len(events) > 0 {
		if err := uc.eventPublisher.PublishBatch(ctx, events); err != nil {
			// Event publishing failure should not fail the use case
			// but could be logged at infrastructure level
		}
		sessionEntity.ClearEvents()
	}

	return &ConnectSessionResult{
		SessionID: sessionEntity.SessionID().Value(),
		Status:    sessionEntity.Status().String(),
		QRCode:    qrCode,
	}, nil
}
