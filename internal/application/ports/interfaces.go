package ports

import (
	"context"
	"time"

	"go.mau.fi/whatsmeow"
)

// ============================================================================
// SESSION MANAGEMENT
// ============================================================================

// SessionManager define operações de gerenciamento de sessões WhatsApp
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

// ============================================================================
// MESSAGE OPERATIONS
// ============================================================================

// ButtonData representa dados de um botão para mensagens interativas
type ButtonData struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// ListItem representa um item de lista para mensagens de lista
type ListItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// ListSection representa uma seção de lista para mensagens de lista
type ListSection struct {
	Title string     `json:"title"`
	Rows  []ListItem `json:"rows"`
}

// MediaMessage representa uma mensagem de mídia
type MediaMessage struct {
	Type     string `json:"type"`
	Data     []byte `json:"data"`
	MimeType string `json:"mime_type"`
	Caption  string `json:"caption,omitempty"`
	Filename string `json:"filename,omitempty"`
}

// ContactData representa dados de contato para envio
type ContactData struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

// MessageSender define operações de envio de mensagens WhatsApp
type MessageSender interface {
	SendTextMessage(ctx context.Context, sessionID, phone, text string) (*whatsmeow.SendResponse, error)
	SendMediaMessage(ctx context.Context, sessionID, phone string, media MediaMessage) (*whatsmeow.SendResponse, error)
	SendImageMessage(ctx context.Context, sessionID, phone string, data []byte, caption, mimeType string) (*whatsmeow.SendResponse, error)
	SendAudioMessage(ctx context.Context, sessionID, phone string, data []byte, mimeType string) (*whatsmeow.SendResponse, error)
	SendVideoMessage(ctx context.Context, sessionID, phone string, data []byte, caption, mimeType string) (*whatsmeow.SendResponse, error)
	SendDocumentMessage(ctx context.Context, sessionID, phone string, data []byte, filename, caption, mimeType string) (*whatsmeow.SendResponse, error)
	SendStickerMessage(ctx context.Context, sessionID, phone string, data []byte, mimeType string) (*whatsmeow.SendResponse, error)

	SendContactsMessage(ctx context.Context, sessionID, phone string, contacts []ContactData) (*whatsmeow.SendResponse, error)
	SendLocationMessage(ctx context.Context, sessionID, phone string, latitude, longitude float64, name, address string) (*whatsmeow.SendResponse, error)
	SendButtonMessage(ctx context.Context, sessionID, phone, title string, buttons []ButtonData) (*whatsmeow.SendResponse, error)
	SendListMessage(ctx context.Context, sessionID, phone, title, description, buttonText, footerText string, sections []ListSection) (*whatsmeow.SendResponse, error)
	SendPollMessage(ctx context.Context, sessionID, phone, name string, options []string, selectableCount int) (*whatsmeow.SendResponse, error)
}

// MessageActions define operações de ação em mensagens WhatsApp
type MessageActions interface {
	MarkAsRead(ctx context.Context, sessionID, phone string, messageIDs []string) error
	SetPresence(ctx context.Context, sessionID, phone, state, media string) error
	ReactToMessage(ctx context.Context, sessionID, phone, messageID, emoji string) error
	DeleteMessage(ctx context.Context, sessionID, phone, messageID string, forEveryone bool) error
	EditMessage(ctx context.Context, sessionID, phone, messageID, newText string) (*whatsmeow.SendResponse, error)
	DownloadMedia(ctx context.Context, sessionID, messageID string) ([]byte, string, error)
}

// ============================================================================
// GROUP MANAGEMENT
// ============================================================================

// GroupInfo representa informações de um grupo WhatsApp
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

// GroupManager define operações de gerenciamento de grupos WhatsApp
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

// ============================================================================
// CONTACT MANAGEMENT
// ============================================================================

// UserCheckResult representa o resultado da verificação de um usuário
type UserCheckResult struct {
	Query        string `json:"query"`
	IsInWhatsapp bool   `json:"is_in_whatsapp"`
	IsInMeow     bool   `json:"is_in_meow"`
	JID          string `json:"jid"`
	VerifiedName string `json:"verified_name,omitempty"`
}

// UserInfoResult representa informações detalhadas de um usuário
type UserInfoResult struct {
	JID          string `json:"jid"`
	Name         string `json:"name,omitempty"`
	Notify       string `json:"notify,omitempty"`
	PushName     string `json:"push_name,omitempty"`
	BusinessName string `json:"business_name,omitempty"`
	IsBlocked    bool   `json:"is_blocked"`
	IsMuted      bool   `json:"is_muted"`
}

// AvatarResult representa informações do avatar de um usuário
type AvatarResult struct {
	Phone     string `json:"phone"`
	JID       string `json:"jid"`
	AvatarURL string `json:"avatar_url,omitempty"`
	PictureID string `json:"picture_id,omitempty"`
}

// ContactResult representa informações de um contato
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

// ContactManager define operações de gerenciamento de contatos WhatsApp
type ContactManager interface {
	CheckUser(ctx context.Context, sessionID string, phones []string) ([]UserCheckResult, error)
	CheckContact(ctx context.Context, sessionID, phone string) (*UserCheckResult, error)
	GetUserInfo(ctx context.Context, sessionID string, phones []string) (map[string]UserInfoResult, error)
	GetAvatar(ctx context.Context, sessionID, phone string) (*AvatarResult, error)
	GetContacts(ctx context.Context, sessionID string, limit, offset int) ([]ContactResult, error)
	SetUserPresence(ctx context.Context, sessionID, state string) error
}

// ============================================================================
// CHAT MANAGEMENT
// ============================================================================

// ChatInfo representa informações de um chat
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

// ChatMessage representa uma mensagem de chat
type ChatMessage struct {
	ID        string    `json:"id"`
	ChatJID   string    `json:"chat_jid"`
	FromJID   string    `json:"from_jid"`
	Text      string    `json:"text"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
}

// ChatManager define operações de gerenciamento de chats WhatsApp
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

// ============================================================================
// NEWSLETTER MANAGEMENT
// ============================================================================

// NewsletterInfo representa informações de um newsletter
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

// NewsletterMessage representa uma mensagem de newsletter
type NewsletterMessage struct {
	ID           string         `json:"id"`
	NewsletterID string         `json:"newsletter_id"`
	Content      string         `json:"content"`
	Timestamp    int64          `json:"timestamp"`
	Views        int            `json:"views"`
	Reactions    map[string]int `json:"reactions,omitempty"`
}

// NewsletterManager define operações de gerenciamento de newsletters WhatsApp
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

// ============================================================================
// PRIVACY MANAGEMENT
// ============================================================================

// PrivacySettings representa configurações de privacidade do WhatsApp
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

// PrivacyManager define operações de gerenciamento de privacidade WhatsApp
type PrivacyManager interface {
	GetPrivacySettings(ctx context.Context, sessionID string) (*PrivacySettings, error)
	SetPrivacySetting(ctx context.Context, sessionID, setting, value string) error
	GetBlocklist(ctx context.Context, sessionID string) ([]string, error)
	UpdateBlocklist(ctx context.Context, sessionID string, action string, contacts []string) error
}

// ============================================================================
// WEBHOOK MANAGEMENT
// ============================================================================

// WebhookManager define operações de gerenciamento de webhooks
type WebhookManager interface {
	UpdateSessionWebhook(sessionID, webhookURL string) error
	UpdateSessionSubscriptions(sessionID string, events []string) error
}

// ============================================================================
// COMBINED INTERFACE
// ============================================================================

// WameowService combina todas as interfaces segregadas
// Mantém compatibilidade com código existente enquanto permite uso das interfaces específicas
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
}

// WhatsAppService é um alias para WameowService para compatibilidade
type WhatsAppService = WameowService

// ============================================================================
// LOGGING
// ============================================================================

// Logger define operações de logging estruturado
type Logger interface {
	Debug(ctx context.Context, msg string, keysAndValues ...interface{})
	Info(ctx context.Context, msg string, keysAndValues ...interface{})
	Warn(ctx context.Context, msg string, keysAndValues ...interface{})
	Error(ctx context.Context, msg string, keysAndValues ...interface{})
	Fatal(ctx context.Context, msg string, keysAndValues ...interface{})
}

// ============================================================================
// UTILITIES
// ============================================================================

// IDGenerator define operações de geração de IDs
type IDGenerator interface {
	Generate() string
	GenerateAPIKey() string
}

// ContactInfo representa informações de contato
type ContactInfo struct {
	Name         string `json:"name"`
	Phone        string `json:"phone"`
	Email        string `json:"email,omitempty"`
	Organization string `json:"organization,omitempty"`
}

// NotificationService define operações de notificação
type NotificationService interface {
	SendNotification(ctx context.Context, message string) error
	SendWebhook(ctx context.Context, url string, payload interface{}) error
}
