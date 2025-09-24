package wmeow

import (
	"context"
	"fmt"
	"time"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/proto/waE2E"
	waTypes "go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

// MessageActions methods - ações sobre mensagens existentes

func (m *MeowService) MarkMessageRead(ctx context.Context, sessionID, chatJID string, messageIDs []string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	jid, err := waTypes.ParseJID(chatJID)
	if err != nil {
		return fmt.Errorf("invalid chat JID %s: %w", chatJID, err)
	}

	// Mark messages as read in WhatsApp
	var msgIDs []waTypes.MessageID
	for _, msgID := range messageIDs {
		msgIDs = append(msgIDs, waTypes.MessageID(msgID))
	}

	err = client.GetClient().MarkRead(msgIDs, time.Now(), jid, jid)
	if err != nil {
		return fmt.Errorf("failed to mark messages as read: %w", err)
	}

	m.logger.Debugf("Marked %d messages as read in chat %s for session %s", len(messageIDs), chatJID, sessionID)
	return nil
}

func (m *MeowService) DeleteMessage(ctx context.Context, sessionID, chatJID, messageID string, forEveryone bool) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	jid, err := waTypes.ParseJID(chatJID)
	if err != nil {
		return fmt.Errorf("invalid chat JID %s: %w", chatJID, err)
	}

	// Use the BuildRevoke method for message deletion
	if forEveryone {
		// Revoke for everyone using the new BuildRevoke method
		revokeMsg := client.GetClient().BuildRevoke(jid, waTypes.EmptyJID, messageID)
		_, err = client.GetClient().SendMessage(ctx, jid, revokeMsg)
	} else {
		// Delete for me only - this is handled locally
		err = m.messageRepo.DeleteMessage(ctx, messageID)
	}
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	deleteType := "for me"
	if forEveryone {
		deleteType = "for everyone"
	}

	m.logger.Debugf("Deleted message %s %s in chat %s for session %s", messageID, deleteType, chatJID, sessionID)
	return nil
}

func (m *MeowService) EditMessage(ctx context.Context, sessionID, chatJID, messageID, newText string) (*whatsmeow.SendResponse, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	jid, err := waTypes.ParseJID(chatJID)
	if err != nil {
		return nil, fmt.Errorf("invalid chat JID %s: %w", chatJID, err)
	}

	// Send a new message with edited content (WhatsApp doesn't support true editing)
	editedMsg := &waProto.Message{
		Conversation: proto.String(newText),
	}

	resp, err := client.GetClient().SendMessage(ctx, jid, editedMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to send edited message: %w", err)
	}

	m.logger.Debugf("Sent edited message %s in chat %s for session %s", messageID, chatJID, sessionID)
	return &resp, nil
}

func (m *MeowService) ReactToMessage(ctx context.Context, sessionID, chatJID, messageID, emoji string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	// For now, just log the reaction - WhatsApp reaction API is complex
	m.logger.Debugf("ReactToMessage: %s to message %s in chat %s for session %s", emoji, messageID, chatJID, sessionID)

	// TODO: Implement proper reaction using whatsmeow when API is stable
	return nil
}

func (m *MeowService) ForwardMessage(ctx context.Context, sessionID, fromChatJID, toChatJID, messageID string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	toJID, err := waTypes.ParseJID(toChatJID)
	if err != nil {
		return fmt.Errorf("invalid to chat JID %s: %w", toChatJID, err)
	}

	// Get the original message from repository
	originalMsg, err := m.messageRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("failed to get original message: %w", err)
	}

	// Create forward message based on original message type
	var forwardMsg *waProto.Message
	if originalMsg.Content != nil && *originalMsg.Content != "" {
		forwardMsg = &waProto.Message{
			Conversation: originalMsg.Content,
		}
	} else {
		// For media messages, create a simple text forward
		forwardMsg = &waProto.Message{
			Conversation: proto.String("Forwarded message"),
		}
	}

	_, err = client.GetClient().SendMessage(ctx, toJID, forwardMsg)
	if err != nil {
		return fmt.Errorf("failed to forward message: %w", err)
	}

	m.logger.Debugf("Forwarded message %s from %s to %s in session %s", messageID, fromChatJID, toChatJID, sessionID)
	return nil
}

func (m *MeowService) DownloadMediaMessage(ctx context.Context, sessionID, messageID string) ([]byte, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	// This is a simplified implementation
	// In a real implementation, you would need to:
	// 1. Find the message by ID
	// 2. Extract media info from the message
	// 3. Download the media using client.Download()

	m.logger.Warnf("DownloadMediaMessage not fully implemented for message %s in session %s", messageID, sessionID)
	return nil, fmt.Errorf("download media message not fully implemented")
}
