package messaging

import (
	"context"
	"fmt"
	"strings"

	"zpmeow/internal/application/common"
	"zpmeow/internal/application/ports"
	"zpmeow/internal/domain/session"
)

type SendTextMessageCommand struct {
	SessionID string
	ChatJID   string
	Message   string
}

func (c SendTextMessageCommand) Validate() error {
	if strings.TrimSpace(c.SessionID) == "" {
		return common.NewValidationError("sessionID", c.SessionID, "session ID is required")
	}

	if strings.TrimSpace(c.ChatJID) == "" {
		return common.NewValidationError("chatJID", c.ChatJID, "chat JID is required")
	}

	if strings.TrimSpace(c.Message) == "" {
		return common.NewValidationError("message", c.Message, "message text is required")
	}

	if len(c.Message) > 4096 {
		return common.NewValidationError("message", c.Message, "message text must not exceed 4096 characters")
	}

	return nil
}

type SendTextMessageResult struct {
	SessionID string
	ChatJID   string
	MessageID string
	Sent      bool
}

type SendTextMessageUseCase struct {
	sessionRepo     session.Repository
	whatsappService ports.WhatsAppService
	logger          ports.Logger
}

func NewSendTextMessageUseCase(
	sessionRepo session.Repository,
	whatsappService ports.WhatsAppService,
	logger ports.Logger,
) *SendTextMessageUseCase {
	return &SendTextMessageUseCase{
		sessionRepo:     sessionRepo,
		whatsappService: whatsappService,
		logger:          logger,
	}
}

func (uc *SendTextMessageUseCase) Handle(ctx context.Context, cmd SendTextMessageCommand) (*SendTextMessageResult, error) {
	if err := cmd.Validate(); err != nil {
		uc.logger.Warn(ctx, "Invalid send text message command", "error", err)
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
			fmt.Sprintf("session must be connected to send messages, current status: %s", sessionEntity.Status()),
		)
	}

	if !sessionEntity.IsAuthenticated() {
		return nil, common.NewBusinessRuleError(
			"session_not_authenticated",
			"session must be authenticated to send messages",
		)
	}

	_, err = uc.whatsappService.SendTextMessage(ctx, cmd.SessionID, cmd.ChatJID, cmd.Message)
	if err != nil {
		uc.logger.Error(ctx, "Failed to send text message",
			"sessionID", cmd.SessionID,
			"chatJID", cmd.ChatJID,
			"error", err)
		return nil, fmt.Errorf("failed to send text message: %w", err)
	}

	uc.logger.Info(ctx, "Text message sent successfully",
		"sessionID", cmd.SessionID,
		"chatJID", cmd.ChatJID,
		"messageLength", len(cmd.Message))

	return &SendTextMessageResult{
		SessionID: cmd.SessionID,
		ChatJID:   cmd.ChatJID,
		MessageID: "",
		Sent:      true,
	}, nil
}
