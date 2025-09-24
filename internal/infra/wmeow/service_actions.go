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

	var msgIDs []waTypes.MessageID
	for _, msgID := range messageIDs {
		msgIDs = append(msgIDs, waTypes.MessageID{
			ID:     msgID,
			FromMe: false,
		})
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

	msgKey := &waProto.MessageKey{
		RemoteJid: &chatJID,
		Id:        &messageID,
		FromMe:    proto.Bool(true),
	}

	var deleteMsg *waProto.Message
	if forEveryone {
		deleteMsg = &waProto.Message{
			ProtocolMessage: &waProto.ProtocolMessage{
				Type: waProto.ProtocolMessage_REVOKE.Enum(),
				Key:  msgKey,
			},
		}
	} else {
		deleteMsg = &waProto.Message{
			ProtocolMessage: &waProto.ProtocolMessage{
				Type: waProto.ProtocolMessage_MESSAGE_DELETE.Enum(),
				Key:  msgKey,
			},
		}
	}

	_, err = client.GetClient().SendMessage(ctx, jid, "", deleteMsg)
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

func (m *MeowService) EditMessage(ctx context.Context, sessionID, chatJID, messageID, newText string) error {
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

	editMsg := &waProto.Message{
		EditedMessage: &waProto.FutureProofMessage{
			Message: &waProto.Message{
				Conversation: &newText,
			},
		},
		ProtocolMessage: &waProto.ProtocolMessage{
			Type: waProto.ProtocolMessage_MESSAGE_EDIT.Enum(),
			Key: &waProto.MessageKey{
				RemoteJid: &chatJID,
				Id:        &messageID,
				FromMe:    proto.Bool(true),
			},
			EditedMessage: &waProto.Message{
				Conversation: &newText,
			},
		},
	}

	_, err = client.GetClient().SendMessage(ctx, jid, "", editMsg)
	if err != nil {
		return fmt.Errorf("failed to edit message: %w", err)
	}

	m.logger.Debugf("Edited message %s in chat %s for session %s", messageID, chatJID, sessionID)
	return nil
}

func (m *MeowService) ReactToMessage(ctx context.Context, sessionID, chatJID, messageID, emoji string) error {
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

	reactionMsg := &waProto.Message{
		ReactionMessage: &waProto.ReactionMessage{
			Key: &waProto.MessageKey{
				RemoteJid: &chatJID,
				Id:        &messageID,
				FromMe:    proto.Bool(false),
			},
			Text:      &emoji,
			SenderJid: proto.String(client.GetClient().Store.ID.String()),
		},
	}

	_, err = client.GetClient().SendMessage(ctx, jid, "", reactionMsg)
	if err != nil {
		return fmt.Errorf("failed to react to message: %w", err)
	}

	m.logger.Debugf("Reacted with %s to message %s in chat %s for session %s", emoji, messageID, chatJID, sessionID)
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

	fromJID, err := waTypes.ParseJID(fromChatJID)
	if err != nil {
		return fmt.Errorf("invalid from chat JID %s: %w", fromChatJID, err)
	}

	toJID, err := waTypes.ParseJID(toChatJID)
	if err != nil {
		return fmt.Errorf("invalid to chat JID %s: %w", toChatJID, err)
	}

	// This is a simplified implementation
	// In a real implementation, you would need to:
	// 1. Find the original message
	// 2. Create a forward message with the original content
	// 3. Send it to the target chat

	m.logger.Warnf("ForwardMessage not fully implemented: forwarding %s from %s to %s in session %s", messageID, fromChatJID, toChatJID, sessionID)
	return fmt.Errorf("forward message not fully implemented")
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
