package wmeow

import (
	"context"
	"fmt"
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

	// For now, just log the presence update
	if phone != "" {
		m.logger.Debugf("SendChatPresence: %s to %s for session %s", state, phone, sessionID)
	} else {
		m.logger.Debugf("SendPresence: %s for session %s", state, sessionID)
	}

	return nil
}
