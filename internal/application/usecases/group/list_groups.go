package group

import (
	"context"
	"fmt"
	"strings"

	"zpmeow/internal/application/common"
	"zpmeow/internal/application/ports"
	"zpmeow/internal/domain/session"
)

type ListGroupsQuery struct {
	SessionID string
	Limit     int
	Offset    int
}

func (q ListGroupsQuery) Validate() error {
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

type ListGroupsResult struct {
	SessionID string
	Groups    []GroupView
	Total     int
	Limit     int
	Offset    int
}

type ListGroupsUseCase struct {
	sessionRepo     session.Repository
	whatsappService ports.WhatsAppService
	logger          ports.Logger
}

func NewListGroupsUseCase(
	sessionRepo session.Repository,
	whatsappService ports.WhatsAppService,
	logger ports.Logger,
) *ListGroupsUseCase {
	return &ListGroupsUseCase{
		sessionRepo:     sessionRepo,
		whatsappService: whatsappService,
		logger:          logger,
	}
}

func (uc *ListGroupsUseCase) Handle(ctx context.Context, query ListGroupsQuery) (*ListGroupsResult, error) {
	if err := query.Validate(); err != nil {
		uc.logger.Warn(ctx, "Invalid list groups query", "error", err)
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
			fmt.Sprintf("session must be connected to list groups, current status: %s", sessionEntity.Status()),
		)
	}

	groups, err := uc.whatsappService.ListGroups(ctx, query.SessionID)
	if err != nil {
		uc.logger.Error(ctx, "Failed to list groups",
			"sessionID", query.SessionID,
			"error", err)
		return nil, fmt.Errorf("failed to list groups: %w", err)
	}

	start := query.Offset
	end := start + query.Limit
	if start > len(groups) {
		start = len(groups)
	}
	if end > len(groups) {
		end = len(groups)
	}

	paginatedGroups := groups[start:end]

	groupViews := make([]GroupView, len(paginatedGroups))
	for i, group := range paginatedGroups {
		groupViews[i] = GroupView{
			JID:          group.JID,
			Name:         group.Name,
			Description:  group.Description,
			Participants: group.Participants,
			Admins:       group.Admins,
			Owner:        group.Owner,
			CreatedAt:    fmt.Sprintf("%d", group.CreatedAt),
			IsAnnounce:   group.IsAnnounce,
			IsLocked:     group.IsLocked,
		}
	}

	uc.logger.Debug(ctx, "Groups listed successfully",
		"sessionID", query.SessionID,
		"totalGroups", len(groups),
		"returnedGroups", len(groupViews))

	return &ListGroupsResult{
		SessionID: query.SessionID,
		Groups:    groupViews,
		Total:     len(groups),
		Limit:     query.Limit,
		Offset:    query.Offset,
	}, nil
}

type GetGroupInfoQuery struct {
	SessionID string
	GroupJID  string
}

func (q GetGroupInfoQuery) Validate() error {
	if strings.TrimSpace(q.SessionID) == "" {
		return common.NewValidationError("sessionID", q.SessionID, "session ID is required")
	}

	if strings.TrimSpace(q.GroupJID) == "" {
		return common.NewValidationError("groupJID", q.GroupJID, "group JID is required")
	}

	return nil
}

type GetGroupInfoUseCase struct {
	sessionRepo     session.Repository
	whatsappService ports.WhatsAppService
	logger          ports.Logger
}

func NewGetGroupInfoUseCase(
	sessionRepo session.Repository,
	whatsappService ports.WhatsAppService,
	logger ports.Logger,
) *GetGroupInfoUseCase {
	return &GetGroupInfoUseCase{
		sessionRepo:     sessionRepo,
		whatsappService: whatsappService,
		logger:          logger,
	}
}

func (uc *GetGroupInfoUseCase) Handle(ctx context.Context, query GetGroupInfoQuery) (*GroupView, error) {
	if err := query.Validate(); err != nil {
		uc.logger.Warn(ctx, "Invalid get group info query", "error", err)
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
			fmt.Sprintf("session must be connected to get group info, current status: %s", sessionEntity.Status()),
		)
	}

	groupInfo, err := uc.whatsappService.GetGroupInfo(ctx, query.SessionID, query.GroupJID)
	if err != nil {
		uc.logger.Error(ctx, "Failed to get group info",
			"sessionID", query.SessionID,
			"groupJID", query.GroupJID,
			"error", err)
		return nil, fmt.Errorf("failed to get group info: %w", err)
	}

	groupView := &GroupView{
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

	uc.logger.Debug(ctx, "Group info retrieved successfully",
		"sessionID", query.SessionID,
		"groupJID", query.GroupJID,
		"groupName", groupInfo.Name)

	return groupView, nil
}
