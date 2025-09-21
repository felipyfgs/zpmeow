package contact

import (
	"context"
	"fmt"
	"strings"

	"zpmeow/internal/application/common"
	"zpmeow/internal/application/ports"
	"zpmeow/internal/domain/session"
)

type GetContactsQuery struct {
	SessionID string
	Limit     int
	Offset    int
}

func (q GetContactsQuery) Validate() error {
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

type ContactView struct {
	JID          string
	Name         string
	Notify       string
	PushName     string
	BusinessName string
	IsBlocked    bool
	IsMuted      bool
	IsContact    bool
	Avatar       string
}

type GetContactsResult struct {
	SessionID string
	Contacts  []ContactView
	Total     int
	Limit     int
	Offset    int
}

type GetContactsUseCase struct {
	sessionRepo     session.Repository
	whatsappService ports.WhatsAppService
	logger          ports.Logger
}

func NewGetContactsUseCase(
	sessionRepo session.Repository,
	whatsappService ports.WhatsAppService,
	logger ports.Logger,
) *GetContactsUseCase {
	return &GetContactsUseCase{
		sessionRepo:     sessionRepo,
		whatsappService: whatsappService,
		logger:          logger,
	}
}

func (uc *GetContactsUseCase) Handle(ctx context.Context, query GetContactsQuery) (*GetContactsResult, error) {
	if err := query.Validate(); err != nil {
		uc.logger.Warn(ctx, "Invalid get contacts query", "error", err)
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
			fmt.Sprintf("session must be connected to get contacts, current status: %s", sessionEntity.Status()),
		)
	}

	contacts, err := uc.whatsappService.GetContacts(ctx, query.SessionID, query.Limit, query.Offset)
	if err != nil {
		uc.logger.Error(ctx, "Failed to get contacts",
			"sessionID", query.SessionID,
			"error", err)
		return nil, fmt.Errorf("failed to get contacts: %w", err)
	}

	contactViews := make([]ContactView, len(contacts))
	for i, contact := range contacts {
		contactViews[i] = ContactView{
			JID:          contact.JID,
			Name:         contact.Name,
			Notify:       contact.Notify,
			PushName:     contact.PushName,
			BusinessName: contact.BusinessName,
			IsBlocked:    contact.IsBlocked,
			IsMuted:      contact.IsMuted,
			IsContact:    contact.IsContact,
			Avatar:       contact.Avatar,
		}
	}

	uc.logger.Debug(ctx, "Contacts retrieved successfully",
		"sessionID", query.SessionID,
		"count", len(contactViews))

	return &GetContactsResult{
		SessionID: query.SessionID,
		Contacts:  contactViews,
		Total:     len(contactViews),
		Limit:     query.Limit,
		Offset:    query.Offset,
	}, nil
}

type CheckContactQuery struct {
	SessionID string
	Phone     string
}

func (q CheckContactQuery) Validate() error {
	if strings.TrimSpace(q.SessionID) == "" {
		return common.NewValidationError("sessionID", q.SessionID, "session ID is required")
	}

	if strings.TrimSpace(q.Phone) == "" {
		return common.NewValidationError("phone", q.Phone, "phone number is required")
	}

	return nil
}

type CheckContactResult struct {
	SessionID    string
	Phone        string
	IsOnWhatsApp bool
	JID          string
}

type CheckContactUseCase struct {
	sessionRepo     session.Repository
	whatsappService ports.WhatsAppService
	logger          ports.Logger
}

func NewCheckContactUseCase(
	sessionRepo session.Repository,
	whatsappService ports.WhatsAppService,
	logger ports.Logger,
) *CheckContactUseCase {
	return &CheckContactUseCase{
		sessionRepo:     sessionRepo,
		whatsappService: whatsappService,
		logger:          logger,
	}
}

func (uc *CheckContactUseCase) Handle(ctx context.Context, query CheckContactQuery) (*CheckContactResult, error) {
	if err := query.Validate(); err != nil {
		uc.logger.Warn(ctx, "Invalid check contact query", "error", err)
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
			fmt.Sprintf("session must be connected to check contacts, current status: %s", sessionEntity.Status()),
		)
	}

	checkResult, err := uc.whatsappService.CheckContact(ctx, query.SessionID, query.Phone)
	if err != nil {
		uc.logger.Error(ctx, "Failed to check contact",
			"sessionID", query.SessionID,
			"phone", query.Phone,
			"error", err)
		return nil, fmt.Errorf("failed to check contact: %w", err)
	}

	isOnWhatsApp := checkResult.IsInWhatsapp
	jid := checkResult.JID

	uc.logger.Debug(ctx, "Contact checked successfully",
		"sessionID", query.SessionID,
		"phone", query.Phone,
		"isOnWhatsApp", isOnWhatsApp)

	return &CheckContactResult{
		SessionID:    query.SessionID,
		Phone:        query.Phone,
		IsOnWhatsApp: isOnWhatsApp,
		JID:          jid,
	}, nil
}

type GetUserInfoQuery struct {
	SessionID string
	UserJID   string
}

func (q GetUserInfoQuery) Validate() error {
	if strings.TrimSpace(q.SessionID) == "" {
		return common.NewValidationError("sessionID", q.SessionID, "session ID is required")
	}

	if strings.TrimSpace(q.UserJID) == "" {
		return common.NewValidationError("userJID", q.UserJID, "user JID is required")
	}

	return nil
}

type UserInfoView struct {
	JID          string
	Name         string
	Notify       string
	PushName     string
	BusinessName string
	Phone        string
	Status       string
	Avatar       string
	IsBlocked    bool
	IsMuted      bool
	IsContact    bool
	LastSeen     string
}

type GetUserInfoUseCase struct {
	sessionRepo     session.Repository
	whatsappService ports.WhatsAppService
	logger          ports.Logger
}

func NewGetUserInfoUseCase(
	sessionRepo session.Repository,
	whatsappService ports.WhatsAppService,
	logger ports.Logger,
) *GetUserInfoUseCase {
	return &GetUserInfoUseCase{
		sessionRepo:     sessionRepo,
		whatsappService: whatsappService,
		logger:          logger,
	}
}

func (uc *GetUserInfoUseCase) Handle(ctx context.Context, query GetUserInfoQuery) (*UserInfoView, error) {
	if err := query.Validate(); err != nil {
		uc.logger.Warn(ctx, "Invalid get user info query", "error", err)
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
			fmt.Sprintf("session must be connected to get user info, current status: %s", sessionEntity.Status()),
		)
	}

	userInfoMap, err := uc.whatsappService.GetUserInfo(ctx, query.SessionID, []string{query.UserJID})
	if err != nil {
		uc.logger.Error(ctx, "Failed to get user info",
			"sessionID", query.SessionID,
			"userJID", query.UserJID,
			"error", err)
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	userInfo, exists := userInfoMap[query.UserJID]
	if !exists {
		return nil, fmt.Errorf("user info not found for JID: %s", query.UserJID)
	}

	phone := ""
	if jidParts := strings.Split(userInfo.JID, "@"); len(jidParts) > 0 {
		phone = jidParts[0]
	}

	userInfoView := &UserInfoView{
		JID:          userInfo.JID,
		Name:         userInfo.Name,
		Notify:       userInfo.Notify,
		PushName:     userInfo.PushName,
		BusinessName: userInfo.BusinessName,
		Phone:        phone,
		Status:       "",
		Avatar:       "",
		IsBlocked:    userInfo.IsBlocked,
		IsMuted:      userInfo.IsMuted,
		IsContact:    false,
		LastSeen:     "",
	}

	uc.logger.Debug(ctx, "User info retrieved successfully",
		"sessionID", query.SessionID,
		"userJID", query.UserJID,
		"userName", userInfo.Name)

	return userInfoView, nil
}
