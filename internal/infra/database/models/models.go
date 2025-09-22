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
	SessionID               string      `db:"session_id" json:"session_id"`
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
