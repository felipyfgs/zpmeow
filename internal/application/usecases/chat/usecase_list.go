package chat

import (
	"context"
	"fmt"
	"strings"

	"zpmeow/internal/application/common"
	"zpmeow/internal/application/ports"
	"zpmeow/internal/domain/session"
)

type GetChatsQuery struct {
	SessionID string
	Limit     int
	Offset    int
}

func (q GetChatsQuery) Validate() error {
	if strings.TrimSpace(q.SessionID) == "" {
		return common.NewValidationError("sessionID", q.SessionID, "session ID is required")
	}

	if q.Limit < 0 {
		return common.NewValidationError("limit", q.Limit, "limit cannot be negative")
	}

	if q.Limit > 1000 {
		return common.NewValidationError("limit", q.Limit, "limit cannot exceed 1000")
	}

	if q.Offset < 0 {
		return common.NewValidationError("offset", q.Offset, "offset cannot be negative")
	}

	return nil
}

type ChatView struct {
	JID           string
	Name          string
	LastMessage   string
	LastMessageAt string
	UnreadCount   int
	IsGroup       bool
	IsMuted       bool
	IsArchived    bool
}

type GetChatsResult struct {
	SessionID string
	Chats     []ChatView
	Total     int
	Limit     int
	Offset    int
}

type GetChatsUseCase struct {
	sessionRepo     session.Repository
	whatsappService ports.WhatsAppService
	logger          ports.Logger
}

func NewGetChatsUseCase(
	sessionRepo session.Repository,
	whatsappService ports.WhatsAppService,
	logger ports.Logger,
) *GetChatsUseCase {
	return &GetChatsUseCase{
		sessionRepo:     sessionRepo,
		whatsappService: whatsappService,
		logger:          logger,
	}
}

func (uc *GetChatsUseCase) Handle(ctx context.Context, query GetChatsQuery) (*GetChatsResult, error) {
	if err := query.Validate(); err != nil {
		uc.logger.Warn(ctx, "Invalid get chats query", "error", err)
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	sessionEntity, err := uc.sessionRepo.GetByID(ctx, query.SessionID)
	if err != nil {
		uc.logger.Error(ctx, "Failed to get session", "sessionID", query.SessionID, "error", err)
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if !sessionEntity.IsConnected() {
		return nil, common.NewBusinessRuleError(
			"session_not_connected",
			fmt.Sprintf("session must be connected to get chats, current status: %s", sessionEntity.Status()),
		)
	}

	chats, err := uc.whatsappService.GetChats(ctx, query.SessionID, query.Limit, query.Offset)
	if err != nil {
		uc.logger.Error(ctx, "Failed to get chats",
			"sessionID", query.SessionID,
			"error", err)
		return nil, fmt.Errorf("failed to get chats: %w", err)
	}

	chatViews := make([]ChatView, len(chats))
	for i, chat := range chats {
		chatViews[i] = ChatView{
			JID:           chat.JID,
			Name:          chat.Name,
			LastMessage:   chat.LastMessage,
			LastMessageAt: chat.LastMessageAt,
			UnreadCount:   chat.UnreadCount,
			IsGroup:       chat.IsGroup,
			IsMuted:       chat.IsMuted,
			IsArchived:    chat.IsArchived,
		}
	}

	uc.logger.Debug(ctx, "Chats retrieved successfully",
		"sessionID", query.SessionID,
		"count", len(chatViews))

	return &GetChatsResult{
		SessionID: query.SessionID,
		Chats:     chatViews,
		Total:     len(chatViews),
		Limit:     query.Limit,
		Offset:    query.Offset,
	}, nil
}
