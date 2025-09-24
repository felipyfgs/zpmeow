package chatwoot

import (
	"fmt"
	"net/http"
	"strings"
)

// URLBuilder constrói URLs para a API Chatwoot
type URLBuilder struct {
	baseURL   string
	accountID string
}

// NewURLBuilder cria um novo construtor de URLs
func NewURLBuilder(baseURL, accountID string) *URLBuilder {
	return &URLBuilder{
		baseURL:   strings.TrimSuffix(baseURL, "/"),
		accountID: accountID,
	}
}

// BuildContactsURL constrói URL para operações de contatos
func (ub *URLBuilder) BuildContactsURL() string {
	return fmt.Sprintf("%s/api/v1/accounts/%s/contacts", ub.baseURL, ub.accountID)
}

// BuildContactURL constrói URL para operação específica de contato
func (ub *URLBuilder) BuildContactURL(contactID int) string {
	return fmt.Sprintf("%s/api/v1/accounts/%s/contacts/%d", ub.baseURL, ub.accountID, contactID)
}

// BuildConversationsURL constrói URL para operações de conversas
func (ub *URLBuilder) BuildConversationsURL() string {
	return fmt.Sprintf("%s/api/v1/accounts/%s/conversations", ub.baseURL, ub.accountID)
}

// BuildConversationURL constrói URL para operação específica de conversa
func (ub *URLBuilder) BuildConversationURL(conversationID int) string {
	return fmt.Sprintf("%s/api/v1/accounts/%s/conversations/%d", ub.baseURL, ub.accountID, conversationID)
}

// BuildMessagesURL constrói URL para operações de mensagens
func (ub *URLBuilder) BuildMessagesURL(conversationID int) string {
	return fmt.Sprintf("%s/api/v1/accounts/%s/conversations/%d/messages", ub.baseURL, ub.accountID, conversationID)
}

// BuildInboxesURL constrói URL para operações de inboxes
func (ub *URLBuilder) BuildInboxesURL() string {
	return fmt.Sprintf("%s/api/v1/accounts/%s/inboxes", ub.baseURL, ub.accountID)
}

// BuildInboxURL constrói URL para operação específica de inbox
func (ub *URLBuilder) BuildInboxURL(inboxID int) string {
	return fmt.Sprintf("%s/api/v1/accounts/%s/inboxes/%d", ub.baseURL, ub.accountID, inboxID)
}

// BuildSearchURL constrói URL para busca
func (ub *URLBuilder) BuildSearchURL(resource string) string {
	return fmt.Sprintf("%s/api/v1/accounts/%s/%s/search", ub.baseURL, ub.accountID, resource)
}

// BuildFilterURL constrói URL para filtros
func (ub *URLBuilder) BuildFilterURL(resource string) string {
	return fmt.Sprintf("%s/api/v1/accounts/%s/%s/filter", ub.baseURL, ub.accountID, resource)
}

// CacheKeyBuilder constrói chaves de cache padronizadas
type CacheKeyBuilder struct{}

// NewCacheKeyBuilder cria um novo construtor de chaves de cache
func NewCacheKeyBuilder() *CacheKeyBuilder {
	return &CacheKeyBuilder{}
}

// ContactKey constrói chave de cache para contato
func (ckb *CacheKeyBuilder) ContactKey(phoneNumber string) string {
	return fmt.Sprintf("contact:%s", phoneNumber)
}

// ConversationKey constrói chave de cache para conversa
func (ckb *CacheKeyBuilder) ConversationKey(contactID int) string {
	return fmt.Sprintf("conversation:%d", contactID)
}

// InboxKey constrói chave de cache para inbox
func (ckb *CacheKeyBuilder) InboxKey(name string) string {
	return fmt.Sprintf("inbox:%s", name)
}

// MessageKey constrói chave de cache para mensagem
func (ckb *CacheKeyBuilder) MessageKey(messageID string) string {
	return fmt.Sprintf("message:%s", messageID)
}

// FileTypeDetector detecta tipos de arquivo
type FileTypeDetector struct{}

// NewFileTypeDetector cria um novo detector de tipos de arquivo
func NewFileTypeDetector() *FileTypeDetector {
	return &FileTypeDetector{}
}

// DetectMimeType detecta o tipo MIME baseado na extensão do arquivo
func (ftd *FileTypeDetector) DetectMimeType(filename string) string {
	ext := strings.ToLower(filename)

	if strings.HasSuffix(ext, ".jpg") || strings.HasSuffix(ext, ".jpeg") {
		return "image/jpeg"
	}
	if strings.HasSuffix(ext, ".png") {
		return "image/png"
	}
	if strings.HasSuffix(ext, ".gif") {
		return "image/gif"
	}
	if strings.HasSuffix(ext, ".webp") {
		return "image/webp"
	}
	if strings.HasSuffix(ext, ".mp4") {
		return "video/mp4"
	}
	if strings.HasSuffix(ext, ".mp3") {
		return "audio/mpeg"
	}
	if strings.HasSuffix(ext, ".ogg") {
		return "audio/ogg"
	}
	if strings.HasSuffix(ext, ".pdf") {
		return "application/pdf"
	}
	if strings.HasSuffix(ext, ".doc") || strings.HasSuffix(ext, ".docx") {
		return "application/msword"
	}

	return "application/octet-stream"
}

// IsImageType verifica se é um tipo de imagem
func (ftd *FileTypeDetector) IsImageType(mimeType string) bool {
	return strings.HasPrefix(mimeType, "image/")
}

// IsVideoType verifica se é um tipo de vídeo
func (ftd *FileTypeDetector) IsVideoType(mimeType string) bool {
	return strings.HasPrefix(mimeType, "video/")
}

// IsAudioType verifica se é um tipo de áudio
func (ftd *FileTypeDetector) IsAudioType(mimeType string) bool {
	return strings.HasPrefix(mimeType, "audio/")
}

// DetectFileType detecta o tipo de arquivo baseado no nome
func (ftd *FileTypeDetector) DetectFileType(filename string) string {
	return ftd.DetectMimeType(filename)
}

// extractFileName extrai o nome do arquivo de um caminho
func (ftd *FileTypeDetector) extractFileName(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return path
}

// PhoneNumberUtils utilitários para números de telefone
type PhoneNumberUtils struct{}

// NewPhoneNumberUtils cria uma nova instância dos utilitários
func NewPhoneNumberUtils() *PhoneNumberUtils {
	return &PhoneNumberUtils{}
}

// ExtractPhoneNumber extrai número de telefone de diferentes formatos
func (pnu *PhoneNumberUtils) ExtractPhoneNumber(from string) (string, bool) {
	// Remove sufixos do WhatsApp
	if idx := strings.Index(from, "@"); idx != -1 {
		from = from[:idx]
	}

	// Remove caracteres não numéricos
	phoneNumber := ""
	for _, char := range from {
		if char >= '0' && char <= '9' {
			phoneNumber += string(char)
		}
	}

	// Verifica se é um número válido (pelo menos 10 dígitos)
	if len(phoneNumber) < 10 {
		return "", false
	}

	return phoneNumber, true
}

// IsGroupJID verifica se o JID é de um grupo
func (pnu *PhoneNumberUtils) IsGroupJID(jid string) bool {
	return strings.Contains(jid, "@g.us")
}

// FormatPhoneNumber formata número de telefone para padrão internacional
func (pnu *PhoneNumberUtils) FormatPhoneNumber(phoneNumber string) string {
	// Remove caracteres não numéricos
	cleaned := ""
	for _, char := range phoneNumber {
		if char >= '0' && char <= '9' {
			cleaned += string(char)
		}
	}

	// Adiciona + se não tiver
	if !strings.HasPrefix(cleaned, "+") {
		cleaned = "+" + cleaned
	}

	return cleaned
}

// HTTPUtils utilitários para HTTP
type HTTPUtils struct{}

// NewHTTPUtils cria uma nova instância dos utilitários HTTP
func NewHTTPUtils() *HTTPUtils {
	return &HTTPUtils{}
}

// IsSuccessStatusCode verifica se o status code indica sucesso
func (hu *HTTPUtils) IsSuccessStatusCode(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}

// IsClientError verifica se o status code indica erro do cliente
func (hu *HTTPUtils) IsClientError(statusCode int) bool {
	return statusCode >= 400 && statusCode < 500
}

// IsServerError verifica se o status code indica erro do servidor
func (hu *HTTPUtils) IsServerError(statusCode int) bool {
	return statusCode >= 500
}

// GetStatusText retorna o texto do status HTTP
func (hu *HTTPUtils) GetStatusText(statusCode int) string {
	return http.StatusText(statusCode)
}

// StringUtils utilitários para strings
type StringUtils struct{}

// NewStringUtils cria uma nova instância dos utilitários de string
func NewStringUtils() *StringUtils {
	return &StringUtils{}
}

// IsEmpty verifica se uma string está vazia ou contém apenas espaços
func (su *StringUtils) IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// Truncate trunca uma string para o tamanho máximo especificado
func (su *StringUtils) Truncate(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength] + "..."
}

// SanitizeForLog sanitiza uma string para logging (remove caracteres especiais)
func (su *StringUtils) SanitizeForLog(s string) string {
	// Remove quebras de linha e caracteres de controle
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\t", " ")

	// Remove múltiplos espaços
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}

	return strings.TrimSpace(s)
}
