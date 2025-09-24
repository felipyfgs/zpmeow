package wmeow

import (
	"context"
	"fmt"
	"time"

	"zpmeow/internal/application/ports"
	"zpmeow/internal/infra/database/models"

	waTypes "go.mau.fi/whatsmeow/types"
)

// ChatManager methods - gestão de conversas e chats

func (m *MeowService) ListChats(ctx context.Context, sessionID, chatType string) ([]ports.ChatInfo, error) {
	if err := m.validateClientConnection(sessionID); err != nil {
		return nil, err
	}

	result, err := m.processChatsFromDatabase(ctx, sessionID, chatType)
	if err != nil {
		m.logger.Debugf("ListChats for session %s, chatType: %s (returning empty - database error)", sessionID, chatType)
		return []ports.ChatInfo{}, nil
	}

	m.logger.Debugf("ListChats for session %s, chatType: %s - found %d chats", sessionID, chatType, len(result))
	return result, nil
}

func (m *MeowService) GetChatHistory(ctx context.Context, sessionID, chatJID string, limit int, offset int) ([]ports.ChatMessage, error) {
	if err := m.validateClientConnection(sessionID); err != nil {
		return nil, err
	}

	limit = m.normalizeLimit(limit)
	chatId := m.getChatIdFromJID(ctx, sessionID, chatJID)
	result, err := m.getMessagesFromDatabase(ctx, chatId, limit, offset)
	if err != nil {
		return nil, err
	}

	m.logger.Debugf("GetChatHistory: %s for session %s (limit: %d) - retrieved %d messages", chatJID, sessionID, limit, len(result))
	return result, nil
}

func (m *MeowService) ArchiveChat(ctx context.Context, sessionID, chatJID string, archive bool) error {
	if err := m.validateClientConnection(sessionID); err != nil {
		return err
	}

	if m.chatRepo == nil {
		return m.logArchiveAction(chatJID, archive, sessionID)
	}

	return m.updateChatArchiveStatus(ctx, sessionID, chatJID, archive)
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
	if m.chatRepo != nil && m.messageRepo != nil {
		chat, err := m.chatRepo.GetChatBySessionAndJID(ctx, sessionID, chatJID)
		if err != nil {
			m.logger.Warnf("Failed to get chat from database: %v", err)
		} else if chat != nil {
			// Delete all messages from this chat
			if err := m.messageRepo.DeleteMessagesByChatId(ctx, chat.ID); err != nil {
				m.logger.Errorf("Failed to delete messages from chat: %v", err)
				return fmt.Errorf("failed to delete messages from chat: %w", err)
			}

			// Optionally delete the chat record itself
			// For now, we'll keep the chat record but clear messages
			m.logger.Debugf("Deleted all messages from chat %s for session %s", chatJID, sessionID)
		}
	}

	m.logger.Debugf("DeleteChat: %s for session %s", chatJID, sessionID)

	m.logger.Debugf("Deleted chat %s for session %s", chatJID, sessionID)
	return nil
}

func (m *MeowService) MuteChat(ctx context.Context, sessionID, chatJID string, mute bool, duration time.Duration) error {
	if err := m.validateClientConnection(sessionID); err != nil {
		return err
	}

	if m.chatRepo == nil {
		return m.logMuteAction(chatJID, mute, duration, sessionID)
	}

	return m.updateChatMuteStatus(ctx, sessionID, chatJID, mute, duration)
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
	return m.MuteChat(ctx, sessionID, chatJID, false, 0)
}

func (m *MeowService) PinChat(ctx context.Context, sessionID, chatJID string, pin bool) error {
	if err := m.validateClientConnection(sessionID); err != nil {
		return err
	}

	if m.chatRepo == nil {
		return m.logPinAction(chatJID, pin, sessionID)
	}

	return m.updateChatPinStatus(ctx, sessionID, chatJID, pin)
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

// Helper functions to reduce complexity

func (m *MeowService) validateClientConnection(sessionID string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}
	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}
	return nil
}

func (m *MeowService) logMuteAction(chatJID string, mute bool, duration time.Duration, sessionID string) error {
	action := "muted"
	if !mute {
		action = "unmuted"
	}
	m.logger.Debugf("Chat %s %s for %v in session %s", chatJID, action, duration, sessionID)
	return nil
}

func (m *MeowService) updateChatMuteStatus(ctx context.Context, sessionID, chatJID string, mute bool, duration time.Duration) error {
	chat, err := m.chatRepo.GetChatBySessionAndJID(ctx, sessionID, chatJID)
	if err != nil {
		m.logger.Warnf("Failed to get chat from database: %v", err)
		return m.logMuteAction(chatJID, mute, duration, sessionID)
	}

	if chat != nil {
		return m.updateExistingChatMute(ctx, chat, mute, duration, chatJID, sessionID)
	}

	return m.createChatWithMuteStatus(ctx, sessionID, chatJID, mute, duration)
}

func (m *MeowService) updateExistingChatMute(ctx context.Context, chat *models.ChatModel, mute bool, duration time.Duration, chatJID, sessionID string) error {
	if chat.Metadata == nil {
		chat.Metadata = make(models.JSONB)
	}

	chat.Metadata["muted"] = mute
	if mute && duration > 0 {
		muteUntil := time.Now().Add(duration)
		chat.Metadata["muteUntil"] = muteUntil.Unix()
	} else {
		delete(chat.Metadata, "muteUntil")
	}

	if err := m.chatRepo.UpdateChat(ctx, chat); err != nil {
		m.logger.Errorf("Failed to update chat mute status: %v", err)
		return fmt.Errorf("failed to update chat mute status: %w", err)
	}

	return m.logMuteAction(chatJID, mute, duration, sessionID)
}

func (m *MeowService) createChatWithMuteStatus(ctx context.Context, sessionID, chatJID string, mute bool, duration time.Duration) error {
	jid, err := waTypes.ParseJID(chatJID)
	if err != nil {
		return fmt.Errorf("invalid chat JID %s: %w", chatJID, err)
	}

	metadata := make(models.JSONB)
	metadata["muted"] = mute
	if mute && duration > 0 {
		muteUntil := time.Now().Add(duration)
		metadata["muteUntil"] = muteUntil.Unix()
	}

	newChat := &models.ChatModel{
		SessionId: sessionID,
		ChatJid:   chatJID,
		IsGroup:   jid.Server == waTypes.GroupServer,
		Metadata:  metadata,
	}

	if err := m.chatRepo.CreateChat(ctx, newChat); err != nil {
		m.logger.Errorf("Failed to create chat with mute status: %v", err)
		return fmt.Errorf("failed to create chat with mute status: %w", err)
	}

	return m.logMuteAction(chatJID, mute, duration, sessionID)
}

func (m *MeowService) logPinAction(chatJID string, pin bool, sessionID string) error {
	action := "pinned"
	if !pin {
		action = "unpinned"
	}
	m.logger.Debugf("Chat %s %s for session %s", chatJID, action, sessionID)
	return nil
}

func (m *MeowService) updateChatPinStatus(ctx context.Context, sessionID, chatJID string, pin bool) error {
	chat, err := m.chatRepo.GetChatBySessionAndJID(ctx, sessionID, chatJID)
	if err != nil {
		m.logger.Warnf("Failed to get chat from database: %v", err)
		return m.logPinAction(chatJID, pin, sessionID)
	}

	if chat != nil {
		return m.updateExistingChatPin(ctx, chat, pin, chatJID, sessionID)
	}

	return m.createChatWithPinStatus(ctx, sessionID, chatJID, pin)
}

func (m *MeowService) updateExistingChatPin(ctx context.Context, chat *models.ChatModel, pin bool, chatJID, sessionID string) error {
	if chat.Metadata == nil {
		chat.Metadata = make(models.JSONB)
	}

	chat.Metadata["pinned"] = pin
	if pin {
		chat.Metadata["pinnedAt"] = time.Now().Unix()
	} else {
		delete(chat.Metadata, "pinnedAt")
	}

	if err := m.chatRepo.UpdateChat(ctx, chat); err != nil {
		m.logger.Errorf("Failed to update chat pin status: %v", err)
		return fmt.Errorf("failed to update chat pin status: %w", err)
	}

	return m.logPinAction(chatJID, pin, sessionID)
}

func (m *MeowService) createChatWithPinStatus(ctx context.Context, sessionID, chatJID string, pin bool) error {
	jid, err := waTypes.ParseJID(chatJID)
	if err != nil {
		return fmt.Errorf("invalid chat JID %s: %w", chatJID, err)
	}

	metadata := make(models.JSONB)
	metadata["pinned"] = pin
	if pin {
		metadata["pinnedAt"] = time.Now().Unix()
	}

	newChat := &models.ChatModel{
		SessionId: sessionID,
		ChatJid:   chatJID,
		IsGroup:   jid.Server == waTypes.GroupServer,
		Metadata:  metadata,
	}

	if err := m.chatRepo.CreateChat(ctx, newChat); err != nil {
		m.logger.Errorf("Failed to create chat with pin status: %v", err)
		return fmt.Errorf("failed to create chat with pin status: %w", err)
	}

	return m.logPinAction(chatJID, pin, sessionID)
}

func (m *MeowService) logArchiveAction(chatJID string, archive bool, sessionID string) error {
	action := "archived"
	if !archive {
		action = "unarchived"
	}
	m.logger.Debugf("Chat %s %s for session %s", chatJID, action, sessionID)
	return nil
}

func (m *MeowService) updateChatArchiveStatus(ctx context.Context, sessionID, chatJID string, archive bool) error {
	chat, err := m.chatRepo.GetChatBySessionAndJID(ctx, sessionID, chatJID)
	if err != nil {
		m.logger.Warnf("Failed to get chat from database: %v", err)
		return m.logArchiveAction(chatJID, archive, sessionID)
	}

	if chat != nil {
		return m.updateExistingChatArchive(ctx, chat, archive, chatJID, sessionID)
	}

	return m.createChatWithArchiveStatus(ctx, sessionID, chatJID, archive)
}

func (m *MeowService) updateExistingChatArchive(ctx context.Context, chat *models.ChatModel, archive bool, chatJID, sessionID string) error {
	chat.IsArchived = archive
	if err := m.chatRepo.UpdateChat(ctx, chat); err != nil {
		m.logger.Errorf("Failed to update chat archive status: %v", err)
		return fmt.Errorf("failed to update chat archive status: %w", err)
	}
	return m.logArchiveAction(chatJID, archive, sessionID)
}

func (m *MeowService) createChatWithArchiveStatus(ctx context.Context, sessionID, chatJID string, archive bool) error {
	jid, err := waTypes.ParseJID(chatJID)
	if err != nil {
		return fmt.Errorf("invalid chat JID %s: %w", chatJID, err)
	}

	newChat := &models.ChatModel{
		SessionId:  sessionID,
		ChatJid:    chatJID,
		IsGroup:    jid.Server == waTypes.GroupServer,
		IsArchived: archive,
	}

	if err := m.chatRepo.CreateChat(ctx, newChat); err != nil {
		m.logger.Errorf("Failed to create chat with archive status: %v", err)
		return fmt.Errorf("failed to create chat with archive status: %w", err)
	}

	return m.logArchiveAction(chatJID, archive, sessionID)
}

func (m *MeowService) getChatIdFromJID(ctx context.Context, sessionID, chatJID string) string {
	if m.chatRepo == nil {
		return ""
	}

	chat, err := m.chatRepo.GetChatBySessionAndJID(ctx, sessionID, chatJID)
	if err != nil {
		m.logger.Warnf("Failed to get chat from database: %v", err)
		return ""
	}

	if chat != nil {
		return chat.ID
	}

	return ""
}

func (m *MeowService) getMessagesFromDatabase(ctx context.Context, chatId string, limit, offset int) ([]ports.ChatMessage, error) {
	if m.messageRepo == nil || chatId == "" {
		return []ports.ChatMessage{}, nil
	}

	messages, err := m.messageRepo.GetMessagesByChatId(ctx, chatId, limit, offset)
	if err != nil {
		m.logger.Warnf("Failed to get messages from database: %v", err)
		return []ports.ChatMessage{}, nil
	}

	return m.formatChatMessages(messages, chatId), nil
}

func (m *MeowService) formatChatMessages(messages []*models.MessageModel, chatId string) []ports.ChatMessage {
	var result []ports.ChatMessage
	for _, msg := range messages {
		chatMessage := ports.ChatMessage{
			ID:        msg.MsgId,
			ChatJID:   chatId,
			FromJID:   msg.SenderJid,
			Text:      getStringValue(msg.Content),
			Content:   getStringValue(msg.Content),
			Type:      msg.MsgType,
			Timestamp: msg.Timestamp,
		}
		result = append(result, chatMessage)
	}
	return result
}

func (m *MeowService) normalizeLimit(limit int) int {
	if limit <= 0 {
		return 50
	}
	return limit
}

func (m *MeowService) getChatsFromDatabase(ctx context.Context, sessionID string) ([]*models.ChatModel, error) {
	if m.chatRepo == nil {
		return nil, fmt.Errorf("chat repository not available")
	}

	chats, err := m.chatRepo.GetChatsBySessionID(ctx, sessionID, 100, 0)
	if err != nil {
		m.logger.Warnf("Failed to get chats from database: %v", err)
		return nil, err
	}

	return chats, nil
}

func (m *MeowService) formatChatInfo(chat *models.ChatModel) ports.ChatInfo {
	var lastMessageAt string
	if chat.LastMsgAt != nil {
		lastMessageAt = chat.LastMsgAt.Format(time.RFC3339)
	}

	return ports.ChatInfo{
		JID:           chat.ChatJid,
		Name:          getStringValue(chat.ChatName),
		IsGroup:       chat.IsGroup,
		UnreadCount:   chat.UnreadCount,
		LastMessageAt: lastMessageAt,
		IsArchived:    chat.IsArchived,
	}
}

func (m *MeowService) shouldIncludeChat(chat *models.ChatModel, chatType string) bool {
	if chatType == "" {
		return true
	}

	if chatType == "group" && !chat.IsGroup {
		return false
	}

	if chatType == "individual" && chat.IsGroup {
		return false
	}

	return true
}

func (m *MeowService) processChatsFromDatabase(ctx context.Context, sessionID, chatType string) ([]ports.ChatInfo, error) {
	chats, err := m.getChatsFromDatabase(ctx, sessionID)
	if err != nil {
		return []ports.ChatInfo{}, nil // Return empty on error, don't fail
	}

	var result []ports.ChatInfo
	for _, chat := range chats {
		if !m.shouldIncludeChat(chat, chatType) {
			continue
		}

		chatInfo := m.formatChatInfo(chat)
		result = append(result, chatInfo)
	}

	return result, nil
}
