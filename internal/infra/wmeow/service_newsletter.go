package wmeow

import (
	"context"
	"fmt"

	"zpmeow/internal/application/ports"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/proto/waE2E"
	waTypes "go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

// NewsletterManager methods - gestão de newsletters

func (m *MeowService) SubscribeNewsletter(ctx context.Context, sessionID, newsletterJID string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	jid, err := waTypes.ParseJID(newsletterJID)
	if err != nil {
		return fmt.Errorf("invalid newsletter JID %s: %w", newsletterJID, err)
	}

	err = client.GetClient().SubscribeNewsletter(jid)
	if err != nil {
		return fmt.Errorf("failed to subscribe to newsletter: %w", err)
	}

	m.logger.Debugf("Subscribed to newsletter %s for session %s", newsletterJID, sessionID)
	return nil
}

func (m *MeowService) UnsubscribeNewsletter(ctx context.Context, sessionID, newsletterJID string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	jid, err := waTypes.ParseJID(newsletterJID)
	if err != nil {
		return fmt.Errorf("invalid newsletter JID %s: %w", newsletterJID, err)
	}

	err = client.GetClient().UnsubscribeNewsletter(jid)
	if err != nil {
		return fmt.Errorf("failed to unsubscribe from newsletter: %w", err)
	}

	m.logger.Debugf("Unsubscribed from newsletter %s for session %s", newsletterJID, sessionID)
	return nil
}

func (m *MeowService) GetNewsletterInfo(ctx context.Context, sessionID, newsletterJID string) (*ports.NewsletterInfo, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	jid, err := waTypes.ParseJID(newsletterJID)
	if err != nil {
		return nil, fmt.Errorf("invalid newsletter JID %s: %w", newsletterJID, err)
	}

	info, err := client.GetClient().GetNewsletterInfo(jid)
	if err != nil {
		return nil, fmt.Errorf("failed to get newsletter info: %w", err)
	}

	result := &ports.NewsletterInfo{
		JID:         info.ID.String(),
		Name:        info.Name,
		Description: info.Description,
		Subscribers: int(info.Subscribers),
		Verified:    info.Verification == waTypes.NewsletterVerificationVerified,
	}

	if info.Picture != nil {
		result.PictureID = info.Picture.ID
	}

	m.logger.Debugf("Retrieved newsletter info for %s in session %s", newsletterJID, sessionID)
	return result, nil
}

func (m *MeowService) SendNewsletterReaction(ctx context.Context, sessionID, newsletterJID, messageID, emoji string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	jid, err := waTypes.ParseJID(newsletterJID)
	if err != nil {
		return fmt.Errorf("invalid newsletter JID %s: %w", newsletterJID, err)
	}

	msgID := waTypes.MessageID{
		ID:     messageID,
		FromMe: false,
	}

	reaction := &whatsmeow.ReactionMessage{
		Key: &waTypes.MessageKey{
			RemoteJID: jid,
			ID:        msgID.ID,
			FromMe:    &msgID.FromMe,
		},
		Text:      emoji,
		GroupJID:  jid,
		SenderJID: client.GetClient().Store.ID,
	}

	_, err = client.GetClient().SendMessage(ctx, jid, "", &waProto.Message{
		ReactionMessage: &waProto.ReactionMessage{
			Key: &waProto.MessageKey{
				RemoteJid: &jid.String(),
				Id:        &msgID.ID,
				FromMe:    &msgID.FromMe,
			},
			Text:      &emoji,
			SenderJid: proto.String(client.GetClient().Store.ID.String()),
		},
	})

	if err != nil {
		return fmt.Errorf("failed to send newsletter reaction: %w", err)
	}

	m.logger.Debugf("Sent reaction %s to message %s in newsletter %s for session %s", emoji, messageID, newsletterJID, sessionID)
	return nil
}

func (m *MeowService) UploadNewsletterMedia(ctx context.Context, sessionID string, data []byte, mimeType string) (*ports.MediaUploadResult, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	uploaded, err := client.GetClient().Upload(ctx, data, whatsmeow.MediaType(mimeType))
	if err != nil {
		return nil, fmt.Errorf("failed to upload newsletter media: %w", err)
	}

	result := &ports.MediaUploadResult{
		URL:       uploaded.URL,
		MediaKey:  uploaded.MediaKey,
		FileEncSHA256: uploaded.FileEncSHA256,
		FileSHA256:    uploaded.FileSHA256,
		FileLength:    uploaded.FileLength,
		MimeType:      mimeType,
	}

	m.logger.Debugf("Uploaded newsletter media for session %s", sessionID)
	return result, nil
}
