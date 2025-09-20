package session

import (
	"context"
	"fmt"
	"strings"

	"zpmeow/internal/application/common"
	"zpmeow/internal/application/ports"
	"zpmeow/internal/domain/session"
)

type DeleteSessionCommand struct {
	SessionID string
	Force     bool // Force deletion even if session is connected
}

func (c DeleteSessionCommand) Validate() error {
	if strings.TrimSpace(c.SessionID) == "" {
		return common.NewValidationError("sessionID", c.SessionID, "session ID is required")
	}

	return nil
}

type DeleteSessionResult struct {
	SessionID string
	Name      string
	Deleted   bool
}

type DeleteSessionUseCase struct {
	sessionRepo     session.Repository
	whatsappService ports.WhatsAppService
	eventPublisher  ports.EventPublisher
	logger          ports.Logger
}

func NewDeleteSessionUseCase(
	sessionRepo session.Repository,
	whatsappService ports.WhatsAppService,
	eventPublisher ports.EventPublisher,
	logger ports.Logger,
) *DeleteSessionUseCase {
	return &DeleteSessionUseCase{
		sessionRepo:     sessionRepo,
		whatsappService: whatsappService,
		eventPublisher:  eventPublisher,
		logger:          logger,
	}
}

func (uc *DeleteSessionUseCase) Handle(ctx context.Context, cmd DeleteSessionCommand) (*DeleteSessionResult, error) {
	if err := cmd.Validate(); err != nil {
		uc.logger.Warn(ctx, "Invalid delete session command", "error", err)
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	sessionEntity, err := uc.sessionRepo.GetByID(ctx, cmd.SessionID)
	if err != nil {
		uc.logger.Error(ctx, "Failed to get session", "sessionID", cmd.SessionID, "error", err)
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if sessionEntity.IsConnected() && !cmd.Force {
		return nil, common.NewBusinessRuleError(
			"session_deletion_not_allowed",
			"cannot delete connected session without force flag",
		)
	}

	if sessionEntity.IsConnected() || sessionEntity.IsConnecting() {
		uc.logger.Info(ctx, "Disconnecting session before deletion", "sessionID", cmd.SessionID)

		if err := uc.whatsappService.DisconnectSession(ctx, cmd.SessionID); err != nil {
			uc.logger.Warn(ctx, "Failed to disconnect session via WhatsApp service", "sessionID", cmd.SessionID, "error", err)
		}

		if err := sessionEntity.Disconnect("deletion requested"); err != nil {
			uc.logger.Warn(ctx, "Failed to disconnect session in domain", "sessionID", cmd.SessionID, "error", err)
		}
	}

	sessionEntity.Delete()

	events := sessionEntity.GetEvents()
	if len(events) > 0 {
		if err := uc.eventPublisher.PublishBatch(ctx, events); err != nil {
			uc.logger.Warn(ctx, "Failed to publish domain events", "sessionID", cmd.SessionID, "error", err)
		}
	}

	if err := uc.sessionRepo.Delete(ctx, cmd.SessionID); err != nil {
		uc.logger.Error(ctx, "Failed to delete session", "sessionID", cmd.SessionID, "error", err)
		return nil, fmt.Errorf("failed to delete session: %w", err)
	}

	uc.logger.Info(ctx, "Session deleted successfully", "sessionID", cmd.SessionID, "name", sessionEntity.Name().Value())

	return &DeleteSessionResult{
		SessionID: sessionEntity.SessionID().Value(),
		Name:      sessionEntity.Name().Value(),
		Deleted:   true,
	}, nil
}
