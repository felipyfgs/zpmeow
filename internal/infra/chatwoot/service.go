package chatwoot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"zpmeow/internal/application/ports"
)

// Service representa o serviço de integração Chatwoot
type Service struct {
	client          *Client
	config          *ChatwootConfig
	logger          *slog.Logger
	cache           map[string]interface{}
	cacheMutex      sync.RWMutex
	inbox           *Inbox
	whatsappService ports.WhatsAppService
	sessionID       string
}

// NewService cria uma nova instância do serviço Chatwoot
func NewService(config *ChatwootConfig, logger *slog.Logger, whatsappService ports.WhatsAppService, sessionID string) (*Service, error) {
	if !config.Enabled {
		return nil, fmt.Errorf("chatwoot integration is disabled")
	}

	client := NewClient(config.URL, config.Token, config.AccountID)

	service := &Service{
		client:          client,
		config:          config,
		logger:          logger,
		cache:           make(map[string]interface{}),
		cacheMutex:      sync.RWMutex{},
		whatsappService: whatsappService,
		sessionID:       sessionID,
	}

	// Inicializa a inbox
	if err := service.initializeInbox(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize inbox: %w", err)
	}

	return service, nil
}

// SetWhatsAppService atualiza o serviço WhatsApp
func (s *Service) SetWhatsAppService(whatsappService ports.WhatsAppService) {
	s.whatsappService = whatsappService
}

// getContentTypeFromWhatsAppMessage determina o tipo de conteúdo baseado no tipo de mensagem do WhatsApp
func (s *Service) getContentTypeFromWhatsAppMessage(msg *WhatsAppMessage) string {
	switch msg.Type {
	case "text", "extendedTextMessage":
		return string(ContentTypeText)
	case "image", "imageMessage":
		return string(ContentTypeImage)
	case "audio", "audioMessage", "ptt":
		return string(ContentTypeAudio)
	case "video", "videoMessage":
		return string(ContentTypeVideo)
	case "document", "documentMessage":
		return string(ContentTypeFile)
	case "sticker", "stickerMessage":
		return string(ContentTypeSticker)
	case "location", "locationMessage":
		return string(ContentTypeLocation)
	case "contact", "contactMessage", "contactsArrayMessage":
		return string(ContentTypeContact)
	default:
		s.logger.Warn("Unknown WhatsApp message type, defaulting to text",
			"type", msg.Type,
			"message_id", msg.ID)
		return string(ContentTypeText)
	}
}

// formatMessageContentByType formata o conteúdo da mensagem baseado no tipo
func (s *Service) formatMessageContentByType(msg *WhatsAppMessage, isGroup bool) string {
	switch msg.Type {
	case "text", "extendedTextMessage":
		return s.formatTextMessage(msg, isGroup)
	case "image", "imageMessage":
		return s.formatMediaMessage(msg, "📷 Imagem", isGroup)
	case "audio", "audioMessage", "ptt":
		return s.formatMediaMessage(msg, "🎵 Áudio", isGroup)
	case "video", "videoMessage":
		return s.formatMediaMessage(msg, "🎬 Vídeo", isGroup)
	case "document", "documentMessage":
		return s.formatDocumentMessage(msg, isGroup)
	case "sticker", "stickerMessage":
		return s.formatMediaMessage(msg, "🎭 Sticker", isGroup)
	case "location", "locationMessage":
		return s.formatLocationMessage(msg, isGroup)
	case "contact", "contactMessage", "contactsArrayMessage":
		return s.formatContactMessage(msg, isGroup)
	default:
		s.logger.Warn("Unknown message type for formatting",
			"type", msg.Type,
			"message_id", msg.ID)
		return s.formatTextMessage(msg, isGroup)
	}
}

// formatTextMessage formata mensagens de texto
func (s *Service) formatTextMessage(msg *WhatsAppMessage, isGroup bool) string {
	content := msg.Body
	if isGroup && msg.PushName != "" {
		content = fmt.Sprintf("*%s:* %s", msg.PushName, content)
	}
	return content
}

// formatMediaMessage formata mensagens de mídia
func (s *Service) formatMediaMessage(msg *WhatsAppMessage, mediaType string, isGroup bool) string {
	content := mediaType
	if msg.Caption != "" {
		content = fmt.Sprintf("%s\n\n%s", mediaType, msg.Caption)
	}
	if isGroup && msg.PushName != "" {
		content = fmt.Sprintf("*%s:* %s", msg.PushName, content)
	}
	return content
}

// formatDocumentMessage formata mensagens de documento
func (s *Service) formatDocumentMessage(msg *WhatsAppMessage, isGroup bool) string {
	content := "📄 Documento"
	if msg.FileName != "" {
		content = fmt.Sprintf("📄 Documento: %s", msg.FileName)
	}
	if msg.Caption != "" {
		content = fmt.Sprintf("%s\n\n%s", content, msg.Caption)
	}
	if isGroup && msg.PushName != "" {
		content = fmt.Sprintf("*%s:* %s", msg.PushName, content)
	}
	return content
}

// formatLocationMessage formata mensagens de localização
func (s *Service) formatLocationMessage(msg *WhatsAppMessage, isGroup bool) string {
	content := "📍 Localização compartilhada"
	if msg.Location != nil {
		content = fmt.Sprintf("📍 Localização: %.6f, %.6f", msg.Location.Latitude, msg.Location.Longitude)
		if msg.Location.Name != "" {
			content = fmt.Sprintf("📍 %s\nCoordenadas: %.6f, %.6f", msg.Location.Name, msg.Location.Latitude, msg.Location.Longitude)
		}
	}
	if isGroup && msg.PushName != "" {
		content = fmt.Sprintf("*%s:* %s", msg.PushName, content)
	}
	return content
}

// formatContactMessage formata mensagens de contato
func (s *Service) formatContactMessage(msg *WhatsAppMessage, isGroup bool) string {
	content := "👤 Contato compartilhado"
	if len(msg.Contacts) > 0 {
		contact := msg.Contacts[0]
		if contact.DisplayName != "" {
			content = fmt.Sprintf("👤 Contato: %s", contact.DisplayName)
		}
		if len(contact.Phones) > 0 {
			content = fmt.Sprintf("%s\nTelefone: %s", content, contact.Phones[0].Number)
		}
	}
	if isGroup && msg.PushName != "" {
		content = fmt.Sprintf("*%s:* %s", msg.PushName, content)
	}
	return content
}

// processMediaAttachment processa anexos de mídia para o Chatwoot
func (s *Service) processMediaAttachment(ctx context.Context, msg *WhatsAppMessage) map[string]interface{} {
	contentAttributes := make(map[string]interface{})

	// Adiciona informações de mídia se disponível
	if msg.MediaURL != "" {
		contentAttributes["media_url"] = msg.MediaURL
	}

	if msg.MimeType != "" {
		contentAttributes["mime_type"] = msg.MimeType
	}

	if msg.FileName != "" {
		contentAttributes["file_name"] = msg.FileName
	}

	// Processa localização
	if msg.Location != nil {
		contentAttributes["location"] = map[string]interface{}{
			"latitude":  msg.Location.Latitude,
			"longitude": msg.Location.Longitude,
			"name":      msg.Location.Name,
			"address":   msg.Location.Address,
		}
	}

	// Processa contatos
	if len(msg.Contacts) > 0 {
		contacts := make([]map[string]interface{}, len(msg.Contacts))
		for i, contact := range msg.Contacts {
			contactData := map[string]interface{}{
				"display_name": contact.DisplayName,
				"first_name":   contact.FirstName,
				"last_name":    contact.LastName,
			}

			if len(contact.Phones) > 0 {
				phones := make([]map[string]interface{}, len(contact.Phones))
				for j, phone := range contact.Phones {
					phones[j] = map[string]interface{}{
						"number": phone.Number,
						"type":   phone.Type,
					}
				}
				contactData["phones"] = phones
			}

			if len(contact.Emails) > 0 {
				emails := make([]map[string]interface{}, len(contact.Emails))
				for j, email := range contact.Emails {
					emails[j] = map[string]interface{}{
						"email": email.Email,
						"type":  email.Type,
					}
				}
				contactData["emails"] = emails
			}

			contacts[i] = contactData
		}
		contentAttributes["contacts"] = contacts
	}

	return contentAttributes
}

// sendMediaToChatwoot envia mídia como anexo real para o Chatwoot
func (s *Service) sendMediaToChatwoot(ctx context.Context, conversationId int, msg *WhatsAppMessage, messageType int, sourceID string) (*Message, error) {
	// Download da mídia usando o serviço zpmeow
	s.logger.Info("📥 DOWNLOADING MEDIA FROM WHATSAPP",
		"message_id", msg.ID,
		"type", msg.Type,
		"mime_type", msg.MimeType)

	mediaData, mimeType, err := s.whatsappService.DownloadMedia(ctx, s.sessionID, msg.ID)
	if err != nil {
		s.logger.Error("❌ FAILED TO DOWNLOAD MEDIA",
			"error", err,
			"message_id", msg.ID)
		return nil, fmt.Errorf("failed to download media: %w", err)
	}

	if len(mediaData) == 0 {
		s.logger.Warn("⚠️ MEDIA DATA IS EMPTY",
			"message_id", msg.ID,
			"type", msg.Type)
		return nil, fmt.Errorf("media data is empty")
	}

	s.logger.Info("✅ MEDIA DOWNLOADED",
		"size", len(mediaData),
		"mime_type", mimeType,
		"message_id", msg.ID)

	// Usa o MIME type retornado pelo download ou o da mensagem
	finalMimeType := mimeType
	if finalMimeType == "" || finalMimeType == "application/octet-stream" {
		finalMimeType = msg.MimeType
	}

	// Determina o nome do arquivo
	fileName := s.getMediaFileName(msg)

	// Envia como anexo para o Chatwoot
	return s.sendMediaAttachmentToChatwoot(ctx, conversationId, mediaData, fileName, finalMimeType, msg.Body, messageType, sourceID)
}

// downloadMediaFromWhatsApp faz download da mídia do WhatsApp
func (s *Service) downloadMediaFromWhatsApp(ctx context.Context, mediaURL string) ([]byte, error) {
	if mediaURL == "" {
		return nil, fmt.Errorf("media URL is empty")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", mediaURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download media: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download media: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read media data: %w", err)
	}

	return data, nil
}

// getMediaFileName determina o nome do arquivo baseado no tipo de mídia
func (s *Service) getMediaFileName(msg *WhatsAppMessage) string {
	if msg.FileName != "" {
		return msg.FileName
	}

	// Gera nome baseado no tipo
	extension := s.getFileExtensionFromMimeType(msg.MimeType)
	timestamp := time.Now().Unix()

	switch msg.Type {
	case "audio", "ptt":
		return fmt.Sprintf("audio_%d.%s", timestamp, extension)
	case "image":
		return fmt.Sprintf("image_%d.%s", timestamp, extension)
	case "video":
		return fmt.Sprintf("video_%d.%s", timestamp, extension)
	case "document":
		return fmt.Sprintf("document_%d.%s", timestamp, extension)
	case "sticker":
		return fmt.Sprintf("sticker_%d.%s", timestamp, extension)
	default:
		return fmt.Sprintf("file_%d.%s", timestamp, extension)
	}
}

// getFileExtensionFromMimeType retorna a extensão baseada no MIME type
func (s *Service) getFileExtensionFromMimeType(mimeType string) string {
	switch mimeType {
	case "audio/ogg", "audio/ogg; codecs=opus":
		return "ogg"
	case "audio/mpeg", "audio/mp3":
		return "mp3"
	case "audio/wav":
		return "wav"
	case "audio/aac":
		return "aac"
	case "image/jpeg":
		return "jpg"
	case "image/png":
		return "png"
	case "image/gif":
		return "gif"
	case "image/webp":
		return "webp"
	case "video/mp4":
		return "mp4"
	case "video/avi":
		return "avi"
	case "video/mov":
		return "mov"
	case "application/pdf":
		return "pdf"
	case "application/msword":
		return "doc"
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return "docx"
	default:
		return "bin"
	}
}

// sendMediaAttachmentToChatwoot envia mídia como anexo usando FormData
func (s *Service) sendMediaAttachmentToChatwoot(ctx context.Context, conversationId int, mediaData []byte, fileName, mimeType, content string, messageType int, sourceID string) (*Message, error) {
	// Cria FormData
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Adiciona o conteúdo de texto se houver
	if content != "" {
		if err := writer.WriteField("content", content); err != nil {
			return nil, fmt.Errorf("failed to write content field: %w", err)
		}
	}

	// Adiciona o message_type baseado na Evolution API
	messageTypeStr := "incoming"
	if messageType == 1 {
		messageTypeStr = "outgoing"
	}
	if err := writer.WriteField("message_type", messageTypeStr); err != nil {
		return nil, fmt.Errorf("failed to write message_type field: %w", err)
	}

	// NÃO enviamos file_type - deixamos o Chatwoot determinar automaticamente baseado no MIME type
	// Isso segue exatamente o mesmo padrão da Evolution API

	// Adiciona o source_id se houver
	if sourceID != "" {
		if err := writer.WriteField("source_id", sourceID); err != nil {
			return nil, fmt.Errorf("failed to write source_id field: %w", err)
		}
	}

	// Adiciona o arquivo de mídia com Content-Type específico
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="attachments[]"; filename="%s"`, fileName))
	h.Set("Content-Type", mimeType)

	part, err := writer.CreatePart(h)
	if err != nil {
		return nil, fmt.Errorf("failed to create form part: %w", err)
	}

	if _, err := part.Write(mediaData); err != nil {
		return nil, fmt.Errorf("failed to write media data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	// Prepara a requisição
	url := fmt.Sprintf("%s/api/v1/accounts/%s/conversations/%d/messages", s.client.baseURL, s.client.accountID, conversationId)

	req, err := http.NewRequestWithContext(ctx, "POST", url, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Define headers
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("api_access_token", s.client.token)

	s.logger.Info("📤 SENDING MEDIA TO CHATWOOT",
		"url", url,
		"file_name", fileName,
		"mime_type", mimeType,
		"size", len(mediaData),
		"conversation_id", conversationId,
		"message_type", messageTypeStr)

	// Envia a requisição
	client := &http.Client{
		Timeout: 60 * time.Second, // Timeout maior para upload de mídia
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Lê a resposta
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		s.logger.Error("❌ CHATWOOT API ERROR",
			"status", resp.StatusCode,
			"response", string(respBody))
		return nil, fmt.Errorf("chatwoot API error: status %d, response: %s", resp.StatusCode, string(respBody))
	}

	s.logger.Info("✅ MEDIA SENT TO CHATWOOT SUCCESSFULLY",
		"conversation_id", conversationId,
		"file_name", fileName,
		"response_size", len(respBody))

	// Parse da resposta
	var response Message
	if err := json.Unmarshal(respBody, &response); err != nil {
		s.logger.Warn("Failed to parse response JSON, but request was successful",
			"error", err,
			"response", string(respBody))
		// Retorna uma resposta básica se não conseguir fazer parse
		return &Message{
			ID: 0, // Será preenchido se conseguirmos extrair
		}, nil
	}

	return &response, nil
}

// initializeInbox inicializa ou encontra a inbox configurada
func (s *Service) initializeInbox(ctx context.Context) error {
	inboxes, err := s.client.ListInboxes(ctx)
	if err != nil {
		return fmt.Errorf("failed to list inboxes: %w", err)
	}

	// Procura por uma inbox existente com o nome configurado
	for _, inbox := range inboxes {
		if inbox.Name == s.config.NameInbox {
			s.inbox = &inbox
			s.logger.Info("Found existing inbox", "name", inbox.Name, "id", inbox.ID)
			return nil
		}
	}

	// Se não encontrou, cria uma nova inbox se autoCreate estiver habilitado
	if s.config.AutoCreate {
		return s.createInbox(ctx)
	}

	return fmt.Errorf("inbox '%s' not found and auto-create is disabled", s.config.NameInbox)
}

// createInbox cria uma nova inbox API
func (s *Service) createInbox(ctx context.Context) error {
	// Usa WebhookURL da configuração se disponível, senão usa SERVER_HOST
	webhookURL := s.config.WebhookURL
	if webhookURL == "" {
		// Usa SERVER_HOST do .env ao invés de localhost
		serverHost := os.Getenv("SERVER_HOST")
		if serverHost == "" {
			serverHost = "localhost:8080" // Fallback apenas se SERVER_HOST não estiver definido
		}

		// Adiciona esquema se não estiver presente
		if !strings.HasPrefix(serverHost, "http://") && !strings.HasPrefix(serverHost, "https://") {
			serverHost = fmt.Sprintf("http://%s", serverHost)
		}

		webhookURL = fmt.Sprintf("%s/chatwoot/webhook/%s", serverHost, s.sessionID)
		s.logger.Info("Generated webhook URL from SERVER_HOST", "webhook", webhookURL, "server_host", os.Getenv("SERVER_HOST"))
	}

	req := InboxCreateRequest{
		Name: s.config.NameInbox,
		Channel: map[string]interface{}{
			"type":        "api",
			"webhook_url": webhookURL,
		},
	}

	inbox, err := s.client.CreateInbox(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create inbox: %w", err)
	}

	s.inbox = inbox
	s.logger.Info("Created new inbox", "name", inbox.Name, "id", inbox.ID, "webhook", webhookURL)
	return nil
}

// getFromCache recupera um item do cache
func (s *Service) getFromCache(key string) (interface{}, bool) {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()
	value, exists := s.cache[key]
	return value, exists
}

// setCache armazena um item no cache
func (s *Service) setCache(key string, value interface{}) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	s.cache[key] = value
}

// findOrCreateContact encontra ou cria um contato
func (s *Service) findOrCreateContact(ctx context.Context, phoneNumber, name, avatarURL string, isGroup bool) (*Contact, error) {
	cacheKey := fmt.Sprintf("contact:%s", phoneNumber)
	
	// Verifica cache primeiro
	if cached, exists := s.getFromCache(cacheKey); exists {
		if contact, ok := cached.(*Contact); ok {
			return contact, nil
		}
	}

	// Busca contato existente
	var searchQuery string
	if isGroup {
		searchQuery = phoneNumber
	} else {
		searchQuery = fmt.Sprintf("+%s", phoneNumber)
	}

	var contacts []Contact
	var err error

	if isGroup {
		contacts, err = s.client.SearchContacts(ctx, searchQuery)
	} else {
		// Para contatos individuais, usar filtro mais robusto
		contacts, err = s.client.FilterContacts(ctx, searchQuery)
	}

	if err != nil {
		s.logger.Error("Failed to search contacts", "error", err)
	}

	// Se encontrou contato, atualiza cache e retorna
	if len(contacts) > 0 {
		contact := s.findBestMatchContact(contacts, searchQuery)
		if contact != nil {
			s.setCache(cacheKey, contact)
			return contact, nil
		}
	}

	// Cria novo contato com fallback para contatos duplicados
	req := ContactCreateRequest{
		InboxID: s.inbox.ID,
		Name:    name,
	}

	if !isGroup {
		req.PhoneNumber = fmt.Sprintf("+%s", phoneNumber)
		req.Identifier = fmt.Sprintf("%s@s.whatsapp.net", phoneNumber)
	} else {
		req.Identifier = phoneNumber
	}

	if avatarURL != "" {
		req.AvatarURL = avatarURL
	}

	contact, err := s.client.CreateContact(ctx, req)
	if err != nil {
		// Se falhou por contato duplicado, tenta buscar novamente
		if strings.Contains(err.Error(), "already been taken") || strings.Contains(err.Error(), "Identifier has already been taken") {
			s.logger.Warn("Contact already exists, searching again", "phoneNumber", phoneNumber, "error", err)

			// Tenta buscar novamente com diferentes métodos
			var retryContacts []Contact
			var retryErr error
			if isGroup {
				retryContacts, retryErr = s.client.SearchContacts(ctx, searchQuery)
			} else {
				retryContacts, retryErr = s.client.FilterContacts(ctx, searchQuery)
			}

			// Se FilterContacts falhar, tenta SearchContacts como último recurso
			if retryErr != nil && !isGroup {
				s.logger.Warn("FilterContacts failed, trying SearchContacts as fallback", "error", retryErr)
				retryContacts, retryErr = s.client.SearchContacts(ctx, searchQuery)
			}

			if retryErr == nil && len(retryContacts) > 0 {
				contact := s.findBestMatchContact(retryContacts, searchQuery)
				if contact != nil {
					s.setCache(cacheKey, contact)
					s.logger.Info("Found existing contact after creation failure", "name", contact.Name, "phone", contact.PhoneNumber, "id", contact.ID)
					return contact, nil
				}
			} else if retryErr != nil {
				s.logger.Error("All contact search methods failed", "error", retryErr)
			}
		}
		return nil, fmt.Errorf("failed to create contact: %w", err)
	}

	s.setCache(cacheKey, contact)
	s.logger.Info("Created new contact", "name", name, "phone", phoneNumber, "id", contact.ID)
	return contact, nil
}

// findBestMatchContact encontra o melhor contato correspondente
func (s *Service) findBestMatchContact(contacts []Contact, query string) *Contact {
	if len(contacts) == 0 {
		return nil
	}

	// Se há apenas um contato, retorna ele
	if len(contacts) == 1 {
		return &contacts[0]
	}

	// Para números brasileiros com 2 contatos, tenta fazer merge se configurado
	if len(contacts) == 2 && s.config.MergeBrazilContacts && strings.HasPrefix(query, "+55") {
		return s.handleBrazilianContacts(contacts, query)
	}

	// Obtém variações do número de telefone
	phoneNumbers := s.getPhoneNumberVariations(query)
	searchableFields := []string{"phone_number"}

	// Procura pelo número mais longo (com 9º dígito para números brasileiros)
	longestPhone := ""
	for _, number := range phoneNumbers {
		if len(number) > len(longestPhone) {
			longestPhone = number
		}
	}

	// Procura contato com o número mais longo
	for _, contact := range contacts {
		if contact.PhoneNumber == longestPhone {
			return &contact
		}
	}

	// Procura por correspondência em qualquer campo pesquisável
	for _, contact := range contacts {
		for _, field := range searchableFields {
			var fieldValue string
			switch field {
			case "phone_number":
				fieldValue = contact.PhoneNumber
			}

			for _, phoneNumber := range phoneNumbers {
				if fieldValue == phoneNumber {
					return &contact
				}
			}
		}
	}

	// Retorna o primeiro contato se não encontrou correspondência exata
	return &contacts[0]
}

// handleBrazilianContacts lida com contatos brasileiros (com e sem 9º dígito)
func (s *Service) handleBrazilianContacts(contacts []Contact, query string) *Contact {
	// Procura contato com número de 14 dígitos (com 9º dígito)
	var contactWith9 *Contact
	var contactWithout9 *Contact

	for i := range contacts {
		contact := &contacts[i]
		if len(contact.PhoneNumber) == 14 {
			contactWith9 = contact
		} else if len(contact.PhoneNumber) == 13 {
			contactWithout9 = contact
		}
	}

	// Se encontrou ambos, prefere o com 9º dígito (mais atual)
	if contactWith9 != nil {
		// TODO: Implementar merge de contatos se necessário
		// Por enquanto, retorna o contato com 9º dígito
		return contactWith9
	}

	// Se não encontrou com 9º dígito, retorna sem 9º dígito
	if contactWithout9 != nil {
		return contactWithout9
	}

	// Fallback: retorna o primeiro contato
	return &contacts[0]
}

// getPhoneNumberVariations retorna variações do número de telefone (especialmente para números brasileiros)
func (s *Service) getPhoneNumberVariations(phoneNumber string) []string {
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

// findOrCreateConversation encontra ou cria uma conversa
func (s *Service) findOrCreateConversation(ctx context.Context, contact *Contact) (*Conversation, error) {
	cacheKey := fmt.Sprintf("conversation:%d:%d", contact.ID, s.inbox.ID)
	
	// Verifica cache primeiro
	if cached, exists := s.getFromCache(cacheKey); exists {
		if conversation, ok := cached.(*Conversation); ok {
			return conversation, nil
		}
	}

	// Busca conversas existentes do contato
	conversations, err := s.client.ListContactConversations(ctx, contact.ID)
	if err != nil {
		s.logger.Error("Failed to list contact conversations", "error", err, "contactID", contact.ID)
	}

	// Procura por conversa aberta na inbox atual
	for _, conv := range conversations {
		if conv.InboxID == s.inbox.ID {
			if s.config.ReopenConversation || conv.Status != string(ConversationStatusResolved) {
				s.setCache(cacheKey, &conv)
				return &conv, nil
			}
		}
	}

	// Cria nova conversa
	req := ConversationCreateRequest{
		ContactID: strconv.Itoa(contact.ID),
		InboxID:   strconv.Itoa(s.inbox.ID),
	}

	if s.config.ConversationPending {
		req.Status = string(ConversationStatusPending)
	}

	conversation, err := s.client.CreateConversation(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	s.setCache(cacheKey, conversation)
	s.logger.Info("Created new conversation", "contactID", contact.ID, "conversationID", conversation.ID)
	return conversation, nil
}

// ProcessWhatsAppMessage processa uma mensagem do WhatsApp e envia para o Chatwoot
func (s *Service) ProcessWhatsAppMessage(ctx context.Context, msg *WhatsAppMessage) error {
	s.logger.Info("📱 PROCESSING WHATSAPP MESSAGE FOR CHATWOOT",
		"message_id", msg.ID,
		"from", msg.From,
		"body", msg.Body,
		"from_me", msg.FromMe,
		"push_name", msg.PushName,
		"chat_name", msg.ChatName,
		"timestamp", msg.Timestamp)

	if !s.config.Enabled {
		s.logger.Info("⏭️ Chatwoot integration disabled, skipping message")
		return nil
	}

	// Verifica se deve ignorar esta mensagem
	if s.shouldIgnoreMessage(msg) {
		s.logger.Info("⏭️ Message ignored by filter rules")
		return nil
	}

	phoneNumber := s.extractPhoneNumber(msg.From)
	isGroup := strings.Contains(msg.From, "@g.us")

	s.logger.Info("📞 EXTRACTED PHONE AND GROUP INFO",
		"phone_number", phoneNumber,
		"is_group", isGroup,
		"original_from", msg.From)
	
	var contactName string
	if isGroup {
		contactName = fmt.Sprintf("%s (GROUP)", msg.ChatName)
	} else {
		contactName = msg.PushName
		if contactName == "" {
			contactName = phoneNumber
		}
	}

	s.logger.Info("👤 CREATING/FINDING CONTACT",
		"contact_name", contactName,
		"phone_number", phoneNumber,
		"is_group", isGroup)

	// Encontra ou cria contato
	contact, err := s.findOrCreateContact(ctx, phoneNumber, contactName, "", isGroup)
	if err != nil {
		s.logger.Error("❌ FAILED TO FIND/CREATE CONTACT", "error", err)
		return fmt.Errorf("failed to find or create contact: %w", err)
	}

	s.logger.Info("✅ CONTACT FOUND/CREATED",
		"contact_id", contact.ID,
		"contact_name", contact.Name,
		"contact_phone", contact.PhoneNumber)

	// Encontra ou cria conversa
	s.logger.Info("💬 CREATING/FINDING CONVERSATION", "contact_id", contact.ID)
	conversation, err := s.findOrCreateConversation(ctx, contact)
	if err != nil {
		s.logger.Error("❌ FAILED TO FIND/CREATE CONVERSATION", "error", err)
		return fmt.Errorf("failed to find or create conversation: %w", err)
	}

	s.logger.Info("✅ CONVERSATION FOUND/CREATED",
		"conversation_id", conversation.ID,
		"inbox_id", conversation.InboxID,
		"status", conversation.Status)

	// Determina tipo de mensagem (0 = incoming, 1 = outgoing)
	messageType := 0 // incoming
	if msg.FromMe {
		messageType = 1 // outgoing
	}

	s.logger.Info("📝 MESSAGE TYPE DETERMINED",
		"message_type", messageType,
		"from_me", msg.FromMe)

	// Determina tipo de conteúdo baseado no tipo de mensagem
	contentType := s.getContentTypeFromWhatsAppMessage(msg)

	// Processa conteúdo da mensagem baseado no tipo
	content := s.formatMessageContentByType(msg, isGroup)

	s.logger.Info("📝 FORMATTED MESSAGE CONTENT",
		"original_body", msg.Body,
		"formatted_content", content,
		"content_type", contentType,
		"message_type_whatsapp", msg.Type,
		"is_group", isGroup)

	// Processa anexos de mídia
	mediaAttributes := s.processMediaAttachment(ctx, msg)

	// Cria mensagem no Chatwoot
	msgReq := MessageCreateRequest{
		Content:     content,
		MessageType: messageType,
		SourceID:    fmt.Sprintf("WAID:%s", msg.ID),
	}

	// Define atributos de conteúdo
	if len(mediaAttributes) > 0 || contentType != string(ContentTypeText) {
		if msgReq.ContentAttributes == nil {
			msgReq.ContentAttributes = make(map[string]interface{})
		}

		// Adiciona tipo de conteúdo
		msgReq.ContentAttributes["content_type"] = contentType

		// Adiciona atributos de mídia
		for key, value := range mediaAttributes {
			msgReq.ContentAttributes[key] = value
		}
	}

	// Adiciona atributos de contexto se for resposta
	if msg.QuotedMessageID != "" {
		msgReq.ContentAttributes = map[string]interface{}{
			"in_reply_to_external_id": msg.QuotedMessageID,
		}
		s.logger.Info("📎 REPLY CONTEXT ADDED", "quoted_message_id", msg.QuotedMessageID)
	}

	// Verifica se é mídia
	isMediaMessage := msg.Type == "audio" || msg.Type == "ptt" || msg.Type == "image" || msg.Type == "video" || msg.Type == "document" || msg.Type == "sticker"

	s.logger.Info("🚀 CREATING MESSAGE IN CHATWOOT",
		"conversation_id", conversation.ID,
		"content", content,
		"message_type", messageType,
		"source_id", msgReq.SourceID,
		"has_media", isMediaMessage,
		"media_type", msg.Type)

	var chatwootMsg *Message
	var msgErr error

	// Se há mídia disponível, envia como anexo
	if msg.Type == "audio" || msg.Type == "ptt" || msg.Type == "image" || msg.Type == "video" || msg.Type == "document" || msg.Type == "sticker" {
		s.logger.Info("📎 SENDING MEDIA MESSAGE TO CHATWOOT",
			"media_url", msg.MediaURL,
			"mime_type", msg.MimeType,
			"type", msg.Type)

		chatwootMsg, msgErr = s.sendMediaToChatwoot(ctx, conversation.ID, msg, messageType, msgReq.SourceID)
		if msgErr != nil {
			s.logger.Error("❌ FAILED TO SEND MEDIA TO CHATWOOT",
				"error", msgErr,
				"conversation_id", conversation.ID,
				"media_url", msg.MediaURL)
			// Fallback: envia como mensagem de texto
			s.logger.Info("🔄 FALLBACK: SENDING AS TEXT MESSAGE")
			chatwootMsg, msgErr = s.client.CreateMessage(ctx, conversation.ID, msgReq)
		}
	} else {
		// Envia como mensagem de texto normal
		chatwootMsg, msgErr = s.client.CreateMessage(ctx, conversation.ID, msgReq)
	}

	if msgErr != nil {
		s.logger.Error("❌ FAILED TO CREATE MESSAGE IN CHATWOOT",
			"error", msgErr,
			"conversation_id", conversation.ID,
			"content", content)
		return fmt.Errorf("failed to create message in chatwoot: %w", msgErr)
	}

	s.logger.Info("✅ MESSAGE SENT TO CHATWOOT SUCCESSFULLY",
		"chatwoot_message_id", chatwootMsg.ID,
		"conversation_id", conversation.ID,
		"source_id", msgReq.SourceID,
		"content", content,
		"has_media", isMediaMessage,
		"media_type", msg.Type)

	return nil
}

// shouldIgnoreMessage verifica se a mensagem deve ser ignorada
func (s *Service) shouldIgnoreMessage(msg *WhatsAppMessage) bool {
	// Ignora mensagens de status
	if strings.Contains(msg.From, "status@broadcast") {
		return true
	}

	// Verifica JIDs ignorados
	for _, ignoreJid := range s.config.IgnoreJids {
		if ignoreJid == "@g.us" && strings.Contains(msg.From, "@g.us") {
			return true
		}
		if ignoreJid == "@s.whatsapp.net" && strings.Contains(msg.From, "@s.whatsapp.net") {
			return true
		}
		if ignoreJid == msg.From {
			return true
		}
	}

	return false
}

// extractPhoneNumber extrai o número de telefone do JID
func (s *Service) extractPhoneNumber(jid string) string {
	// Remove sufixos como @s.whatsapp.net ou @g.us
	parts := strings.Split(jid, "@")
	if len(parts) > 0 {
		// Remove possível sufixo de timestamp (ex: :1234567890)
		phoneNumber := regexp.MustCompile(`:\d+`).ReplaceAllString(parts[0], "")
		return phoneNumber
	}
	return jid
}

// formatMessageContent formata o conteúdo da mensagem para o Chatwoot
func (s *Service) formatMessageContent(msg *WhatsAppMessage, isGroup bool) string {
	content := msg.Body

	// Para grupos, adiciona informações do participante
	if isGroup && !msg.FromMe {
		participantPhone := s.extractPhoneNumber(msg.Participant)
		participantName := msg.PushName
		if participantName == "" {
			participantName = participantPhone
		}
		
		// Formata número brasileiro se aplicável
		formattedPhone := s.formatBrazilianPhone(participantPhone)
		content = fmt.Sprintf("**%s - %s:**\n\n%s", formattedPhone, participantName, content)
	}

	// Converte formatação do WhatsApp para Markdown
	content = s.convertWhatsAppFormatting(content)

	return content
}

// formatBrazilianPhone formata número brasileiro
func (s *Service) formatBrazilianPhone(phone string) string {
	if len(phone) == 13 && strings.HasPrefix(phone, "55") {
		// Formato: 5511999999999 -> +55 (11) 99999-9999
		return fmt.Sprintf("+%s (%s) %s-%s", 
			phone[:2], phone[2:4], phone[4:9], phone[9:])
	}
	return fmt.Sprintf("+%s", phone)
}

// convertWhatsAppFormatting converte formatação do WhatsApp para Markdown
func (s *Service) convertWhatsAppFormatting(text string) string {
	// *texto* -> **texto** (negrito)
	text = regexp.MustCompile(`\*([^*\n]+)\*`).ReplaceAllString(text, "**$1**")
	
	// _texto_ -> *texto* (itálico)
	text = regexp.MustCompile(`_([^_\n]+)_`).ReplaceAllString(text, "*$1*")
	
	// ~texto~ -> ~~texto~~ (riscado)
	text = regexp.MustCompile(`~([^~\n]+)~`).ReplaceAllString(text, "~~$1~~")
	
	return text
}

// ProcessWebhook processa webhooks recebidos do Chatwoot
func (s *Service) ProcessWebhook(ctx context.Context, payload *WebhookPayload) error {
	if !s.config.Enabled {
		return nil
	}

	s.logger.Info("🔄 Processing Chatwoot webhook",
		"event", payload.Event,
		"message_type", payload.MessageType,
		"content", payload.Content,
		"source_id", payload.SourceID,
		"private", payload.Private)

	// Log detalhado do payload completo
	s.logger.Info("📋 WEBHOOK PAYLOAD DETAILS",
		"contact_exists", payload.Contact != nil,
		"conversation_exists", payload.Conversation != nil)

	if payload.Contact != nil {
		s.logger.Info("👤 CONTACT DETAILS",
			"id", payload.Contact.ID,
			"name", payload.Contact.Name,
			"phone", payload.Contact.PhoneNumber,
			"identifier", payload.Contact.Identifier)
	}

	if payload.Conversation != nil {
		s.logger.Info("💬 CONVERSATION DETAILS",
			"id", payload.Conversation.ID,
			"inbox_id", payload.Conversation.InboxID,
			"status", payload.Conversation.Status)
	}

	// Processa apenas mensagens de saída (outgoing) de agentes
	if payload.Event == "message_created" && payload.MessageType == "outgoing" {
		s.logger.Info("✅ Processing outgoing message for WhatsApp")
		return s.processOutgoingMessage(ctx, payload)
	}

	s.logger.Info("⏭️ Skipping webhook - not an outgoing message",
		"event", payload.Event,
		"message_type", payload.MessageType)
	return nil
}

// processOutgoingMessage processa mensagens de saída do Chatwoot para WhatsApp
func (s *Service) processOutgoingMessage(ctx context.Context, payload *WebhookPayload) error {
	s.logger.Info("📤 PROCESSING OUTGOING MESSAGE")

	if payload.Conversation == nil || payload.Contact == nil {
		s.logger.Error("❌ Missing conversation or contact in webhook payload",
			"conversation_exists", payload.Conversation != nil,
			"contact_exists", payload.Contact != nil)
		return fmt.Errorf("missing conversation or contact in webhook payload")
	}

	// Extrai número de telefone do contato
	phoneNumber := s.extractPhoneFromContact(payload.Contact)
	s.logger.Info("📞 EXTRACTED PHONE NUMBER",
		"raw_phone", phoneNumber,
		"contact_phone", payload.Contact.PhoneNumber,
		"contact_identifier", payload.Contact.Identifier)

	if phoneNumber == "" {
		s.logger.Error("❌ Could not extract phone number from contact")
		return fmt.Errorf("could not extract phone number from contact")
	}

	// Remove o prefixo + do número se presente
	cleanPhoneNumber := strings.TrimPrefix(phoneNumber, "+")

	s.logger.Info("📨 SENDING MESSAGE TO WHATSAPP",
		"to", cleanPhoneNumber,
		"content", payload.Content,
		"session_id", s.sessionID,
		"whatsapp_service_available", s.whatsappService != nil,
		"conversationID", payload.Conversation.ID,
		"messageType", payload.MessageType)

	// Envia mensagem via WhatsApp usando o serviço zpmeow
	if s.whatsappService != nil {
		s.logger.Info("🚀 CALLING WHATSAPP SERVICE",
			"session_id", s.sessionID,
			"to", cleanPhoneNumber,
			"content", payload.Content,
			"content_type", payload.ContentType,
			"has_attachments", len(payload.Attachments) > 0)

		// Determina o tipo de mensagem e envia adequadamente
		var err error

		// Verifica se há anexos na mensagem
		if len(payload.Attachments) > 0 {
			_, err = s.sendAttachmentMessage(ctx, cleanPhoneNumber, payload)
		} else if payload.Content != "" {
			_, err = s.whatsappService.SendTextMessage(ctx, s.sessionID, cleanPhoneNumber, payload.Content)
		} else {
			s.logger.Warn("⚠️ MESSAGE WITH NO CONTENT OR ATTACHMENTS",
				"to", cleanPhoneNumber,
				"content_type", payload.ContentType)
			return fmt.Errorf("message has no content or attachments")
		}

		if err != nil {
			s.logger.Error("❌ FAILED TO SEND MESSAGE TO WHATSAPP",
				"error", err,
				"to", cleanPhoneNumber,
				"session_id", s.sessionID,
				"content_type", payload.ContentType)
			return fmt.Errorf("failed to send message to WhatsApp: %w", err)
		}

		// Log de sucesso
		s.logger.Info("✅ MESSAGE SENT TO WHATSAPP SUCCESSFULLY",
			"to", cleanPhoneNumber,
			"content", payload.Content,
			"session_id", s.sessionID,
			"content_type", payload.ContentType,
			"has_attachments", len(payload.Attachments) > 0)

		return nil
	}

	s.logger.Warn("⚠️ WHATSAPP SERVICE NOT AVAILABLE",
		"to", cleanPhoneNumber,
		"session_id", s.sessionID,
		"content", payload.Content)

	return nil
}

// extractPhoneFromContact extrai número de telefone do contato
func (s *Service) extractPhoneFromContact(contact *Contact) string {
	if contact.PhoneNumber != "" {
		// Remove + e espaços
		phone := strings.ReplaceAll(contact.PhoneNumber, "+", "")
		phone = strings.ReplaceAll(phone, " ", "")
		phone = strings.ReplaceAll(phone, "(", "")
		phone = strings.ReplaceAll(phone, ")", "")
		phone = strings.ReplaceAll(phone, "-", "")
		return phone
	}

	if contact.Identifier != "" {
		return s.extractPhoneNumber(contact.Identifier)
	}

	return ""
}

// sendAttachmentMessage envia mensagens com anexos
func (s *Service) sendAttachmentMessage(ctx context.Context, phoneNumber string, payload *WebhookPayload) (interface{}, error) {
	if len(payload.Attachments) == 0 {
		return nil, fmt.Errorf("no attachments found")
	}

	attachment := payload.Attachments[0] // Processa o primeiro anexo

	s.logger.Info("📎 SENDING ATTACHMENT MESSAGE",
		"file_type", attachment.FileType,
		"data_url", attachment.DataURL,
		"to", phoneNumber)

	// Download do arquivo
	fileData, err := s.downloadAttachment(ctx, attachment.DataURL)
	if err != nil {
		s.logger.Error("❌ FAILED TO DOWNLOAD ATTACHMENT",
			"error", err,
			"data_url", attachment.DataURL)
		return nil, fmt.Errorf("failed to download attachment: %w", err)
	}

	// Envia baseado no tipo de arquivo
	switch attachment.FileType {
	case "audio":
		s.logger.Info("🎵 SENDING AUDIO MESSAGE",
			"size", len(fileData),
			"to", phoneNumber)
		// Use SendAudioMessageWithPTT com PTT=true para mensagens de voz do Chatwoot
		return s.whatsappService.SendAudioMessageWithPTT(ctx, s.sessionID, phoneNumber, fileData, "audio/ogg", true)
	case "image":
		s.logger.Info("📷 SENDING IMAGE MESSAGE",
			"size", len(fileData),
			"to", phoneNumber)
		return s.whatsappService.SendImageMessage(ctx, s.sessionID, phoneNumber, fileData, payload.Content, "image/jpeg")
	case "video":
		s.logger.Info("🎬 SENDING VIDEO MESSAGE",
			"size", len(fileData),
			"to", phoneNumber)
		return s.whatsappService.SendVideoMessage(ctx, s.sessionID, phoneNumber, fileData, payload.Content, "video/mp4")
	case "document":
		s.logger.Info("📄 SENDING DOCUMENT MESSAGE",
			"size", len(fileData),
			"to", phoneNumber)
		fileName := attachment.Fallback
		if fileName == "" {
			fileName = "document"
		}
		return s.whatsappService.SendDocumentMessage(ctx, s.sessionID, phoneNumber, fileData, fileName, payload.Content, "application/octet-stream")
	default:
		s.logger.Warn("⚠️ UNSUPPORTED ATTACHMENT TYPE, SENDING AS TEXT",
			"file_type", attachment.FileType)
		return s.whatsappService.SendTextMessage(ctx, s.sessionID, phoneNumber,
			fmt.Sprintf("📎 Anexo enviado (%s)\n\n%s", attachment.FileType, payload.Content))
	}
}

// downloadAttachment faz download do anexo do Chatwoot
func (s *Service) downloadAttachment(ctx context.Context, dataURL string) ([]byte, error) {
	if dataURL == "" {
		return nil, fmt.Errorf("data URL is empty")
	}

	s.logger.Info("⬇️ DOWNLOADING ATTACHMENT", "url", dataURL)

	req, err := http.NewRequestWithContext(ctx, "GET", dataURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download file: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read file data: %w", err)
	}

	s.logger.Info("✅ ATTACHMENT DOWNLOADED",
		"size", len(data),
		"url", dataURL)

	return data, nil
}

// sendMediaMessage envia mensagens de mídia baseado no tipo de conteúdo (mantido para compatibilidade)
func (s *Service) sendMediaMessage(ctx context.Context, phoneNumber string, payload *WebhookPayload) (interface{}, error) {
	s.logger.Info("📎 SENDING MEDIA MESSAGE",
		"content_type", payload.ContentType,
		"to", phoneNumber)

	// Por enquanto, para tipos de mídia que não conseguimos processar diretamente,
	// enviamos como mensagem de texto com descrição
	switch payload.ContentType {
	case "image":
		// TODO: Implementar download e envio de imagem
		return s.whatsappService.SendTextMessage(ctx, s.sessionID, phoneNumber,
			fmt.Sprintf("📷 Imagem enviada\n\n%s", payload.Content))
	case "audio":
		// TODO: Implementar download e envio de áudio
		return s.whatsappService.SendTextMessage(ctx, s.sessionID, phoneNumber,
			fmt.Sprintf("🎵 Áudio enviado\n\n%s", payload.Content))
	case "video":
		// TODO: Implementar download e envio de vídeo
		return s.whatsappService.SendTextMessage(ctx, s.sessionID, phoneNumber,
			fmt.Sprintf("🎬 Vídeo enviado\n\n%s", payload.Content))
	case "file":
		// TODO: Implementar download e envio de documento
		return s.whatsappService.SendTextMessage(ctx, s.sessionID, phoneNumber,
			fmt.Sprintf("📄 Documento enviado\n\n%s", payload.Content))
	case "sticker":
		// TODO: Implementar download e envio de sticker
		return s.whatsappService.SendTextMessage(ctx, s.sessionID, phoneNumber,
			fmt.Sprintf("🎭 Sticker enviado\n\n%s", payload.Content))
	case "location":
		// TODO: Implementar envio de localização
		return s.whatsappService.SendTextMessage(ctx, s.sessionID, phoneNumber,
			fmt.Sprintf("📍 Localização enviada\n\n%s", payload.Content))
	case "contact":
		// TODO: Implementar envio de contato
		return s.whatsappService.SendTextMessage(ctx, s.sessionID, phoneNumber,
			fmt.Sprintf("👤 Contato enviado\n\n%s", payload.Content))
	default:
		s.logger.Warn("Unknown content type, sending as text",
			"content_type", payload.ContentType)
		return s.whatsappService.SendTextMessage(ctx, s.sessionID, phoneNumber, payload.Content)
	}
}
