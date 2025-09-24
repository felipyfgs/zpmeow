package wmeow

import (
	"context"
	"fmt"

	waTypes "go.mau.fi/whatsmeow/types"
)

// ProfileManager methods - gestão de perfil do usuário

func (m *MeowService) UpdateProfile(ctx context.Context, sessionID, name, about string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	if name != "" {
		err := client.GetClient().SetStatusMessage(name)
		if err != nil {
			m.logger.Warnf("Failed to set profile name for session %s: %v", sessionID, err)
		}
	}

	if about != "" {
		err := client.GetClient().SetStatusMessage(about)
		if err != nil {
			return fmt.Errorf("failed to set status message: %w", err)
		}
	}

	m.logger.Debugf("Updated profile for session %s", sessionID)
	return nil
}

func (m *MeowService) SetUserPresence(ctx context.Context, sessionID, state string) error {
	return m.SetPresence(ctx, sessionID, "", state, "")
}

func (m *MeowService) SetPresence(ctx context.Context, sessionID, phone, state, media string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	var chatJID waTypes.JID
	if phone != "" {
		jid, err := parsePhoneToJID(phone)
		if err != nil {
			return fmt.Errorf("invalid phone number: %w", err)
		}
		chatJID = jid
	}

	var presence waTypes.Presence
	switch state {
	case "available":
		presence = waTypes.PresenceAvailable
	case "unavailable":
		presence = waTypes.PresenceUnavailable
	case "composing":
		presence = waTypes.PresenceComposing
	case "recording":
		presence = waTypes.PresenceRecording
	case "paused":
		presence = waTypes.PresencePaused
	default:
		return fmt.Errorf("invalid presence state: %s", state)
	}

	var mediaType waTypes.ChatPresenceMedia
	switch media {
	case "text":
		mediaType = waTypes.ChatPresenceMediaText
	case "audio":
		mediaType = waTypes.ChatPresenceMediaAudio
	default:
		mediaType = waTypes.ChatPresenceMediaText
	}

	if phone != "" {
		err := client.GetClient().SendChatPresence(chatJID, presence, mediaType)
		if err != nil {
			return fmt.Errorf("failed to send chat presence: %w", err)
		}
		m.logger.Debugf("Set presence %s for chat %s in session %s", state, phone, sessionID)
	} else {
		err := client.GetClient().SendPresence(presence)
		if err != nil {
			return fmt.Errorf("failed to send presence: %w", err)
		}
		m.logger.Debugf("Set global presence %s for session %s", state, sessionID)
	}

	return nil
}
