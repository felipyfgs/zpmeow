package ports

import (
	"context"
	"fmt"
	"time"

	"go.mau.fi/whatsmeow"
	"zpmeow/internal/domain/session"
)

type SessionManager interface {
	StartClient(sessionID string) error
	StopClient(sessionID string) error
	LogoutClient(sessionID string) error
	GetQRCode(sessionID string) (string, error)
	PairPhone(sessionID, phoneNumber string) (string, error)
	IsClientConnected(sessionID string) bool

	ConnectOnStartup(ctx context.Context) error
	ConnectSession(ctx context.Context, sessionID string) (string, error)
	DisconnectSession(ctx context.Context, sessionID string) error
}

type ButtonData struct {
	ID   string `json:"id"`
	Text string `json:"text"`
	Type string `json:"type,omitempty"`
}

type ListItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type ListSection struct {
	Title string     `json:"title"`
	Rows  []ListItem `json:"rows"`
}

type MediaMessage struct {
	Type     string `json:"type"`
	Data     []byte `json:"data"`
	MimeType string `json:"mime_type"`
	Caption  string `json:"caption,omitempty"`
	Filename string `json:"filename,omitempty"`
}

type ContactData struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

type MessageSender interface {
	SendTextMessage(ctx context.Context, sessionID, phone, text string) (*whatsmeow.SendResponse, error)
	SendMediaMessage(ctx context.Context, sessionID, phone string, media MediaMessage) (*whatsmeow.SendResponse, error)
	SendImageMessage(ctx context.Context, sessionID, phone string, data []byte, caption, mimeType string) (*whatsmeow.SendResponse, error)
	SendAudioMessage(ctx context.Context, sessionID, phone string, data []byte, mimeType string) (*whatsmeow.SendResponse, error)
	SendAudioMessageWithPTT(ctx context.Context, sessionID, phone string, data []byte, mimeType string, ptt bool) (*whatsmeow.SendResponse, error)
	SendVideoMessage(ctx context.Context, sessionID, phone string, data []byte, caption, mimeType string) (*whatsmeow.SendResponse, error)
	SendDocumentMessage(ctx context.Context, sessionID, phone string, data []byte, filename, caption, mimeType string) (*whatsmeow.SendResponse, error)
	SendStickerMessage(ctx context.Context, sessionID, phone string, data []byte, mimeType string) (*whatsmeow.SendResponse, error)

	SendContactsMessage(ctx context.Context, sessionID, phone string, contacts []ContactData) (*whatsmeow.SendResponse, error)
	SendLocationMessage(ctx context.Context, sessionID, phone string, latitude, longitude float64, name, address string) (*whatsmeow.SendResponse, error)
	SendButtonMessage(ctx context.Context, sessionID, phone, title string, buttons []ButtonData) (*whatsmeow.SendResponse, error)
	SendListMessage(ctx context.Context, sessionID, phone, title, description, buttonText, footerText string, sections []ListSection) (*whatsmeow.SendResponse, error)
	SendPollMessage(ctx context.Context, sessionID, phone, name string, options []string, selectableCount int) (*whatsmeow.SendResponse, error)
}

type MessageActions interface {
	MarkAsRead(ctx context.Context, sessionID, phone string, messageIDs []string) error
	SetPresence(ctx context.Context, sessionID, phone, state, media string) error
	ReactToMessage(ctx context.Context, sessionID, phone, messageID, emoji string) error
	DeleteMessage(ctx context.Context, sessionID, phone, messageID string, forEveryone bool) error
	EditMessage(ctx context.Context, sessionID, phone, messageID, newText string) (*whatsmeow.SendResponse, error)
	DownloadMedia(ctx context.Context, sessionID, messageID string) ([]byte, string, error)
}

type GroupInfo struct {
	JID          string   `json:"jid"`
	Name         string   `json:"name"`
	Topic        string   `json:"topic,omitempty"`
	Description  string   `json:"description,omitempty"`
	Participants []string `json:"participants"`
	Admins       []string `json:"admins"`
	Owner        string   `json:"owner,omitempty"`
	IsAnnounce   bool     `json:"is_announce"`
	IsLocked     bool     `json:"is_locked"`
	IsEphemeral  bool     `json:"is_ephemeral"`
	CreatedAt    int64    `json:"created_at"`
	CreatedBy    string   `json:"created_by,omitempty"`
}

type GroupManager interface {
	CreateGroup(ctx context.Context, sessionID, name string, participants []string) (*GroupInfo, error)
	ListGroups(ctx context.Context, sessionID string) ([]GroupInfo, error)
	GetGroupInfo(ctx context.Context, sessionID, groupJID string) (*GroupInfo, error)
	JoinGroup(ctx context.Context, sessionID, inviteLink string) (*GroupInfo, error)
	JoinGroupWithInvite(ctx context.Context, sessionID, groupJID, inviter, code string, expiration int64) (*GroupInfo, error)
	LeaveGroup(ctx context.Context, sessionID, groupJID string) error
	GetInviteLink(ctx context.Context, sessionID, groupJID string, reset bool) (string, error)
	GetGroupInviteLink(ctx context.Context, sessionID, groupJID string) (string, error)
	GetInviteInfo(ctx context.Context, sessionID, inviteLink string) (*GroupInfo, error)
	GetGroupInfoFromInvite(ctx context.Context, sessionID, groupJID, inviter, code string, expiration int64) (*GroupInfo, error)
	UpdateParticipants(ctx context.Context, sessionID, groupJID, action string, participants []string) error
	AddParticipants(ctx context.Context, sessionID, groupJID string, participants []string) error
	RemoveParticipants(ctx context.Context, sessionID, groupJID string, participants []string) error
	SetGroupName(ctx context.Context, sessionID, groupJID, name string) error
	SetGroupTopic(ctx context.Context, sessionID, groupJID, topic string) error
	SetGroupPhoto(ctx context.Context, sessionID, groupJID string, photo []byte) error
	RemoveGroupPhoto(ctx context.Context, sessionID, groupJID string) error
	SetGroupAnnounce(ctx context.Context, sessionID, groupJID string, announceOnly bool) error
	SetGroupLocked(ctx context.Context, sessionID, groupJID string, locked bool) error
	SetGroupEphemeral(ctx context.Context, sessionID, groupJID string, ephemeral bool, duration int) error
	SetGroupJoinApprovalMode(ctx context.Context, sessionID, groupJID string, requireApproval bool) error
	SetGroupJoinApproval(ctx context.Context, sessionID, groupJID string, requireApproval bool) error
	SetGroupMemberAddMode(ctx context.Context, sessionID, groupJID string, mode string) error
	GetGroupRequestParticipants(ctx context.Context, sessionID, groupJID string) ([]string, error)
	UpdateGroupRequestParticipants(ctx context.Context, sessionID, groupJID, action string, participants []string) error
	LinkGroup(ctx context.Context, sessionID, communityJID, groupJID string) error
	UnlinkGroup(ctx context.Context, sessionID, communityJID, groupJID string) error
	GetSubGroups(ctx context.Context, sessionID, communityJID string) ([]string, error)
	GetLinkedGroupsParticipants(ctx context.Context, sessionID, communityJID string) ([]string, error)
}

type UserCheckResult struct {
	Query        string `json:"query"`
	IsInWhatsapp bool   `json:"is_in_whatsapp"`
	IsInMeow     bool   `json:"is_in_meow"`
	JID          string `json:"jid"`
	VerifiedName string `json:"verified_name,omitempty"`
}

type UserInfoResult struct {
	JID          string `json:"jid"`
	Name         string `json:"name,omitempty"`
	Notify       string `json:"notify,omitempty"`
	PushName     string `json:"push_name,omitempty"`
	BusinessName string `json:"business_name,omitempty"`
	IsBlocked    bool   `json:"is_blocked"`
	IsMuted      bool   `json:"is_muted"`
}

type AvatarResult struct {
	Phone     string `json:"phone"`
	JID       string `json:"jid"`
	AvatarURL string `json:"avatar_url,omitempty"`
	PictureID string `json:"picture_id,omitempty"`
}

type ContactResult struct {
	JID          string `json:"jid"`
	Name         string `json:"name,omitempty"`
	Notify       string `json:"notify,omitempty"`
	PushName     string `json:"push_name,omitempty"`
	BusinessName string `json:"business_name,omitempty"`
	IsBlocked    bool   `json:"is_blocked"`
	IsMuted      bool   `json:"is_muted"`
	IsContact    bool   `json:"is_contact"`
	Avatar       string `json:"avatar,omitempty"`
}

type ContactManager interface {
	CheckUser(ctx context.Context, sessionID string, phones []string) ([]UserCheckResult, error)
	CheckContact(ctx context.Context, sessionID, phone string) (*UserCheckResult, error)
	GetUserInfo(ctx context.Context, sessionID string, phones []string) (map[string]UserInfoResult, error)
	GetAvatar(ctx context.Context, sessionID, phone string) (*AvatarResult, error)
	GetContacts(ctx context.Context, sessionID string, limit, offset int) ([]ContactResult, error)
	SetUserPresence(ctx context.Context, sessionID, state string) error
}

type ChatInfo struct {
	JID           string    `json:"jid"`
	Name          string    `json:"name,omitempty"`
	Type          string    `json:"type"`
	IsGroup       bool      `json:"is_group"`
	IsPinned      bool      `json:"is_pinned"`
	IsMuted       bool      `json:"is_muted"`
	IsArchived    bool      `json:"is_archived"`
	UnreadCount   int       `json:"unread_count"`
	LastMessage   string    `json:"last_message,omitempty"`
	LastMessageAt string    `json:"last_message_at,omitempty"`
	LastSeen      time.Time `json:"last_seen,omitempty"`
	CreatedAt     time.Time `json:"created_at,omitempty"`
}

type ChatMessage struct {
	ID        string    `json:"id"`
	ChatJID   string    `json:"chat_jid"`
	FromJID   string    `json:"from_jid"`
	Text      string    `json:"text"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
}

type ChatManager interface {
	SetDisappearingTimer(ctx context.Context, sessionID, chatJID string, duration time.Duration) error
	ListChats(ctx context.Context, sessionID, chatType string) ([]ChatInfo, error)
	GetChats(ctx context.Context, sessionID string, limit, offset int) ([]ChatInfo, error)
	GetChatInfo(ctx context.Context, sessionID, chatJID string) (*ChatInfo, error)
	PinChat(ctx context.Context, sessionID, chatJID string, pinned bool) error
	MuteChat(ctx context.Context, sessionID, chatJID string, muted bool, duration time.Duration) error
	ArchiveChat(ctx context.Context, sessionID, chatJID string, archived bool) error
	GetChatHistory(ctx context.Context, sessionID, chatJID string, limit, offset int) ([]ChatMessage, error)
}

type NewsletterInfo struct {
	ID              string `json:"id"`
	JID             string `json:"jid"`
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	Subscribers     int    `json:"subscribers"`
	SubscriberCount int    `json:"subscriber_count"`
	Verified        bool   `json:"verified"`
	IsVerified      bool   `json:"is_verified"`
	IsSubscribed    bool   `json:"is_subscribed"`
	Muted           bool   `json:"muted"`
	Following       bool   `json:"following"`
	CreatedAt       int64  `json:"created_at"`
	ServerID        string `json:"server_id,omitempty"`
	Timestamp       int64  `json:"timestamp"`
}

type NewsletterMessage struct {
	ID           string         `json:"id"`
	NewsletterID string         `json:"newsletter_id"`
	Content      string         `json:"content"`
	Timestamp    int64          `json:"timestamp"`
	Views        int            `json:"views"`
	Reactions    map[string]int `json:"reactions,omitempty"`
}

type NewsletterManager interface {
	GetNewsletterMessageUpdates(ctx context.Context, sessionID, newsletterID string) ([]NewsletterMessage, error)
	NewsletterMarkViewed(ctx context.Context, sessionID, newsletterID string, messageIDs []string) error
	NewsletterSendReaction(ctx context.Context, sessionID, newsletterID, messageID, reaction string) error
	NewsletterToggleMute(ctx context.Context, sessionID, newsletterID string, muted bool) error
	NewsletterSubscribeLiveUpdates(ctx context.Context, sessionID, newsletterID string) error
	UploadNewsletter(ctx context.Context, sessionID string, data []byte) error
	GetNewsletterInfoWithInvite(ctx context.Context, sessionID, inviteCode string) (*NewsletterInfo, error)
	CreateNewsletter(ctx context.Context, sessionID, name, description string) (*NewsletterInfo, error)
	GetNewsletterInfo(ctx context.Context, sessionID, newsletterID string) (*NewsletterInfo, error)
	GetSubscribedNewsletters(ctx context.Context, sessionID string) ([]NewsletterInfo, error)
	FollowNewsletter(ctx context.Context, sessionID, newsletterID string) error
	UnfollowNewsletter(ctx context.Context, sessionID, newsletterID string) error
	SendNewsletterMessage(ctx context.Context, sessionID, newsletterID, message string) error
	GetNewsletterMessages(ctx context.Context, sessionID, newsletterID string) ([]NewsletterMessage, error)
}

type PrivacySettings struct {
	LastSeen             string `json:"last_seen"`
	ProfilePhoto         string `json:"profile_photo"`
	About                string `json:"about"`
	Status               string `json:"status"`
	ReadReceipts         bool   `json:"read_receipts"`
	GroupsAddMe          string `json:"groups_add_me"`
	CallsAddMe           string `json:"calls_add_me"`
	DisappearingMessages string `json:"disappearing_messages"`
}

type PrivacyManager interface {
	GetPrivacySettings(ctx context.Context, sessionID string) (*PrivacySettings, error)
	SetPrivacySetting(ctx context.Context, sessionID, setting, value string) error
	GetBlocklist(ctx context.Context, sessionID string) ([]string, error)
	UpdateBlocklist(ctx context.Context, sessionID string, action string, contacts []string) error
}

type WebhookManager interface {
	UpdateSessionWebhook(sessionID, webhookURL string) error
	UpdateSessionSubscriptions(sessionID string, events []string) error
}

type ProfileManager interface {
	UpdateProfile(ctx context.Context, sessionID, name, about string) error
	SetProfilePicture(ctx context.Context, sessionID string, imageData []byte) error
	RemoveProfilePicture(ctx context.Context, sessionID string) error
	GetUserStatus(ctx context.Context, sessionID, phone string) (string, error)
	SetStatus(ctx context.Context, sessionID, status string) error
}

type MediaManager interface {
	UploadMedia(ctx context.Context, sessionID string, data []byte, mediaType string) (string, error)
	GetMediaInfo(ctx context.Context, sessionID, mediaID string) (map[string]interface{}, error)
	DeleteMedia(ctx context.Context, sessionID, mediaID string) error
	ListMedia(ctx context.Context, sessionID string, limit, offset int) ([]map[string]interface{}, error)
	GetMediaProgress(ctx context.Context, sessionID, mediaID string) (map[string]interface{}, error)
	ConvertMedia(ctx context.Context, sessionID, mediaID, targetFormat string) (string, error)
	CompressMedia(ctx context.Context, sessionID, mediaID string, quality int) (string, error)
	GetMediaMetadata(ctx context.Context, sessionID, mediaID string) (map[string]interface{}, error)
}

type WameowService interface {
	SessionManager
	MessageSender
	MessageActions
	GroupManager
	ContactManager
	ChatManager
	NewsletterManager
	PrivacyManager
	WebhookManager
	ProfileManager
	MediaManager
}

type WhatsAppService = WameowService

type Logger interface {
	Debug(ctx context.Context, msg string, keysAndValues ...interface{})
	Info(ctx context.Context, msg string, keysAndValues ...interface{})
	Warn(ctx context.Context, msg string, keysAndValues ...interface{})
	Error(ctx context.Context, msg string, keysAndValues ...interface{})
	Fatal(ctx context.Context, msg string, keysAndValues ...interface{})
}

type IDGenerator interface {
	Generate() string
	GenerateAPIKey() string
}

// ChatwootService defines the interface for Chatwoot integration service
type ChatwootService interface {
	ProcessWebhook(ctx context.Context, sessionID string, payload []byte) error
	SendMessageToWhatsApp(ctx context.Context, sessionID, phone, content string) error
	ProcessWhatsAppMessage(ctx context.Context, msg *WhatsAppMessage) error
	SetWhatsAppService(service WhatsAppService)
}

// ChatwootClient defines the interface for Chatwoot API operations
type ChatwootClient interface {
	// Contact operations
	CreateContact(ctx context.Context, request ContactCreateRequest) (*ContactResponse, error)
	GetContact(ctx context.Context, contactID int) (*ContactResponse, error)
	SearchContacts(ctx context.Context, query string) ([]*ContactResponse, error)
	FilterContacts(ctx context.Context, query string) ([]*ContactResponse, error)

	// Conversation operations
	CreateConversation(ctx context.Context, request ConversationCreateRequest) (*ConversationResponse, error)
	GetConversation(ctx context.Context, conversationID int) (*ConversationResponse, error)
	ListContactConversations(ctx context.Context, contactID int) ([]*ConversationResponse, error)

	// Message operations
	CreateMessage(ctx context.Context, conversationID int, request MessageCreateRequest) (*MessageResponse, error)
	CreateMessageWithAttachment(ctx context.Context, conversationID int, content, messageType string, attachment []byte, filename, sourceID string) (*MessageResponse, error)

	// Inbox operations
	CreateInbox(ctx context.Context, request InboxCreateRequest) (*InboxResponse, error)
	ListInboxes(ctx context.Context) ([]*InboxResponse, error)
	GetInbox(ctx context.Context, inboxID int) (*InboxResponse, error)
}

// ChatwootContactManager manages contact operations
type ChatwootContactManager interface {
	FindOrCreateContact(ctx context.Context, phoneNumber, name, avatarURL string, isGroup bool, inboxID int) (*ContactResponse, error)
	SearchExistingContact(ctx context.Context, phoneNumber string, isGroup bool) (*ContactResponse, error)
	CreateNewContact(ctx context.Context, phoneNumber, name, avatarURL string, isGroup bool, inboxID int) (*ContactResponse, error)
	ValidateContact(contact *ContactResponse) error
}

// ChatwootMessageProcessor handles message processing between WhatsApp and Chatwoot
type ChatwootMessageProcessor interface {
	ProcessIncomingMessage(ctx context.Context, msg *WhatsAppMessage, conversationID int) (*MessageResponse, error)
	ProcessOutgoingMessage(ctx context.Context, payload *WebhookPayload) error
	FormatMessageContent(msg *WhatsAppMessage) string
	GetContentType(msg *WhatsAppMessage) string
	HasMediaContent(msg *WhatsAppMessage) bool
}

// ChatwootConversationManager manages conversation operations
type ChatwootConversationManager interface {
	GetOrCreateConversation(ctx context.Context, contactID int, inboxID int) (*ConversationResponse, error)
	FindActiveConversation(ctx context.Context, contactID int, inboxID int) (*ConversationResponse, error)
	CreateNewConversation(ctx context.Context, contactID int, inboxID int) (*ConversationResponse, error)
	MapConversation(ctx context.Context, chatJID string, contactID int, conversationID int) error
}

// ChatwootInboxManager manages inbox operations
type ChatwootInboxManager interface {
	InitializeInbox(ctx context.Context, config *ChatwootConfig) (*InboxResponse, error)
	FindInboxByName(ctx context.Context, name string) (*InboxResponse, error)
	CreateInbox(ctx context.Context, name, webhookURL string) (*InboxResponse, error)
	ValidateInbox(ctx context.Context, inbox *InboxResponse) error
	GenerateWebhookURL(sessionID string) string
}

// ChatwootCacheManager manages caching for Chatwoot operations
type ChatwootCacheManager interface {
	// Contact cache
	GetContact(phoneNumber string) (*ContactResponse, bool)
	SetContact(phoneNumber string, contact *ContactResponse, ttl time.Duration)
	DeleteContact(phoneNumber string)

	// Conversation cache
	GetConversation(contactID int) (*ConversationResponse, bool)
	SetConversation(contactID int, conversation *ConversationResponse, ttl time.Duration)
	DeleteConversation(contactID int)

	// General cache operations
	Clear()
	Size() int
	Close()
}

// ChatwootErrorHandler handles error processing and logging
type ChatwootErrorHandler interface {
	HandleContactError(err error, phoneNumber string) error
	HandleMessageError(err error, messageID string) error
	HandleConversationError(err error, conversationID int) error
	WrapError(err error, operation string, context map[string]interface{}) error
}

// ChatwootLogger provides structured logging for Chatwoot operations
type ChatwootLogger interface {
	LogContactOperation(operation string, phoneNumber string, success bool, details map[string]interface{})
	LogMessageOperation(operation string, messageID string, success bool, details map[string]interface{})
	LogConversationOperation(operation string, conversationID int, success bool, details map[string]interface{})
	LogAPICall(method, endpoint string, statusCode int, duration time.Duration)
}

// ChatwootValidator validates data for Chatwoot operations
type ChatwootValidator interface {
	ValidatePhoneNumber(phoneNumber string) error
	ValidateContactData(name, phoneNumber string, isGroup bool) error
	ValidateMessageContent(content string, contentType string) error
	ValidateWebhookPayload(payload []byte) error
	ValidateConversationRequest(req *ConversationCreateRequest) error
	ValidateMessageRequest(req *MessageCreateRequest) error
	ValidateInboxRequest(req *InboxCreateRequest) error
	ValidateURL(url string) error
	ValidateToken(token string) error
	ValidateAccountID(accountID int) error
}

// Chatwoot data types for interfaces
type WhatsAppMessage struct {
	ID        string                 `json:"id"`
	From      string                 `json:"from"`
	To        string                 `json:"to"`
	Body      string                 `json:"body"`
	Type      string                 `json:"type"`
	Timestamp float64                `json:"timestamp"`
	FromMe    bool                   `json:"from_me"`
	PushName  string                 `json:"push_name"`
	ChatName  string                 `json:"chat_name"`
	Caption   string                 `json:"caption"`
	FileName  string                 `json:"file_name"`
	MediaURL  string                 `json:"media_url"`
	MimeType  string                 `json:"mime_type"`
	Data      map[string]interface{} `json:"data"`
}

type WebhookPayload struct {
	Event        string                 `json:"event"`
	Account      map[string]interface{} `json:"account"`
	Conversation map[string]interface{} `json:"conversation"`
	Message      map[string]interface{} `json:"message"`
	Contact      map[string]interface{} `json:"contact"`
	Inbox        map[string]interface{} `json:"inbox"`
}

type ChatwootConfig struct {
	IsActive   bool     `json:"is_active"`
	URL        string   `json:"url"`
	Token      string   `json:"token"`
	AccountID  int      `json:"account_id"`
	NameInbox  string   `json:"name_inbox"`
	AutoCreate bool     `json:"auto_create"`
	IgnoreJids []string `json:"ignore_jids"`
}

// Request types
type ContactCreateRequest struct {
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number,omitempty"`
	Email       string `json:"email,omitempty"`
	Identifier  string `json:"identifier,omitempty"`
	InboxID     int    `json:"inbox_id"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

type ConversationCreateRequest struct {
	ContactID int    `json:"contact_id"`
	InboxID   int    `json:"inbox_id"`
	Status    string `json:"status,omitempty"`
}

type MessageCreateRequest struct {
	Content     string `json:"content"`
	MessageType int    `json:"message_type"`
	SourceID    string `json:"source_id,omitempty"`
}

type InboxCreateRequest struct {
	Name    string                 `json:"name"`
	Channel map[string]interface{} `json:"channel"`
}

// Response types
type ContactResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email"`
	Identifier  string `json:"identifier"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type ConversationResponse struct {
	ID        int    `json:"id"`
	InboxID   int    `json:"inbox_id"`
	Status    string `json:"status"`
	ContactID int    `json:"contact_id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type MessageResponse struct {
	ID             int    `json:"id"`
	Content        string `json:"content"`
	MessageType    int    `json:"message_type"`
	ConversationID int    `json:"conversation_id"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

type InboxResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	ChannelType string `json:"channel_type"`
	WebhookURL  string `json:"webhook_url"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type ContactInfo struct {
	Name         string `json:"name"`
	Phone        string `json:"phone"`
	Email        string `json:"email,omitempty"`
	Organization string `json:"organization,omitempty"`
}

type NotificationService interface {
	SendNotification(ctx context.Context, message string) error
	SendWebhook(ctx context.Context, url string, payload interface{}) error
}

// HTTP Client interface for external requests
type HTTPClient interface {
	Post(ctx context.Context, url string, payload interface{}, headers map[string]string) error
	Get(ctx context.Context, url string, headers map[string]string) ([]byte, error)
	Put(ctx context.Context, url string, payload interface{}, headers map[string]string) error
	Delete(ctx context.Context, url string, headers map[string]string) error
}

// Media operations interface
type MediaUploader interface {
	UploadMedia(ctx context.Context, data []byte, mediaType string) (*MediaUploadResult, error)
}

type MediaUploadResult struct {
	URL      string `json:"url"`
	MediaKey string `json:"media_key"`
	FileSize int64  `json:"file_size"`
}

// File operations interface
type FileDownloader interface {
	Download(ctx context.Context, url string) ([]byte, error)
	DownloadToFile(ctx context.Context, url, filePath string) error
}

// Validation interfaces - consolidated from infra layers
type MessageValidator interface {
	ValidateTextMessage(content string) error
	ValidateMediaMessage(data []byte, mediaType string) error
	ValidatePhoneNumber(phone string) error
	ValidateClient(client interface{}) error
	ValidateRecipient(to string) error
}

type SessionValidator interface {
	ValidateSessionID(sessionID string) error
	ValidateSessionName(name string) error
}

type PhoneValidator interface {
	ValidatePhoneNumber(phone string) error
	NormalizePhoneNumber(phone string) (string, error)
	FormatPhoneNumber(phone string) string
}

type URLValidator interface {
	ValidateURL(url string) error
	ValidateScheme(url string, allowedSchemes []string) error
	ExtractScheme(url string) string
	HasHost(url string) bool
}

// Cache interface - consolidated from cache.go
type CacheManager interface {
	// Generic cache operations
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error

	// Session-specific cache operations
	GetSession(ctx context.Context, sessionID string) (*session.Session, error)
	SetSession(ctx context.Context, sessionID string, sess *session.Session, ttl time.Duration) error
	DeleteSession(ctx context.Context, sessionID string) error

	// QR Code cache operations
	GetQRCode(ctx context.Context, sessionID string) (string, error)
	SetQRCode(ctx context.Context, sessionID string, qrCode string, ttl time.Duration) error
	DeleteQRCode(ctx context.Context, sessionID string) error
	GetQRCodeBase64(ctx context.Context, sessionID string) (string, error)
	SetQRCodeBase64(ctx context.Context, sessionID string, qrCodeBase64 string, ttl time.Duration) error

	// Additional cache methods needed by implementations
	GetSessionByName(ctx context.Context, name string) (*session.Session, error)
	SetSessionByName(ctx context.Context, name string, sess *session.Session, ttl time.Duration) error
	DeleteSessionByName(ctx context.Context, name string) error
	SetDeviceJID(ctx context.Context, sessionID, jid string, ttl time.Duration) error
	DeleteDeviceJID(ctx context.Context, sessionID string) error
	DeleteSessionStatus(ctx context.Context, sessionID string) error
	GetStats(ctx context.Context) (*CacheStats, error)
	Ping(ctx context.Context) error
}

// Cache error helper
func NewCacheError(operation, key string, err error) error {
	return fmt.Errorf("cache %s failed for key '%s': %w", operation, key, err)
}

// Cache stats
type CacheStats struct {
	Hits        int64 `json:"hits"`
	Misses      int64 `json:"misses"`
	Keys        int64 `json:"keys"`
	Memory      int64 `json:"memory"`
	Connections int   `json:"connections"`
}

// Cache key prefixes
const (
	SessionKeyPrefix       = "session:"
	SessionNameKeyPrefix   = "session_name:"
	QRCodeKeyPrefix        = "qr:"
	QRCodeBase64KeyPrefix  = "qr_base64:"
	DeviceJIDKeyPrefix     = "device_jid:"
	SessionStatusKeyPrefix = "session_status:"
)

// Configuration interfaces for application layer
type ConfigProvider interface {
	GetDatabase() DatabaseConfig
	GetServer() ServerConfig
	GetAuth() AuthConfig
	GetWebhook() WebhookConfig
	GetSecurity() SecurityConfig
	GetCache() CacheConfig
}

type DatabaseConfig interface {
	GetURL() string
	GetMaxOpenConns() int
	GetMaxIdleConns() int
	GetConnMaxLifetime() time.Duration
}

type ServerConfig interface {
	GetPort() string
	GetReadTimeout() time.Duration
	GetWriteTimeout() time.Duration
	GetIdleTimeout() time.Duration
}

type AuthConfig interface {
	GetGlobalAPIKey() string
	GetSessionTimeout() time.Duration
	GetTokenExpiration() time.Duration
}

type WebhookConfig interface {
	GetTimeout() time.Duration
	GetMaxRetries() int
	GetInitialBackoff() time.Duration
	GetMaxBackoff() time.Duration
	GetBackoffMultiplier() float64
}

type SecurityConfig interface {
	GetRateLimitEnabled() bool
	GetRateLimitRPS() int
	GetRequestTimeout() time.Duration
	GetMaxRequestSize() int64
}

type CacheConfig interface {
	GetCacheEnabled() bool
	GetSessionTTL() time.Duration
	GetQRCodeTTL() time.Duration
}

// Time provider interface for testability
type TimeProvider interface {
	Now() time.Time
	Unix() int64
}

// Event interfaces - EventPublisher is defined in events.go
type DomainEvent interface {
	EventID() string
	EventType() string
	AggregateID() string
	OccurredAt() time.Time
	EventData() interface{}
}
