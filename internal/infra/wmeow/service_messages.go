package wmeow

import (
	"context"
	"fmt"

	"zpmeow/internal/application/ports"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/proto/waE2E"
)

// MessageSender methods - envio de mensagens de todos os tipos

func (m *MeowService) SendTextMessage(ctx context.Context, sessionID, phone, text string) (*whatsmeow.SendResponse, error) {
	client, err := m.validateAndGetClientForSending(sessionID)
	if err != nil {
		return nil, err
	}
	return m.sendTextMessage(client.GetClient(), phone, text)
}

func (m *MeowService) sendTextMessage(client *whatsmeow.Client, to, text string) (*whatsmeow.SendResponse, error) {
	validator := m.getValidator()
	if err := validator.ValidateMessageInput(client, to); err != nil {
		return nil, err
	}

	builder := m.getMessageBuilder()
	message, err := builder.BuildTextMessage(text)
	if err != nil {
		return nil, err
	}

	return m.messageSender.SendToJID(client, to, message)
}

func (m *MeowService) SendImageMessage(ctx context.Context, sessionID, phone string, data []byte, caption, mimeType string) (*whatsmeow.SendResponse, error) {
	client, err := m.validateAndGetClientForSending(sessionID)
	if err != nil {
		return nil, err
	}
	return m.sendImageMessage(client.GetClient(), phone, data, caption)
}

func (m *MeowService) sendImageMessage(client *whatsmeow.Client, to string, data []byte, caption string) (*whatsmeow.SendResponse, error) {
	if err := m.getValidator().ValidateMessageInput(client, to); err != nil {
		return nil, err
	}

	builder := m.getMessageBuilder()
	message, err := builder.BuildImageMessage(data, caption)
	if err != nil {
		return nil, err
	}

	return m.messageSender.SendToJID(client, to, message)
}

func (m *MeowService) SendAudioMessage(ctx context.Context, sessionID, phone string, data []byte, mimeType string) (*whatsmeow.SendResponse, error) {
	client, err := m.validateAndGetClientForSending(sessionID)
	if err != nil {
		return nil, err
	}
	return m.sendAudioMessage(client.GetClient(), phone, data, mimeType, true) // Default PTT to true
}

func (m *MeowService) SendAudioMessageWithPTT(ctx context.Context, sessionID, phone string, data []byte, mimeType string, ptt bool) (*whatsmeow.SendResponse, error) {
	client, err := m.validateAndGetClientForSending(sessionID)
	if err != nil {
		return nil, err
	}
	return m.sendAudioMessage(client.GetClient(), phone, data, mimeType, ptt)
}

func (m *MeowService) sendAudioMessage(client *whatsmeow.Client, to string, data []byte, mimeType string, ptt bool) (*whatsmeow.SendResponse, error) {
	if err := m.getValidator().ValidateMessageInput(client, to); err != nil {
		return nil, err
	}

	builder := m.getMessageBuilder()
	message, err := builder.BuildAudioMessage(data, mimeType, ptt)
	if err != nil {
		return nil, err
	}

	return m.messageSender.SendToJID(client, to, message)
}

func (m *MeowService) SendVideoMessage(ctx context.Context, sessionID, phone string, data []byte, caption, mimeType string) (*whatsmeow.SendResponse, error) {
	client, err := m.validateAndGetClientForSending(sessionID)
	if err != nil {
		return nil, err
	}
	return m.sendVideoMessage(client.GetClient(), phone, data, caption, mimeType)
}

func (m *MeowService) sendVideoMessage(client *whatsmeow.Client, to string, data []byte, caption, mimeType string) (*whatsmeow.SendResponse, error) {
	if err := m.getValidator().ValidateMessageInput(client, to); err != nil {
		return nil, err
	}

	builder := m.getMessageBuilder()
	message, err := builder.BuildVideoMessage(data, caption, mimeType)
	if err != nil {
		return nil, err
	}

	return m.messageSender.SendToJID(client, to, message)
}

func (m *MeowService) SendDocumentMessage(ctx context.Context, sessionID, phone string, data []byte, filename, mimeType string) (*whatsmeow.SendResponse, error) {
	client, err := m.validateAndGetClientForSending(sessionID)
	if err != nil {
		return nil, err
	}
	return m.sendDocumentMessage(client.GetClient(), phone, data, filename, mimeType)
}

func (m *MeowService) sendDocumentMessage(client *whatsmeow.Client, to string, data []byte, filename, mimeType string) (*whatsmeow.SendResponse, error) {
	if err := m.getValidator().ValidateMessageInput(client, to); err != nil {
		return nil, err
	}

	builder := m.getMessageBuilder()
	message, err := builder.BuildDocumentMessage(data, filename, mimeType)
	if err != nil {
		return nil, err
	}

	return m.messageSender.SendToJID(client, to, message)
}

func (m *MeowService) SendStickerMessage(ctx context.Context, sessionID, phone string, data []byte, mimeType string) (*whatsmeow.SendResponse, error) {
	client, err := m.validateAndGetClientForSending(sessionID)
	if err != nil {
		return nil, err
	}
	return m.sendStickerMessage(client.GetClient(), phone, data, mimeType)
}

func (m *MeowService) sendStickerMessage(client *whatsmeow.Client, to string, data []byte, mimeType string) (*whatsmeow.SendResponse, error) {
	if err := m.getValidator().ValidateMessageInput(client, to); err != nil {
		return nil, err
	}

	builder := m.getMessageBuilder()
	message, err := builder.BuildStickerMessage(data, mimeType)
	if err != nil {
		return nil, err
	}

	return m.messageSender.SendToJID(client, to, message)
}

func (m *MeowService) SendContactMessage(ctx context.Context, sessionID, phone, name, contactPhone string) (*whatsmeow.SendResponse, error) {
	client, err := m.validateAndGetClientForSending(sessionID)
	if err != nil {
		return nil, err
	}
	return m.sendContactMessage(client.GetClient(), phone, name, contactPhone)
}

func (m *MeowService) sendContactMessage(client *whatsmeow.Client, to, name, phone string) (*whatsmeow.SendResponse, error) {
	if err := m.getValidator().ValidateMessageInput(client, to); err != nil {
		return nil, err
	}

	vcard := fmt.Sprintf("BEGIN:VCARD\nVERSION:3.0\nFN:%s\nTEL:%s\nEND:VCARD", name, phone)
	message := &waProto.Message{
		ContactMessage: &waProto.ContactMessage{
			DisplayName: &name,
			Vcard:       &vcard,
		},
	}

	return m.messageSender.SendToJID(client, to, message)
}

func (m *MeowService) SendContactsMessage(ctx context.Context, sessionID, phone string, contacts []ports.ContactData) (*whatsmeow.SendResponse, error) {
	client, err := m.validateAndGetClientForSending(sessionID)
	if err != nil {
		return nil, err
	}
	return m.sendContactsMessage(client.GetClient(), phone, contacts)
}

func (m *MeowService) sendContactsMessage(client *whatsmeow.Client, to string, contacts []ports.ContactData) (*whatsmeow.SendResponse, error) {
	if err := m.getValidator().ValidateMessageInput(client, to); err != nil {
		return nil, err
	}

	builder := m.getMessageBuilder()
	message, err := builder.BuildContactsMessage(contacts)
	if err != nil {
		return nil, err
	}

	return m.messageSender.SendToJID(client, to, message)
}

func (m *MeowService) SendLocationMessage(ctx context.Context, sessionID, phone string, latitude, longitude float64, name, address string) (*whatsmeow.SendResponse, error) {
	client, err := m.validateAndGetClientForSending(sessionID)
	if err != nil {
		return nil, err
	}
	return m.sendLocationMessage(client.GetClient(), phone, latitude, longitude, name, address)
}

func (m *MeowService) sendLocationMessage(client *whatsmeow.Client, to string, latitude, longitude float64, name, address string) (*whatsmeow.SendResponse, error) {
	if err := m.getValidator().ValidateMessageInput(client, to); err != nil {
		return nil, err
	}

	builder := m.getMessageBuilder()
	message, err := builder.BuildLocationMessage(latitude, longitude, name, address)
	if err != nil {
		return nil, err
	}

	return m.messageSender.SendToJID(client, to, message)
}

func (m *MeowService) SendMediaMessage(ctx context.Context, sessionID, phone string, media ports.MediaMessage) (*whatsmeow.SendResponse, error) {
	client, err := m.validateAndGetClientForSending(sessionID)
	if err != nil {
		return nil, err
	}

	switch media.Type {
	case "image":
		return m.sendImageMessage(client.GetClient(), phone, media.Data, media.Caption)
	case "audio":
		return m.sendAudioMessage(client.GetClient(), phone, media.Data, media.MimeType, true)
	case "video":
		return m.sendVideoMessage(client.GetClient(), phone, media.Data, media.Caption, media.MimeType)
	case "document":
		return m.sendDocumentMessage(client.GetClient(), phone, media.Data, media.Filename, media.MimeType)
	case "sticker":
		return m.sendStickerMessage(client.GetClient(), phone, media.Data, media.MimeType)
	default:
		return nil, fmt.Errorf("unsupported media type: %s", media.Type)
	}
}

// Helper methods for message sending

func (m *MeowService) validateAndGetClientForSending(sessionID string) (*WameowClient, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	return client, nil
}

func (m *MeowService) getValidator() MessageValidator {
	return NewMessageValidator()
}

func (m *MeowService) getMessageBuilder() MessageBuilder {
	return NewMessageBuilder()
}
