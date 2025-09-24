package messaging

import (
	"context"
	"fmt"
	"strings"

	"zpmeow/internal/application/common"
	"zpmeow/internal/application/ports"
	"zpmeow/internal/domain/session"
)

type MarkAsReadCommand struct {
	SessionID  string
	ChatJID    string
	MessageIDs []string
}

func (c MarkAsReadCommand) Validate() error {
	if strings.TrimSpace(c.SessionID) == "" {
		return common.NewValidationError("sessionID", c.SessionID, "session ID is required")
	}

	if strings.TrimSpace(c.ChatJID) == "" {
		return common.NewValidationError("chatJID", c.ChatJID, "chat JID is required")
	}

	if len(c.MessageIDs) == 0 {
		return common.NewValidationError("messageIDs", "", "at least one message ID is required")
	}

	for i, msgID := range c.MessageIDs {
		if strings.TrimSpace(msgID) == "" {
			return common.NewValidationError("messageIDs", msgID, fmt.Sprintf("message ID at index %d is required", i))
		}
	}

	return nil
}

type ReactToMessageCommand struct {
	SessionID string
	ChatJID   string
	MessageID string
	Emoji     string
	Remove    bool
}

func (c ReactToMessageCommand) Validate() error {
	if strings.TrimSpace(c.SessionID) == "" {
		return common.NewValidationError("sessionID", c.SessionID, "session ID is required")
	}

	if strings.TrimSpace(c.ChatJID) == "" {
		return common.NewValidationError("chatJID", c.ChatJID, "chat JID is required")
	}

	if strings.TrimSpace(c.MessageID) == "" {
		return common.NewValidationError("messageID", c.MessageID, "message ID is required")
	}

	if !c.Remove && strings.TrimSpace(c.Emoji) == "" {
		return common.NewValidationError("emoji", c.Emoji, "emoji is required when not removing reaction")
	}

	return nil
}

type EditMessageCommand struct {
	SessionID  string
	ChatJID    string
	MessageID  string
	NewContent string
}

func (c EditMessageCommand) Validate() error {
	if strings.TrimSpace(c.SessionID) == "" {
		return common.NewValidationError("sessionID", c.SessionID, "session ID is required")
	}

	if strings.TrimSpace(c.ChatJID) == "" {
		return common.NewValidationError("chatJID", c.ChatJID, "chat JID is required")
	}

	if strings.TrimSpace(c.MessageID) == "" {
		return common.NewValidationError("messageID", c.MessageID, "message ID is required")
	}

	if strings.TrimSpace(c.NewContent) == "" {
		return common.NewValidationError("newContent", c.NewContent, "new content is required")
	}

	if len(c.NewContent) > 4096 {
		return common.NewValidationError("newContent", c.NewContent, "new content must not exceed 4096 characters")
	}

	return nil
}

type DeleteMessageCommand struct {
	SessionID   string
	ChatJID     string
	MessageID   string
	ForEveryone bool
}

func (c DeleteMessageCommand) Validate() error {
	if strings.TrimSpace(c.SessionID) == "" {
		return common.NewValidationError("sessionID", c.SessionID, "session ID is required")
	}

	if strings.TrimSpace(c.ChatJID) == "" {
		return common.NewValidationError("chatJID", c.ChatJID, "chat JID is required")
	}

	if strings.TrimSpace(c.MessageID) == "" {
		return common.NewValidationError("messageID", c.MessageID, "message ID is required")
	}

	return nil
}

type MessageActionResult struct {
	SessionID string
	ChatJID   string
	MessageID string
	Action    string
	Success   bool
	Message   string
}

type MarkAsReadUseCase struct {
	sessionRepo     session.Repository
	whatsappService ports.WhatsAppService
	logger          ports.Logger
}

func NewMarkAsReadUseCase(
	sessionRepo session.Repository,
	whatsappService ports.WhatsAppService,
	logger ports.Logger,
) *MarkAsReadUseCase {
	return &MarkAsReadUseCase{
		sessionRepo:     sessionRepo,
		whatsappService: whatsappService,
		logger:          logger,
	}
}

func (uc *MarkAsReadUseCase) Handle(ctx context.Context, cmd MarkAsReadCommand) (*MessageActionResult, error) {
	if err := cmd.Validate(); err != nil {
		uc.logger.Warn(ctx, "Invalid mark as read command", "error", err)
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
			fmt.Sprintf("session must be connected to mark messages as read, current status: %s", sessionEntity.Status()),
		)
	}

	if err := uc.whatsappService.MarkAsRead(ctx, cmd.SessionID, cmd.ChatJID, cmd.MessageIDs); err != nil {
		uc.logger.Error(ctx, "Failed to mark messages as read",
			"sessionID", cmd.SessionID,
			"chatJID", cmd.ChatJID,
			"messageIDs", cmd.MessageIDs,
			"error", err)
		return nil, fmt.Errorf("failed to mark messages as read: %w", err)
	}

	uc.logger.Info(ctx, "Messages marked as read successfully",
		"sessionID", cmd.SessionID,
		"chatJID", cmd.ChatJID,
		"messageIDs", cmd.MessageIDs)

	return &MessageActionResult{
		SessionID: cmd.SessionID,
		ChatJID:   cmd.ChatJID,
		MessageID: strings.Join(cmd.MessageIDs, ","),
		Action:    "mark_as_read",
		Success:   true,
		Message:   fmt.Sprintf("%d messages marked as read successfully", len(cmd.MessageIDs)),
	}, nil
}

type ReactToMessageUseCase struct {
	sessionRepo     session.Repository
	whatsappService ports.WhatsAppService
	logger          ports.Logger
}

func NewReactToMessageUseCase(
	sessionRepo session.Repository,
	whatsappService ports.WhatsAppService,
	logger ports.Logger,
) *ReactToMessageUseCase {
	return &ReactToMessageUseCase{
		sessionRepo:     sessionRepo,
		whatsappService: whatsappService,
		logger:          logger,
	}
}

func (uc *ReactToMessageUseCase) Handle(ctx context.Context, cmd ReactToMessageCommand) (*MessageActionResult, error) {
	if err := cmd.Validate(); err != nil {
		uc.logger.Warn(ctx, "Invalid react to message command", "error", err)
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
			fmt.Sprintf("session must be connected to react to messages, current status: %s", sessionEntity.Status()),
		)
	}

	emoji := cmd.Emoji
	if cmd.Remove {
		emoji = ""
	}

	if err := uc.whatsappService.ReactToMessage(ctx, cmd.SessionID, cmd.ChatJID, cmd.MessageID, emoji); err != nil {
		uc.logger.Error(ctx, "Failed to react to message",
			"sessionID", cmd.SessionID,
			"chatJID", cmd.ChatJID,
			"messageID", cmd.MessageID,
			"emoji", cmd.Emoji,
			"error", err)
		return nil, fmt.Errorf("failed to react to message: %w", err)
	}

	action := "add_reaction"
	if cmd.Remove {
		action = "remove_reaction"
	}

	uc.logger.Info(ctx, "Message reaction updated successfully",
		"sessionID", cmd.SessionID,
		"chatJID", cmd.ChatJID,
		"messageID", cmd.MessageID,
		"action", action)

	return &MessageActionResult{
		SessionID: cmd.SessionID,
		ChatJID:   cmd.ChatJID,
		MessageID: cmd.MessageID,
		Action:    action,
		Success:   true,
		Message:   fmt.Sprintf("Message reaction %s successfully", action),
	}, nil
}

type EditMessageUseCase struct {
	sessionRepo     session.Repository
	whatsappService ports.WhatsAppService
	logger          ports.Logger
}

func NewEditMessageUseCase(
	sessionRepo session.Repository,
	whatsappService ports.WhatsAppService,
	logger ports.Logger,
) *EditMessageUseCase {
	return &EditMessageUseCase{
		sessionRepo:     sessionRepo,
		whatsappService: whatsappService,
		logger:          logger,
	}
}

func (uc *EditMessageUseCase) Handle(ctx context.Context, cmd EditMessageCommand) (*MessageActionResult, error) {
	if err := cmd.Validate(); err != nil {
		uc.logger.Warn(ctx, "Invalid edit message command", "error", err)
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
			fmt.Sprintf("session must be connected to edit messages, current status: %s", sessionEntity.Status()),
		)
	}

	sendResp, err := uc.whatsappService.EditMessage(ctx, cmd.SessionID, cmd.ChatJID, cmd.MessageID, cmd.NewContent)
	if err != nil {
		uc.logger.Error(ctx, "Failed to edit message",
			"sessionID", cmd.SessionID,
			"chatJID", cmd.ChatJID,
			"messageID", cmd.MessageID,
			"error", err)
		return nil, fmt.Errorf("failed to edit message: %w", err)
	}

	uc.logger.Info(ctx, "Message edited successfully",
		"sessionID", cmd.SessionID,
		"chatJID", cmd.ChatJID,
		"messageID", cmd.MessageID,
		"editMessageID", sendResp.ID,
		"timestamp", sendResp.Timestamp)

	return &MessageActionResult{
		SessionID: cmd.SessionID,
		ChatJID:   cmd.ChatJID,
		MessageID: cmd.MessageID,
		Action:    "edit",
		Success:   true,
		Message:   "Message edited successfully",
	}, nil
}

type DeleteMessageUseCase struct {
	sessionRepo     session.Repository
	whatsappService ports.WhatsAppService
	logger          ports.Logger
}

func NewDeleteMessageUseCase(
	sessionRepo session.Repository,
	whatsappService ports.WhatsAppService,
	logger ports.Logger,
) *DeleteMessageUseCase {
	return &DeleteMessageUseCase{
		sessionRepo:     sessionRepo,
		whatsappService: whatsappService,
		logger:          logger,
	}
}

func (uc *DeleteMessageUseCase) Handle(ctx context.Context, cmd DeleteMessageCommand) (*MessageActionResult, error) {
	if err := cmd.Validate(); err != nil {
		uc.logger.Warn(ctx, "Invalid delete message command", "error", err)
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
			fmt.Sprintf("session must be connected to delete messages, current status: %s", sessionEntity.Status()),
		)
	}

	if err := uc.whatsappService.DeleteMessage(ctx, cmd.SessionID, cmd.ChatJID, cmd.MessageID, cmd.ForEveryone); err != nil {
		uc.logger.Error(ctx, "Failed to delete message",
			"sessionID", cmd.SessionID,
			"chatJID", cmd.ChatJID,
			"messageID", cmd.MessageID,
			"forEveryone", cmd.ForEveryone,
			"error", err)
		return nil, fmt.Errorf("failed to delete message: %w", err)
	}

	action := "delete_for_me"
	if cmd.ForEveryone {
		action = "delete_for_everyone"
	}

	uc.logger.Info(ctx, "Message deleted successfully",
		"sessionID", cmd.SessionID,
		"chatJID", cmd.ChatJID,
		"messageID", cmd.MessageID,
		"action", action)

	return &MessageActionResult{
		SessionID: cmd.SessionID,
		ChatJID:   cmd.ChatJID,
		MessageID: cmd.MessageID,
		Action:    action,
		Success:   true,
		Message:   fmt.Sprintf("Message %s successfully", action),
	}, nil
}
