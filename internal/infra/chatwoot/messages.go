package chatwoot

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"

	"zpmeow/internal/application/ports"
)

// MessageService gerencia operações relacionadas a mensagens
type MessageService struct {
	client          *Client
	logger          *slog.Logger
	whatsappService ports.WhatsAppService
	errorHandler    ports.ChatwootErrorHandler
	validator       ports.ChatwootValidator
	contentMapper   *ContentTypeMapper
	sessionID       string
	adapter         *MessageAdapter
}

// NewMessageService cria um novo serviço de mensagens
func NewMessageService(client *Client, logger *slog.Logger, whatsappService ports.WhatsAppService, sessionID string) *MessageService {
	return &MessageService{
		client:          client,
		logger:          logger,
		whatsappService: whatsappService,
		errorHandler:    NewErrorHandler(),
		validator:       NewValidator(),
		contentMapper:   NewContentTypeMapper(),
		sessionID:       sessionID,
		adapter:         NewMessageAdapter(),
	}
}

// ProcessIncomingMessage processa mensagem recebida do WhatsApp
func (ms *MessageService) ProcessIncomingMessage(ctx context.Context, msg *ports.WhatsAppMessage, conversationID int) (*ports.MessageResponse, error) {
	// Valida a mensagem
	if err := ms.validator.ValidateMessageContent(msg.Body, msg.Type); err != nil {
		return nil, ms.errorHandler.HandleMessageError(err, msg.ID)
	}

	// Determina tipo de mensagem
	messageType := ms.getMessageType(msg.FromMe)

	// Determina tipo de conteúdo
	contentType := ms.getContentType(msg)

	// Formata conteúdo da mensagem
	content := ms.formatMessageContent(msg)

	ms.logger.Info("Processing incoming message",
		"message_id", msg.ID,
		"content_type", contentType,
		"message_type", messageType,
		"conversation_id", conversationID,
		"from_me", msg.FromMe)

	// Cria request da mensagem
	msgReq := MessageCreateRequest{
		Content:     content,
		MessageType: messageType,
		SourceID:    fmt.Sprintf("WAID:%s", msg.ID),
	}

	// Processa anexos de mídia se existirem
	if ms.hasMediaContent(msg) {
		return ms.processMediaMessage(ctx, msg, conversationID, msgReq)
	}

	// Envia como mensagem de texto
	message, err := ms.client.CreateMessage(ctx, conversationID, msgReq)
	if err != nil {
		return nil, ms.errorHandler.HandleMessageError(err, msg.ID)
	}

	ms.logger.Info("Successfully sent text message to Chatwoot",
		"chatwoot_message_id", message.ID,
		"conversation_id", conversationID,
		"content", content)

	// Converte para tipo da interface
	return ms.adapter.ToPortsMessage(message), nil
}

// ProcessOutgoingMessage processa mensagem enviada do Chatwoot para WhatsApp
func (ms *MessageService) ProcessOutgoingMessage(ctx context.Context, payload *ports.WebhookPayload) error {
	// Extrai informações da mensagem
	messageInfo := ms.extractMessageInfo(payload)

	// Extrai número do telefone do contato
	phoneNumber := ms.extractPhoneNumber(payload)
	if phoneNumber == "" {
		return fmt.Errorf("could not extract phone number from payload")
	}

	ms.logger.Info("Processing outgoing message",
		"to", phoneNumber,
		"content", messageInfo.Content,
		"has_attachments", len(messageInfo.Attachments) > 0)

	// Envia mensagem via WhatsApp
	return ms.sendToWhatsApp(ctx, phoneNumber, messageInfo)
}

// getMessageType determina o tipo de mensagem baseado na origem
func (ms *MessageService) getMessageType(fromMe bool) int {
	if fromMe {
		return MessageTypeOutgoing
	}
	return MessageTypeIncoming
}

// getContentType determina o tipo de conteúdo baseado no tipo de mensagem do WhatsApp
func (ms *MessageService) getContentType(msg *ports.WhatsAppMessage) string {
	return ms.contentMapper.WhatsAppToChatwoot(msg.Type)
}

// formatMessageContent formata o conteúdo da mensagem
func (ms *MessageService) formatMessageContent(msg *ports.WhatsAppMessage) string {
	switch msg.Type {
	case "text", "extendedTextMessage":
		return msg.Body
	case "image", "imageMessage":
		if msg.Caption != "" {
			return msg.Caption
		}
		return "📷 Imagem"
	case "audio", "audioMessage":
		return "🎵 Áudio"
	case "ptt":
		return "🎤 Áudio"
	case "video", "videoMessage":
		if msg.Caption != "" {
			return msg.Caption
		}
		return "🎥 Vídeo"
	case "document", "documentMessage":
		if msg.FileName != "" {
			return fmt.Sprintf("📄 %s", msg.FileName)
		}
		return "📄 Documento"
	case "sticker", "stickerMessage":
		return "🎭 Sticker"
	case "location", "locationMessage":
		return "📍 Localização"
	case "contact", "contactMessage":
		return "👤 Contato"
	default:
		if msg.Body != "" {
			return msg.Body
		}
		return fmt.Sprintf("Mensagem do tipo: %s", msg.Type)
	}
}

// hasMediaContent verifica se a mensagem tem conteúdo de mídia
func (ms *MessageService) hasMediaContent(msg *ports.WhatsAppMessage) bool {
	mediaTypes := []string{"audio", "ptt", "image", "video", "document", "sticker"}
	for _, mediaType := range mediaTypes {
		if msg.Type == mediaType {
			return true
		}
	}
	return false
}

// processMediaMessage processa mensagem com mídia
func (ms *MessageService) processMediaMessage(ctx context.Context, msg *ports.WhatsAppMessage, conversationID int, msgReq MessageCreateRequest) (*ports.MessageResponse, error) {
	ms.logger.Info("Processing media message",
		"media_url", msg.MediaURL,
		"mime_type", msg.MimeType,
		"type", msg.Type)

	// Tenta enviar como anexo de mídia
	chatwootMsg, err := ms.sendMediaToChatwoot(ctx, conversationID, msg, msgReq.SourceID)
	if err != nil {
		ms.logger.Error("Failed to send media to Chatwoot, falling back to text",
			"error", err,
			"conversation_id", conversationID,
			"media_url", msg.MediaURL)

		// Fallback: envia como mensagem de texto
		message, fallbackErr := ms.client.CreateMessage(ctx, conversationID, msgReq)
		if fallbackErr != nil {
			return nil, ms.errorHandler.HandleMessageError(fallbackErr, msg.ID)
		}
		return ms.adapter.ToPortsMessage(message), nil
	}

	return chatwootMsg, nil
}

// sendMediaToChatwoot envia mídia para o Chatwoot
func (ms *MessageService) sendMediaToChatwoot(ctx context.Context, conversationID int, msg *ports.WhatsAppMessage, sourceID string) (*ports.MessageResponse, error) {
	if ms.whatsappService == nil {
		return nil, fmt.Errorf("WhatsApp service not available")
	}

	// Download da mídia usando o serviço WhatsApp
	ms.logger.Info("Downloading media from WhatsApp",
		"message_id", msg.ID,
		"type", msg.Type,
		"mime_type", msg.MimeType)

	mediaData, _, err := ms.whatsappService.DownloadMedia(ctx, ms.sessionID, msg.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to download media: %w", err)
	}

	// Determina nome do arquivo
	fileName := ms.generateFileName(msg)

	// Converte []byte para io.Reader
	mediaReader := bytes.NewReader(mediaData)

	// Envia para Chatwoot
	message, err := ms.client.CreateMessageWithAttachment(ctx, conversationID, msg.Caption, "0", mediaReader, fileName, sourceID)
	if err != nil {
		return nil, err
	}

	// Converte para tipo da interface
	return ms.adapter.ToPortsMessage(message), nil
}

// generateFileName gera nome do arquivo baseado no tipo de mídia
func (ms *MessageService) generateFileName(msg *ports.WhatsAppMessage) string {
	if msg.FileName != "" {
		return msg.FileName
	}

	extension := ms.getFileExtension(msg.Type, msg.MimeType)
	return fmt.Sprintf("%s%s", msg.ID, extension)
}

// getFileExtension retorna a extensão do arquivo baseada no tipo
func (ms *MessageService) getFileExtension(msgType, mimeType string) string {
	switch msgType {
	case "image":
		if strings.Contains(mimeType, "jpeg") {
			return ".jpg"
		}
		return ".png"
	case "audio", "ptt":
		return ".ogg"
	case "video":
		return ".mp4"
	case "document":
		return ".pdf"
	case "sticker":
		return ".webp"
	default:
		return ""
	}
}

// MessageInfo contém informações extraídas de uma mensagem
type MessageInfo struct {
	Content     string
	ContentType string
	Attachments []interface{}
}

// extractMessageInfo extrai informações da mensagem do payload
func (ms *MessageService) extractMessageInfo(payload *ports.WebhookPayload) *MessageInfo {
	info := &MessageInfo{}

	if payload.Message != nil {
		if content, ok := payload.Message["content"].(string); ok {
			info.Content = content
		}
		if contentType, ok := payload.Message["content_type"].(string); ok {
			info.ContentType = contentType
		}
		if attachments, ok := payload.Message["attachments"].([]interface{}); ok {
			info.Attachments = attachments
		}
	}

	return info
}

// extractPhoneNumber extrai o número do telefone do payload
func (ms *MessageService) extractPhoneNumber(payload *ports.WebhookPayload) string {
	if payload.Conversation == nil {
		return ""
	}

	// Tenta extrair do meta.sender
	if meta, ok := payload.Conversation["meta"].(map[string]interface{}); ok {
		if sender, ok := meta["sender"].(map[string]interface{}); ok {
			if phone, ok := sender["phone_number"].(string); ok {
				return ms.cleanPhoneNumber(phone)
			}
		}
	}

	return ""
}

// cleanPhoneNumber limpa o número do telefone
func (ms *MessageService) cleanPhoneNumber(phone string) string {
	// Remove caracteres não numéricos exceto +
	cleaned := strings.ReplaceAll(phone, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")

	// Remove + do início se existir
	cleaned = strings.TrimPrefix(cleaned, "+")

	return cleaned
}

// sendToWhatsApp envia mensagem para WhatsApp
func (ms *MessageService) sendToWhatsApp(ctx context.Context, phoneNumber string, messageInfo *MessageInfo) error {
	if ms.whatsappService == nil {
		ms.logger.Warn("WhatsApp service not available", "to", phoneNumber)
		return nil
	}

	// Verifica se há anexos
	if len(messageInfo.Attachments) > 0 {
		return ms.sendAttachmentMessage(ctx, phoneNumber, messageInfo.Content, messageInfo.Attachments)
	}

	// Envia mensagem de texto
	if messageInfo.Content != "" {
		_, err := ms.whatsappService.SendTextMessage(ctx, ms.sessionID, phoneNumber, messageInfo.Content)
		return err
	}

	return fmt.Errorf("message has no content or attachments")
}

// sendAttachmentMessage envia mensagem com anexo
func (ms *MessageService) sendAttachmentMessage(ctx context.Context, phoneNumber, content string, attachments []interface{}) error {
	// Implementação simplificada - processa apenas o primeiro anexo
	if len(attachments) == 0 {
		return fmt.Errorf("no attachments provided")
	}

	// Por enquanto, envia como mensagem de texto com informação do anexo
	attachmentInfo := "📎 Anexo recebido"
	if content != "" {
		content = fmt.Sprintf("%s\n\n%s", content, attachmentInfo)
	} else {
		content = attachmentInfo
	}

	_, err := ms.whatsappService.SendTextMessage(ctx, ms.sessionID, phoneNumber, content)
	return err
}

// ContentTypeMapper mapeia tipos de conteúdo
type ContentTypeMapper struct{}

// NewContentTypeMapper cria um novo mapeador de tipos de conteúdo
func NewContentTypeMapper() *ContentTypeMapper {
	return &ContentTypeMapper{}
}

// WhatsAppToChatwoot converte tipo de conteúdo do WhatsApp para Chatwoot
func (ctm *ContentTypeMapper) WhatsAppToChatwoot(whatsappType string) string {
	switch whatsappType {
	case "text", "extendedTextMessage":
		return "text"
	case "image", "imageMessage":
		return "image"
	case "audio", "audioMessage", "ptt":
		return "audio"
	case "video", "videoMessage":
		return "video"
	case "document", "documentMessage":
		return "file"
	case "sticker", "stickerMessage":
		return "sticker"
	case "location", "locationMessage":
		return "location"
	case "contact", "contactMessage", "contactsArrayMessage":
		return "contact"
	default:
		return "text"
	}
}
