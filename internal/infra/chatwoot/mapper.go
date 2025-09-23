package chatwoot

import (
	"fmt"
	"regexp"
	"strings"
)

// MessageMapper mapeia mensagens entre WhatsApp e Chatwoot
type MessageMapper struct {
	config *ChatwootConfig
}

// NewMessageMapper cria uma nova instância do mapper
func NewMessageMapper(config *ChatwootConfig) *MessageMapper {
	return &MessageMapper{
		config: config,
	}
}

// WhatsAppToChatwoot converte uma mensagem do WhatsApp para formato Chatwoot
func (m *MessageMapper) WhatsAppToChatwoot(msg *WhatsAppMessage) (*MessageCreateRequest, error) {
	// 0 = incoming, 1 = outgoing
	messageType := 0 // incoming
	if msg.FromMe {
		messageType = 1 // outgoing
	}

	content := m.formatWhatsAppContent(msg)

	req := &MessageCreateRequest{
		Content:  content,
		MsgType:  messageType,
		SourceID: fmt.Sprintf("WAID:%s", msg.ID),
	}

	// Adiciona atributos de contexto para mensagens citadas
	if msg.QuotedMessageID != "" {
		req.ContentAttributes = map[string]interface{}{
			"in_reply_to_external_id": msg.QuotedMessageID,
		}
	}

	return req, nil
}

// formatWhatsAppContent formata o conteúdo da mensagem do WhatsApp
func (m *MessageMapper) formatWhatsAppContent(msg *WhatsAppMessage) string {
	content := msg.Body

	// Processa diferentes tipos de mensagem
	switch msg.Type {
	case "text":
		content = m.formatTextMessage(msg)
	case "image":
		content = m.formatMediaMessage(msg, "📷 Imagem")
	case "video":
		content = m.formatMediaMessage(msg, "🎥 Vídeo")
	case "audio":
		content = m.formatMediaMessage(msg, "🎵 Áudio")
	case "document":
		content = m.formatDocumentMessage(msg)
	case "sticker":
		content = "🎭 Sticker"
	case "location":
		content = m.formatLocationMessage(msg)
	case "contact":
		content = m.formatContactMessage(msg)
	case "list":
		content = m.formatListMessage(msg)
	case "button":
		content = m.formatButtonMessage(msg)
	case "reaction":
		content = m.formatReactionMessage(msg)
	default:
		if content == "" {
			content = fmt.Sprintf("📎 Mensagem do tipo: %s", msg.Type)
		}
	}

	// Para grupos, adiciona informações do participante
	if strings.Contains(msg.From, "@g.us") && !msg.FromMe {
		participantInfo := m.formatParticipantInfo(msg)
		content = fmt.Sprintf("%s\n\n%s", participantInfo, content)
	}

	// Converte formatação do WhatsApp para Markdown
	content = m.convertWhatsAppFormatting(content)

	return content
}

// formatTextMessage formata mensagem de texto
func (m *MessageMapper) formatTextMessage(msg *WhatsAppMessage) string {
	content := msg.Body

	// Processa links com preview
	if msg.LinkPreview != nil {
		content = m.formatLinkPreview(msg, content)
	}

	return content
}

// formatMediaMessage formata mensagem de mídia
func (m *MessageMapper) formatMediaMessage(msg *WhatsAppMessage, mediaType string) string {
	content := mediaType

	if msg.Caption != "" {
		content = fmt.Sprintf("%s\n\n%s", content, msg.Caption)
	}

	if msg.MediaURL != "" {
		content = fmt.Sprintf("%s\n\n🔗 [Visualizar mídia](%s)", content, msg.MediaURL)
	}

	return content
}

// formatDocumentMessage formata mensagem de documento
func (m *MessageMapper) formatDocumentMessage(msg *WhatsAppMessage) string {
	content := "📄 Documento"

	if msg.FileName != "" {
		content = fmt.Sprintf("📄 **%s**", msg.FileName)
	}

	if msg.Caption != "" {
		content = fmt.Sprintf("%s\n\n%s", content, msg.Caption)
	}

	if msg.MediaURL != "" {
		content = fmt.Sprintf("%s\n\n🔗 [Baixar documento](%s)", content, msg.MediaURL)
	}

	return content
}

// formatLocationMessage formata mensagem de localização
func (m *MessageMapper) formatLocationMessage(msg *WhatsAppMessage) string {
	if msg.Location == nil {
		return "📍 Localização"
	}

	content := "📍 **Localização:**\n\n"
	content += fmt.Sprintf("_Latitude:_ %f\n", msg.Location.Latitude)
	content += fmt.Sprintf("_Longitude:_ %f\n", msg.Location.Longitude)

	if msg.Location.Name != "" {
		content += fmt.Sprintf("_Nome:_ %s\n", msg.Location.Name)
	}

	if msg.Location.Address != "" {
		content += fmt.Sprintf("_Endereço:_ %s\n", msg.Location.Address)
	}

	// Adiciona link do Google Maps
	mapsURL := fmt.Sprintf("https://www.google.com/maps/search/?api=1&query=%f,%f",
		msg.Location.Latitude, msg.Location.Longitude)
	content += fmt.Sprintf("\n🗺️ [Ver no Google Maps](%s)", mapsURL)

	return content
}

// formatContactMessage formata mensagem de contato
func (m *MessageMapper) formatContactMessage(msg *WhatsAppMessage) string {
	if len(msg.Contacts) == 0 {
		return "👤 Contato"
	}

	content := "👤 **Contato(s):**\n\n"

	for i, contact := range msg.Contacts {
		if i > 0 {
			content += "\n---\n\n"
		}

		content += fmt.Sprintf("**%s**\n", contact.Name)

		for j, phone := range contact.Phones {
			content += fmt.Sprintf("📞 %s", phone.Number)
			if phone.Type != "" {
				content += fmt.Sprintf(" (%s)", phone.Type)
			}
			if j < len(contact.Phones)-1 {
				content += "\n"
			}
		}
	}

	return content
}

// formatListMessage formata mensagem de lista
func (m *MessageMapper) formatListMessage(msg *WhatsAppMessage) string {
	if msg.List == nil {
		return "📋 Lista"
	}

	content := "📋 **Lista:**\n\n"

	if msg.List.Title != "" {
		content += fmt.Sprintf("**%s**\n\n", msg.List.Title)
	}

	if msg.List.Description != "" {
		content += fmt.Sprintf("%s\n\n", msg.List.Description)
	}

	for i, section := range msg.List.Sections {
		if section.Title != "" {
			content += fmt.Sprintf("**%s**\n", section.Title)
		}

		for _, row := range section.Rows {
			content += fmt.Sprintf("• %s", row.Title)
			if row.Description != "" {
				content += fmt.Sprintf(" - %s", row.Description)
			}
			content += "\n"
		}

		if i < len(msg.List.Sections)-1 {
			content += "\n"
		}
	}

	return content
}

// formatButtonMessage formata mensagem com botões
func (m *MessageMapper) formatButtonMessage(msg *WhatsAppMessage) string {
	content := msg.Body

	if len(msg.Buttons) > 0 {
		content += "\n\n**Opções:**\n"
		for _, button := range msg.Buttons {
			content += fmt.Sprintf("• %s\n", button.Text)
		}
	}

	return content
}

// formatReactionMessage formata mensagem de reação
func (m *MessageMapper) formatReactionMessage(msg *WhatsAppMessage) string {
	if msg.Reaction == nil {
		return "👍 Reação"
	}

	return fmt.Sprintf("👍 Reagiu com: %s", msg.Reaction.Emoji)
}

// formatLinkPreview formata preview de link
func (m *MessageMapper) formatLinkPreview(msg *WhatsAppMessage, content string) string {
	if msg.LinkPreview == nil {
		return content
	}

	preview := "\n\n---\n"
	preview += "🔗 **Preview do Link:**\n"

	if msg.LinkPreview.Title != "" {
		preview += fmt.Sprintf("**%s**\n", msg.LinkPreview.Title)
	}

	if msg.LinkPreview.Description != "" {
		preview += fmt.Sprintf("%s\n", msg.LinkPreview.Description)
	}

	if msg.LinkPreview.URL != "" {
		preview += fmt.Sprintf("[%s](%s)", msg.LinkPreview.URL, msg.LinkPreview.URL)
	}

	return content + preview
}

// formatParticipantInfo formata informações do participante em grupos
func (m *MessageMapper) formatParticipantInfo(msg *WhatsAppMessage) string {
	participantPhone := extractPhoneNumber(msg.Participant)
	participantName := msg.PushName

	if participantName == "" {
		participantName = participantPhone
	}

	formattedPhone := formatBrazilianPhone(participantPhone)
	return fmt.Sprintf("**%s - %s:**", formattedPhone, participantName)
}

// convertWhatsAppFormatting converte formatação do WhatsApp para Markdown
func (m *MessageMapper) convertWhatsAppFormatting(text string) string {
	// *texto* -> **texto** (negrito)
	text = regexp.MustCompile(`\*([^*\n]+)\*`).ReplaceAllString(text, "**$1**")

	// _texto_ -> *texto* (itálico)
	text = regexp.MustCompile(`_([^_\n]+)_`).ReplaceAllString(text, "*$1*")

	// ~texto~ -> ~~texto~~ (riscado)
	text = regexp.MustCompile(`~([^~\n]+)~`).ReplaceAllString(text, "~~$1~~")

	// ```texto``` -> `texto` (código)
	text = regexp.MustCompile("```([^`]+)```").ReplaceAllString(text, "`$1`")

	return text
}

// ChatwootToWhatsApp converte mensagem do Chatwoot para formato WhatsApp
func (m *MessageMapper) ChatwootToWhatsApp(msg *Message) (*OutgoingMessage, error) {
	content := m.formatChatwootContent(msg)

	outMsg := &OutgoingMessage{
		To:      "", // Será preenchido pelo serviço
		Type:    "text",
		Content: content,
	}

	// Processa anexos se existirem
	if len(msg.Attachments) > 0 {
		attachment := msg.Attachments[0] // Pega o primeiro anexo
		outMsg.Type = m.getWhatsAppMediaType(attachment.FileType)
		outMsg.MediaURL = attachment.DataURL
		outMsg.Caption = content
		outMsg.FileName = attachment.Fallback
	}

	return outMsg, nil
}

// formatChatwootContent formata conteúdo do Chatwoot para WhatsApp
func (m *MessageMapper) formatChatwootContent(msg *Message) string {
	content := msg.Content

	// Remove formatação Markdown se necessário
	content = m.convertMarkdownToWhatsApp(content)

	// Adiciona assinatura se configurado
	if m.config.SignMsg && m.config.SignDelimiter != "" {
		content = fmt.Sprintf("%s%s", content, m.config.SignDelimiter)
	}

	return content
}

// convertMarkdownToWhatsApp converte Markdown para formatação WhatsApp
func (m *MessageMapper) convertMarkdownToWhatsApp(text string) string {
	// **texto** -> *texto* (negrito)
	text = regexp.MustCompile(`\*\*([^*]+)\*\*`).ReplaceAllString(text, "*$1*")

	// *texto* -> _texto_ (itálico)
	text = regexp.MustCompile(`\*([^*]+)\*`).ReplaceAllString(text, "_$1_")

	// ~~texto~~ -> ~texto~ (riscado)
	text = regexp.MustCompile(`~~([^~]+)~~`).ReplaceAllString(text, "~$1~")

	// Remove links markdown [texto](url) -> texto (url)
	text = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`).ReplaceAllString(text, "$1 ($2)")

	return text
}

// getWhatsAppMediaType mapeia tipo de arquivo para tipo de mídia WhatsApp
func (m *MessageMapper) getWhatsAppMediaType(fileType string) string {
	switch {
	case strings.HasPrefix(fileType, "image/"):
		return "image"
	case strings.HasPrefix(fileType, "video/"):
		return "video"
	case strings.HasPrefix(fileType, "audio/"):
		return "audio"
	default:
		return "document"
	}
}

// ContactMapper mapeia contatos entre WhatsApp e Chatwoot
type ContactMapper struct{}

// NewContactMapper cria uma nova instância do mapper de contatos
func NewContactMapper() *ContactMapper {
	return &ContactMapper{}
}

// WhatsAppToChatwoot converte contato do WhatsApp para formato Chatwoot
func (cm *ContactMapper) WhatsAppToChatwoot(contact *WhatsAppContact, inboxID int, isGroup bool) *ContactCreateRequest {
	req := &ContactCreateRequest{
		InboxID: inboxID,
		Name:    contact.Name,
	}

	if contact.ProfilePictureURL != "" {
		req.AvatarURL = contact.ProfilePictureURL
	}

	if !isGroup {
		req.PhoneNumber = fmt.Sprintf("+%s", contact.Phone)
		req.Identifier = fmt.Sprintf("%s@s.whatsapp.net", contact.Phone)
	} else {
		req.Identifier = contact.JID
	}

	return req
}

// Funções auxiliares

// extractPhoneNumber extrai número de telefone do JID
func extractPhoneNumber(jid string) string {
	parts := strings.Split(jid, "@")
	if len(parts) > 0 {
		phoneNumber := regexp.MustCompile(`:\d+`).ReplaceAllString(parts[0], "")
		return phoneNumber
	}
	return jid
}

// formatBrazilianPhone formata número brasileiro
func formatBrazilianPhone(phone string) string {
	if len(phone) == 13 && strings.HasPrefix(phone, "55") {
		return fmt.Sprintf("+%s (%s) %s-%s",
			phone[:2], phone[2:4], phone[4:9], phone[9:])
	}
	if len(phone) == 12 && strings.HasPrefix(phone, "55") {
		return fmt.Sprintf("+%s (%s) %s-%s",
			phone[:2], phone[2:4], phone[4:8], phone[8:])
	}
	return fmt.Sprintf("+%s", phone)
}
