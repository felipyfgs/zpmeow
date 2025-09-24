package wmeow

import (
	"context"
	"fmt"
	"strings"
	"time"

	"zpmeow/internal/application/ports"

	waTypes "go.mau.fi/whatsmeow/types"
)

// ChatManager methods - gestão de conversas e chats

func (m *MeowService) ListChats(ctx context.Context, sessionID, phone string, limit int) ([]ports.ChatInfo, error) {
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

	// If phone is provided, filter by phone
	if phone != "" {
		cleanPhone := strings.TrimPrefix(phone, "+")
		chats, err := m.chatRepo.GetChatsByPhoneNumber(ctx, sessionID, cleanPhone)
		if err != nil {
			return nil, fmt.Errorf("failed to get chats by phone number: %w", err)
		}

		var result []ports.ChatInfo
		for _, chat := range chats {
			chatInfo := ports.ChatInfo{
				JID:       chat.ChatJid,
				Name:      chat.ChatName,
				Phone:     phone,
				IsGroup:   strings.Contains(chat.ChatJid, "@g.us"),
				Timestamp: chat.CreatedAt,
			}

			if chat.LastMsgAt != nil {
				chatInfo.LastMessageTime = *chat.LastMsgAt
			}

			result = append(result, chatInfo)
		}

		// Apply limit
		if len(result) > limit {
			result = result[:limit]
		}

		m.logger.Debugf("Retrieved %d chats for phone %s in session %s", len(result), phone, sessionID)
		return result, nil
	}

	// Get all chats from WhatsApp
	conversations, err := client.GetClient().GetConversations(limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}

	var result []ports.ChatInfo
	for _, conv := range conversations {
		chatInfo := ports.ChatInfo{
			JID:             conv.ID.String(),
			IsGroup:         conv.ID.Server == waTypes.GroupServer,
			Timestamp:       time.Unix(int64(conv.ConversationTimestamp), 0),
			LastMessageTime: time.Unix(int64(conv.ConversationTimestamp), 0),
		}

		// Get contact info for name
		if !chatInfo.IsGroup {
			contact, err := client.GetClient().Store.Contacts.GetContact(conv.ID)
			if err == nil && contact.PushName != "" {
				chatInfo.Name = contact.PushName
			}

			// Extract phone number
			if conv.ID.Server == waTypes.DefaultUserServer {
				phone := conv.ID.User
				if !strings.HasPrefix(phone, "+") {
					phone = "+" + phone
				}
				chatInfo.Phone = phone
			}
		} else {
			// For groups, get group info
			groupInfo, err := client.GetClient().GetGroupInfo(conv.ID)
			if err == nil {
				chatInfo.Name = groupInfo.Name
			}
		}

		result = append(result, chatInfo)
	}

	m.logger.Debugf("Retrieved %d chats for session %s", len(result), sessionID)
	return result, nil
}

func (m *MeowService) GetChatHistory(ctx context.Context, sessionID, chatJID string, limit int, before string) ([]MessageInfo, error) {
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

	if limit <= 0 {
		limit = 50
	}

	var beforeID *waTypes.MessageID
	if before != "" {
		beforeID = &waTypes.MessageID{
			ID:     before,
			FromMe: false,
		}
	}

	messages, err := client.GetClient().GetMessageHistory(jid, beforeID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat history: %w", err)
	}

	var result []MessageInfo
	for _, msg := range messages {
		msgInfo := MessageInfo{
			ID:        msg.Info.ID,
			FromMe:    msg.Info.IsFromMe,
			Timestamp: msg.Info.Timestamp,
			ChatJID:   chatJID,
		}

		if msg.Info.Sender != nil {
			msgInfo.SenderJID = msg.Info.Sender.String()
		}

		// Extract message content
		if msg.Message.GetConversation() != "" {
			msgInfo.Content = msg.Message.GetConversation()
			msgInfo.Type = "text"
		} else if msg.Message.GetImageMessage() != nil {
			msgInfo.Content = msg.Message.GetImageMessage().GetCaption()
			msgInfo.Type = "image"
		} else if msg.Message.GetVideoMessage() != nil {
			msgInfo.Content = msg.Message.GetVideoMessage().GetCaption()
			msgInfo.Type = "video"
		} else if msg.Message.GetAudioMessage() != nil {
			msgInfo.Type = "audio"
		} else if msg.Message.GetDocumentMessage() != nil {
			msgInfo.Content = msg.Message.GetDocumentMessage().GetFileName()
			msgInfo.Type = "document"
		} else {
			msgInfo.Type = "unknown"
		}

		result = append(result, msgInfo)
	}

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

	jid, err := waTypes.ParseJID(chatJID)
	if err != nil {
		return fmt.Errorf("invalid chat JID %s: %w", chatJID, err)
	}

	err = client.GetClient().SetArchived(jid, archive)
	if err != nil {
		return fmt.Errorf("failed to archive chat: %w", err)
	}

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

	jid, err := waTypes.ParseJID(chatJID)
	if err != nil {
		return fmt.Errorf("invalid chat JID %s: %w", chatJID, err)
	}

	err = client.GetClient().DeleteChat(jid)
	if err != nil {
		return fmt.Errorf("failed to delete chat: %w", err)
	}

	m.logger.Debugf("Deleted chat %s for session %s", chatJID, sessionID)
	return nil
}

func (m *MeowService) MuteChat(ctx context.Context, sessionID, chatJID string, duration time.Duration) error {
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

	var muteEndTime time.Time
	if duration > 0 {
		muteEndTime = time.Now().Add(duration)
	} else {
		// Mute forever
		muteEndTime = time.Unix(0, 0)
	}

	err = client.GetClient().SetMuted(jid, muteEndTime)
	if err != nil {
		return fmt.Errorf("failed to mute chat: %w", err)
	}

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

	jid, err := waTypes.ParseJID(chatJID)
	if err != nil {
		return fmt.Errorf("invalid chat JID %s: %w", chatJID, err)
	}

	err = client.GetClient().SetMuted(jid, time.Time{})
	if err != nil {
		return fmt.Errorf("failed to unmute chat: %w", err)
	}

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

	jid, err := waTypes.ParseJID(chatJID)
	if err != nil {
		return fmt.Errorf("invalid chat JID %s: %w", chatJID, err)
	}

	err = client.GetClient().SetPinned(jid, pin)
	if err != nil {
		return fmt.Errorf("failed to pin chat: %w", err)
	}

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

	err = client.GetClient().SetDisappearingTimer(jid, timer)
	if err != nil {
		return fmt.Errorf("failed to set disappearing timer: %w", err)
	}

	m.logger.Debugf("Set disappearing timer to %v for chat %s in session %s", timer, chatJID, sessionID)
	return nil
}

// Additional methods required by ChatManager interface

func (m *MeowService) GetChats(ctx context.Context, sessionID string, limit, offset int) ([]ports.ChatInfo, error) {
	return m.ListChats(ctx, sessionID, "", limit)
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
		contact, err := client.GetClient().Store.Contacts.GetContact(jid)
		if err == nil && contact.PushName != "" {
			chatInfo.Name = contact.PushName
		}

		if jid.Server == waTypes.DefaultUserServer {
			phone := jid.User
			if !strings.HasPrefix(phone, "+") {
				phone = "+" + phone
			}
			chatInfo.Phone = phone
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
