package group

import (
	"context"
	"fmt"
	"strings"

	"zpmeow/internal/application/common"
	"zpmeow/internal/application/ports"
	"zpmeow/internal/domain/session"
)

type ManageParticipantsCommand struct {
	SessionID    string
	GroupJID     string
	Participants []string
	Action       string // "add" or "remove"
}

func (c ManageParticipantsCommand) Validate() error {
	if strings.TrimSpace(c.SessionID) == "" {
		return common.NewValidationError("sessionID", c.SessionID, "session ID is required")
	}

	if strings.TrimSpace(c.GroupJID) == "" {
		return common.NewValidationError("groupJID", c.GroupJID, "group JID is required")
	}

	if len(c.Participants) == 0 {
		return common.NewValidationError("participants", "", "at least one participant is required")
	}

	if len(c.Participants) > 50 {
		return common.NewValidationError("participants", "", "maximum 50 participants allowed per operation")
	}

	if c.Action != "add" && c.Action != "remove" {
		return common.NewValidationError("action", c.Action, "action must be 'add' or 'remove'")
	}

	for i, participant := range c.Participants {
		if strings.TrimSpace(participant) == "" {
			return common.NewValidationError("participants", participant, fmt.Sprintf("participant %d cannot be empty", i))
		}
	}

	return nil
}

type ManageParticipantsResult struct {
	SessionID          string
	GroupJID           string
	Action             string
	RequestedCount     int
	SuccessfulCount    int
	FailedParticipants []string
	Success            bool
	Message            string
}

type ManageParticipantsUseCase struct {
	sessionRepo     session.Repository
	whatsappService ports.WhatsAppService
	logger          ports.Logger
}

func NewManageParticipantsUseCase(
	sessionRepo session.Repository,
	whatsappService ports.WhatsAppService,
	logger ports.Logger,
) *ManageParticipantsUseCase {
	return &ManageParticipantsUseCase{
		sessionRepo:     sessionRepo,
		whatsappService: whatsappService,
		logger:          logger,
	}
}

func (uc *ManageParticipantsUseCase) Handle(ctx context.Context, cmd ManageParticipantsCommand) (*ManageParticipantsResult, error) {
	if err := cmd.Validate(); err != nil {
		uc.logger.Warn(ctx, "Invalid manage participants command", "error", err)
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	sessionEntity, err := uc.sessionRepo.GetByID(ctx, cmd.SessionID)
	if err != nil {
		uc.logger.Error(ctx, "Failed to get session", "sessionID", cmd.SessionID, "error", err)
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if !sessionEntity.IsConnected() {
		return nil, common.NewBusinessRuleError(
			"session_not_connected",
			fmt.Sprintf("session must be connected to manage participants, current status: %s", sessionEntity.Status()),
		)
	}

	if !sessionEntity.IsAuthenticated() {
		return nil, common.NewBusinessRuleError(
			"session_not_authenticated",
			"session must be authenticated to manage participants",
		)
	}

	var failedParticipants []string
	successfulCount := 0

	if cmd.Action == "add" {
		err = uc.whatsappService.AddParticipants(ctx, cmd.SessionID, cmd.GroupJID, cmd.Participants)
	} else {
		err = uc.whatsappService.RemoveParticipants(ctx, cmd.SessionID, cmd.GroupJID, cmd.Participants)
	}

	if err != nil {
		uc.logger.Error(ctx, "Failed to manage participants",
			"sessionID", cmd.SessionID,
			"groupJID", cmd.GroupJID,
			"action", cmd.Action,
			"participantCount", len(cmd.Participants),
			"error", err)

		failedParticipants = cmd.Participants
		successfulCount = 0
	} else {
		successfulCount = len(cmd.Participants)
		failedParticipants = []string{}
	}

	success := successfulCount > 0
	message := fmt.Sprintf("Successfully %sed %d participants", cmd.Action, successfulCount)
	if len(failedParticipants) > 0 {
		message += fmt.Sprintf(", failed to %s %d participants", cmd.Action, len(failedParticipants))
	}

	uc.logger.Info(ctx, "Participants management completed",
		"sessionID", cmd.SessionID,
		"groupJID", cmd.GroupJID,
		"action", cmd.Action,
		"requested", len(cmd.Participants),
		"successful", successfulCount,
		"failed", len(failedParticipants))

	return &ManageParticipantsResult{
		SessionID:          cmd.SessionID,
		GroupJID:           cmd.GroupJID,
		Action:             cmd.Action,
		RequestedCount:     len(cmd.Participants),
		SuccessfulCount:    successfulCount,
		FailedParticipants: failedParticipants,
		Success:            success,
		Message:            message,
	}, nil
}

type GetInviteLinkCommand struct {
	SessionID string
	GroupJID  string
}

func (c GetInviteLinkCommand) Validate() error {
	if strings.TrimSpace(c.SessionID) == "" {
		return common.NewValidationError("sessionID", c.SessionID, "session ID is required")
	}

	if strings.TrimSpace(c.GroupJID) == "" {
		return common.NewValidationError("groupJID", c.GroupJID, "group JID is required")
	}

	return nil
}

type GetInviteLinkResult struct {
	SessionID  string
	GroupJID   string
	InviteLink string
	Success    bool
}

type GetInviteLinkUseCase struct {
	sessionRepo     session.Repository
	whatsappService ports.WhatsAppService
	logger          ports.Logger
}

func NewGetInviteLinkUseCase(
	sessionRepo session.Repository,
	whatsappService ports.WhatsAppService,
	logger ports.Logger,
) *GetInviteLinkUseCase {
	return &GetInviteLinkUseCase{
		sessionRepo:     sessionRepo,
		whatsappService: whatsappService,
		logger:          logger,
	}
}

func (uc *GetInviteLinkUseCase) Handle(ctx context.Context, cmd GetInviteLinkCommand) (*GetInviteLinkResult, error) {
	if err := cmd.Validate(); err != nil {
		uc.logger.Warn(ctx, "Invalid get invite link command", "error", err)
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	sessionEntity, err := uc.sessionRepo.GetByID(ctx, cmd.SessionID)
	if err != nil {
		uc.logger.Error(ctx, "Failed to get session", "sessionID", cmd.SessionID, "error", err)
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if !sessionEntity.IsConnected() {
		return nil, common.NewBusinessRuleError(
			"session_not_connected",
			fmt.Sprintf("session must be connected to get invite link, current status: %s", sessionEntity.Status()),
		)
	}

	inviteLink, err := uc.whatsappService.GetGroupInviteLink(ctx, cmd.SessionID, cmd.GroupJID)
	if err != nil {
		uc.logger.Error(ctx, "Failed to get group invite link",
			"sessionID", cmd.SessionID,
			"groupJID", cmd.GroupJID,
			"error", err)
		return nil, fmt.Errorf("failed to get group invite link: %w", err)
	}

	uc.logger.Info(ctx, "Group invite link retrieved successfully",
		"sessionID", cmd.SessionID,
		"groupJID", cmd.GroupJID)

	return &GetInviteLinkResult{
		SessionID:  cmd.SessionID,
		GroupJID:   cmd.GroupJID,
		InviteLink: inviteLink,
		Success:    true,
	}, nil
}
