package application

import (
	"context"
	"fmt"

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

func (s *SessionApp) GetSession(ctx context.Context, sessionIDOrName string) (*session.Session, error) {
	sess, err := s.sessionRepo.GetByID(ctx, sessionIDOrName)
	if err == nil {
		return sess, nil
	}

	return s.sessionRepo.GetByName(ctx, sessionIDOrName)
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

	// Set webhook events
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
		// Messages and Communication
		"Message",
		"UndecryptableMessage",
		"Receipt",
		"MediaRetry",
		"ReadReceipt",

		// Groups and Contacts
		"GroupInfo",
		"JoinedGroup",
		"Picture",
		"BlocklistChange",
		"Blocklist",

		// Connection and Session
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
		"PairSuccess",
		"PairError",
		"QR",
		"QRScannedWithoutMultidevice",

		// Privacy and Settings
		"PrivacySettings",
		"PushNameSetting",
		"UserAbout",

		// Synchronization and State
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

		// Presence and Activity
		"Presence",
		"ChatPresence",

		// Identity
		"IdentityChange",

		// Errors
		"CATRefreshError",

		// Newsletter (WhatsApp Channels)
		"NewsletterJoin",
		"NewsletterLeave",
		"NewsletterMuteChange",
		"NewsletterLiveUpdate",

		// Facebook/Meta Bridge
		"FBMessage",

		// Special - receives all events
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
