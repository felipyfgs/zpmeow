package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type SessionModel struct {
	ID            string    `db:"id" json:"id"`
	Name          string    `db:"name" json:"name"`
	DeviceJid     string    `db:"deviceJid" json:"deviceJid"` // camelCase exato com aspas duplas
	Status        string    `db:"status" json:"status"`
	QrCode        string    `db:"qrCode" json:"qrCode"`               // camelCase exato com aspas duplas
	ProxyUrl      string    `db:"proxyUrl" json:"proxyUrl"`           // camelCase exato com aspas duplas
	WebhookUrl    string    `db:"webhookUrl" json:"webhookUrl"`       // camelCase exato com aspas duplas
	WebhookEvents string    `db:"webhookEvents" json:"webhookEvents"` // camelCase exato com aspas duplas
	Connected     bool      `db:"connected" json:"connected"`         // Campo necessário para o repositório
	ApiKey        string    `db:"apiKey" json:"apiKey"`               // camelCase exato com aspas duplas
	CreatedAt     time.Time `db:"createdAt" json:"createdAt"`         // camelCase exato com aspas duplas
	UpdatedAt     time.Time `db:"updatedAt" json:"updatedAt"`         // camelCase exato com aspas duplas
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

// ChatwootModel representa a configuração Chatwoot no banco de dados (OTIMIZADA)
type ChatwootModel struct {
	ID         string     `db:"id" json:"id"`
	SessionId  string     `db:"sessionId" json:"sessionId"` // camelCase exato com aspas duplas
	Enabled    bool       `db:"enabled" json:"enabled"`
	AccountId  *string    `db:"accountId" json:"accountId"` // camelCase exato com aspas duplas
	Token      *string    `db:"token" json:"token"`
	URL        *string    `db:"url" json:"url"`
	NameInbox  *string    `db:"nameInbox" json:"nameInbox"` // camelCase exato com aspas duplas
	Number     string     `db:"number" json:"number"`
	InboxId    *int       `db:"inboxId" json:"inboxId"`       // camelCase exato com aspas duplas
	Config     JSONB      `db:"config" json:"config"`         // configurações específicas agrupadas
	SyncStatus string     `db:"syncStatus" json:"syncStatus"` // camelCase exato com aspas duplas
	LastSync   *time.Time `db:"lastSync" json:"lastSync"`     // camelCase exato com aspas duplas
	CreatedAt  time.Time  `db:"createdAt" json:"createdAt"`   // camelCase exato com aspas duplas
	UpdatedAt  time.Time  `db:"updatedAt" json:"updatedAt"`   // camelCase exato com aspas duplas
	// Contadores removidos - calcular dinamicamente se necessário
	// Configurações específicas movidas para Config JSONB
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

// ChatModel representa um chat/conversa no banco de dados (OTIMIZADA)
type ChatModel struct {
	ID          string     `db:"id" json:"id"`
	SessionId   string     `db:"sessionId" json:"sessionId"`     // camelCase exato com aspas duplas
	ChatJid     string     `db:"chatJid" json:"chatJid"`         // camelCase exato com aspas duplas
	ChatName    *string    `db:"chatName" json:"chatName"`       // camelCase exato com aspas duplas
	PhoneNumber *string    `db:"phoneNumber" json:"phoneNumber"` // camelCase exato com aspas duplas
	IsGroup     bool       `db:"isGroup" json:"isGroup"`         // camelCase exato com aspas duplas
	LastMsgAt   *time.Time `db:"lastMsgAt" json:"lastMsgAt"`     // camelCase exato com aspas duplas
	UnreadCount int        `db:"unreadCount" json:"unreadCount"` // camelCase exato com aspas duplas
	IsArchived  bool       `db:"isArchived" json:"isArchived"`   // camelCase exato com aspas duplas
	Metadata    JSONB      `db:"metadata" json:"metadata"`       // groupSubject, groupDescription movidos aqui
	CreatedAt   time.Time  `db:"createdAt" json:"createdAt"`     // camelCase exato com aspas duplas
	UpdatedAt   time.Time  `db:"updatedAt" json:"updatedAt"`     // camelCase exato com aspas duplas
	// ChatType removido - redundante com IsGroup
	// Campos Chatwoot removidos - usar relação separada
	// GroupSubject, GroupDescription movidos para Metadata
}

func (ChatModel) TableName() string {
	return "chats"
}

// MessageModel representa uma mensagem no banco de dados (OTIMIZADA)
type MessageModel struct {
	ID            string     `db:"id" json:"id"`
	SessionId     string     `db:"sessionId" json:"sessionId"` // camelCase exato com aspas duplas, MANTIDO - essencial!
	ChatId        string     `db:"chatId" json:"chatId"`       // camelCase exato com aspas duplas
	MsgId         string     `db:"msgId" json:"msgId"`         // WhatsApp ID (nome encurtado)
	MsgType       string     `db:"msgType" json:"msgType"`     // camelCase exato com aspas duplas
	Content       *string    `db:"content" json:"content"`
	MediaInfo     JSONB      `db:"mediaInfo" json:"mediaInfo"`         // camelCase exato com aspas duplas
	SenderJid     string     `db:"senderJid" json:"senderJid"`         // camelCase exato com aspas duplas
	SenderName    *string    `db:"senderName" json:"senderName"`       // camelCase exato com aspas duplas
	IsFromMe      bool       `db:"isFromMe" json:"isFromMe"`           // camelCase exato com aspas duplas
	IsForwarded   bool       `db:"isForwarded" json:"isForwarded"`     // camelCase exato com aspas duplas
	IsBroadcast   bool       `db:"isBroadcast" json:"isBroadcast"`     // camelCase exato com aspas duplas
	QuotedMsgId   *string    `db:"quotedMsgId" json:"quotedMsgId"`     // camelCase exato com aspas duplas
	QuotedContent *string    `db:"quotedContent" json:"quotedContent"` // camelCase exato com aspas duplas
	Status        string     `db:"status" json:"status"`
	Timestamp     time.Time  `db:"timestamp" json:"timestamp"`
	EditTimestamp *time.Time `db:"editTimestamp" json:"editTimestamp"` // camelCase exato com aspas duplas
	IsDeleted     bool       `db:"isDeleted" json:"isDeleted"`         // camelCase exato com aspas duplas
	DeletedAt     *time.Time `db:"deletedAt" json:"deletedAt"`         // camelCase exato com aspas duplas
	Reaction      *string    `db:"reaction" json:"reaction"`           // emoji reaction
	Metadata      JSONB      `db:"metadata" json:"metadata"`           // outros metadados
	CreatedAt     time.Time  `db:"createdAt" json:"createdAt"`         // camelCase exato com aspas duplas
	UpdatedAt     time.Time  `db:"updatedAt" json:"updatedAt"`         // camelCase exato com aspas duplas
	// Campos de mídia agrupados em MediaInfo JSONB
	// EditTimestamp, Reaction movidos para Metadata
}

func (MessageModel) TableName() string {
	return "messages"
}

// ZpCwMessageModel representa a relação entre mensagens zpmeow e Chatwoot (OTIMIZADA)
type ZpCwMessageModel struct {
	ID             string    `db:"id" json:"id"`
	SessionId      string    `db:"sessionId" json:"sessionId"`           // camelCase exato com aspas duplas, MANTIDO - essencial!
	MsgId          string    `db:"msgId" json:"msgId"`                   // camelCase exato com aspas duplas
	ChatwootMsgId  int64     `db:"chatwootMsgId" json:"chatwootMsgId"`   // camelCase exato com aspas duplas
	ChatwootConvId int64     `db:"chatwootConvId" json:"chatwootConvId"` // camelCase exato com aspas duplas
	Direction      string    `db:"direction" json:"direction"`           // encurtado: 'in', 'out'
	SyncStatus     string    `db:"syncStatus" json:"syncStatus"`         // camelCase exato com aspas duplas
	SourceId       *string   `db:"sourceId" json:"sourceId"`             // camelCase exato com aspas duplas
	Metadata       JSONB     `db:"metadata" json:"metadata"`             // syncError movido aqui
	CreatedAt      time.Time `db:"createdAt" json:"createdAt"`           // camelCase exato com aspas duplas
	UpdatedAt      time.Time `db:"updatedAt" json:"updatedAt"`           // camelCase exato com aspas duplas
	// ChatwootAccountID removido - obter via config
	// SyncError, LastSyncAt, ChatwootEchoID movidos para Metadata
}

func (ZpCwMessageModel) TableName() string {
	return "zp_cw_messages"
}
