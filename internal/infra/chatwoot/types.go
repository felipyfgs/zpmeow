package chatwoot

import (
	"time"
)

// API Request types
type CreateContactRequest struct {
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email,omitempty"`
	Identifier  string `json:"identifier,omitempty"`
	InboxID     int    `json:"inbox_id,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

type CreateConversationRequest struct {
	ContactID int    `json:"contact_id"`
	InboxID   int    `json:"inbox_id"`
	Status    string `json:"status,omitempty"`
}

type SendMessageRequest struct {
	Content           string                 `json:"content"`
	MessageType       int                    `json:"message_type"`
	Private           bool                   `json:"private,omitempty"`
	SourceID          string                 `json:"source_id,omitempty"`
	ContentAttributes map[string]interface{} `json:"content_attributes,omitempty"`
}

// Aliases for compatibility
type MessageCreateRequest = SendMessageRequest

// Configuration types
type ChatwootConfig struct {
	IsActive                bool     `json:"isActive" yaml:"isActive"`
	AccountID               string   `json:"accountId" yaml:"accountId"`
	Token                   string   `json:"token" yaml:"token"`
	URL                     string   `json:"url" yaml:"url"`
	NameInbox               string   `json:"nameInbox" yaml:"nameInbox"`
	WebhookURL              string   `json:"webhookUrl" yaml:"webhookUrl"`
	SignMsg                 bool     `json:"signMsg" yaml:"signMsg"`
	SignDelimiter           string   `json:"signDelimiter" yaml:"signDelimiter"`
	Number                  string   `json:"number" yaml:"number"`
	ReopenConversation      bool     `json:"reopenConversation" yaml:"reopenConversation"`
	ConversationPending     bool     `json:"conversationPending" yaml:"conversationPending"`
	MergeBrazilContacts     bool     `json:"mergeBrazilContacts" yaml:"mergeBrazilContacts"`
	ImportContacts          bool     `json:"importContacts" yaml:"importContacts"`
	ImportMessages          bool     `json:"importMessages" yaml:"importMessages"`
	DaysLimitImportMessages int      `json:"daysLimitImportMessages" yaml:"daysLimitImportMessages"`
	AutoCreate              bool     `json:"autoCreate" yaml:"autoCreate"`
	Organization            string   `json:"organization" yaml:"organization"`
	Logo                    string   `json:"logo" yaml:"logo"`
	IgnoreJids              []string `json:"ignoreJids" yaml:"ignoreJids"`
}

// API DTOs
type APIResponse struct {
	Payload []map[string]interface{} `json:"payload"`
	Meta    map[string]interface{}   `json:"meta,omitempty"`
}

type ErrorResponse struct {
	Message string   `json:"message"`
	Errors  []string `json:"errors,omitempty"`
}

// Webhook payload types
type WebhookPayload struct {
	Event        string                 `json:"event"`
	Account      map[string]interface{} `json:"account,omitempty"`
	Contact      map[string]interface{} `json:"contact,omitempty"`
	Conversation map[string]interface{} `json:"conversation,omitempty"`
	Message      map[string]interface{} `json:",inline"`
}

// Attachment represents file attachment in Chatwoot
type Attachment struct {
	ID       int    `json:"id"`
	FileType string `json:"file_type"`
	DataURL  string `json:"data_url"`
	FileSize int    `json:"file_size"`
	Fallback string `json:"fallback,omitempty"`
}

// API entity types (simplified)
type Contact struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email"`
	Identifier  string `json:"identifier"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type Conversation struct {
	ID             int       `json:"id"`
	ContactID      int       `json:"contact_id"`
	InboxID        int       `json:"inbox_id"`
	Status         string    `json:"status"`
	UnreadCount    int       `json:"unread_count"`
	LastActivityAt time.Time `json:"last_activity_at"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Inbox struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	ChannelType string `json:"channel_type"`
	WebhookURL  string `json:"webhook_url"`
}

// Additional types needed by client
type ContactCreateRequest = CreateContactRequest
type InboxCreateRequest struct {
	Name    string                 `json:"name"`
	Channel map[string]interface{} `json:"channel"`
}

type ConversationCreateRequest = CreateConversationRequest

type Message struct {
	ID             int          `json:"id"`
	Content        string       `json:"content"`
	MessageType    int          `json:"message_type"`
	ConversationID int          `json:"conversation_id"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
	Attachments    []Attachment `json:"attachments,omitempty"`
}

// ConversationStatus represents conversation status
type ConversationStatus string

const (
	ConversationStatusOpen     ConversationStatus = "open"
	ConversationStatusResolved ConversationStatus = "resolved"
	ConversationStatusPending  ConversationStatus = "pending"
)

// WhatsApp integration types
type WhatsAppMessage struct {
	ID              string                 `json:"id"`
	From            string                 `json:"from"`
	To              string                 `json:"to"`
	Body            string                 `json:"body"`
	Type            string                 `json:"type"`
	FromMe          bool                   `json:"fromMe"`
	PushName        string                 `json:"pushName"`
	ChatName        string                 `json:"chatName"`
	Participant     string                 `json:"participant"`
	Timestamp       float64                `json:"timestamp"`
	MediaURL        string                 `json:"mediaUrl,omitempty"`
	FileName        string                 `json:"fileName,omitempty"`
	Caption         string                 `json:"caption,omitempty"`
	MimeType        string                 `json:"mimeType,omitempty"`
	QuotedMessageID string                 `json:"quotedMessageId,omitempty"`
	LinkPreview     *LinkPreviewInfo       `json:"linkPreview,omitempty"`
	Location        *LocationInfo          `json:"location,omitempty"`
	Contacts        []ContactInfo          `json:"contacts,omitempty"`
	List            *ListInfo              `json:"list,omitempty"`
	Buttons         []ButtonInfo           `json:"buttons,omitempty"`
	Reaction        *ReactionInfo          `json:"reaction,omitempty"`
	Extra           map[string]interface{} `json:"extra,omitempty"`
}

// Supporting types for WhatsAppMessage
type LinkPreviewInfo struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Image       string `json:"image,omitempty"`
}

type LocationInfo struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Name      string  `json:"name,omitempty"`
	Address   string  `json:"address,omitempty"`
}

type ContactInfo struct {
	Name        string      `json:"name"`
	DisplayName string      `json:"displayName,omitempty"`
	FirstName   string      `json:"firstName,omitempty"`
	LastName    string      `json:"lastName,omitempty"`
	Phones      []PhoneInfo `json:"phones,omitempty"`
	Emails      []EmailInfo `json:"emails,omitempty"`
}

type PhoneInfo struct {
	Number string `json:"number"`
	Type   string `json:"type,omitempty"`
}

type EmailInfo struct {
	Email string `json:"email"`
	Type  string `json:"type,omitempty"`
}

type ListInfo struct {
	Title       string        `json:"title"`
	Description string        `json:"description"`
	ButtonText  string        `json:"buttonText"`
	Sections    []ListSection `json:"sections"`
}

type ListSection struct {
	Title string     `json:"title"`
	Rows  []ListItem `json:"rows"`
}

type ListItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

type ButtonInfo struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type ReactionInfo struct {
	Text      string `json:"text"`
	MessageID string `json:"messageId"`
	Emoji     string `json:"emoji"`
}

// Content type constants
const (
	ContentTypeText     = "text"
	ContentTypeImage    = "image"
	ContentTypeAudio    = "audio"
	ContentTypeVideo    = "video"
	ContentTypeFile     = "file"
	ContentTypeSticker  = "sticker"
	ContentTypeLocation = "location"
	ContentTypeContact  = "contact"
)

// Additional types needed by other files
type OutgoingMessage struct {
	To       string `json:"to"`
	Content  string `json:"content"`
	Type     string `json:"type"`
	MediaURL string `json:"mediaUrl,omitempty"`
	Caption  string `json:"caption,omitempty"`
	FileName string `json:"fileName,omitempty"`
}

type WhatsAppContact struct {
	Name              string `json:"name"`
	PhoneNumber       string `json:"phoneNumber"`
	Phone             string `json:"phone"`
	JID               string `json:"jid"`
	ProfilePictureURL string `json:"profilePictureUrl,omitempty"`
	IsGroup           bool   `json:"isGroup"`
}
