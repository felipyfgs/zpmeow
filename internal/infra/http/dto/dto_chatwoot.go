package dto

// ChatwootConfigRequest representa a requisição para configurar a integração Chatwoot
// Campos obrigatórios: isActive
// Quando isActive=true, também são obrigatórios: accountId, token, url
// Todos os outros campos são opcionais
type ChatwootConfigRequest struct {
	// IsActive indica se a integração Chatwoot deve estar ativa (obrigatório)
	IsActive *bool `json:"isActive" validate:"required" example:"true" swaggertype:"boolean"`

	// AccountID é o ID da conta no Chatwoot (obrigatório quando isActive=true)
	AccountID string `json:"accountId" validate:"required_if=IsActive true" example:"1"`

	// Token é o token de API do Chatwoot (obrigatório quando isActive=true)
	Token string `json:"token" validate:"required_if=IsActive true" example:"your-chatwoot-api-token"`

	// URL é a URL da instância Chatwoot (obrigatório quando isActive=true)
	URL string `json:"url" validate:"required_if=IsActive true,url" example:"https://chatwoot.example.com"`

	// NameInbox é o nome da inbox no Chatwoot (opcional)
	NameInbox string `json:"nameInbox,omitempty" example:"WhatsApp Inbox"`

	// SignMsg indica se deve assinar mensagens (opcional)
	SignMsg *bool `json:"signMsg,omitempty" example:"true" swaggertype:"boolean"`

	// SignDelimiter é o delimitador usado na assinatura das mensagens (opcional)
	SignDelimiter string `json:"signDelimiter,omitempty" example:"\n\n---\nSent via WhatsApp"`

	// Number é o número do WhatsApp (opcional)
	Number string `json:"number,omitempty" example:"5511999999999"`

	// ReopenConversation indica se deve reabrir conversas (opcional)
	ReopenConversation *bool `json:"reopenConversation,omitempty" example:"false" swaggertype:"boolean"`

	// ConversationPending indica se conversas devem ficar pendentes (opcional)
	ConversationPending *bool `json:"conversationPending,omitempty" example:"true" swaggertype:"boolean"`

	// MergeBrazilContacts indica se deve mesclar contatos brasileiros (opcional)
	MergeBrazilContacts *bool `json:"mergeBrazilContacts,omitempty" example:"true" swaggertype:"boolean"`

	// ImportContacts indica se deve importar contatos (opcional)
	ImportContacts *bool `json:"importContacts,omitempty" example:"false" swaggertype:"boolean"`

	// ImportMessages indica se deve importar mensagens (opcional)
	ImportMessages *bool `json:"importMessages,omitempty" example:"false" swaggertype:"boolean"`

	// DaysLimitImportMessages é o limite de dias para importar mensagens (opcional, 1-365)
	DaysLimitImportMessages *int `json:"daysLimitImportMessages,omitempty" validate:"omitempty,min=1,max=365" example:"30"`

	// AutoCreate indica se deve criar automaticamente (opcional)
	AutoCreate *bool `json:"autoCreate,omitempty" example:"true" swaggertype:"boolean"`

	// Organization é o nome da organização (opcional)
	Organization string `json:"organization,omitempty" example:"My Company"`

	// Logo é a URL do logo da organização (opcional)
	Logo string `json:"logo,omitempty" validate:"omitempty,url" example:"https://example.com/logo.png"`

	// IgnoreJids é uma lista de JIDs para ignorar (opcional)
	IgnoreJids []string `json:"ignoreJids,omitempty" example:"554988989314@s.whatsapp.net,559999999999@s.whatsapp.net"`
}

// ChatwootConfigResponse representa a resposta da configuração Chatwoot
type ChatwootConfigResponse struct {
	Enabled                 bool     `json:"enabled"`
	AccountID               string   `json:"accountId,omitempty"`
	URL                     string   `json:"url,omitempty"`
	NameInbox               string   `json:"nameInbox,omitempty"`
	SignMsg                 bool     `json:"signMsg"`
	SignDelimiter           string   `json:"signDelimiter,omitempty"`
	Number                  string   `json:"number,omitempty"`
	ReopenConversation      bool     `json:"reopenConversation"`
	ConversationPending     bool     `json:"conversationPending"`
	MergeBrazilContacts     bool     `json:"mergeBrazilContacts"`
	ImportContacts          bool     `json:"importContacts"`
	ImportMessages          bool     `json:"importMessages"`
	DaysLimitImportMessages int      `json:"daysLimitImportMessages"`
	AutoCreate              bool     `json:"autoCreate"`
	Organization            string   `json:"organization,omitempty"`
	Logo                    string   `json:"logo,omitempty"`
	IgnoreJids              []string `json:"ignoreJids,omitempty"`
}

// ChatwootStatusResponse representa o status da integração Chatwoot
type ChatwootStatusResponse struct {
	Enabled       bool   `json:"enabled"`
	Connected     bool   `json:"connected"`
	InboxID       *int   `json:"inboxId,omitempty"`
	InboxName     string `json:"inboxName,omitempty"`
	LastSync      string `json:"lastSync,omitempty"`
	MessagesCount int    `json:"messagesCount"`
	ContactsCount int    `json:"contactsCount"`
	ErrorMessage  string `json:"errorMessage,omitempty"`
}

// ChatwootTestConnectionRequest representa a requisição para testar conexão
type ChatwootTestConnectionRequest struct {
	AccountID string `json:"accountId" validate:"required"`
	Token     string `json:"token" validate:"required"`
	URL       string `json:"url" validate:"required,url"`
}

// ChatwootTestConnectionResponse representa a resposta do teste de conexão
type ChatwootTestConnectionResponse struct {
	Success      bool                   `json:"success"`
	Message      string                 `json:"message"`
	AccountInfo  *ChatwootAccountInfo   `json:"accountInfo,omitempty"`
	InboxesCount int                    `json:"inboxesCount,omitempty"`
	ErrorDetails map[string]interface{} `json:"errorDetails,omitempty"`
}

// ChatwootAccountInfo representa informações da conta
type ChatwootAccountInfo struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Timezone string `json:"timezone,omitempty"`
	Locale   string `json:"locale,omitempty"`
}

// ChatwootSyncRequest representa a requisição para sincronizar dados
type ChatwootSyncRequest struct {
	SyncContacts bool `json:"syncContacts"`
	SyncMessages bool `json:"syncMessages"`
	DaysLimit    *int `json:"daysLimit,omitempty" validate:"omitempty,min=1,max=365"`
}

// ChatwootSyncResponse representa a resposta da sincronização
type ChatwootSyncResponse struct {
	Success        bool   `json:"success"`
	Message        string `json:"message"`
	ContactsSynced int    `json:"contactsSynced,omitempty"`
	MessagesSynced int    `json:"messagesSynced,omitempty"`
	ErrorMessage   string `json:"errorMessage,omitempty"`
}

// ChatwootMetricsResponse representa métricas da integração
type ChatwootMetricsResponse struct {
	TotalMessages      int `json:"totalMessages"`
	MessagesToday      int `json:"messagesToday"`
	TotalContacts      int `json:"totalContacts"`
	ContactsToday      int `json:"contactsToday"`
	TotalConversations int `json:"totalConversations"`
	OpenConversations  int `json:"openConversations"`
	AvgResponseTime    int `json:"avgResponseTime"` // em segundos
}

// ChatwootLogsRequest representa a requisição para obter logs
type ChatwootLogsRequest struct {
	Level     string `json:"level,omitempty" validate:"omitempty,oneof=debug info warn error"`
	StartDate string `json:"startDate,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	EndDate   string `json:"endDate,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	Limit     *int   `json:"limit,omitempty" validate:"omitempty,min=1,max=1000"`
}

// ChatwootLogsResponse representa a resposta dos logs
type ChatwootLogsResponse struct {
	Logs       []ChatwootLogEntry `json:"logs"`
	TotalCount int                `json:"totalCount"`
	HasMore    bool               `json:"hasMore"`
}

// ChatwootLogEntry representa uma entrada de log
type ChatwootLogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

// ChatwootHealthResponse representa a resposta de health check
type ChatwootHealthResponse struct {
	Status       string `json:"status"`
	Service      string `json:"service"`
	Version      string `json:"version,omitempty"`
	Uptime       string `json:"uptime,omitempty"`
	LastActivity string `json:"lastActivity,omitempty"`
}

// ChatwootWebhookPayload representa o payload de webhook do Chatwoot (interno)
type ChatwootWebhookPayload struct {
	Event             string                 `json:"event" validate:"required"`
	MsgType           interface{}            `json:"message_type,omitempty"` // pode ser string ou número
	ID                int                    `json:"id,omitempty"`
	Content           string                 `json:"content,omitempty"`
	CreatedAt         interface{}            `json:"created_at,omitempty"` // pode ser string ou número
	Private           bool                   `json:"private,omitempty"`
	SourceID          string                 `json:"source_id,omitempty"`
	ContentType       string                 `json:"content_type,omitempty"`
	ContentAttributes map[string]interface{} `json:"content_attributes,omitempty"`
	Sender            *ChatwootContact       `json:"sender,omitempty"`
	Contact           *ChatwootContact       `json:"contact,omitempty"`
	Conversation      *ChatwootConversation  `json:"conversation,omitempty"`
	Account           *ChatwootAccount       `json:"account,omitempty"`
	Inbox             *ChatwootInbox         `json:"inbox,omitempty"`
	Attachments       []ChatwootAttachment   `json:"attachments,omitempty"`
}

// ChatwootContact representa um contato no webhook
type ChatwootContact struct {
	ID               int                    `json:"id"`
	Name             string                 `json:"name"`
	Avatar           string                 `json:"avatar,omitempty"`
	AvatarURL        string                 `json:"avatar_url,omitempty"`
	PhoneNumber      string                 `json:"phone_number,omitempty"`
	Email            string                 `json:"email,omitempty"`
	Identifier       string                 `json:"identifier,omitempty"`
	Thumbnail        string                 `json:"thumbnail,omitempty"`
	CustomAttributes map[string]interface{} `json:"custom_attributes,omitempty"`
}

// ChatwootConversation representa uma conversa no webhook
type ChatwootConversation struct {
	ID                   int                    `json:"id"`
	AccountID            int                    `json:"account_id"`
	InboxID              int                    `json:"inbox_id"`
	Status               string                 `json:"status"`
	Timestamp            int64                  `json:"timestamp"`
	UnreadCount          int                    `json:"unread_count"`
	AdditionalAttributes map[string]interface{} `json:"additional_attributes,omitempty"`
	CustomAttributes     map[string]interface{} `json:"custom_attributes,omitempty"`
	Contact              *ChatwootContact       `json:"contact,omitempty"`
	Assignee             *ChatwootAgent         `json:"assignee,omitempty"`
	Team                 *ChatwootTeam          `json:"team,omitempty"`
	Meta                 *ChatwootMeta          `json:"meta,omitempty"`
}

// ChatwootAgent representa um agente no webhook
type ChatwootAgent struct {
	ID          int    `json:"id"`
	UID         string `json:"uid"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	AccountID   int    `json:"account_id"`
	Role        string `json:"role"`
}

// ChatwootTeam representa uma equipe no webhook
type ChatwootTeam struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	AllowAutoAssign bool   `json:"allow_auto_assign"`
	AccountID       int    `json:"account_id"`
}

// ChatwootMeta representa metadados da conversa no webhook
type ChatwootMeta struct {
	Sender   *ChatwootContact `json:"sender,omitempty"`
	Assignee *ChatwootAgent   `json:"assignee,omitempty"`
	Team     *ChatwootTeam    `json:"team,omitempty"`
}

// ChatwootAccount representa uma conta no webhook
type ChatwootAccount struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// ChatwootInbox representa uma inbox no webhook
type ChatwootInbox struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	ChannelID   int    `json:"channel_id"`
	ChannelType string `json:"channel_type"`
	PhoneNumber string `json:"phone_number,omitempty"`
}

// ChatwootAttachment representa um anexo no webhook
type ChatwootAttachment struct {
	ID        int    `json:"id"`
	MessageID int    `json:"message_id"`
	FileType  string `json:"file_type"`
	AccountID int    `json:"account_id"`
	Extension string `json:"extension"`
	DataURL   string `json:"data_url"`
	ThumbURL  string `json:"thumb_url,omitempty"`
	FileSize  int64  `json:"file_size"`
	Fallback  string `json:"fallback,omitempty"`
}
