package session

import (
	"context"
	"fmt"
	"strings"

	"zpmeow/internal/application/common"
	"zpmeow/internal/application/ports"
	"zpmeow/internal/domain/session"
)

type DisconnectSessionCommand struct {
	SessionID string
	Reason    string
}

func (c DisconnectSessionCommand) Validate() error {
	if strings.TrimSpace(c.SessionID) == "" {
		return common.NewValidationError("sessionID", c.SessionID, "session ID is required")
	}

	if len(c.Reason) > 500 {
		return common.NewValidationError("reason", c.Reason, "reason must not exceed 500 characters")
	}

	return nil
}

type DisconnectSessionResult struct {
	SessionID string
	Status    string
	Reason    string
}

type DisconnectSessionUseCase struct {
	sessionRepo     session.Repository
	whatsappService ports.WhatsAppService
	eventPublisher  ports.EventPublisher
	logger          ports.Logger
}

func NewDisconnectSessionUseCase(
	sessionRepo session.Repository,
	whatsappService ports.WhatsAppService,
	eventPublisher ports.EventPublisher,
	logger ports.Logger,
) *DisconnectSessionUseCase {
	return &DisconnectSessionUseCase{
		sessionRepo:     sessionRepo,
		whatsappService: whatsappService,
		eventPublisher:  eventPublisher,
		logger:          logger,
	}
}

func (uc *DisconnectSessionUseCase) Handle(ctx context.Context, cmd DisconnectSessionCommand) (*DisconnectSessionResult, error) {
	if err := cmd.Validate(); err != nil {
		uc.logger.Warn(ctx, "Invalid disconnect session command", "error", err)
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	sessionEntity, err := uc.sessionRepo.GetByID(ctx, cmd.SessionID)
	if err != nil {
		uc.logger.Error(ctx, "Failed to get session", "sessionID", cmd.SessionID, "error", err)
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if sessionEntity.IsDisconnected() {
		uc.logger.Info(ctx, "Session already disconnected", "sessionID", cmd.SessionID)
		return &DisconnectSessionResult{
			SessionID: sessionEntity.SessionID().Value(),
			Status:    sessionEntity.Status().String(),
			Reason:    "already disconnected",
		}, nil
	}

	reason := cmd.Reason
	if reason == "" {
		reason = "user requested"
	}

	if err := uc.whatsappService.DisconnectSession(ctx, cmd.SessionID); err != nil {
		uc.logger.Warn(ctx, "Failed to disconnect session via WhatsApp service", "sessionID", cmd.SessionID, "error", err)
	}

	if err := sessionEntity.Disconnect(reason); err != nil {
		uc.logger.Error(ctx, "Failed to disconnect session in domain", "sessionID", cmd.SessionID, "error", err)
		return nil, fmt.Errorf("failed to disconnect session: %w", err)
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

	uc.logger.Info(ctx, "Session disconnected successfully", "sessionID", cmd.SessionID, "reason", reason)

	return &DisconnectSessionResult{
		SessionID: sessionEntity.SessionID().Value(),
		Status:    sessionEntity.Status().String(),
		Reason:    reason,
	}, nil
}
