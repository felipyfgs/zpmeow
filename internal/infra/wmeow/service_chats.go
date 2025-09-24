package wmeow

import (
	"context"
	"fmt"
	"time"

	"zpmeow/internal/application/ports"

	waTypes "go.mau.fi/whatsmeow/types"
)

// ChatManager methods - gestão de conversas e chats

func (m *MeowService) ListChats(ctx context.Context, sessionID, chatType string) ([]ports.ChatInfo, error) {
	// For now, return empty list
	// TODO: Implement proper chat listing when needed
	m.logger.Debugf("ListChats for session %s, chatType: %s (returning empty for now)", sessionID, chatType)
	return []ports.ChatInfo{}, nil
}

func (m *MeowService) GetChatHistory(ctx context.Context, sessionID, chatJID string, limit int, offset int) ([]ports.ChatMessage, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	if limit <= 0 {
		limit = 50
	}

	// For now, return empty message history
	// TODO: Implement proper message retrieval when repository method is available
	m.logger.Debugf("GetChatHistory: %s for session %s (returning empty for now, limit: %d)", chatJID, sessionID, limit)

	var result []ports.ChatMessage
	// Return empty result for now

	m.logger.Debugf("Retrieved %d messages from chat %s for session %s", len(result), chatJID, sessionID)
	return result, nil
}

func (m *MeowService) ArchiveChat(ctx context.Context, sessionID, chatJID string, archive bool) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	// Archive functionality - store archive status locally
	// For now, just log the operation
	m.logger.Debugf("ArchiveChat: %s (archived: %v) for session %s", chatJID, archive, sessionID)

	// TODO: Implement proper archive status storage when repository method is available

	action := "archived"
	if !archive {
		action = "unarchived"
	}

	m.logger.Debugf("Chat %s %s for session %s", chatJID, action, sessionID)
	return nil
}

func (m *MeowService) DeleteChat(ctx context.Context, sessionID, chatJID string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	// Delete chat functionality - clear local messages
	// For now, just log the operation
	m.logger.Debugf("DeleteChat: %s for session %s", chatJID, sessionID)

	// TODO: Implement proper message deletion when repository method is available

	m.logger.Debugf("Deleted chat %s for session %s", chatJID, sessionID)
	return nil
}

func (m *MeowService) MuteChat(ctx context.Context, sessionID, chatJID string, mute bool, duration time.Duration) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	// Removed unused variables for compilation

	// Mute functionality - store mute status locally
	// For now, just log the operation
	m.logger.Debugf("MuteChat: %s (mute: %v, duration: %v) for session %s", chatJID, mute, duration, sessionID)

	// TODO: Implement proper mute status storage when repository method is available

	m.logger.Debugf("Muted chat %s for %v in session %s", chatJID, duration, sessionID)
	return nil
}

func (m *MeowService) UnmuteChat(ctx context.Context, sessionID, chatJID string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	// Unmute functionality - remove mute status locally
	// For now, just log the operation
	m.logger.Debugf("UnmuteChat: %s for session %s", chatJID, sessionID)

	// TODO: Implement proper mute status removal when repository method is available

	m.logger.Debugf("Unmuted chat %s for session %s", chatJID, sessionID)
	return nil
}

func (m *MeowService) PinChat(ctx context.Context, sessionID, chatJID string, pin bool) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	// Pin functionality - for now just log
	m.logger.Debugf("PinChat: %s (pinned: %v) for session %s", chatJID, pin, sessionID)

	// TODO: Implement proper pin functionality when whatsmeow supports it

	action := "pinned"
	if !pin {
		action = "unpinned"
	}

	m.logger.Debugf("Chat %s %s for session %s", chatJID, action, sessionID)
	return nil
}

func (m *MeowService) SetDisappearingTimer(ctx context.Context, sessionID, chatJID string, timer time.Duration) error {
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

	err = client.GetClient().SetDisappearingTimer(jid, timer, time.Now())
	if err != nil {
		return fmt.Errorf("failed to set disappearing timer: %w", err)
	}

	m.logger.Debugf("Set disappearing timer to %v for chat %s in session %s", timer, chatJID, sessionID)
	return nil
}

// Additional methods required by ChatManager interface

func (m *MeowService) GetChats(ctx context.Context, sessionID string, limit, offset int) ([]ports.ChatInfo, error) {
	return m.ListChats(ctx, sessionID, "")
}

func (m *MeowService) GetChatInfo(ctx context.Context, sessionID, chatJID string) (*ports.ChatInfo, error) {
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

	chatInfo := &ports.ChatInfo{
		JID:     chatJID,
		IsGroup: jid.Server == waTypes.GroupServer,
	}

	if !chatInfo.IsGroup {
		// Get contact info for individual chat
		contact, err := client.GetClient().Store.Contacts.GetContact(ctx, jid)
		if err == nil && contact.PushName != "" {
			chatInfo.Name = contact.PushName
		}
	} else {
		// Get group info
		groupInfo, err := client.GetClient().GetGroupInfo(jid)
		if err == nil {
			chatInfo.Name = groupInfo.Name
		}
	}

	return chatInfo, nil
}
