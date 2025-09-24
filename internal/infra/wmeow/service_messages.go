package wmeow

import (
	"context"
	"fmt"

	"zpmeow/internal/application/ports"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/proto/waE2E"
	waTypes "go.mau.fi/whatsmeow/types"
)

// MessageSender methods - envio de mensagens

func (m *MeowService) SendTextMessage(ctx context.Context, sessionID, to, text string) (*whatsmeow.SendResponse, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	jid, err := waTypes.ParseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID %s: %w", to, err)
	}

	message := &waProto.Message{
		Conversation: &text,
	}

	resp, err := client.GetClient().SendMessage(ctx, jid, message)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	return &resp, nil
}

func (m *MeowService) SendImageMessage(ctx context.Context, sessionID, to string, data []byte, caption, mimeType string) (*whatsmeow.SendResponse, error) {
	// For now, send as text message
	return m.SendTextMessage(ctx, sessionID, to, caption)
}

func (m *MeowService) SendAudioMessage(ctx context.Context, sessionID, to string, data []byte, mimeType string) (*whatsmeow.SendResponse, error) {
	// For now, send as text message
	return m.SendTextMessage(ctx, sessionID, to, "Audio message")
}

func (m *MeowService) SendVideoMessage(ctx context.Context, sessionID, to string, data []byte, caption, mimeType string) (*whatsmeow.SendResponse, error) {
	// For now, send as text message
	return m.SendTextMessage(ctx, sessionID, to, caption)
}

func (m *MeowService) SendDocumentMessage(ctx context.Context, sessionID, to string, data []byte, filename, mimetype, caption string) (*whatsmeow.SendResponse, error) {
	// For now, send as text message
	return m.SendTextMessage(ctx, sessionID, to, "Document: "+filename)
}

func (m *MeowService) SendStickerMessage(ctx context.Context, sessionID, to string, data []byte, mimeType string) (*whatsmeow.SendResponse, error) {
	// For now, send as text message
	return m.SendTextMessage(ctx, sessionID, to, "Sticker")
}

func (m *MeowService) SendContactMessage(ctx context.Context, sessionID, to string, contacts []ports.ContactInfo) (*whatsmeow.SendResponse, error) {
	// For now, send as text message
	contactText := fmt.Sprintf("Contact: %d contacts", len(contacts))
	return m.SendTextMessage(ctx, sessionID, to, contactText)
}

func (m *MeowService) SendLocationMessage(ctx context.Context, sessionID, to string, latitude, longitude float64, name, address string) (*whatsmeow.SendResponse, error) {
	// For now, send as text message
	locationText := fmt.Sprintf("Location: %s at %f,%f", name, latitude, longitude)
	return m.SendTextMessage(ctx, sessionID, to, locationText)
}

func (m *MeowService) SendTemplateMessage(ctx context.Context, sessionID, to string, template map[string]interface{}) (*whatsmeow.SendResponse, error) {
	// For now, send as text message
	return m.SendTextMessage(ctx, sessionID, to, "Template message")
}

func (m *MeowService) SendButtonMessage(ctx context.Context, sessionID, to, text string, buttons []ports.ButtonData) (*whatsmeow.SendResponse, error) {
	// For now, send as text message
	buttonText := fmt.Sprintf("%s (with %d buttons)", text, len(buttons))
	return m.SendTextMessage(ctx, sessionID, to, buttonText)
}

func (m *MeowService) SendListMessage(ctx context.Context, sessionID, to, text, buttonText, footer, title string, sections []ports.ListSection) (*whatsmeow.SendResponse, error) {
	// For now, send as text message
	listText := fmt.Sprintf("%s (with %d sections)", text, len(sections))
	return m.SendTextMessage(ctx, sessionID, to, listText)
}

func (m *MeowService) SendPollMessage(ctx context.Context, sessionID, to, question string, options []string, maxSelections int) (*whatsmeow.SendResponse, error) {
	// For now, send as text message
	pollText := fmt.Sprintf("Poll: %s (with %d options, max %d selections)", question, len(options), maxSelections)
	return m.SendTextMessage(ctx, sessionID, to, pollText)
}

// ForwardMessage is already implemented in service_actions.go

func (m *MeowService) SendReaction(ctx context.Context, sessionID, chatJID, messageID, emoji string) error {
	// For now, just log
	m.logger.Debugf("SendReaction: %s to message %s in chat %s for session %s", emoji, messageID, chatJID, sessionID)
	return nil
}

func (m *MeowService) RemoveReaction(ctx context.Context, sessionID, chatJID, messageID string) error {
	// For now, just log
	m.logger.Debugf("RemoveReaction: message %s in chat %s for session %s", messageID, chatJID, sessionID)
	return nil
}

func (m *MeowService) MarkAsRead(ctx context.Context, sessionID, chatJID string, messageIDs []string) error {
	// For now, just log
	m.logger.Debugf("MarkAsRead: %d messages in chat %s for session %s", len(messageIDs), chatJID, sessionID)
	return nil
}

// DeleteMessage is already implemented in service_actions.go

func (m *MeowService) GetMessage(ctx context.Context, sessionID, chatJID, messageID string) (*ports.ChatMessage, error) {
	// For now, return mock message
	m.logger.Debugf("GetMessage: %s in chat %s for session %s", messageID, chatJID, sessionID)

	return &ports.ChatMessage{
		ID:      messageID,
		ChatJID: chatJID,
		Type:    "text",
		Content: "Mock message content",
	}, nil
}

func (m *MeowService) GetMessages(ctx context.Context, sessionID, chatJID string, limit int, before string) ([]ports.ChatMessage, error) {
	// For now, return empty list
	m.logger.Debugf("GetMessages: chat %s for session %s (returning empty for now)", chatJID, sessionID)
	return []ports.ChatMessage{}, nil
}

func (m *MeowService) SearchMessages(ctx context.Context, sessionID, query string, chatJID string, limit int) ([]ports.ChatMessage, error) {
	// For now, return empty list
	m.logger.Debugf("SearchMessages: query '%s' in chat %s for session %s (returning empty for now)", query, chatJID, sessionID)
	return []ports.ChatMessage{}, nil
}

func (m *MeowService) GetMessageHistory(ctx context.Context, sessionID, chatJID string, limit int, offset int) ([]ports.ChatMessage, error) {
	// For now, return empty list
	m.logger.Debugf("GetMessageHistory: chat %s for session %s (returning empty for now)", chatJID, sessionID)
	return []ports.ChatMessage{}, nil
}

func (m *MeowService) GetUnreadMessages(ctx context.Context, sessionID string) ([]ports.ChatMessage, error) {
	// For now, return empty list
	m.logger.Debugf("GetUnreadMessages for session %s (returning empty for now)", sessionID)
	return []ports.ChatMessage{}, nil
}

func (m *MeowService) GetMessagesByType(ctx context.Context, sessionID, chatJID, messageType string, limit int) ([]ports.ChatMessage, error) {
	// For now, return empty list
	m.logger.Debugf("GetMessagesByType: type %s in chat %s for session %s (returning empty for now)", messageType, chatJID, sessionID)
	return []ports.ChatMessage{}, nil
}

func (m *MeowService) GetMessageCount(ctx context.Context, sessionID, chatJID string) (int, error) {
	// For now, return 0
	m.logger.Debugf("GetMessageCount: chat %s for session %s (returning 0 for now)", chatJID, sessionID)
	return 0, nil
}

func (m *MeowService) GetUnreadCount(ctx context.Context, sessionID, chatJID string) (int, error) {
	// For now, return 0
	m.logger.Debugf("GetUnreadCount: chat %s for session %s (returning 0 for now)", chatJID, sessionID)
	return 0, nil
}

func (m *MeowService) StarMessage(ctx context.Context, sessionID, chatJID, messageID string) error {
	// For now, just log
	m.logger.Debugf("StarMessage: %s in chat %s for session %s", messageID, chatJID, sessionID)
	return nil
}

func (m *MeowService) UnstarMessage(ctx context.Context, sessionID, chatJID, messageID string) error {
	// For now, just log
	m.logger.Debugf("UnstarMessage: %s in chat %s for session %s", messageID, chatJID, sessionID)
	return nil
}

func (m *MeowService) GetStarredMessages(ctx context.Context, sessionID string) ([]ports.ChatMessage, error) {
	// For now, return empty list
	m.logger.Debugf("GetStarredMessages for session %s (returning empty for now)", sessionID)
	return []ports.ChatMessage{}, nil
}

func (m *MeowService) PinMessage(ctx context.Context, sessionID, chatJID, messageID string) error {
	// For now, just log
	m.logger.Debugf("PinMessage: %s in chat %s for session %s", messageID, chatJID, sessionID)
	return nil
}

func (m *MeowService) UnpinMessage(ctx context.Context, sessionID, chatJID, messageID string) error {
	// For now, just log
	m.logger.Debugf("UnpinMessage: %s in chat %s for session %s", messageID, chatJID, sessionID)
	return nil
}

func (m *MeowService) GetPinnedMessages(ctx context.Context, sessionID, chatJID string) ([]ports.ChatMessage, error) {
	// For now, return empty list
	m.logger.Debugf("GetPinnedMessages: chat %s for session %s (returning empty for now)", chatJID, sessionID)
	return []ports.ChatMessage{}, nil
}
