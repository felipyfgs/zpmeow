package chat

import (
	"context"
	"fmt"
	"strings"
	"time"

	"zpmeow/internal/application/common"
	"zpmeow/internal/application/ports"
	"zpmeow/internal/domain/session"
)

type MuteChatCommand struct {
	SessionID string
	ChatJID   string
	Mute      bool
	Duration  time.Duration // Duration to mute (0 for permanent)
}

func (c MuteChatCommand) Validate() error {
	if strings.TrimSpace(c.SessionID) == "" {
		return common.NewValidationError("sessionID", c.SessionID, "session ID is required")
	}

	if strings.TrimSpace(c.ChatJID) == "" {
		return common.NewValidationError("chatJID", c.ChatJID, "chat JID is required")
	}

	if c.Mute && c.Duration < 0 {
		return common.NewValidationError("duration", c.Duration, "duration cannot be negative")
	}

	return nil
}

type ArchiveChatCommand struct {
	SessionID string
	ChatJID   string
	Archive   bool
}

func (c ArchiveChatCommand) Validate() error {
	if strings.TrimSpace(c.SessionID) == "" {
		return common.NewValidationError("sessionID", c.SessionID, "session ID is required")
	}

	if strings.TrimSpace(c.ChatJID) == "" {
		return common.NewValidationError("chatJID", c.ChatJID, "chat JID is required")
	}

	return nil
}

type BlockChatCommand struct {
	SessionID string
	ChatJID   string
	Block     bool
}

func (c BlockChatCommand) Validate() error {
	if strings.TrimSpace(c.SessionID) == "" {
		return common.NewValidationError("sessionID", c.SessionID, "session ID is required")
	}

	if strings.TrimSpace(c.ChatJID) == "" {
		return common.NewValidationError("chatJID", c.ChatJID, "chat JID is required")
	}

	return nil
}

type ChatManagementResult struct {
	SessionID string
	ChatJID   string
	Action    string
	Success   bool
}

type MuteChatUseCase struct {
	sessionRepo     session.Repository
	whatsappService ports.WhatsAppService
	logger          ports.Logger
}

func NewMuteChatUseCase(
	sessionRepo session.Repository,
	whatsappService ports.WhatsAppService,
	logger ports.Logger,
) *MuteChatUseCase {
	return &MuteChatUseCase{
		sessionRepo:     sessionRepo,
		whatsappService: whatsappService,
		logger:          logger,
	}
}

func (uc *MuteChatUseCase) Handle(ctx context.Context, cmd MuteChatCommand) (*ChatManagementResult, error) {
	if err := cmd.Validate(); err != nil {
		uc.logger.Warn(ctx, "Invalid mute chat command", "error", err)
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
			fmt.Sprintf("session must be connected to manage chats, current status: %s", sessionEntity.Status()),
		)
	}

	if err := uc.whatsappService.MuteChat(ctx, cmd.SessionID, cmd.ChatJID, cmd.Mute, cmd.Duration); err != nil {
		uc.logger.Error(ctx, "Failed to mute/unmute chat",
			"sessionID", cmd.SessionID,
			"chatJID", cmd.ChatJID,
			"mute", cmd.Mute,
			"error", err)
		return nil, fmt.Errorf("failed to mute/unmute chat: %w", err)
	}

	action := "unmute"
	if cmd.Mute {
		action = "mute"
	}

	uc.logger.Info(ctx, "Chat mute status changed successfully",
		"sessionID", cmd.SessionID,
		"chatJID", cmd.ChatJID,
		"action", action)

	return &ChatManagementResult{
		SessionID: cmd.SessionID,
		ChatJID:   cmd.ChatJID,
		Action:    action,
		Success:   true,
	}, nil
}

type ArchiveChatUseCase struct {
	sessionRepo     session.Repository
	whatsappService ports.WhatsAppService
	logger          ports.Logger
}

func NewArchiveChatUseCase(
	sessionRepo session.Repository,
	whatsappService ports.WhatsAppService,
	logger ports.Logger,
) *ArchiveChatUseCase {
	return &ArchiveChatUseCase{
		sessionRepo:     sessionRepo,
		whatsappService: whatsappService,
		logger:          logger,
	}
}

func (uc *ArchiveChatUseCase) Handle(ctx context.Context, cmd ArchiveChatCommand) (*ChatManagementResult, error) {
	if err := cmd.Validate(); err != nil {
		uc.logger.Warn(ctx, "Invalid archive chat command", "error", err)
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
			fmt.Sprintf("session must be connected to manage chats, current status: %s", sessionEntity.Status()),
		)
	}

	if err := uc.whatsappService.ArchiveChat(ctx, cmd.SessionID, cmd.ChatJID, cmd.Archive); err != nil {
		uc.logger.Error(ctx, "Failed to archive/unarchive chat",
			"sessionID", cmd.SessionID,
			"chatJID", cmd.ChatJID,
			"archive", cmd.Archive,
			"error", err)
		return nil, fmt.Errorf("failed to archive/unarchive chat: %w", err)
	}

	action := "unarchive"
	if cmd.Archive {
		action = "archive"
	}

	uc.logger.Info(ctx, "Chat archive status changed successfully",
		"sessionID", cmd.SessionID,
		"chatJID", cmd.ChatJID,
		"action", action)

	return &ChatManagementResult{
		SessionID: cmd.SessionID,
		ChatJID:   cmd.ChatJID,
		Action:    action,
		Success:   true,
	}, nil
}
