package chatwoot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

// Client representa o cliente HTTP para a API do Chatwoot
type Client struct {
	baseURL    string
	token      string
	accountID  string
	httpClient *http.Client
}

// NewClient cria uma nova instância do cliente Chatwoot
func NewClient(baseURL, token, accountID string) *Client {
	return &Client{
		baseURL:   strings.TrimSuffix(baseURL, "/"),
		token:     token,
		accountID: accountID,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// makeRequest executa uma requisição HTTP para a API do Chatwoot
func (c *Client) makeRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	url := fmt.Sprintf("%s/api/v1/accounts/%s%s", c.baseURL, c.accountID, endpoint)
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api_access_token", c.token)

	return c.httpClient.Do(req)
}

// makeMultipartRequest executa uma requisição multipart para upload de arquivos
func (c *Client) makeMultipartRequest(ctx context.Context, method, endpoint string, fields map[string]string, file io.Reader, filename string) (*http.Response, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Adiciona campos de texto
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return nil, fmt.Errorf("failed to write field %s: %w", key, err)
		}
	}

	// Adiciona arquivo se fornecido
	if file != nil && filename != "" {
		part, err := writer.CreateFormFile("attachments[]", filename)
		if err != nil {
			return nil, fmt.Errorf("failed to create form file: %w", err)
		}

		if _, err := io.Copy(part, file); err != nil {
			return nil, fmt.Errorf("failed to copy file: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/accounts/%s%s", c.baseURL, c.accountID, endpoint)
	req, err := http.NewRequestWithContext(ctx, method, url, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("api_access_token", c.token)

	return c.httpClient.Do(req)
}

// parseResponse analisa a resposta HTTP e decodifica o JSON
func parseResponse(resp *http.Response, result interface{}) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
		}
		return fmt.Errorf("API error: %s", errResp.Message)
	}

	if result != nil {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// CreateContact cria um novo contato
func (c *Client) CreateContact(ctx context.Context, req ContactCreateRequest) (*Contact, error) {
	resp, err := c.makeRequest(ctx, "POST", "/contacts", req)
	if err != nil {
		return nil, err
	}

	var apiResp APIResponse
	if err := parseResponse(resp, &apiResp); err != nil {
		return nil, err
	}

	// Converte o payload para Contact
	contactData, err := json.Marshal(apiResp.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal contact data: %w", err)
	}

	var contact Contact
	if err := json.Unmarshal(contactData, &contact); err != nil {
		return nil, fmt.Errorf("failed to unmarshal contact: %w", err)
	}

	return &contact, nil
}

// GetContact busca um contato por ID
func (c *Client) GetContact(ctx context.Context, contactID int) (*Contact, error) {
	endpoint := fmt.Sprintf("/contacts/%d", contactID)
	resp, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var apiResp APIResponse
	if err := parseResponse(resp, &apiResp); err != nil {
		return nil, err
	}

	contactData, err := json.Marshal(apiResp.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal contact data: %w", err)
	}

	var contact Contact
	if err := json.Unmarshal(contactData, &contact); err != nil {
		return nil, fmt.Errorf("failed to unmarshal contact: %w", err)
	}

	return &contact, nil
}

// SearchContacts busca contatos por query
func (c *Client) SearchContacts(ctx context.Context, query string) ([]Contact, error) {
	endpoint := fmt.Sprintf("/contacts/search?q=%s", query)
	resp, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var apiResp APIResponse
	if err := parseResponse(resp, &apiResp); err != nil {
		return nil, err
	}

	contactsData, err := json.Marshal(apiResp.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal contacts data: %w", err)
	}

	var contacts []Contact
	if err := json.Unmarshal(contactsData, &contacts); err != nil {
		return nil, fmt.Errorf("failed to unmarshal contacts: %w", err)
	}

	return contacts, nil
}

// FilterContacts busca contatos usando filtro (mais robusto que search)
func (c *Client) FilterContacts(ctx context.Context, phoneNumber string) ([]Contact, error) {
	// Cria payload de filtro baseado na Evolution API
	filterPayload := c.createFilterPayload(phoneNumber)

	requestBody := map[string]interface{}{
		"payload": filterPayload,
	}

	resp, err := c.makeRequest(ctx, "POST", "/contacts/filter", requestBody)
	if err != nil {
		return nil, err
	}

	var apiResp APIResponse
	if err := parseResponse(resp, &apiResp); err != nil {
		return nil, err
	}

	contactsData, err := json.Marshal(apiResp.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal contacts data: %w", err)
	}

	var contacts []Contact
	if err := json.Unmarshal(contactsData, &contacts); err != nil {
		return nil, fmt.Errorf("failed to unmarshal contacts: %w", err)
	}

	return contacts, nil
}

// createFilterPayload cria o payload de filtro para busca de contatos
func (c *Client) createFilterPayload(phoneNumber string) []map[string]interface{} {
	numbers := c.getPhoneNumberVariations(phoneNumber)
	fieldsToSearch := []string{"phone_number"}

	var filterPayload []map[string]interface{}

	for i, field := range fieldsToSearch {
		for j, number := range numbers {
			// Remove o + do número para a busca
			searchNumber := strings.TrimPrefix(number, "+")

			// Determina se é o último item para definir query_operator
			var queryOperator *string
			if !(i == len(fieldsToSearch)-1 && j == len(numbers)-1) {
				op := "OR"
				queryOperator = &op
			}

			filter := map[string]interface{}{
				"attribute_key":   field,
				"filter_operator": "equal_to",
				"values":          []string{searchNumber},
			}

			if queryOperator != nil {
				filter["query_operator"] = *queryOperator
			}

			filterPayload = append(filterPayload, filter)
		}
	}

	return filterPayload
}

// getPhoneNumberVariations retorna variações do número de telefone (especialmente para números brasileiros)
func (c *Client) getPhoneNumberVariations(phoneNumber string) []string {
	numbers := []string{phoneNumber}

	// Para números brasileiros, adiciona variação com/sem 9º dígito
	if strings.HasPrefix(phoneNumber, "+55") && len(phoneNumber) == 14 {
		// Remove o 9º dígito
		withoutNine := phoneNumber[:5] + phoneNumber[6:]
		numbers = append(numbers, withoutNine)
	} else if strings.HasPrefix(phoneNumber, "+55") && len(phoneNumber) == 13 {
		// Adiciona o 9º dígito
		withNine := phoneNumber[:5] + "9" + phoneNumber[5:]
		numbers = append(numbers, withNine)
	}

	return numbers
}

// UpdateContact atualiza um contato existente
func (c *Client) UpdateContact(ctx context.Context, contactID int, updates map[string]interface{}) (*Contact, error) {
	endpoint := fmt.Sprintf("/contacts/%d", contactID)
	resp, err := c.makeRequest(ctx, "PATCH", endpoint, updates)
	if err != nil {
		return nil, err
	}

	var apiResp APIResponse
	if err := parseResponse(resp, &apiResp); err != nil {
		return nil, err
	}

	contactData, err := json.Marshal(apiResp.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal contact data: %w", err)
	}

	var contact Contact
	if err := json.Unmarshal(contactData, &contact); err != nil {
		return nil, fmt.Errorf("failed to unmarshal contact: %w", err)
	}

	return &contact, nil
}

// ListInboxes lista todas as inboxes da conta
func (c *Client) ListInboxes(ctx context.Context) ([]Inbox, error) {
	resp, err := c.makeRequest(ctx, "GET", "/inboxes", nil)
	if err != nil {
		return nil, err
	}

	var apiResp APIResponse
	if err := parseResponse(resp, &apiResp); err != nil {
		return nil, err
	}

	inboxesData, err := json.Marshal(apiResp.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal inboxes data: %w", err)
	}

	var inboxes []Inbox
	if err := json.Unmarshal(inboxesData, &inboxes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal inboxes: %w", err)
	}

	return inboxes, nil
}

// CreateInbox cria uma nova inbox
func (c *Client) CreateInbox(ctx context.Context, req InboxCreateRequest) (*Inbox, error) {
	resp, err := c.makeRequest(ctx, "POST", "/inboxes", req)
	if err != nil {
		return nil, err
	}

	var inbox Inbox
	if err := parseResponse(resp, &inbox); err != nil {
		return nil, err
	}

	return &inbox, nil
}

// CreateConversation cria uma nova conversa
func (c *Client) CreateConversation(ctx context.Context, req ConversationCreateRequest) (*Conversation, error) {
	resp, err := c.makeRequest(ctx, "POST", "/conversations", req)
	if err != nil {
		return nil, err
	}

	var conversation Conversation
	if err := parseResponse(resp, &conversation); err != nil {
		return nil, err
	}

	return &conversation, nil
}

// GetConversation busca uma conversa por ID
func (c *Client) GetConversation(ctx context.Context, conversationID int) (*Conversation, error) {
	endpoint := fmt.Sprintf("/conversations/%d", conversationID)
	resp, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var conversation Conversation
	if err := parseResponse(resp, &conversation); err != nil {
		return nil, err
	}

	return &conversation, nil
}

// ListContactConversations lista conversas de um contato
func (c *Client) ListContactConversations(ctx context.Context, contactID int) ([]Conversation, error) {
	endpoint := fmt.Sprintf("/contacts/%d/conversations", contactID)
	resp, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var apiResp APIResponse
	if err := parseResponse(resp, &apiResp); err != nil {
		return nil, err
	}

	conversationsData, err := json.Marshal(apiResp.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal conversations data: %w", err)
	}

	var conversations []Conversation
	if err := json.Unmarshal(conversationsData, &conversations); err != nil {
		return nil, fmt.Errorf("failed to unmarshal conversations: %w", err)
	}

	return conversations, nil
}

// CreateMessage cria uma nova mensagem
func (c *Client) CreateMessage(ctx context.Context, conversationID int, req MessageCreateRequest) (*Message, error) {
	endpoint := fmt.Sprintf("/conversations/%d/messages", conversationID)
	resp, err := c.makeRequest(ctx, "POST", endpoint, req)
	if err != nil {
		return nil, err
	}

	var message Message
	if err := parseResponse(resp, &message); err != nil {
		return nil, err
	}

	return &message, nil
}

// CreateMessageWithAttachment cria uma mensagem com anexo
func (c *Client) CreateMessageWithAttachment(ctx context.Context, conversationID int, content, messageType string, file io.Reader, filename string, sourceID string) (*Message, error) {
	endpoint := fmt.Sprintf("/conversations/%d/messages", conversationID)

	fields := map[string]string{
		"message_type": messageType,
	}

	if content != "" {
		fields["content"] = content
	}

	if sourceID != "" {
		fields["source_id"] = sourceID
	}

	resp, err := c.makeMultipartRequest(ctx, "POST", endpoint, fields, file, filename)
	if err != nil {
		return nil, err
	}

	var message Message
	if err := parseResponse(resp, &message); err != nil {
		return nil, err
	}

	return &message, nil
}

// UpdateConversationStatus atualiza o status de uma conversa
func (c *Client) UpdateConversationStatus(ctx context.Context, conversationID int, status ConversationStatus) error {
	endpoint := fmt.Sprintf("/conversations/%d/toggle_status", conversationID)
	req := map[string]string{"status": string(status)}

	resp, err := c.makeRequest(ctx, "POST", endpoint, req)
	if err != nil {
		return err
	}

	return parseResponse(resp, nil)
}

// DeleteMessage deleta uma mensagem
func (c *Client) DeleteMessage(ctx context.Context, conversationID, messageID int) error {
	endpoint := fmt.Sprintf("/conversations/%d/messages/%d", conversationID, messageID)
	resp, err := c.makeRequest(ctx, "DELETE", endpoint, nil)
	if err != nil {
		return err
	}

	return parseResponse(resp, nil)
}
