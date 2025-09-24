package chatwoot

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"zpmeow/internal/application/ports"
)

// Validator implementa a interface ChatwootValidator
type Validator struct {
	phoneRegex *regexp.Regexp
}

// NewValidator cria um novo validador
func NewValidator() ports.ChatwootValidator {
	return &Validator{
		phoneRegex: regexp.MustCompile(`^\+?[1-9]\d{1,14}$`), // E.164 format
	}
}

// ValidatePhoneNumber valida um número de telefone
func (v *Validator) ValidatePhoneNumber(phoneNumber string) error {
	if phoneNumber == "" {
		return fmt.Errorf("phone number is empty")
	}

	// Remove espaços e caracteres especiais
	cleaned := strings.ReplaceAll(phoneNumber, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")

	// Verifica se tem pelo menos 10 dígitos
	digitCount := 0
	for _, char := range cleaned {
		if char >= '0' && char <= '9' {
			digitCount++
		}
	}

	if digitCount < 10 {
		return fmt.Errorf("phone number must have at least 10 digits")
	}

	if digitCount > 15 {
		return fmt.Errorf("phone number cannot have more than 15 digits")
	}

	return nil
}

// ValidateContactData valida dados de contato
func (v *Validator) ValidateContactData(name, phoneNumber string, isGroup bool) error {
	if name == "" {
		return fmt.Errorf("contact name is required")
	}

	if len(name) > 255 {
		return fmt.Errorf("contact name is too long (max 255 characters)")
	}

	if !isGroup {
		if err := v.ValidatePhoneNumber(phoneNumber); err != nil {
			return fmt.Errorf("invalid phone number: %w", err)
		}
	} else {
		if phoneNumber == "" {
			return fmt.Errorf("group identifier is required")
		}
	}

	return nil
}

// ValidateMessageContent valida conteúdo de mensagem
func (v *Validator) ValidateMessageContent(content string, contentType string) error {
	if content == "" && contentType == "text" {
		return fmt.Errorf("text message content cannot be empty")
	}

	if len(content) > 4096 {
		return fmt.Errorf("message content is too long (max 4096 characters)")
	}

	// Valida tipos de conteúdo suportados
	validContentTypes := []string{
		"text", "image", "audio", "video", "file", "sticker", "location", "contact",
	}

	isValidType := false
	for _, validType := range validContentTypes {
		if contentType == validType {
			isValidType = true
			break
		}
	}

	if !isValidType {
		return fmt.Errorf("unsupported content type: %s", contentType)
	}

	return nil
}

// ValidateWebhookPayload valida payload de webhook
func (v *Validator) ValidateWebhookPayload(payload []byte) error {
	if len(payload) == 0 {
		return fmt.Errorf("webhook payload is empty")
	}

	// Verifica se é JSON válido
	var webhookData map[string]interface{}
	if err := json.Unmarshal(payload, &webhookData); err != nil {
		return fmt.Errorf("invalid JSON payload: %w", err)
	}

	// Verifica se tem campo event
	event, exists := webhookData["event"]
	if !exists {
		return fmt.Errorf("webhook payload missing 'event' field")
	}

	eventStr, ok := event.(string)
	if !ok {
		return fmt.Errorf("webhook 'event' field must be a string")
	}

	if eventStr == "" {
		return fmt.Errorf("webhook 'event' field cannot be empty")
	}

	// Valida eventos suportados
	validEvents := []string{
		"message_created", "message_updated", "conversation_created",
		"conversation_updated", "conversation_status_changed",
	}

	isValidEvent := false
	for _, validEvent := range validEvents {
		if eventStr == validEvent {
			isValidEvent = true
			break
		}
	}

	if !isValidEvent {
		return fmt.Errorf("unsupported webhook event: %s", eventStr)
	}

	return nil
}

// ValidateConversationRequest valida request de criação de conversa
func (v *Validator) ValidateConversationRequest(req *ports.ConversationCreateRequest) error {
	if req == nil {
		return fmt.Errorf("conversation request is nil")
	}

	if req.ContactID <= 0 {
		return fmt.Errorf("contact ID must be positive")
	}

	if req.InboxID <= 0 {
		return fmt.Errorf("inbox ID must be positive")
	}

	if req.Status != "" {
		validStatuses := []string{"open", "pending", "resolved", "snoozed"}
		isValidStatus := false
		for _, status := range validStatuses {
			if req.Status == status {
				isValidStatus = true
				break
			}
		}
		if !isValidStatus {
			return fmt.Errorf("invalid conversation status: %s", req.Status)
		}
	}

	return nil
}

// ValidateMessageRequest valida request de criação de mensagem
func (v *Validator) ValidateMessageRequest(req *ports.MessageCreateRequest) error {
	if req == nil {
		return fmt.Errorf("message request is nil")
	}

	if err := v.ValidateMessageContent(req.Content, "text"); err != nil {
		return err
	}

	if req.MessageType < 0 || req.MessageType > 2 {
		return fmt.Errorf("invalid message type: %d (must be 0=incoming, 1=outgoing, 2=activity)", req.MessageType)
	}

	return nil
}

// ValidateInboxRequest valida request de criação de inbox
func (v *Validator) ValidateInboxRequest(req *ports.InboxCreateRequest) error {
	if req == nil {
		return fmt.Errorf("inbox request is nil")
	}

	if req.Name == "" {
		return fmt.Errorf("inbox name is required")
	}

	if len(req.Name) > 100 {
		return fmt.Errorf("inbox name is too long (max 100 characters)")
	}

	if req.Channel == nil {
		return fmt.Errorf("inbox channel configuration is required")
	}

	channelType, exists := req.Channel["type"]
	if !exists {
		return fmt.Errorf("inbox channel type is required")
	}

	channelTypeStr, ok := channelType.(string)
	if !ok {
		return fmt.Errorf("inbox channel type must be a string")
	}

	if channelTypeStr != "api" {
		return fmt.Errorf("unsupported channel type: %s", channelTypeStr)
	}

	return nil
}

// ValidateURL valida uma URL
func (v *Validator) ValidateURL(url string) error {
	if url == "" {
		return fmt.Errorf("URL is empty")
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("URL must start with http:// or https://")
	}

	if len(url) > 2048 {
		return fmt.Errorf("URL is too long (max 2048 characters)")
	}

	return nil
}

// ValidateSessionID valida um ID de sessão
func (v *Validator) ValidateSessionID(sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("session ID is empty")
	}

	if len(sessionID) < 3 {
		return fmt.Errorf("session ID is too short (min 3 characters)")
	}

	if len(sessionID) > 100 {
		return fmt.Errorf("session ID is too long (max 100 characters)")
	}

	// Verifica se contém apenas caracteres válidos
	validChars := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validChars.MatchString(sessionID) {
		return fmt.Errorf("session ID contains invalid characters (only alphanumeric, underscore and dash allowed)")
	}

	return nil
}

// ValidateTimeout valida um timeout
func (v *Validator) ValidateTimeout(timeout time.Duration) error {
	if timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	if timeout > 5*time.Minute {
		return fmt.Errorf("timeout is too long (max 5 minutes)")
	}

	return nil
}

// ValidateAccountID valida um ID de conta
func (v *Validator) ValidateAccountID(accountID int) error {
	if accountID <= 0 {
		return fmt.Errorf("account ID must be positive")
	}

	return nil
}

// ValidateToken valida um token de API
func (v *Validator) ValidateToken(token string) error {
	if token == "" {
		return fmt.Errorf("API token is empty")
	}

	if len(token) < 10 {
		return fmt.Errorf("API token is too short")
	}

	if len(token) > 500 {
		return fmt.Errorf("API token is too long")
	}

	return nil
}
