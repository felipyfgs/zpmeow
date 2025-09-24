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
	"strings"
	"sync"
	"time"

	"zpmeow/internal/application/ports"
	"zpmeow/internal/infra/database/models"
	"zpmeow/internal/infra/database/repository"
)

// Service representa o serviço de integração Chatwoot
type Service struct {
	client             *Client
	config             *ChatwootConfig
	logger             *slog.Logger
	cache              map[string]interface{}
	cacheMutex         sync.RWMutex
	inbox              *Inbox
	whatsappService    ports.WhatsAppService
	sessionID          string
	messageRepo        *repository.MessageRepository
	zpCwRepo           *repository.ZpCwMessageRepository
	chatRepo           *repository.ChatRepository
	conversationMapper *ConversationMapper
}

// NewService cria uma nova instância do serviço Chatwoot
func NewService(config *ChatwootConfig, logger *slog.Logger, whatsappService ports.WhatsAppService, sessionID string, messageRepo *repository.MessageRepository, zpCwRepo *repository.ZpCwMessageRepository, chatRepo *repository.ChatRepository) (*Service, error) {
	if !config.IsActive {
		return nil, fmt.Errorf("chatwoot integration is disabled")
	}

	client := NewClient(config.URL, config.Token, config.AccountID, nil)

	service := &Service{
		client:          client,
		config:          config,
		logger:          logger,
		cache:           make(map[string]interface{}),
		cacheMutex:      sync.RWMutex{},
		whatsappService: whatsappService,
		sessionID:       sessionID,
		messageRepo:     messageRepo,
		zpCwRepo:        zpCwRepo,
		chatRepo:        chatRepo,
	}

	// Inicializa a inbox
	if err := service.initializeInbox(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize inbox: %w", err)
	}

	// Inicializa o conversation mapper após ter a inbox
	service.conversationMapper = NewConversationMapper(
		client,
		chatRepo,
		logger,
		sessionID,
		service.inbox.ID,
	)

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
func (s *Service) _formatMessageContentByType(msg *WhatsAppMessage, isGroup bool) string {
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
func (s *Service) processMediaAttachment(_ context.Context, msg *WhatsAppMessage) map[string]interface{} {
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
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			s.logger.Error("Failed to close response body", "error", closeErr)
		}
	}()

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
	cacheKeyBuilder := NewCacheKeyBuilder()
	cacheKey := cacheKeyBuilder.ContactKey(phoneNumber)

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

	s.logger.Info("🚀 [CHATWOOT SERVICE DEBUG] Calling CreateContact API", "request", req)
	contact, err := s.client.CreateContact(ctx, req)
	if err != nil {
		s.logger.Error("❌ [CHATWOOT SERVICE DEBUG] CreateContact failed", "error", err, "request", req)
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
	s.logger.Info("✅ [CHATWOOT SERVICE DEBUG] Successfully created contact",
		"name", name, "phone", phoneNumber, "id", contact.ID,
		"identifier", contact.Identifier, "contact_obj", contact)
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
func (s *Service) handleBrazilianContacts(contacts []Contact, _ string) *Contact {
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

// findOrCreateConversationWithEvolutionStrategy implementa estratégia Evolution API melhorada
func (s *Service) findOrCreateConversationWithEvolutionStrategy(ctx context.Context, contact *Contact) (*Conversation, error) {
	s.logger.Info("🔍 [EVOLUTION STRATEGY] Starting Evolution API strategy",
		"contact_id", contact.ID,
		"inbox_id", s.inbox.ID)

	// Lista conversas do contato
	conversations, err := s.client.ListContactConversations(ctx, contact.ID)
	if err != nil {
		s.logger.Error("❌ [EVOLUTION STRATEGY] Failed to list conversations", "error", err)
		return nil, fmt.Errorf("failed to list conversations: %w", err)
	}

	s.logger.Info("🔍 [EVOLUTION STRATEGY] Analyzing conversations",
		"total_conversations", len(conversations),
		"contact_id", contact.ID,
		"target_inbox_id", s.inbox.ID)

	// ESTRATÉGIA EVOLUTION: Encontra melhor conversa baseada na última atividade
	var bestConversation *Conversation
	var latestActivity float64 = 0

	for _, conv := range conversations {
		// Só considera conversas da inbox correta
		if conv.InboxID == s.inbox.ID {
			// Só considera conversas não resolvidas (ou reabre se configurado)
			if s.config.ReopenConversation || conv.Status != string(ConversationStatusResolved) {

				// Converte LastActivityAt para float64
				var convActivity float64
				if activityVal, ok := conv.LastActivityAt.(float64); ok {
					convActivity = activityVal
				} else if activityVal, ok := conv.LastActivityAt.(int); ok {
					convActivity = float64(activityVal)
				}

				s.logger.Info("🔍 [EVOLUTION STRATEGY] Evaluating conversation",
					"conversation_id", conv.ID,
					"status", conv.Status,
					"last_activity", convActivity,
					"inbox_id", conv.InboxID)

				// Escolhe conversa com atividade mais recente
				if convActivity > latestActivity {
					latestActivity = convActivity
					bestConversation = &conv
					s.logger.Info("✅ [EVOLUTION STRATEGY] Found better conversation",
						"conversation_id", conv.ID,
						"last_activity", convActivity,
						"status", conv.Status)
				} else if bestConversation == nil {
					// Fallback: primeira conversa válida encontrada
					bestConversation = &conv
					s.logger.Info("📝 [EVOLUTION STRATEGY] Using fallback conversation",
						"conversation_id", conv.ID,
						"status", conv.Status)
				}
			} else {
				s.logger.Info("⏭️ [EVOLUTION STRATEGY] Skipping resolved conversation",
					"conversation_id", conv.ID,
					"status", conv.Status)
			}
		} else {
			s.logger.Info("⏭️ [EVOLUTION STRATEGY] Skipping conversation from different inbox",
				"conversation_id", conv.ID,
				"conversation_inbox_id", conv.InboxID,
				"target_inbox_id", s.inbox.ID)
		}
	}

	if bestConversation != nil {
		s.logger.Info("🎯 [EVOLUTION STRATEGY] Selected best conversation",
			"conversation_id", bestConversation.ID,
			"status", bestConversation.Status,
			"last_activity", latestActivity,
			"inbox_id", bestConversation.InboxID)
		return bestConversation, nil
	}

	// Se não encontrou nenhuma conversa, cria nova
	s.logger.Info("📝 [EVOLUTION STRATEGY] No suitable conversation found, creating new one",
		"contact_id", contact.ID,
		"inbox_id", s.inbox.ID)

	conversation, err := s.client.CreateConversation(ctx, ConversationCreateRequest{
		ContactID: contact.ID,
		InboxID:   s.inbox.ID,
	})
	if err != nil {
		s.logger.Error("❌ [EVOLUTION STRATEGY] Failed to create conversation", "error", err)
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	s.logger.Info("✅ [EVOLUTION STRATEGY] Created new conversation",
		"conversation_id", conversation.ID,
		"contact_id", contact.ID,
		"inbox_id", s.inbox.ID)

	return conversation, nil
}

// saveConversationMappingAsync salva mapeamento de conversa de forma assíncrona
func (s *Service) saveConversationMappingAsync(_ context.Context, chatJid, phoneNumber string, contactID, conversationID int) {
	s.logger.Info("💾 [ASYNC MAPPING] Starting async conversation mapping save",
		"chat_jid", chatJid,
		"phone_number", phoneNumber,
		"contact_id", contactID,
		"conversation_id", conversationID)

	// Executa em goroutine separada para não bloquear processamento
	go func() {
		// Cria contexto com timeout para operação assíncrona
		asyncCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Tenta salvar mapeamento via ConversationMapper
		err := s.conversationMapper.SaveConversationMapping(asyncCtx, chatJid, phoneNumber, contactID, conversationID)
		if err != nil {
			s.logger.Error("❌ [ASYNC MAPPING] Failed to save conversation mapping",
				"error", err,
				"chat_jid", chatJid,
				"conversation_id", conversationID)
		} else {
			s.logger.Info("✅ [ASYNC MAPPING] Successfully saved conversation mapping",
				"chat_jid", chatJid,
				"phone_number", phoneNumber,
				"contact_id", contactID,
				"conversation_id", conversationID)
		}
	}()
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

	if !s.config.IsActive {
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

	// 🎯 ESTRATÉGIA HÍBRIDA: Mapeamento Persistente + Evolution API Fallback
	s.logger.Info("🔍 [HYBRID MAPPING] Starting hybrid conversation mapping",
		"contact_id", contact.ID,
		"chat_jid", msg.From,
		"phone_number", phoneNumber)

	// FASE 1: Tenta ConversationMapper (usa colunas zpChats)
	mapping, err := s.conversationMapper.GetOrCreateConversationMapping(ctx, msg.From, phoneNumber)
	if err != nil {
		s.logger.Error("❌ [HYBRID MAPPING] ConversationMapper failed, using Evolution fallback", "error", err)
		mapping = nil // Força fallback
	}

	var conversationID int

	if mapping != nil && mapping.IsValid {
		// SUCESSO: Usa mapeamento persistente
		conversationID = mapping.ConversationID
		s.logger.Info("✅ [HYBRID MAPPING] Using persistent mapping",
			"chat_id", mapping.ChatID,
			"contact_id", mapping.ContactID,
			"conversation_id", conversationID,
			"source", "persistent_mapping")
	} else {
		// FALLBACK: Usa Evolution API strategy
		s.logger.Info("🔄 [HYBRID MAPPING] Using Evolution API fallback",
			"reason", "no_persistent_mapping")

		conversation, err := s.findOrCreateConversationWithEvolutionStrategy(ctx, contact)
		if err != nil {
			s.logger.Error("❌ [HYBRID MAPPING] Evolution fallback failed", "error", err)
			return fmt.Errorf("failed to find conversation with hybrid strategy: %w", err)
		}

		conversationID = conversation.ID

		s.logger.Info("✅ [HYBRID MAPPING] Evolution fallback successful",
			"conversation_id", conversationID,
			"inbox_id", conversation.InboxID,
			"status", conversation.Status,
			"source", "evolution_fallback")

		// IMPORTANTE: Salva mapeamento para próximas mensagens
		go s.saveConversationMappingAsync(ctx, msg.From, phoneNumber, contact.ID, conversationID)
	}

	s.logger.Info("🎯 [HYBRID MAPPING] Final conversation selected",
		"conversation_id", conversationID,
		"contact_id", contact.ID)

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
	content := s._formatMessageContentByType(msg, isGroup)

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
		"conversation_id", conversationID,
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

		chatwootMsg, msgErr = s.sendMediaToChatwoot(ctx, conversationID, msg, messageType, msgReq.SourceID)
		if msgErr != nil {
			s.logger.Error("❌ FAILED TO SEND MEDIA TO CHATWOOT",
				"error", msgErr,
				"conversation_id", conversationID,
				"media_url", msg.MediaURL)
			// Fallback: envia como mensagem de texto
			s.logger.Info("🔄 FALLBACK: SENDING AS TEXT MESSAGE")
			chatwootMsg, msgErr = s.client.CreateMessage(ctx, conversationID, msgReq)
		}
	} else {
		// Envia como mensagem de texto normal
		chatwootMsg, msgErr = s.client.CreateMessage(ctx, conversationID, msgReq)
	}

	if msgErr != nil {
		s.logger.Error("❌ FAILED TO CREATE MESSAGE IN CHATWOOT",
			"error", msgErr,
			"conversation_id", conversationID,
			"content", content)
		return fmt.Errorf("failed to create message in chatwoot: %w", msgErr)
	}

	s.logger.Info("✅ MESSAGE SENT TO CHATWOOT SUCCESSFULLY",
		"chatwoot_message_id", chatwootMsg.ID,
		"conversation_id", conversationID,
		"source_id", msgReq.SourceID,
		"content", content,
		"has_media", isMediaMessage,
		"media_type", msg.Type)

	// Salvar relação zpmeow-chatwoot (criar objeto conversation temporário)
	tempConversation := &Conversation{ID: conversationID}
	if err := s.saveZpCwRelation(ctx, msg, chatwootMsg, tempConversation); err != nil {
		s.logger.Error("❌ FAILED TO SAVE ZP-CW RELATION",
			"error", err,
			"zpmeow_message_id", msg.ID,
			"chatwoot_message_id", chatwootMsg.ID)
		// Não retorna erro para não quebrar o fluxo principal
	}

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

// ProcessWebhook processa webhooks recebidos do Chatwoot
func (s *Service) ProcessWebhook(ctx context.Context, payload *WebhookPayload) error {
	if !s.config.IsActive {
		return nil
	}

	// Extract message fields from the Message map
	messageType := ""
	content := ""
	sourceID := ""
	private := false

	if payload.Message != nil {
		if mt, ok := payload.Message["message_type"].(string); ok {
			messageType = mt
		}
		if c, ok := payload.Message["content"].(string); ok {
			content = c
		}
		if sid, ok := payload.Message["source_id"].(string); ok {
			sourceID = sid
		}
		if p, ok := payload.Message["private"].(bool); ok {
			private = p
		}
	}

	s.logger.Info("🔄 Processing Chatwoot webhook",
		"event", payload.Event,
		"message_type", messageType,
		"content", content,
		"source_id", sourceID,
		"private", private)

	// Log detalhado do payload completo
	s.logger.Info("📋 WEBHOOK PAYLOAD DETAILS",
		"contact_exists", payload.Contact != nil,
		"conversation_exists", payload.Conversation != nil)

	if payload.Contact != nil {
		contactID := 0
		contactName := ""
		contactPhone := ""
		contactIdentifier := ""

		if id, ok := payload.Contact["id"].(float64); ok {
			contactID = int(id)
		}
		if name, ok := payload.Contact["name"].(string); ok {
			contactName = name
		}
		if phone, ok := payload.Contact["phone_number"].(string); ok {
			contactPhone = phone
		}
		if identifier, ok := payload.Contact["identifier"].(string); ok {
			contactIdentifier = identifier
		}

		s.logger.Info("👤 CONTACT DETAILS",
			"id", contactID,
			"name", contactName,
			"phone", contactPhone,
			"identifier", contactIdentifier)
	}

	if payload.Conversation != nil {
		conversationID := 0
		inboxID := 0

		if id, ok := payload.Conversation["id"].(float64); ok {
			conversationID = int(id)
		}
		if iid, ok := payload.Conversation["inbox_id"].(float64); ok {
			inboxID = int(iid)
		}

		conversationStatus := ""
		if status, ok := payload.Conversation["status"].(string); ok {
			conversationStatus = status
		}

		s.logger.Info("💬 CONVERSATION DETAILS",
			"id", conversationID,
			"inbox_id", inboxID,
			"status", conversationStatus)
	}

	// Processa apenas mensagens de saída (outgoing) de agentes
	if payload.Event == "message_created" && messageType == "outgoing" {
		s.logger.Info("✅ Processing outgoing message for WhatsApp")
		return s.processOutgoingMessage(ctx, payload)
	}

	s.logger.Info("⏭️ Skipping webhook - not an outgoing message",
		"event", payload.Event,
		"message_type", messageType)
	return nil
}

// processOutgoingMessage processa mensagens de saída do Chatwoot para WhatsApp
func (s *Service) processOutgoingMessage(ctx context.Context, payload *WebhookPayload) error {
	s.logger.Info("📤 PROCESSING OUTGOING MESSAGE")

	// Extract message fields from the Message map
	messageType := ""
	content := ""
	contentType := ""
	attachments := []interface{}{}

	if payload.Message != nil {
		if mt, ok := payload.Message["message_type"].(string); ok {
			messageType = mt
		}
		if c, ok := payload.Message["content"].(string); ok {
			content = c
		}
		if ct, ok := payload.Message["content_type"].(string); ok {
			contentType = ct
		}
		if att, ok := payload.Message["attachments"].([]interface{}); ok {
			attachments = att
		}
	}

	if payload.Conversation == nil || payload.Contact == nil {
		s.logger.Error("❌ Missing conversation or contact in webhook payload",
			"conversation_exists", payload.Conversation != nil,
			"contact_exists", payload.Contact != nil)
		return fmt.Errorf("missing conversation or contact in webhook payload")
	}

	// Extrai número de telefone do contato
	phoneNumber := s.extractPhoneFromContactMap(payload.Contact)

	contactPhone := ""
	contactIdentifier := ""
	if phone, ok := payload.Contact["phone_number"].(string); ok {
		contactPhone = phone
	}
	if identifier, ok := payload.Contact["identifier"].(string); ok {
		contactIdentifier = identifier
	}

	s.logger.Info("📞 EXTRACTED PHONE NUMBER",
		"raw_phone", phoneNumber,
		"contact_phone", contactPhone,
		"contact_identifier", contactIdentifier)

	if phoneNumber == "" {
		s.logger.Error("❌ Could not extract phone number from contact")
		return fmt.Errorf("could not extract phone number from contact")
	}

	// Remove o prefixo + do número se presente
	cleanPhoneNumber := strings.TrimPrefix(phoneNumber, "+")

	// Extract conversation ID
	conversationID := 0
	if payload.Conversation != nil {
		if id, ok := payload.Conversation["id"].(float64); ok {
			conversationID = int(id)
		}
	}

	s.logger.Info("📨 SENDING MESSAGE TO WHATSAPP",
		"to", cleanPhoneNumber,
		"content", content,
		"session_id", s.sessionID,
		"whatsapp_service_available", s.whatsappService != nil,
		"conversationID", conversationID,
		"messageType", messageType)

	// Envia mensagem via WhatsApp usando o serviço zpmeow
	if s.whatsappService != nil {
		// contentType and attachments already extracted above

		s.logger.Info("🚀 CALLING WHATSAPP SERVICE",
			"session_id", s.sessionID,
			"to", cleanPhoneNumber,
			"content", content,
			"content_type", contentType,
			"has_attachments", len(attachments) > 0)

		// Determina o tipo de mensagem e envia adequadamente
		var err error

		// Verifica se há anexos na mensagem
		if len(attachments) > 0 {
			_, err = s.sendAttachmentMessage(ctx, cleanPhoneNumber, content, attachments)
		} else if content != "" {
			_, err = s.whatsappService.SendTextMessage(ctx, s.sessionID, cleanPhoneNumber, content)
		} else {
			s.logger.Warn("⚠️ MESSAGE WITH NO CONTENT OR ATTACHMENTS",
				"to", cleanPhoneNumber,
				"content_type", contentType)
			return fmt.Errorf("message has no content or attachments")
		}

		if err != nil {
			s.logger.Error("❌ FAILED TO SEND MESSAGE TO WHATSAPP",
				"error", err,
				"to", cleanPhoneNumber,
				"session_id", s.sessionID,
				"content_type", contentType)
			return fmt.Errorf("failed to send message to WhatsApp: %w", err)
		}

		// Log de sucesso
		s.logger.Info("✅ MESSAGE SENT TO WHATSAPP SUCCESSFULLY",
			"to", cleanPhoneNumber,
			"content", content,
			"session_id", s.sessionID,
			"content_type", contentType,
			"has_attachments", len(attachments) > 0)

		return nil
	}

	s.logger.Warn("⚠️ WHATSAPP SERVICE NOT AVAILABLE",
		"to", cleanPhoneNumber,
		"session_id", s.sessionID,
		"content", content)

	return nil
}

// extractPhoneFromContactMap extrai número de telefone do contato a partir de um map
func (s *Service) extractPhoneFromContactMap(contact map[string]interface{}) string {
	phoneNumber := ""
	identifier := ""

	if phone, ok := contact["phone_number"].(string); ok {
		phoneNumber = phone
	}
	if id, ok := contact["identifier"].(string); ok {
		identifier = id
	}

	if phoneNumber != "" {
		// Remove + e espaços
		phone := strings.ReplaceAll(phoneNumber, "+", "")
		phone = strings.ReplaceAll(phone, " ", "")
		phone = strings.ReplaceAll(phone, "(", "")
		phone = strings.ReplaceAll(phone, ")", "")
		phone = strings.ReplaceAll(phone, "-", "")
		return phone
	}

	if identifier != "" {
		return s.extractPhoneNumber(identifier)
	}

	return ""
}

// sendAttachmentMessage envia mensagens com anexos
func (s *Service) sendAttachmentMessage(ctx context.Context, phoneNumber, content string, attachments []interface{}) (interface{}, error) {
	if len(attachments) == 0 {
		return nil, fmt.Errorf("no attachments found")
	}

	s.logger.Info("📎 PROCESSING MULTIPLE ATTACHMENTS",
		"total_attachments", len(attachments),
		"to", phoneNumber)

	var results []interface{}
	var lastResult interface{}

	// Processa todos os anexos
	for i, attachment := range attachments {
		// Convert interface{} to map[string]interface{}
		attachmentMap, ok := attachment.(map[string]interface{})
		if !ok {
			s.logger.Error("❌ INVALID ATTACHMENT FORMAT", "attachment_index", i+1)
			continue
		}

		fileType := ""
		dataURL := ""
		if ft, ok := attachmentMap["file_type"].(string); ok {
			fileType = ft
		}
		if du, ok := attachmentMap["data_url"].(string); ok {
			dataURL = du
		}

		s.logger.Info("📎 SENDING ATTACHMENT MESSAGE",
			"attachment_index", i+1,
			"total_attachments", len(attachments),
			"file_type", fileType,
			"data_url", dataURL,
			"to", phoneNumber)

		// Download do arquivo
		fileData, err := s.downloadAttachment(ctx, dataURL)
		if err != nil {
			s.logger.Error("❌ FAILED TO DOWNLOAD ATTACHMENT",
				"error", err,
				"attachment_index", i+1,
				"data_url", dataURL)
			continue // Continua com o próximo anexo
		}

		// Determina o tipo real do arquivo baseado na URL ou extensão
		actualFileType := s.determineActualFileTypeFromMap(attachmentMap)

		// Envia baseado no tipo de arquivo
		var result interface{}
		switch actualFileType {
		case "audio":
			s.logger.Info("🎵 SENDING AUDIO MESSAGE",
				"size", len(fileData),
				"to", phoneNumber,
				"attachment_index", i+1,
				"original_type", fileType,
				"detected_type", actualFileType)
			// Use SendAudioMessageWithPTT com PTT=true para mensagens de voz do Chatwoot
			result, err = s.whatsappService.SendAudioMessageWithPTT(ctx, s.sessionID, phoneNumber, fileData, "audio/ogg", true)
		case "image":
			s.logger.Info("📷 SENDING IMAGE MESSAGE",
				"size", len(fileData),
				"to", phoneNumber,
				"attachment_index", i+1,
				"original_type", fileType,
				"detected_type", actualFileType)
			result, err = s.whatsappService.SendImageMessage(ctx, s.sessionID, phoneNumber, fileData, content, "image/jpeg")
		case "video":
			s.logger.Info("🎬 SENDING VIDEO MESSAGE",
				"size", len(fileData),
				"to", phoneNumber,
				"attachment_index", i+1,
				"original_type", fileType,
				"detected_type", actualFileType)
			result, err = s.whatsappService.SendVideoMessage(ctx, s.sessionID, phoneNumber, fileData, content, "video/mp4")
		case "document", "file":
			s.logger.Info("📄 SENDING DOCUMENT MESSAGE",
				"size", len(fileData),
				"to", phoneNumber,
				"attachment_index", i+1,
				"original_type", fileType,
				"detected_type", actualFileType)
			fileName := s.getFileNameFromAttachmentMap(attachmentMap)
			result, err = s.whatsappService.SendDocumentMessage(ctx, s.sessionID, phoneNumber, fileData, fileName, content, "application/octet-stream")
		default:
			// Para qualquer outro tipo, tenta enviar como documento
			s.logger.Info("📎 SENDING UNKNOWN TYPE AS DOCUMENT",
				"size", len(fileData),
				"to", phoneNumber,
				"attachment_index", i+1,
				"original_type", fileType,
				"detected_type", actualFileType)
			fileName := s.getFileNameFromAttachmentMap(attachmentMap)
			result, err = s.whatsappService.SendDocumentMessage(ctx, s.sessionID, phoneNumber, fileData, fileName, content, "application/octet-stream")
		}

		if err != nil {
			s.logger.Error("❌ FAILED TO SEND ATTACHMENT",
				"error", err,
				"attachment_index", i+1,
				"file_type", fileType)
			continue // Continua com o próximo anexo
		}

		results = append(results, result)
		lastResult = result

		s.logger.Info("✅ ATTACHMENT SENT SUCCESSFULLY",
			"attachment_index", i+1,
			"total_attachments", len(attachments),
			"file_type", fileType)
	}

	s.logger.Info("📎 FINISHED PROCESSING ALL ATTACHMENTS",
		"total_attachments", len(attachments),
		"successful_sends", len(results),
		"to", phoneNumber)

	// Retorna o último resultado para compatibilidade
	if lastResult != nil {
		return lastResult, nil
	}

	return nil, fmt.Errorf("failed to send any attachments")
}

// determineActualFileTypeFromMap determina o tipo real do arquivo baseado na URL, extensão ou MIME type
func (s *Service) determineActualFileTypeFromMap(attachmentMap map[string]interface{}) string {
	// Extrai apenas a URL dos dados que é o que precisamos
	var dataURL string
	if url, ok := attachmentMap["data_url"].(string); ok {
		dataURL = url
	}

	detector := NewFileTypeDetector()
	return detector.DetectFileType(dataURL)
}

// getFileNameFromAttachmentMap extrai o nome do arquivo do anexo a partir de um map
func (s *Service) getFileNameFromAttachmentMap(attachmentMap map[string]interface{}) string {
	// Extrai apenas a URL dos dados que é o que precisamos
	var dataURL string
	if url, ok := attachmentMap["data_url"].(string); ok {
		dataURL = url
	}

	detector := NewFileTypeDetector()
	return detector.extractFileName(dataURL)
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
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			s.logger.Error("Failed to close response body", "error", closeErr)
		}
	}()

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

// saveZpCwRelation salva a relação entre mensagem zpmeow e Chatwoot
func (s *Service) saveZpCwRelation(ctx context.Context, whatsappMsg *WhatsAppMessage, chatwootMsg *Message, conversation *Conversation) error {
	if s.messageRepo == nil || s.zpCwRepo == nil {
		s.logger.Warn("🔗 [ZP-CW RELATION] Repositories not available, skipping relation save")
		return nil
	}

	s.logger.Info("🔗 [ZP-CW RELATION] Starting to save zpmeow-chatwoot relation",
		"msgId", whatsappMsg.ID,
		"chatwoot_message_id", chatwootMsg.ID,
		"conversation_id", conversation.ID)

	// Buscar mensagem zpmeow pelo WhatsApp message ID
	zpmeowMessage, err := s.messageRepo.GetMessageByWhatsAppID(ctx, s.sessionID, whatsappMsg.ID)
	if err != nil {
		s.logger.Error("🔗 [ZP-CW RELATION ERROR] Failed to find zpmeow message",
			"msgId", whatsappMsg.ID,
			"error", err)
		return fmt.Errorf("failed to find zpmeow message: %w", err)
	}

	if zpmeowMessage == nil {
		s.logger.Error("🔗 [ZP-CW RELATION ERROR] Zpmeow message not found",
			"msgId", whatsappMsg.ID)
		return fmt.Errorf("zpmeow message not found for WhatsApp ID: %s", whatsappMsg.ID)
	}

	// Determinar direção da mensagem
	direction := "incoming"
	if whatsappMsg.FromMe {
		direction = "outgoing"
	}

	// AccountID removido - não mais necessário na relação

	// Criar relação (campos otimizados)
	relation := &models.ZpCwMessageModel{
		SessionId:      s.sessionID,
		MsgId:          zpmeowMessage.ID,
		ChatwootMsgId:  int64(chatwootMsg.ID),
		ChatwootConvId: int64(conversation.ID),
		Direction:      direction,
		SyncStatus:     "synced",
		SourceId:       &whatsappMsg.ID,
		Metadata:       models.JSONB{},
	}

	if err := s.zpCwRepo.CreateRelation(ctx, relation); err != nil {
		s.logger.Error("🔗 [ZP-CW RELATION ERROR] Failed to create relation",
			"zpmeow_message_id", zpmeowMessage.ID,
			"chatwoot_message_id", chatwootMsg.ID,
			"error", err)
		return fmt.Errorf("failed to create zp-cw relation: %w", err)
	}

	s.logger.Info("🔗 [ZP-CW RELATION SUCCESS] Successfully saved zpmeow-chatwoot relation",
		"relation_id", relation.ID,
		"zpmeow_message_id", zpmeowMessage.ID,
		"chatwoot_message_id", chatwootMsg.ID,
		"direction", direction)

	return nil
}
