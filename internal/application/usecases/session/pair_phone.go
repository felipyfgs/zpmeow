package session

import (
	"context"
	"fmt"
	"strings"

	"zpmeow/internal/application/common"
	"zpmeow/internal/application/ports"
	"zpmeow/internal/domain/session"
)

type PairPhoneCommand struct {
	SessionID   string
	PhoneNumber string
}

func (c PairPhoneCommand) Validate() error {
	if strings.TrimSpace(c.SessionID) == "" {
		return common.NewValidationError("sessionID", c.SessionID, "session ID is required")
	}

	if strings.TrimSpace(c.PhoneNumber) == "" {
		return common.NewValidationError("phoneNumber", c.PhoneNumber, "phone number is required")
	}

	phoneNumber := strings.TrimSpace(c.PhoneNumber)
	if len(phoneNumber) < 10 || len(phoneNumber) > 15 {
		return common.NewValidationError("phoneNumber", c.PhoneNumber, "phone number must be between 10 and 15 digits")
	}

	return nil
}

type PairPhoneResult struct {
	SessionID   string
	PhoneNumber string
	PairCode    string
	Success     bool
	Message     string
}

type PairPhoneUseCase struct {
	sessionRepo     session.Repository
	whatsappService ports.WhatsAppService
	eventPublisher  ports.EventPublisher
	logger          ports.Logger
}

func NewPairPhoneUseCase(
	sessionRepo session.Repository,
	whatsappService ports.WhatsAppService,
	eventPublisher ports.EventPublisher,
	logger ports.Logger,
) *PairPhoneUseCase {
	return &PairPhoneUseCase{
		sessionRepo:     sessionRepo,
		whatsappService: whatsappService,
		eventPublisher:  eventPublisher,
		logger:          logger,
	}
}

func (uc *PairPhoneUseCase) Handle(ctx context.Context, cmd PairPhoneCommand) (*PairPhoneResult, error) {
	if err := cmd.Validate(); err != nil {
		uc.logger.Warn(ctx, "Invalid pair phone command", "error", err)
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	sessionEntity, err := uc.sessionRepo.GetByID(ctx, cmd.SessionID)
	if err != nil {
		uc.logger.Error(ctx, "Failed to get session", "sessionID", cmd.SessionID, "error", err)
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if !sessionEntity.CanConnect() {
		return nil, common.NewBusinessRuleError(
			"session_pairing_not_allowed",
			fmt.Sprintf("session cannot be paired from current status: %s", sessionEntity.Status()),
		)
	}

	pairCode, err := uc.whatsappService.PairPhone(cmd.SessionID, cmd.PhoneNumber)
	if err != nil {
		uc.logger.Error(ctx, "Failed to pair phone with session",
			"sessionID", cmd.SessionID,
			"phoneNumber", cmd.PhoneNumber,
			"error", err)

		sessionEntity.SetError("phone pairing failed: " + err.Error())

		if updateErr := uc.sessionRepo.Update(ctx, sessionEntity); updateErr != nil {
			uc.logger.Error(ctx, "Failed to update session after pairing error", "sessionID", cmd.SessionID, "error", updateErr)
		}

		return nil, fmt.Errorf("failed to pair phone with session: %w", err)
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

	uc.logger.Info(ctx, "Phone paired successfully",
		"sessionID", cmd.SessionID,
		"phoneNumber", cmd.PhoneNumber)

	return &PairPhoneResult{
		SessionID:   cmd.SessionID,
		PhoneNumber: cmd.PhoneNumber,
		PairCode:    pairCode,
		Success:     true,
		Message:     "Phone paired successfully",
	}, nil
}
