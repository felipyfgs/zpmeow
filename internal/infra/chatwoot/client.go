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
)

// Client implements Chatwoot API client
type Client struct {
	baseURL     string
	token       string
	accountID   string
	httpClient  *http.Client
	urlBuilder  *URLBuilder
	validator   *ResponseValidator
	errorHelper *ErrorHelper
}

// NewClient creates a new Chatwoot client
func NewClient(baseURL, token, accountID string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	return &Client{
		baseURL:     strings.TrimSuffix(baseURL, "/"),
		token:       token,
		accountID:   accountID,
		httpClient:  httpClient,
		urlBuilder:  NewURLBuilder(baseURL, accountID),
		validator:   NewResponseValidator(),
		errorHelper: NewErrorHelper(),
	}
}

// makeRequest executes HTTP request to Chatwoot API
func (c *Client) makeRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Response, error) {
	reqBody, err := c.prepareRequestBody(body)
	if err != nil {
		return nil, c.errorHelper.WrapError(err, "failed to prepare request body")
	}

	url := c.buildURL(endpoint)
	req, err := c.createRequest(ctx, method, url, reqBody)
	if err != nil {
		return nil, c.errorHelper.WrapError(err, "failed to create request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, c.errorHelper.WrapError(err, "failed to execute request")
	}

	return resp, nil
}

// prepareRequestBody prepares request body for JSON requests
func (c *Client) prepareRequestBody(body interface{}) (io.Reader, error) {
	if body == nil {
		return nil, nil
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(jsonBody), nil
}

// buildURL builds full URL for API endpoint
func (c *Client) buildURL(endpoint string) string {
	return fmt.Sprintf("%s/api/v1/accounts/%s%s", c.baseURL, c.accountID, endpoint)
}

// createRequest creates HTTP request with headers
func (c *Client) createRequest(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api_access_token", c.token)

	return req, nil
}

// makeMultipartRequest executes multipart request for file uploads
func (c *Client) makeMultipartRequest(ctx context.Context, method, endpoint string, fields map[string]string, file io.Reader, filename string) (*http.Response, error) {
	body, contentType, err := c.prepareMultipartBody(fields, file, filename)
	if err != nil {
		return nil, c.errorHelper.WrapError(err, "failed to prepare multipart body")
	}

	url := c.buildURL(endpoint)
	req, err := c.createMultipartRequest(ctx, method, url, body, contentType)
	if err != nil {
		return nil, c.errorHelper.WrapError(err, "failed to create multipart request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, c.errorHelper.WrapError(err, "failed to execute multipart request")
	}

	return resp, nil
}

// prepareMultipartBody prepares multipart form data
func (c *Client) prepareMultipartBody(fields map[string]string, file io.Reader, filename string) (io.Reader, string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	if err := c.addFormFields(writer, fields); err != nil {
		return nil, "", err
	}

	if err := c.addFormFile(writer, file, filename); err != nil {
		return nil, "", err
	}

	contentType := writer.FormDataContentType()
	if err := writer.Close(); err != nil {
		return nil, "", err
	}

	return &buf, contentType, nil
}

// addFormFields adds text fields to multipart form
func (c *Client) addFormFields(writer *multipart.Writer, fields map[string]string) error {
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return fmt.Errorf("failed to write field %s: %w", key, err)
		}
	}
	return nil
}

// addFormFile adds file to multipart form
func (c *Client) addFormFile(writer *multipart.Writer, file io.Reader, filename string) error {
	if file == nil || filename == "" {
		return nil
	}

	part, err := writer.CreateFormFile("attachments[]", filename)
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

// createMultipartRequest creates HTTP request for multipart data
func (c *Client) createMultipartRequest(ctx context.Context, method, url string, body io.Reader, contentType string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("api_access_token", c.token)

	return req, nil
}

// parseResponse analisa a resposta HTTP e decodifica o JSON
func parseResponse(resp *http.Response, result interface{}) error {
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close response body in parseResponse: %v\n", closeErr)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	fmt.Printf("🔍 [CHATWOOT API DEBUG] Response Status: %d\n", resp.StatusCode)
	fmt.Printf("📄 [CHATWOOT API DEBUG] FULL RESPONSE PAYLOAD: %s\n", string(body))

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			fmt.Printf("❌ [CHATWOOT API DEBUG] Failed to parse error response: %v\n", err)
			return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
		}
		fmt.Printf("❌ [CHATWOOT API DEBUG] API Error: %s\n", errResp.Message)
		return fmt.Errorf("API error: %s", errResp.Message)
	}

	if result != nil {
		fmt.Printf("🔄 [CHATWOOT API DEBUG] Unmarshaling to type: %T\n", result)
		if err := json.Unmarshal(body, result); err != nil {
			fmt.Printf("❌ [CHATWOOT API DEBUG] Failed to unmarshal: %v\n", err)
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
		fmt.Printf("✅ [CHATWOOT API DEBUG] Successfully unmarshaled result: %+v\n", result)
	}

	return nil
}

// CreateContact cria um novo contato
func (c *Client) CreateContact(ctx context.Context, req ContactCreateRequest) (*Contact, error) {
	fmt.Printf("🚀 [CHATWOOT API DEBUG] Creating contact: %+v\n", req)

	resp, err := c.makeRequest(ctx, "POST", "/contacts", req)
	if err != nil {
		fmt.Printf("❌ [CHATWOOT API DEBUG] Failed to make request: %v\n", err)
		return nil, err
	}

	// Vamos ler a resposta manualmente primeiro para debug
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ [CHATWOOT API DEBUG] Failed to read response body: %v\n", err)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Recriar o body para parseResponse
	resp.Body = io.NopCloser(bytes.NewReader(respBody))

	fmt.Printf("📋 [CHATWOOT API DEBUG] RAW RESPONSE BODY FOR CONTACT CREATION: %s\n", string(respBody))

	// Parse the response structure which has nested contact data
	var responseData struct {
		Payload struct {
			Contact Contact `json:"contact"`
		} `json:"payload"`
	}

	if err := json.Unmarshal(respBody, &responseData); err != nil {
		fmt.Printf("❌ [CHATWOOT API DEBUG] Failed to parse contact creation response: %v\n", err)
		return nil, fmt.Errorf("failed to parse contact creation response: %w", err)
	}

	contact := responseData.Payload.Contact

	// Validate that the contact was parsed correctly
	if contact.ID == 0 {
		fmt.Printf("❌ [CHATWOOT API DEBUG] Contact ID is 0, parsing may have failed\n")
		return nil, fmt.Errorf("contact creation response parsing failed: contact ID is 0")
	}

	fmt.Printf("✅ [CHATWOOT API DEBUG] Successfully parsed contact directly from response\n")

	fmt.Printf("✅ [CHATWOOT API DEBUG] Created contact - ID: %d, Name: %s, Phone: %s\n", contact.ID, contact.Name, contact.PhoneNumber)
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
			if i != len(fieldsToSearch)-1 || j != len(numbers)-1 {
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
	fmt.Printf("🚀 [CHATWOOT API DEBUG] Creating conversation: %+v\n", req)

	resp, err := c.makeRequest(ctx, "POST", "/conversations", req)
	if err != nil {
		fmt.Printf("❌ [CHATWOOT API DEBUG] Failed to create conversation: %v\n", err)
		return nil, err
	}

	// Ler resposta para debug
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ [CHATWOOT API DEBUG] Failed to read conversation response body: %v\n", err)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Recriar o body
	resp.Body = io.NopCloser(bytes.NewReader(respBody))

	fmt.Printf("📋 [CHATWOOT API DEBUG] RAW RESPONSE BODY FOR CONVERSATION CREATION: %s\n", string(respBody))

	var conversation Conversation
	if err := parseResponse(resp, &conversation); err != nil {
		fmt.Printf("❌ [CHATWOOT API DEBUG] Failed to parse conversation response: %v\n", err)
		return nil, err
	}

	fmt.Printf("✅ [CHATWOOT API DEBUG] Created conversation - ID: %d, Status: %s\n", conversation.ID, conversation.Status)
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
	fmt.Printf("🔍 [CHATWOOT API DEBUG] Listing conversations for contact ID: %d\n", contactID)

	endpoint := fmt.Sprintf("/contacts/%d/conversations", contactID)
	resp, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		fmt.Printf("❌ [CHATWOOT API DEBUG] Failed to list conversations: %v\n", err)
		return nil, err
	}

	// Ler resposta para debug
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ [CHATWOOT API DEBUG] Failed to read conversations response body: %v\n", err)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Recriar o body
	resp.Body = io.NopCloser(bytes.NewReader(respBody))

	fmt.Printf("📋 [CHATWOOT API DEBUG] RAW RESPONSE BODY FOR LIST CONVERSATIONS: %s\n", string(respBody))

	var apiResp APIResponse
	if err := parseResponse(resp, &apiResp); err != nil {
		fmt.Printf("❌ [CHATWOOT API DEBUG] Failed to parse conversations response: %v\n", err)
		return nil, err
	}

	fmt.Printf("🔄 [CHATWOOT API DEBUG] Conversations APIResponse payload: %+v\n", apiResp.Payload)

	conversationsData, err := json.Marshal(apiResp.Payload)
	if err != nil {
		fmt.Printf("❌ [CHATWOOT API DEBUG] Failed to marshal conversations: %v\n", err)
		return nil, fmt.Errorf("failed to marshal conversations data: %w", err)
	}

	fmt.Printf("🔄 [CHATWOOT API DEBUG] Conversations data JSON: %s\n", string(conversationsData))

	var conversations []Conversation
	if err := json.Unmarshal(conversationsData, &conversations); err != nil {
		fmt.Printf("❌ [CHATWOOT API DEBUG] Failed to unmarshal conversations: %v\n", err)
		return nil, fmt.Errorf("failed to unmarshal conversations: %w", err)
	}

	fmt.Printf("✅ [CHATWOOT API DEBUG] Found %d conversations for contact %d\n", len(conversations), contactID)
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
