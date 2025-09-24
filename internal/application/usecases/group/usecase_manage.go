package group

import (
	"context"
	"fmt"
	"strings"

	"zpmeow/internal/application/common"
	"zpmeow/internal/application/ports"
	"zpmeow/internal/domain/session"
)

type JoinGroupCommand struct {
	SessionID  string
	InviteLink string
}

func (c JoinGroupCommand) Validate() error {
	if strings.TrimSpace(c.SessionID) == "" {
		return common.NewValidationError("sessionID", c.SessionID, "session ID is required")
	}

	if strings.TrimSpace(c.InviteLink) == "" {
		return common.NewValidationError("inviteLink", c.InviteLink, "invite link is required")
	}

	if !strings.Contains(c.InviteLink, "chat.whatsapp.com") {
		return common.NewValidationError("inviteLink", c.InviteLink, "invalid WhatsApp invite link format")
	}

	return nil
}

type LeaveGroupCommand struct {
	SessionID string
	GroupJID  string
}

func (c LeaveGroupCommand) Validate() error {
	if strings.TrimSpace(c.SessionID) == "" {
		return common.NewValidationError("sessionID", c.SessionID, "session ID is required")
	}

	if strings.TrimSpace(c.GroupJID) == "" {
		return common.NewValidationError("groupJID", c.GroupJID, "group JID is required")
	}

	return nil
}

type GroupManagementResult struct {
	SessionID string
	GroupJID  string
	Action    string
	Success   bool
	Message   string
}

type JoinGroupUseCase struct {
	sessionRepo     session.Repository
	whatsappService ports.WhatsAppService
	logger          ports.Logger
}

func NewJoinGroupUseCase(
	sessionRepo session.Repository,
	whatsappService ports.WhatsAppService,
	logger ports.Logger,
) *JoinGroupUseCase {
	return &JoinGroupUseCase{
		sessionRepo:     sessionRepo,
		whatsappService: whatsappService,
		logger:          logger,
	}
}

func (uc *JoinGroupUseCase) Handle(ctx context.Context, cmd JoinGroupCommand) (*GroupManagementResult, error) {
	if err := cmd.Validate(); err != nil {
		uc.logger.Warn(ctx, "Invalid join group command", "error", err)
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
			fmt.Sprintf("session must be connected to join groups, current status: %s", sessionEntity.Status()),
		)
	}

	groupInfo, err := uc.whatsappService.JoinGroup(ctx, cmd.SessionID, cmd.InviteLink)
	if err != nil {
		uc.logger.Error(ctx, "Failed to join group",
			"sessionID", cmd.SessionID,
			"inviteLink", cmd.InviteLink,
			"error", err)
		return nil, fmt.Errorf("failed to join group: %w", err)
	}

	uc.logger.Info(ctx, "Successfully joined group",
		"sessionID", cmd.SessionID,
		"inviteLink", cmd.InviteLink)

	return &GroupManagementResult{
		SessionID: cmd.SessionID,
		GroupJID:  groupInfo.JID,
		Action:    "join",
		Success:   true,
		Message:   "Successfully joined group",
	}, nil
}

type LeaveGroupUseCase struct {
	sessionRepo     session.Repository
	whatsappService ports.WhatsAppService
	logger          ports.Logger
}

func NewLeaveGroupUseCase(
	sessionRepo session.Repository,
	whatsappService ports.WhatsAppService,
	logger ports.Logger,
) *LeaveGroupUseCase {
	return &LeaveGroupUseCase{
		sessionRepo:     sessionRepo,
		whatsappService: whatsappService,
		logger:          logger,
	}
}

func (uc *LeaveGroupUseCase) Handle(ctx context.Context, cmd LeaveGroupCommand) (*GroupManagementResult, error) {
	if err := cmd.Validate(); err != nil {
		uc.logger.Warn(ctx, "Invalid leave group command", "error", err)
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
			fmt.Sprintf("session must be connected to leave groups, current status: %s", sessionEntity.Status()),
		)
	}

	if err := uc.whatsappService.LeaveGroup(ctx, cmd.SessionID, cmd.GroupJID); err != nil {
		uc.logger.Error(ctx, "Failed to leave group",
			"sessionID", cmd.SessionID,
			"groupJID", cmd.GroupJID,
			"error", err)
		return nil, fmt.Errorf("failed to leave group: %w", err)
	}

	uc.logger.Info(ctx, "Successfully left group",
		"sessionID", cmd.SessionID,
		"groupJID", cmd.GroupJID)

	return &GroupManagementResult{
		SessionID: cmd.SessionID,
		GroupJID:  cmd.GroupJID,
		Action:    "leave",
		Success:   true,
		Message:   "Successfully left group",
	}, nil
}
