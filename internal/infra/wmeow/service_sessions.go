package wmeow

import (
	"context"
	"fmt"
	"time"

	"zpmeow/internal/domain/session"
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
		m.logger.Warnf("StopClient: Client not found for session %s", sessionID)
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	m.logger.Debugf("StopClient: Found client for session %s, disconnecting...", sessionID)
	client.Disconnect()

	m.mu.Lock()
	delete(m.clients, sessionID)
	m.mu.Unlock()

	m.logger.Infof("Client stopped and removed for session %s", sessionID)
	return nil
}

func (m *MeowService) LogoutClient(sessionID string) error {
	m.logger.Infof("Logging out client for session %s", sessionID)

	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if err := client.Logout(); err != nil {
		return fmt.Errorf("failed to logout client for session %s: %w", sessionID, err)
	}

	m.mu.Lock()
	delete(m.clients, sessionID)
	m.mu.Unlock()

	m.logger.Infof("Client logged out and removed for session %s", sessionID)
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

func (m *MeowService) ConnectOnStartup(ctx context.Context) error {
	m.logger.Info("Connecting sessions on startup")

	sessions, err := m.sessions.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to get sessions: %w", err)
	}

	for _, sess := range sessions {
		if sess.IsConnected() || sess.IsConnecting() {
			m.logger.Infof("Auto-connecting session %s", sess.ID().Value())
			
			if err := m.StartClient(sess.ID().Value()); err != nil {
				m.logger.Errorf("Failed to auto-connect session %s: %v", sess.ID().Value(), err)
				
				// Update session status to error
				if updateErr := sess.SetError(fmt.Sprintf("Auto-connect failed: %v", err)); updateErr != nil {
					m.logger.Errorf("Failed to update session status to error: %v", updateErr)
				}
				
				if saveErr := m.sessions.Update(ctx, sess); saveErr != nil {
					m.logger.Errorf("Failed to save session with error status: %v", saveErr)
				}
				continue
			}
			
			// Give some time between connections to avoid overwhelming
			time.Sleep(1 * time.Second)
		}
	}

	m.logger.Info("Startup connection process completed")
	return nil
}

func (m *MeowService) ConnectSession(ctx context.Context, sessionID string) (string, error) {
	m.logger.Infof("Connecting session %s", sessionID)

	if err := m.StartClient(sessionID); err != nil {
		return "", fmt.Errorf("failed to start client for session %s: %w", sessionID, err)
	}

	qrCode, err := m.GetQRCode(sessionID)
	if err != nil {
		m.logger.Warnf("Failed to get QR code for session %s: %v", sessionID, err)
		return "", nil
	}

	return qrCode, nil
}

func (m *MeowService) DisconnectSession(ctx context.Context, sessionID string) error {
	m.logger.Infof("Disconnecting session %s", sessionID)

	if err := m.StopClient(sessionID); err != nil {
		return fmt.Errorf("failed to stop client for session %s: %w", sessionID, err)
	}

	return nil
}

// Helper methods for session management

func (m *MeowService) getClient(sessionID string) *WameowClient {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.clients[sessionID]
}

func (m *MeowService) getOrCreateClient(sessionID string) *WameowClient {
	client := m.getClient(sessionID)
	if client != nil {
		return client
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if client, exists := m.clients[sessionID]; exists {
		return client
	}

	client = m.createNewClient(sessionID)
	if client != nil {
		m.clients[sessionID] = client
	}

	return client
}

func (m *MeowService) createNewClient(sessionID string) *WameowClient {
	sessionConfig := m.loadSessionConfiguration(sessionID)

	var eventProcessor *EventProcessor
	if m.chatwootIntegration != nil && m.chatwootRepo != nil {
		eventProcessor = NewEventProcessorWithChatwoot(sessionID, m.sessions, m.chatwootIntegration, m.chatwootRepo, m.messageRepo, m.chatRepo, m.webhookRepo)
	} else {
		eventProcessor = NewEventProcessor(sessionID, m.sessions, m.messageRepo, m.chatRepo, m.webhookRepo)
	}

	client, err := NewWameowClientWithDeviceJID(
		sessionID,
		sessionConfig.deviceJID,
		m.container,
		m.waLogger,
		eventProcessor,
		m.sessions,
	)
	if err != nil {
		m.logger.Errorf("Failed to create WameowClient for session %s: %v", sessionID, err)
		return nil
	}

	return client
}

type sessionConfiguration struct {
	deviceJID string
}

func (m *MeowService) loadSessionConfiguration(sessionID string) *sessionConfiguration {
	config := &sessionConfiguration{}

	// Try to load session from repository
	sess, err := m.sessions.GetByID(context.Background(), session.SessionID{})
	if err != nil {
		m.logger.Warnf("Failed to load session %s from repository: %v", sessionID, err)
		return config
	}

	if sess != nil && sess.IsAuthenticated() {
		config.deviceJID = sess.DeviceJID().Value()
	}

	return config
}
