package chatwoot

import (
	"time"
)

// WhatsAppMessage representa uma mensagem do WhatsApp para integração
type WhatsAppMessage struct {
	ID          string                 `json:"id"`
	From        string                 `json:"from"`
	To          string                 `json:"to"`
	Body        string                 `json:"body"`
	Type        string                 `json:"type"`
	FromMe      bool                   `json:"fromMe"`
	PushName    string                 `json:"pushName"`
	ChatName    string                 `json:"chatName"`
	Participant string                 `json:"participant"`
	Timestamp   int64                  `json:"timestamp"`
	MediaURL    string                 `json:"mediaUrl,omitempty"`
	FileName    string                 `json:"fileName,omitempty"`
	Caption     string                 `json:"caption,omitempty"`
	MimeType    string                 `json:"mimeType,omitempty"`
	Location    *LocationInfo          `json:"location,omitempty"`
	Contacts        []ContactInfo          `json:"contacts,omitempty"`
	QuotedMsg       *WhatsAppMessage       `json:"quotedMsg,omitempty"`
	QuotedMessageID string                 `json:"quotedMessageId,omitempty"`
	Reaction        *ReactionInfo          `json:"reaction,omitempty"`
	LinkPreview     *LinkPreviewInfo       `json:"linkPreview,omitempty"`
	List            *ListInfo              `json:"list,omitempty"`
	Buttons         []ButtonInfo           `json:"buttons,omitempty"`
	Extra           map[string]interface{} `json:"extra,omitempty"`
}

// LocationInfo representa informações de localização
type LocationInfo struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Name      string  `json:"name,omitempty"`
	Address   string  `json:"address,omitempty"`
}

// ContactInfo representa informações de contato
type ContactInfo struct {
	Name        string      `json:"name"`
	DisplayName string      `json:"displayName,omitempty"`
	FirstName   string      `json:"firstName,omitempty"`
	LastName    string      `json:"lastName,omitempty"`
	Phones      []PhoneInfo `json:"phones,omitempty"`
	Emails      []EmailInfo `json:"emails,omitempty"`
}

// PhoneInfo representa informações de telefone
type PhoneInfo struct {
	Number string `json:"number"`
	Type   string `json:"type,omitempty"`
}

// EmailInfo representa informações de email
type EmailInfo struct {
	Email string `json:"email"`
	Type  string `json:"type,omitempty"`
}

// ReactionInfo representa informações de reação
type ReactionInfo struct {
	Text      string `json:"text"`
	MessageID string `json:"messageId"`
	Emoji     string `json:"emoji"`
}

// LinkPreviewInfo representa informações de preview de link
type LinkPreviewInfo struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Image       string `json:"image,omitempty"`
}

// ListInfo representa informações de lista
type ListInfo struct {
	Title       string        `json:"title"`
	Description string        `json:"description"`
	ButtonText  string        `json:"buttonText"`
	Sections    []ListSection `json:"sections"`
}

// ListSection representa uma seção da lista
type ListSection struct {
	Title string     `json:"title"`
	Rows  []ListItem `json:"rows"`
}

// ListItem representa um item da lista
type ListItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

// ButtonInfo representa informações de botão
type ButtonInfo struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// OutgoingMessage representa uma mensagem de saída para WhatsApp
type OutgoingMessage struct {
	To       string `json:"to"`
	Content  string `json:"content"`
	Type     string `json:"type"`
	MediaURL string `json:"mediaUrl,omitempty"`
	Caption  string `json:"caption,omitempty"`
	FileName string `json:"fileName,omitempty"`
}

// WhatsAppContact representa um contato do WhatsApp
type WhatsAppContact struct {
	Name               string `json:"name"`
	PhoneNumber        string `json:"phoneNumber"`
	Phone              string `json:"phone"` // Alias para PhoneNumber
	JID                string `json:"jid"`
	ProfilePictureURL  string `json:"profilePictureUrl,omitempty"`
	IsGroup            bool   `json:"isGroup"`
}

// ChatwootConfig representa a configuração da integração Chatwoot
type ChatwootConfig struct {
	Enabled                     bool     `json:"enabled" yaml:"enabled"`
	AccountID                   string   `json:"accountId" yaml:"accountId"`
	Token                       string   `json:"token" yaml:"token"`
	URL                         string   `json:"url" yaml:"url"`
	NameInbox                   string   `json:"nameInbox" yaml:"nameInbox"`
	WebhookURL                  string   `json:"webhookUrl" yaml:"webhookUrl"`
	SignMsg                     bool     `json:"signMsg" yaml:"signMsg"`
	SignDelimiter               string   `json:"signDelimiter" yaml:"signDelimiter"`
	Number                      string   `json:"number" yaml:"number"`
	ReopenConversation          bool     `json:"reopenConversation" yaml:"reopenConversation"`
	ConversationPending         bool     `json:"conversationPending" yaml:"conversationPending"`
	MergeBrazilContacts         bool     `json:"mergeBrazilContacts" yaml:"mergeBrazilContacts"`
	ImportContacts              bool     `json:"importContacts" yaml:"importContacts"`
	ImportMessages              bool     `json:"importMessages" yaml:"importMessages"`
	DaysLimitImportMessages     int      `json:"daysLimitImportMessages" yaml:"daysLimitImportMessages"`
	AutoCreate                  bool     `json:"autoCreate" yaml:"autoCreate"`
	Organization                string   `json:"organization" yaml:"organization"`
	Logo                        string   `json:"logo" yaml:"logo"`
	IgnoreJids                  []string `json:"ignoreJids" yaml:"ignoreJids"`
}

// Contact representa um contato no Chatwoot
type Contact struct {
	ID           int                    `json:"id"`
	Name         string                 `json:"name"`
	Avatar       string                 `json:"avatar"`
	AvatarURL    string                 `json:"avatar_url"`
	PhoneNumber  string                 `json:"phone_number"`
	Email        string                 `json:"email"`
	Identifier   string                 `json:"identifier"`
	Thumbnail    string                 `json:"thumbnail"`
	CustomAttributes map[string]interface{} `json:"custom_attributes"`
	CreatedAt    int64                  `json:"created_at"`
	UpdatedAt    int64                  `json:"updated_at"`
}

// ContactCreateRequest representa a requisição para criar um contato
type ContactCreateRequest struct {
	InboxID     int    `json:"inbox_id"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number,omitempty"`
	Email       string `json:"email,omitempty"`
	Identifier  string `json:"identifier,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

// Inbox representa uma caixa de entrada no Chatwoot
type Inbox struct {
	ID                int                    `json:"id"`
	Name              string                 `json:"name"`
	ChannelID         int                    `json:"channel_id"`
	ChannelType       string                 `json:"channel_type"`
	GreetingEnabled   bool                   `json:"greeting_enabled"`
	GreetingMessage   string                 `json:"greeting_message"`
	WorkingHoursEnabled bool                 `json:"working_hours_enabled"`
	EnableEmailCollect bool                  `json:"enable_email_collect"`
	CsatSurveyEnabled bool                   `json:"csat_survey_enabled"`
	EnableAutoAssignment bool                `json:"enable_auto_assignment"`
	WebsiteURL        string                 `json:"website_url"`
	WelcomeTitle      string                 `json:"welcome_title"`
	WelcomeTagline    string                 `json:"welcome_tagline"`
	WebsiteToken      string                 `json:"website_token"`
	ForwardToEmail    string                 `json:"forward_to_email"`
	PhoneNumber       string                 `json:"phone_number"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

// InboxCreateRequest representa a requisição para criar uma inbox
type InboxCreateRequest struct {
	Name    string                 `json:"name"`
	Channel map[string]interface{} `json:"channel"`
}

// Conversation representa uma conversa no Chatwoot
type Conversation struct {
	ID                    int                    `json:"id"`
	Messages              []Message              `json:"messages"`
	AccountID             int                    `json:"account_id"`
	InboxID               int                    `json:"inbox_id"`
	Status                string                 `json:"status"`
	Timestamp             int64                  `json:"timestamp"`
	ContactLastSeenAt     int64                  `json:"contact_last_seen_at"`
	AgentLastSeenAt       int64                  `json:"agent_last_seen_at"`
	UnreadCount           int                    `json:"unread_count"`
	AdditionalAttributes  map[string]interface{} `json:"additional_attributes"`
	CustomAttributes      map[string]interface{} `json:"custom_attributes"`
	Contact               Contact                `json:"contact"`
	Assignee              *Agent                 `json:"assignee"`
	Team                  *Team                  `json:"team"`
	Meta                  ConversationMeta       `json:"meta"`
	CreatedAt             int64                  `json:"created_at"`
	UpdatedAt             int64                  `json:"updated_at"`
}

// ConversationCreateRequest representa a requisição para criar uma conversa
type ConversationCreateRequest struct {
	ContactID string `json:"contact_id"`
	InboxID   string `json:"inbox_id"`
	Status    string `json:"status,omitempty"`
}

// ConversationMeta representa metadados da conversa
type ConversationMeta struct {
	Sender    Contact `json:"sender"`
	Channel   string  `json:"channel"`
	Assignee  *Agent  `json:"assignee"`
	Team      *Team   `json:"team"`
	HmacVerified bool `json:"hmac_verified"`
}

// Message representa uma mensagem no Chatwoot
type Message struct {
	ID                  int                    `json:"id"`
	Content             string                 `json:"content"`
	MessageType         int                    `json:"message_type"` // 0 = incoming, 1 = outgoing
	ContentType         string                 `json:"content_type"`
	ContentAttributes   map[string]interface{} `json:"content_attributes"`
	CreatedAt           int64                  `json:"created_at"`
	Private             bool                   `json:"private"`
	SourceID            string                 `json:"source_id"`
	Sender              *Contact               `json:"sender"`
	ConversationID      int                    `json:"conversation_id"`
	InboxID             int                    `json:"inbox_id"`
	Attachments         []Attachment           `json:"attachments"`
	ExternalSourceIds   map[string]string      `json:"external_source_ids"`
}

// MessageCreateRequest representa a requisição para criar uma mensagem
type MessageCreateRequest struct {
	Content           string                 `json:"content"`
	MessageType       int                    `json:"message_type"` // 0 = incoming, 1 = outgoing
	Private           bool                   `json:"private,omitempty"`
	ContentType       string                 `json:"content_type,omitempty"`
	ContentAttributes map[string]interface{} `json:"content_attributes,omitempty"`
	SourceID          string                 `json:"source_id,omitempty"`
	SourceReplyID     string                 `json:"source_reply_id,omitempty"`
}

// Attachment representa um anexo de mensagem
type Attachment struct {
	ID           int    `json:"id"`
	MessageID    int    `json:"message_id"`
	FileType     string `json:"file_type"`
	AccountID    int    `json:"account_id"`
	Extension    string `json:"extension"`
	DataURL      string `json:"data_url"`
	ThumbURL     string `json:"thumb_url"`
	FileSize     int64  `json:"file_size"`
	Fallback     string `json:"fallback"`
}

// Agent representa um agente no Chatwoot
type Agent struct {
	ID                int                    `json:"id"`
	UID               string                 `json:"uid"`
	Name              string                 `json:"name"`
	DisplayName       string                 `json:"display_name"`
	Email             string                 `json:"email"`
	AccountID         int                    `json:"account_id"`
	Role              string                 `json:"role"`
	ConfirmedAt       time.Time              `json:"confirmed_at"`
	CustomAttributes  map[string]interface{} `json:"custom_attributes"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

// Team representa uma equipe no Chatwoot
type Team struct {
	ID                int       `json:"id"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	AllowAutoAssign   bool      `json:"allow_auto_assign"`
	AccountID         int       `json:"account_id"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// WebhookPayload representa o payload de um webhook do Chatwoot
type WebhookPayload struct {
	Event       string                 `json:"event"`
	MessageType string                 `json:"message_type"` // "incoming" ou "outgoing"
	ID          int                    `json:"id"`
	Content     string                 `json:"content"`
	CreatedAt   time.Time              `json:"created_at"`
	Private     bool                   `json:"private"`
	SourceID    string                 `json:"source_id"`
	ContentType string                 `json:"content_type"`
	ContentAttributes map[string]interface{} `json:"content_attributes"`
	Sender      *Contact               `json:"sender"`
	Contact     *Contact               `json:"contact"`
	Conversation *Conversation         `json:"conversation"`
	Account     *Account               `json:"account"`
	Inbox       *Inbox                 `json:"inbox"`
	Attachments []Attachment           `json:"attachments"`
}

// Account representa uma conta no Chatwoot
type Account struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// APIResponse representa uma resposta genérica da API do Chatwoot
type APIResponse struct {
	Payload interface{} `json:"payload"`
	Meta    interface{} `json:"meta"`
}

// ErrorResponse representa uma resposta de erro da API do Chatwoot
type ErrorResponse struct {
	Message string                 `json:"message"`
	Errors  map[string]interface{} `json:"errors"`
}

// ContactInbox representa a relação entre contato e inbox
type ContactInbox struct {
	SourceID string `json:"source_id"`
	InboxID  int    `json:"inbox_id"`
	ContactID int   `json:"contact_id"`
}

// ConversationStatus representa os possíveis status de uma conversa
type ConversationStatus string

const (
	ConversationStatusOpen     ConversationStatus = "open"
	ConversationStatusResolved ConversationStatus = "resolved"
	ConversationStatusPending  ConversationStatus = "pending"
)

// MessageType representa os tipos de mensagem
type MessageType string

const (
	MessageTypeIncoming MessageType = "incoming"
	MessageTypeOutgoing MessageType = "outgoing"
	MessageTypeActivity MessageType = "activity"
	MessageTypeTemplate MessageType = "template"
)

// ContentType representa os tipos de conteúdo
type ContentType string

const (
	ContentTypeText        ContentType = "text"
	ContentTypeInputEmail  ContentType = "input_email"
	ContentTypeCards       ContentType = "cards"
	ContentTypeInputSelect ContentType = "input_select"
	ContentTypeForm        ContentType = "form"
	ContentTypeArticle     ContentType = "article"
	ContentTypeImage       ContentType = "image"
	ContentTypeAudio       ContentType = "audio"
	ContentTypeVideo       ContentType = "video"
	ContentTypeFile        ContentType = "file"
	ContentTypeLocation    ContentType = "location"
	ContentTypeContact     ContentType = "contact"
	ContentTypeSticker     ContentType = "sticker"
)



// MessageCreateRequestWithAttachment representa uma requisição de criação de mensagem com anexo
type MessageCreateRequestWithAttachment struct {
	Content           string                 `json:"content"`
	MessageType       int                    `json:"message_type"`
	SourceID          string                 `json:"source_id"`
	ContentType       string                 `json:"content_type"`
	ContentAttributes map[string]interface{} `json:"content_attributes,omitempty"`
	Attachments       []AttachmentData       `json:"attachments,omitempty"`
}

// AttachmentData representa dados de anexo para upload
type AttachmentData struct {
	FileType string `json:"file_type"`
	DataURL  string `json:"data_url"`
	FileName string `json:"file_name,omitempty"`
}
