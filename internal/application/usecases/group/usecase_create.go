package group

import (
	"context"
	"fmt"
	"strings"

	"zpmeow/internal/application/common"
	"zpmeow/internal/application/ports"
	"zpmeow/internal/domain/session"
)

type CreateGroupCommand struct {
	SessionID    string
	Name         string
	Description  string
	Participants []string
}

func (c CreateGroupCommand) Validate() error {
	if strings.TrimSpace(c.SessionID) == "" {
		return common.NewValidationError("sessionID", c.SessionID, "session ID is required")
	}

	if strings.TrimSpace(c.Name) == "" {
		return common.NewValidationError("name", c.Name, "group name is required")
	}

	if len(c.Name) > 100 {
		return common.NewValidationError("name", c.Name, "group name must not exceed 100 characters")
	}

	if len(c.Description) > 500 {
		return common.NewValidationError("description", c.Description, "group description must not exceed 500 characters")
	}

	if len(c.Participants) == 0 {
		return common.NewValidationError("participants", "", "at least one participant is required")
	}

	if len(c.Participants) > 256 {
		return common.NewValidationError("participants", "", "maximum 256 participants allowed")
	}

	for i, participant := range c.Participants {
		if strings.TrimSpace(participant) == "" {
			return common.NewValidationError("participants", participant, fmt.Sprintf("participant %d cannot be empty", i))
		}
	}

	return nil
}

type GroupView struct {
	JID          string
	Name         string
	Description  string
	Participants []string
	Admins       []string
	Owner        string
	CreatedAt    string
	IsAnnounce   bool
	IsLocked     bool
}

type CreateGroupResult struct {
	SessionID string
	Group     GroupView
	Success   bool
}

type CreateGroupUseCase struct {
	sessionRepo     session.Repository
	whatsappService ports.WhatsAppService
	logger          ports.Logger
}

func NewCreateGroupUseCase(
	sessionRepo session.Repository,
	whatsappService ports.WhatsAppService,
	logger ports.Logger,
) *CreateGroupUseCase {
	return &CreateGroupUseCase{
		sessionRepo:     sessionRepo,
		whatsappService: whatsappService,
		logger:          logger,
	}
}

func (uc *CreateGroupUseCase) Handle(ctx context.Context, cmd CreateGroupCommand) (*CreateGroupResult, error) {
	if err := cmd.Validate(); err != nil {
		uc.logger.Warn(ctx, "Invalid create group command", "error", err)
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
			fmt.Sprintf("session must be connected to create groups, current status: %s", sessionEntity.Status()),
		)
	}

	if !sessionEntity.IsAuthenticated() {
		return nil, common.NewBusinessRuleError(
			"session_not_authenticated",
			"session must be authenticated to create groups",
		)
	}

	groupJID, err := uc.whatsappService.CreateGroup(ctx, cmd.SessionID, cmd.Name, cmd.Participants)
	if err != nil {
		uc.logger.Error(ctx, "Failed to create group",
			"sessionID", cmd.SessionID,
			"groupName", cmd.Name,
			"participantCount", len(cmd.Participants),
			"error", err)
		return nil, fmt.Errorf("failed to create group: %w", err)
	}

	groupInfo, err := uc.whatsappService.GetGroupInfo(ctx, cmd.SessionID, groupJID.JID)
	if err != nil {
		uc.logger.Warn(ctx, "Failed to get group info after creation",
			"sessionID", cmd.SessionID,
			"groupJID", groupJID,
			"error", err)
		groupInfo = &ports.GroupInfo{
			JID:          groupJID.JID,
			Name:         cmd.Name,
			Description:  cmd.Description,
			Participants: cmd.Participants,
		}
	}

	groupView := GroupView{
		JID:          groupInfo.JID,
		Name:         groupInfo.Name,
		Description:  groupInfo.Description,
		Participants: groupInfo.Participants,
		Admins:       groupInfo.Admins,
		Owner:        groupInfo.Owner,
		CreatedAt:    fmt.Sprintf("%d", groupInfo.CreatedAt),
		IsAnnounce:   groupInfo.IsAnnounce,
		IsLocked:     groupInfo.IsLocked,
	}

	uc.logger.Info(ctx, "Group created successfully",
		"sessionID", cmd.SessionID,
		"groupJID", groupJID,
		"groupName", cmd.Name,
		"participantCount", len(cmd.Participants))

	return &CreateGroupResult{
		SessionID: cmd.SessionID,
		Group:     groupView,
		Success:   true,
	}, nil
}
