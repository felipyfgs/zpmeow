package application

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"zpmeow/internal/application/ports"
	"zpmeow/internal/domain/session"
)

type SessionApp struct {
	sessionRepo   session.Repository
	domainService session.Service
}

func NewSessionApp(sessionRepo session.Repository, domainService session.Service) *SessionApp {
	return &SessionApp{
		sessionRepo:   sessionRepo,
		domainService: domainService,
	}
}

// isUUID checks if a string looks like a UUID
func isUUID(s string) bool {
	// UUID pattern: 8-4-4-4-12 hexadecimal digits
	uuidPattern := `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`
	matched, _ := regexp.MatchString(uuidPattern, s)
	return matched
}

func (s *SessionApp) GetSession(ctx context.Context, sessionIDOrName string) (*session.Session, error) {
	// Smart detection: try the most likely option first
	if isUUID(sessionIDOrName) {
		// Looks like UUID, try ID first
		sess, err := s.sessionRepo.GetByID(ctx, sessionIDOrName)
		if err == nil {
			return sess, nil
		}
		// Fallback to name (edge case)
		return s.sessionRepo.GetByName(ctx, sessionIDOrName)
	} else {
		// Looks like name, try name first
		sess, err := s.sessionRepo.GetByName(ctx, sessionIDOrName)
		if err == nil {
			return sess, nil
		}
		// Fallback to ID (edge case)
		return s.sessionRepo.GetByID(ctx, sessionIDOrName)
	}
}

func (s *SessionApp) GetAllSessions(ctx context.Context) ([]*session.Session, error) {
	return s.sessionRepo.GetAll(ctx)
}

func (s *SessionApp) CreateSessionWithRequest(ctx context.Context, req CreateSessionRequest) (*session.Session, error) {
	sess, err := session.NewSession("", req.Name)
	if err != nil {
		return nil, err
	}

	generatedID, err := s.sessionRepo.CreateWithGeneratedID(ctx, sess)
	if err != nil {
		return nil, err
	}

	return s.sessionRepo.GetByID(ctx, generatedID)
}

func (s *SessionApp) DeleteSession(ctx context.Context, sessionID string) error {
	return s.sessionRepo.Delete(ctx, sessionID)
}

func (s *SessionApp) GetSessionByDeviceJID(ctx context.Context, deviceJID string) (*session.Session, error) {
	return nil, fmt.Errorf("GetSessionByDeviceJID not implemented")
}

type CreateSessionRequest struct {
	SessionID string `json:"session_id"`
	Name      string `json:"name"`
}

type WebhookApp struct {
	sessionRepo session.Repository
}

func NewWebhookApp(sessionRepo session.Repository) *WebhookApp {
	return &WebhookApp{
		sessionRepo: sessionRepo,
	}
}

func (w *WebhookApp) SetWebhook(ctx context.Context, sessionID, webhookURL string, events []string) error {
	sess, err := w.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return err
	}

	if err := sess.SetWebhookEndpoint(webhookURL); err != nil {
		return err
	}

	sess.SetWebhookEvents(events)

	return w.sessionRepo.Update(ctx, sess)
}

func (w *WebhookApp) GetWebhook(ctx context.Context, sessionID string) (string, []string, error) {
	sess, err := w.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return "", nil, err
	}

	webhookURL := sess.GetWebhookEndpointString()
	events := []string{}

	return webhookURL, events, nil
}

func (w *WebhookApp) ListEvents(ctx context.Context) ([]string, error) {
	events := []string{
		// Messages
		"Message",
		"UndecryptableMessage",
		"Receipt",
		"MediaRetry",
		"MediaRetryError",

		// Groups
		"GroupInfo",
		"JoinedGroup",

		// Contacts and Profiles
		"Contact",
		"Picture",
		"BusinessName",
		"PushName",
		"PushNameSetting",

		// Chat Management
		"Archive",
		"Pin",
		"Mute",
		"Star",
		"DeleteChat",
		"ClearChat",
		"DeleteForMe",
		"MarkChatAsRead",

		// Blocklist
		"Blocklist",
		"BlocklistChange",

		// Labels
		"LabelAssociationChat",
		"LabelAssociationMessage",
		"LabelEdit",

		// Connection Events
		"Connected",
		"Disconnected",
		"ConnectFailure",
		"KeepAliveRestored",
		"KeepAliveTimeout",
		"LoggedOut",
		"ClientOutdated",
		"TemporaryBan",
		"StreamError",
		"StreamReplaced",

		// Pairing
		"PairSuccess",
		"PairError",
		"QR",
		"QRScannedWithoutMultidevice",

		// Settings
		"PrivacySettings",
		"UserAbout",
		"UnarchiveChatsSetting",
		"UserStatusMute",

		// Sync Events
		"AppState",
		"AppStateSyncComplete",
		"HistorySync",
		"OfflineSyncCompleted",
		"OfflineSyncPreview",

		// Calls
		"CallOffer",
		"CallAccept",
		"CallTerminate",
		"CallOfferNotice",
		"CallRelayLatency",
		"CallPreAccept",
		"CallReject",
		"CallTransport",
		"UnknownCallEvent",

		// Presence
		"Presence",
		"ChatPresence",

		// Security
		"IdentityChange",
		"CATRefreshError",

		// Newsletters
		"NewsletterJoin",
		"NewsletterLeave",
		"NewsletterMuteChange",
		"NewsletterLiveUpdate",
		"NewsletterMessageMeta",

		// Other Platforms
		"FBMessage",

		// Special
		"ManualLoginReconnect",

		// All events
		"All",
	}
	return events, nil
}

type MessageApp struct {
	sessionRepo   session.Repository
	messageSender ports.MessageSender
}

func NewMessageApp(sessionRepo session.Repository, messageSender ports.MessageSender) *MessageApp {
	return &MessageApp{
		sessionRepo:   sessionRepo,
		messageSender: messageSender,
	}
}

type ChatApp struct {
	sessionRepo session.Repository
	chatManager ports.ChatManager
}

func NewChatApp(sessionRepo session.Repository, chatManager ports.ChatManager) *ChatApp {
	return &ChatApp{
		sessionRepo: sessionRepo,
		chatManager: chatManager,
	}
}

type GetChatHistoryRequest struct {
	SessionID string
	Phone     string
	Limit     int
	Offset    int
}

type GetChatHistoryResponse struct {
	SessionID string
	Phone     string
	Messages  []ChatMessage
	Count     int
	Limit     int
	Offset    int
}

type ChatMessage struct {
	ID        string
	ChatJID   string
	FromJID   string
	Content   string
	Type      string
	Timestamp string
	IsFromMe  bool
	IsRead    bool
	MediaURL  string
	Caption   string
}

func (app *ChatApp) GetChatHistory(ctx context.Context, req GetChatHistoryRequest) (*GetChatHistoryResponse, error) {
	if req.Limit <= 0 {
		req.Limit = 50
	}
	if req.Limit > 1000 {
		req.Limit = 1000
	}

	chatJID := req.Phone + "@s.whatsapp.net"
	if strings.Contains(req.Phone, "@") {
		chatJID = req.Phone
	}

	messages, err := app.chatManager.GetChatHistory(ctx, req.SessionID, chatJID, req.Limit, req.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat history: %w", err)
	}

	chatMessages := make([]ChatMessage, len(messages))
	for i, message := range messages {
		chatMessages[i] = ChatMessage{
			ID:        message.ID,
			ChatJID:   message.ChatJID,
			FromJID:   message.FromJID,
			Content:   message.Content,
			Type:      message.Type,
			Timestamp: fmt.Sprintf("%d", message.Timestamp.Unix()),
			IsFromMe:  false,
			IsRead:    false,
			MediaURL:  "",
			Caption:   "",
		}
	}

	return &GetChatHistoryResponse{
		SessionID: req.SessionID,
		Phone:     req.Phone,
		Messages:  chatMessages,
		Count:     len(chatMessages),
		Limit:     req.Limit,
		Offset:    req.Offset,
	}, nil
}

type GroupApp struct {
	sessionRepo  session.Repository
	groupManager ports.GroupManager
}

func NewGroupApp(sessionRepo session.Repository, groupManager ports.GroupManager) *GroupApp {
	return &GroupApp{
		sessionRepo:  sessionRepo,
		groupManager: groupManager,
	}
}

type ListGroupsRequest struct {
	SessionID string
}

type ListGroupsResponse struct {
	SessionID string
	Groups    []GroupInfo
	Count     int
}

type GroupInfo struct {
	JID          string
	Name         string
	Description  string
	Participants []string
	Admins       []string
	Owner        string
	IsAnnounce   bool
	IsLocked     bool
	CreatedAt    string
}

func (app *GroupApp) ListGroups(ctx context.Context, req ListGroupsRequest) (*ListGroupsResponse, error) {
	groups, err := app.groupManager.ListGroups(ctx, req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list groups: %w", err)
	}

	groupInfos := make([]GroupInfo, len(groups))
	for i, group := range groups {
		groupInfos[i] = GroupInfo{
			JID:          group.JID,
			Name:         group.Name,
			Description:  group.Description,
			Participants: group.Participants,
			Admins:       group.Admins,
			Owner:        group.Owner,
			IsAnnounce:   group.IsAnnounce,
			IsLocked:     group.IsLocked,
			CreatedAt:    fmt.Sprintf("%d", group.CreatedAt),
		}
	}

	return &ListGroupsResponse{
		SessionID: req.SessionID,
		Groups:    groupInfos,
		Count:     len(groupInfos),
	}, nil
}

type ContactApp struct {
	sessionRepo    session.Repository
	contactManager ports.ContactManager
}

func NewContactApp(sessionRepo session.Repository, contactManager ports.ContactManager) *ContactApp {
	return &ContactApp{
		sessionRepo:    sessionRepo,
		contactManager: contactManager,
	}
}

type GetContactsRequest struct {
	SessionID string
	Limit     int
	Offset    int
}

type GetContactsResponse struct {
	SessionID string
	Contacts  []ContactInfo
	Total     int
	Limit     int
	Offset    int
}

type ContactInfo struct {
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

func (app *ContactApp) GetContacts(ctx context.Context, req GetContactsRequest) (*GetContactsResponse, error) {
	if req.Limit <= 0 {
		req.Limit = 100
	}
	if req.Limit > 1000 {
		req.Limit = 1000
	}

	contacts, err := app.contactManager.GetContacts(ctx, req.SessionID, req.Limit, req.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get contacts: %w", err)
	}

	contactInfos := make([]ContactInfo, len(contacts))
	for i, contact := range contacts {
		contactInfos[i] = ContactInfo{
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

	return &GetContactsResponse{
		SessionID: req.SessionID,
		Contacts:  contactInfos,
		Total:     len(contactInfos),
		Limit:     req.Limit,
		Offset:    req.Offset,
	}, nil
}

type CheckContactRequest struct {
	SessionID string
	Phones    []string
}

type CheckContactResponse struct {
	SessionID string
	Results   []ContactCheckResult
}

type ContactCheckResult struct {
	Query        string
	IsInWhatsapp bool
	IsInMeow     bool
	JID          string
	VerifiedName string
}

func (app *ContactApp) CheckContact(ctx context.Context, req CheckContactRequest) (*CheckContactResponse, error) {
	if len(req.Phones) == 0 {
		return nil, fmt.Errorf("at least one phone number is required")
	}

	results, err := app.contactManager.CheckUser(ctx, req.SessionID, req.Phones)
	if err != nil {
		return nil, fmt.Errorf("failed to check contacts: %w", err)
	}

	checkResults := make([]ContactCheckResult, len(results))
	for i, result := range results {
		checkResults[i] = ContactCheckResult{
			Query:        result.Query,
			IsInWhatsapp: result.IsInWhatsapp,
			IsInMeow:     result.IsInMeow,
			JID:          result.JID,
			VerifiedName: result.VerifiedName,
		}
	}

	return &CheckContactResponse{
		SessionID: req.SessionID,
		Results:   checkResults,
	}, nil
}

type NewsletterApp struct {
	sessionRepo       session.Repository
	newsletterManager ports.NewsletterManager
}

func NewNewsletterApp(sessionRepo session.Repository, newsletterManager ports.NewsletterManager) *NewsletterApp {
	return &NewsletterApp{
		sessionRepo:       sessionRepo,
		newsletterManager: newsletterManager,
	}
}
