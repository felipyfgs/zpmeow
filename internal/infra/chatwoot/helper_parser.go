package chatwoot

import (
	"encoding/json"
	"fmt"
	"net/http"

	"zpmeow/internal/application/ports"
)

// ResponseParser centraliza o parsing de responses da API Chatwoot
type ResponseParser struct {
	errorHandler *ErrorHandler
}

// NewResponseParser cria um novo parser de responses
func NewResponseParser() *ResponseParser {
	return &ResponseParser{
		errorHandler: &ErrorHandler{},
	}
}

// ParseResponse faz o parsing genérico de uma response HTTP
func (rp *ResponseParser) ParseResponse(resp *http.Response, target interface{}) error {
	defer func() { _ = resp.Body.Close() }()

	// Verifica se a response é de erro
	if resp.StatusCode >= 400 {
		return rp.handleErrorResponse(resp)
	}

	// Faz o parsing do JSON
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// ParseContactResponse faz o parsing específico para responses de contato
func (rp *ResponseParser) ParseContactResponse(resp *http.Response) (*ports.ContactResponse, error) {
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return nil, rp.handleErrorResponse(resp)
	}

	// Para criação de contato, a response tem estrutura aninhada
	if resp.StatusCode == http.StatusCreated {
		return rp.parseContactCreationResponse(resp)
	}

	// Para outras operações, parsing direto
	var contact ports.ContactResponse
	if err := json.NewDecoder(resp.Body).Decode(&contact); err != nil {
		return nil, fmt.Errorf("failed to decode contact response: %w", err)
	}

	return &contact, nil
}

// parseContactCreationResponse faz o parsing da response de criação de contato
func (rp *ResponseParser) parseContactCreationResponse(resp *http.Response) (*ports.ContactResponse, error) {
	var responseData struct {
		Payload struct {
			Contact ports.ContactResponse `json:"contact"`
		} `json:"payload"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return nil, fmt.Errorf("failed to decode contact creation response: %w", err)
	}

	contact := responseData.Payload.Contact

	// Valida se o contato foi criado corretamente
	if contact.ID == 0 {
		return nil, fmt.Errorf("contact creation response parsing failed: contact ID is 0")
	}

	return &contact, nil
}

// ParseContactListResponse faz o parsing de lista de contatos
func (rp *ResponseParser) ParseContactListResponse(resp *http.Response) ([]*ports.ContactResponse, error) {
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return nil, rp.handleErrorResponse(resp)
	}

	var apiResponse struct {
		Payload []ports.ContactResponse `json:"payload"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("failed to decode contact list response: %w", err)
	}

	// Converte para slice de ponteiros
	contacts := make([]*ports.ContactResponse, len(apiResponse.Payload))
	for i := range apiResponse.Payload {
		contacts[i] = &apiResponse.Payload[i]
	}

	return contacts, nil
}

// ParseConversationResponse faz o parsing de response de conversa
func (rp *ResponseParser) ParseConversationResponse(resp *http.Response) (*ports.ConversationResponse, error) {
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return nil, rp.handleErrorResponse(resp)
	}

	var conversation ports.ConversationResponse
	if err := json.NewDecoder(resp.Body).Decode(&conversation); err != nil {
		return nil, fmt.Errorf("failed to decode conversation response: %w", err)
	}

	return &conversation, nil
}

// ParseConversationListResponse faz o parsing de lista de conversas
func (rp *ResponseParser) ParseConversationListResponse(resp *http.Response) ([]*ports.ConversationResponse, error) {
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return nil, rp.handleErrorResponse(resp)
	}

	var apiResponse struct {
		Payload []ports.ConversationResponse `json:"payload"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("failed to decode conversation list response: %w", err)
	}

	// Converte para slice de ponteiros
	conversations := make([]*ports.ConversationResponse, len(apiResponse.Payload))
	for i := range apiResponse.Payload {
		conversations[i] = &apiResponse.Payload[i]
	}

	return conversations, nil
}

// ParseMessageResponse faz o parsing de response de mensagem
func (rp *ResponseParser) ParseMessageResponse(resp *http.Response) (*ports.MessageResponse, error) {
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return nil, rp.handleErrorResponse(resp)
	}

	var message ports.MessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&message); err != nil {
		return nil, fmt.Errorf("failed to decode message response: %w", err)
	}

	return &message, nil
}

// ParseInboxResponse faz o parsing de response de inbox
func (rp *ResponseParser) ParseInboxResponse(resp *http.Response) (*ports.InboxResponse, error) {
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return nil, rp.handleErrorResponse(resp)
	}

	var inbox ports.InboxResponse
	if err := json.NewDecoder(resp.Body).Decode(&inbox); err != nil {
		return nil, fmt.Errorf("failed to decode inbox response: %w", err)
	}

	return &inbox, nil
}

// ParseInboxListResponse faz o parsing de lista de inboxes
func (rp *ResponseParser) ParseInboxListResponse(resp *http.Response) ([]*ports.InboxResponse, error) {
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return nil, rp.handleErrorResponse(resp)
	}

	var inboxes []ports.InboxResponse
	if err := json.NewDecoder(resp.Body).Decode(&inboxes); err != nil {
		return nil, fmt.Errorf("failed to decode inbox list response: %w", err)
	}

	// Converte para slice de ponteiros
	result := make([]*ports.InboxResponse, len(inboxes))
	for i := range inboxes {
		result[i] = &inboxes[i]
	}

	return result, nil
}

// handleErrorResponse trata responses de erro
func (rp *ResponseParser) handleErrorResponse(resp *http.Response) error {
	body, err := readResponseBody(resp)
	if err != nil {
		return fmt.Errorf("failed to read error response body: %w", err)
	}

	// Tenta fazer parsing da estrutura de erro do Chatwoot
	var errorResponse struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(body, &errorResponse); err == nil {
		message := errorResponse.Error
		if message == "" {
			message = errorResponse.Message
		}
		if message != "" {
			return NewAPIError(resp.StatusCode, message, resp.Request.URL.Path, resp.Request.Method)
		}
	}

	// Fallback para mensagem genérica
	return NewAPIError(resp.StatusCode, string(body), resp.Request.URL.Path, resp.Request.Method)
}

// readResponseBody lê o corpo da response de forma segura
func readResponseBody(resp *http.Response) ([]byte, error) {
	if resp.Body == nil {
		return []byte{}, nil
	}

	body, err := json.Marshal(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	return body, nil
}

// ResponseValidator valida responses da API
type ResponseValidator struct{}

// NewResponseValidator cria um novo validador de responses
func NewResponseValidator() *ResponseValidator {
	return &ResponseValidator{}
}

// ValidateContact valida uma response de contato
func (rv *ResponseValidator) ValidateContact(contact *ports.ContactResponse) error {
	if contact == nil {
		return fmt.Errorf("contact response is nil")
	}

	if contact.ID == 0 {
		return fmt.Errorf("contact ID is invalid")
	}

	if contact.Name == "" {
		return fmt.Errorf("contact name is empty")
	}

	return nil
}

// ValidateConversation valida uma response de conversa
func (rv *ResponseValidator) ValidateConversation(conversation *ports.ConversationResponse) error {
	if conversation == nil {
		return fmt.Errorf("conversation response is nil")
	}

	if conversation.ID == 0 {
		return fmt.Errorf("conversation ID is invalid")
	}

	if conversation.ContactID == 0 {
		return fmt.Errorf("conversation contact ID is invalid")
	}

	if conversation.InboxID == 0 {
		return fmt.Errorf("conversation inbox ID is invalid")
	}

	return nil
}

// ValidateMessage valida uma response de mensagem
func (rv *ResponseValidator) ValidateMessage(message *ports.MessageResponse) error {
	if message == nil {
		return fmt.Errorf("message response is nil")
	}

	if message.ID == 0 {
		return fmt.Errorf("message ID is invalid")
	}

	if message.ConversationID == 0 {
		return fmt.Errorf("message conversation ID is invalid")
	}

	return nil
}

// ValidateInbox valida uma response de inbox
func (rv *ResponseValidator) ValidateInbox(inbox *ports.InboxResponse) error {
	if inbox == nil {
		return fmt.Errorf("inbox response is nil")
	}

	if inbox.ID == 0 {
		return fmt.Errorf("inbox ID is invalid")
	}

	if inbox.Name == "" {
		return fmt.Errorf("inbox name is empty")
	}

	return nil
}
