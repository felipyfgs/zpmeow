package wmeow

import (
	"context"
	"fmt"
	"time"

	"zpmeow/internal/domain/session"

	"go.mau.fi/whatsmeow"
)

// SessionManager methods - gestão de sessões e conexões

func (m *MeowService) StartClient(sessionID string) error {
	m.logger.Infof("Starting client for session %s", sessionID)
	client := m.getOrCreateClient(sessionID)
	if client == nil {
		return fmt.Errorf("failed to create or get client for session %s", sessionID)
	}
	return client.Connect()
}

func (m *MeowService) StopClient(sessionID string) error {
	m.logger.Infof("Stopping client for session %s", sessionID)
	m.logger.Debugf("StopClient: Looking for client for session %s", sessionID)

	client := m.getClient(sessionID)
	if client == nil {
		m.logger.Debugf("StopClient: No client found for session %s", sessionID)
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	m.logger.Debugf("StopClient: Found client for session %s, calling Disconnect()", sessionID)
	if err := client.Disconnect(); err != nil {
		m.logger.Debugf("StopClient: Disconnect() failed for session %s: %v", sessionID, err)
		return fmt.Errorf("failed to disconnect client: %w", err)
	}

	m.logger.Debugf("StopClient: Disconnect() succeeded for session %s, removing client", sessionID)
	m.removeClient(sessionID)
	m.logger.Debugf("StopClient: Completed successfully for session %s", sessionID)
	return nil
}

func (m *MeowService) LogoutClient(sessionID string) error {
	m.logger.Infof("Logging out client for session %s", sessionID)
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if err := client.Logout(); err != nil {
		return fmt.Errorf("failed to logout client: %w", err)
	}

	m.removeClient(sessionID)
	return nil
}

func (m *MeowService) GetQRCode(sessionID string) (string, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return "", fmt.Errorf("client not found for session %s", sessionID)
	}

	qrCode, err := client.GetQRCode()
	if err != nil {
		return "", fmt.Errorf("failed to get QR code for session %s: %w", sessionID, err)
	}

	return qrCode, nil
}

func (m *MeowService) PairPhone(sessionID, phoneNumber string) (string, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return "", fmt.Errorf("client not found for session %s", sessionID)
	}

	code, err := client.PairPhone(phoneNumber)
	if err != nil {
		return "", fmt.Errorf("failed to pair phone for session %s: %w", sessionID, err)
	}

	return code, nil
}

func (m *MeowService) IsClientConnected(sessionID string) bool {
	client := m.getClient(sessionID)
	if client == nil {
		return false
	}
	return client.IsConnected()
}

func (m *MeowService) ConnectSession(ctx context.Context, sessionID string) (string, error) {
	m.logger.Infof("Connecting session %s", sessionID)

	client := m.getOrCreateClient(sessionID)
	if client == nil {
		return fmt.Errorf("failed to create client for session %s", sessionID)
	}

	if err := client.Connect(); err != nil {
		return "", fmt.Errorf("failed to connect session %s: %w", sessionID, err)
	}

	m.logger.Infof("Session %s connected successfully", sessionID)
	return "connected", nil
}

func (m *MeowService) DisconnectSession(ctx context.Context, sessionID string) error {
	m.logger.Infof("Disconnecting session %s", sessionID)

	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if err := client.Disconnect(); err != nil {
		return fmt.Errorf("failed to disconnect session %s: %w", sessionID, err)
	}

	m.logger.Infof("Session %s disconnected successfully", sessionID)
	return nil
}

// Internal helper for session configuration (different from service.go)
func (m *MeowService) loadSessionConfigurationInternal(sessionID string) *sessionConfiguration {
	config := &sessionConfiguration{}

	// Try to load session from repository
	sess, err := m.sessions.GetByID(context.Background(), session.SessionID{})
	if err != nil {
		m.logger.Warnf("Failed to load session %s from repository: %v", sessionID, err)
		return config
	}

	if sess != nil && sess.IsAuthenticated() {
		config.deviceJID = sess.ID().Value()
	}

	return config
}
