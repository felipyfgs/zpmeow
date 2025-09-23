package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type SessionModel struct {
	ID         string    `db:"id" json:"id"`
	Name       string    `db:"name" json:"name"`
	DeviceJID  string    `db:"device_jid" json:"device_jid"`
	Status     string    `db:"status" json:"status"`
	QRCode     string    `db:"qr_code" json:"qr_code"`
	ProxyURL   string    `db:"proxy_url" json:"proxy_url"`
	WebhookURL string    `db:"webhook_url" json:"webhook_url"`
	Events     string    `db:"webhook_events" json:"webhook_events"`
	Connected  bool      `db:"connected" json:"connected"`
	ApiKey     string    `db:"apikey" json:"apikey"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

func (SessionModel) TableName() string {
	return "sessions"
}

// StringArray representa um array de strings para PostgreSQL JSONB
type StringArray []string

// Scan implementa o driver.Valuer interface para StringArray
func (sa *StringArray) Scan(value interface{}) error {
	if value == nil {
		*sa = StringArray{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, sa)
	case string:
		return json.Unmarshal([]byte(v), sa)
	default:
		return fmt.Errorf("cannot scan %T into StringArray", value)
	}
}

// Value implementa o driver.Valuer interface para StringArray
func (sa StringArray) Value() (driver.Value, error) {
	if len(sa) == 0 {
		return "[]", nil
	}
	return json.Marshal(sa)
}

// ChatwootModel representa a configuração Chatwoot no banco de dados
type ChatwootModel struct {
	ID                      string      `db:"id" json:"id"`
	SessionID               string      `db:"session_id" json:"session_id"` // UUID da sessão
	Enabled                 bool        `db:"enabled" json:"enabled"`
	AccountID               *string     `db:"account_id" json:"account_id"`
	Token                   *string     `db:"token" json:"token"`
	URL                     *string     `db:"url" json:"url"`
	NameInbox               *string     `db:"name_inbox" json:"name_inbox"`
	SignMsg                 bool        `db:"sign_msg" json:"sign_msg"`
	SignDelimiter           string      `db:"sign_delimiter" json:"sign_delimiter"`
	Number                  string      `db:"number" json:"number"`
	ReopenConversation      bool        `db:"reopen_conversation" json:"reopen_conversation"`
	ConversationPending     bool        `db:"conversation_pending" json:"conversation_pending"`
	MergeBrazilContacts     bool        `db:"merge_brazil_contacts" json:"merge_brazil_contacts"`
	ImportContacts          bool        `db:"import_contacts" json:"import_contacts"`
	ImportMessages          bool        `db:"import_messages" json:"import_messages"`
	DaysLimitImportMessages int         `db:"days_limit_import_messages" json:"days_limit_import_messages"`
	AutoCreate              bool        `db:"auto_create" json:"auto_create"`
	Organization            string      `db:"organization" json:"organization"`
	Logo                    string      `db:"logo" json:"logo"`
	IgnoreJids              StringArray `db:"ignore_jids" json:"ignore_jids"`
	InboxID                 *int        `db:"inbox_id" json:"inbox_id"`
	InboxName               *string     `db:"inbox_name" json:"inbox_name"`
	LastSync                *time.Time  `db:"last_sync" json:"last_sync"`
	SyncStatus              string      `db:"sync_status" json:"sync_status"`
	ErrorMessage            *string     `db:"error_message" json:"error_message"`
	MessagesCount           int         `db:"messages_count" json:"messages_count"`
	ContactsCount           int         `db:"contacts_count" json:"contacts_count"`
	ConversationsCount      int         `db:"conversations_count" json:"conversations_count"`
	CreatedAt               time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt               time.Time   `db:"updated_at" json:"updated_at"`
}

func (ChatwootModel) TableName() string {
	return "chatwoot"
}

// JSONB representa um campo JSONB do PostgreSQL
type JSONB map[string]interface{}

// Scan implementa o driver.Valuer interface para JSONB
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = JSONB{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, j)
	case string:
		return json.Unmarshal([]byte(v), j)
	default:
		return fmt.Errorf("cannot scan %T into JSONB", value)
	}
}

// Value implementa o driver.Valuer interface para JSONB
func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		return "{}", nil
	}
	return json.Marshal(j)
}

// ChatModel representa um chat/conversa no banco de dados
type ChatModel struct {
	ID                     string     `db:"id" json:"id"`
	SessionID              string     `db:"session_id" json:"session_id"`
	ChatJID                string     `db:"chat_jid" json:"chat_jid"`
	ChatName               *string    `db:"chat_name" json:"chat_name"`
	ChatType               string     `db:"chat_type" json:"chat_type"`
	PhoneNumber            *string    `db:"phone_number" json:"phone_number"`
	IsGroup                bool       `db:"is_group" json:"is_group"`
	GroupSubject           *string    `db:"group_subject" json:"group_subject"`
	GroupDescription       *string    `db:"group_description" json:"group_description"`
	ChatwootConversationID *int64     `db:"chatwoot_conversation_id" json:"chatwoot_conversation_id"`
	ChatwootContactID      *int64     `db:"chatwoot_contact_id" json:"chatwoot_contact_id"`
	LastMessageAt          *time.Time `db:"last_message_at" json:"last_message_at"`
	UnreadCount            int        `db:"unread_count" json:"unread_count"`
	IsArchived             bool       `db:"is_archived" json:"is_archived"`
	Metadata               JSONB      `db:"metadata" json:"metadata"`
	CreatedAt              time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt              time.Time  `db:"updated_at" json:"updated_at"`
}

func (ChatModel) TableName() string {
	return "chats"
}

// MessageModel representa uma mensagem no banco de dados
type MessageModel struct {
	ID                 string     `db:"id" json:"id"`
	ChatID             string     `db:"chat_id" json:"chat_id"`
	SessionID          string     `db:"session_id" json:"session_id"`
	WhatsAppMessageID  string     `db:"whatsapp_message_id" json:"whatsapp_message_id"`
	MessageType        string     `db:"message_type" json:"message_type"`
	Content            *string    `db:"content" json:"content"`
	MediaURL           *string    `db:"media_url" json:"media_url"`
	MediaMimeType      *string    `db:"media_mime_type" json:"media_mime_type"`
	MediaSize          *int64     `db:"media_size" json:"media_size"`
	MediaFilename      *string    `db:"media_filename" json:"media_filename"`
	ThumbnailURL       *string    `db:"thumbnail_url" json:"thumbnail_url"`
	SenderJID          string     `db:"sender_jid" json:"sender_jid"`
	SenderName         *string    `db:"sender_name" json:"sender_name"`
	IsFromMe           bool       `db:"is_from_me" json:"is_from_me"`
	IsForwarded        bool       `db:"is_forwarded" json:"is_forwarded"`
	IsBroadcast        bool       `db:"is_broadcast" json:"is_broadcast"`
	QuotedMessageID    *string    `db:"quoted_message_id" json:"quoted_message_id"`
	QuotedContent      *string    `db:"quoted_content" json:"quoted_content"`
	Status             string     `db:"status" json:"status"`
	Timestamp          time.Time  `db:"timestamp" json:"timestamp"`
	EditTimestamp      *time.Time `db:"edit_timestamp" json:"edit_timestamp"`
	IsDeleted          bool       `db:"is_deleted" json:"is_deleted"`
	DeletedAt          *time.Time `db:"deleted_at" json:"deleted_at"`
	Reaction           *string    `db:"reaction" json:"reaction"`
	Metadata           JSONB      `db:"metadata" json:"metadata"`
	CreatedAt          time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at" json:"updated_at"`
}

func (MessageModel) TableName() string {
	return "messages"
}

// ZpCwMessageModel representa a relação entre mensagens zpmeow e Chatwoot
type ZpCwMessageModel struct {
	ID                      string     `db:"id" json:"id"`
	SessionID               string     `db:"session_id" json:"session_id"`
	ZpmeowMessageID         string     `db:"zpmeow_message_id" json:"zpmeow_message_id"`
	ChatwootMessageID       int64      `db:"chatwoot_message_id" json:"chatwoot_message_id"`
	ChatwootConversationID  int64      `db:"chatwoot_conversation_id" json:"chatwoot_conversation_id"`
	ChatwootAccountID       int64      `db:"chatwoot_account_id" json:"chatwoot_account_id"`
	Direction               string     `db:"direction" json:"direction"`
	SyncStatus              string     `db:"sync_status" json:"sync_status"`
	SyncError               *string    `db:"sync_error" json:"sync_error"`
	LastSyncAt              *time.Time `db:"last_sync_at" json:"last_sync_at"`
	ChatwootSourceID        *string    `db:"chatwoot_source_id" json:"chatwoot_source_id"`
	ChatwootEchoID          *string    `db:"chatwoot_echo_id" json:"chatwoot_echo_id"`
	Metadata                JSONB      `db:"metadata" json:"metadata"`
	CreatedAt               time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt               time.Time  `db:"updated_at" json:"updated_at"`
}

func (ZpCwMessageModel) TableName() string {
	return "zp_cw_messages"
}
