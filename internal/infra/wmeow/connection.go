package wmeow

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"zpmeow/internal/domain/session"
	"zpmeow/internal/infra/logging"

	"github.com/mdp/qrterminal/v3"
	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waTypes "go.mau.fi/whatsmeow/types"
)

// connectionManager implements ConnectionManager interface
type connectionManager struct {
	logger logging.Logger
}

func NewConnectionManager(logger logging.Logger) *connectionManager {
	return &connectionManager{
		logger: logger,
	}
}

func (c *connectionManager) SafeConnect(client *whatsmeow.Client, sessionID string) error {
	if err := ValidateClientAndStore(client, sessionID); err != nil {
		return NewConnectionError(sessionID, "connect", err)
	}

	if client.IsConnected() {
		c.logger.Debugf("Client already connected for session %s", sessionID)
		return nil
	}

	c.logger.Infof("Connecting client for session %s", sessionID)
	err := client.Connect()
	if err != nil {
		return NewConnectionError(sessionID, "connect", err)
	}

	return nil
}

func (c *connectionManager) SafeDisconnect(client *whatsmeow.Client, sessionID string) {
	if client == nil {
		c.logger.Warnf("Cannot disconnect nil client for session %s", sessionID)
		return
	}

	if !client.IsConnected() {
		c.logger.Debugf("Client already disconnected for session %s", sessionID)
		return
	}

	c.logger.Infof("Disconnecting client for session %s", sessionID)
	client.Disconnect()
}

// qrCodeGenerator implements QRCodeGenerator interface
type qrCodeGenerator struct {
	logger logging.Logger
}

func NewQRCodeGenerator(logger logging.Logger) *qrCodeGenerator {
	return &qrCodeGenerator{
		logger: logger,
	}
}

func (q *qrCodeGenerator) GenerateQRCodeImage(qrText string) string {
	if qrText == "" {
		q.logger.Warn("Empty QR text provided")
		return ""
	}

	png, err := qrcode.Encode(qrText, qrcode.Medium, 256)
	if err != nil {
		q.logger.Errorf("Failed to generate QR code: %v", err)
		return ""
	}

	base64String := base64.StdEncoding.EncodeToString(png)
	return "data:image/png;base64," + base64String
}

func (q *qrCodeGenerator) DisplayQRCodeInTerminal(qrCode, sessionID string) {
	if qrCode == "" {
		q.logger.Warnf("Empty QR code for session %s", sessionID)
		return
	}

	q.logger.Infof("QR Code for session %s:", sessionID)
	qrterminal.Generate(qrCode, qrterminal.L, nil)
}

// sessionManager implements SessionStateManager interface
type sessionManager struct {
	sessionRepo session.Repository
	logger      logging.Logger
}

func NewSessionManager(sessionRepo session.Repository, logger logging.Logger) *sessionManager {
	return &sessionManager{
		sessionRepo: sessionRepo,
		logger:      logger,
	}
}

func (s *sessionManager) UpdateStatus(sessionID string, status session.Status) {
	s.logger.Debugf("Updating session %s status to %s", sessionID, status)

	if s.sessionRepo == nil {
		s.logger.Warnf("No session repository available for session %s", sessionID)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sessionEntity, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		s.logger.Errorf("Failed to get session %s: %v", sessionID, err)
		return
	}

	// Check if status is already the same
	currentStatus := sessionEntity.Status()
	if currentStatus == status {
		s.logger.Debugf("Session %s already has status %s, skipping update", sessionID, status)
		return
	}

	// Validate status transition
	if err := session.ValidateSessionStatus(currentStatus, status); err != nil {
		s.logger.Warnf("Invalid status transition for session %s: %s -> %s: %v", sessionID, currentStatus, status, err)
		return
	}

	if err := sessionEntity.SetStatus(status); err != nil {
		s.logger.Errorf("Failed to set status for session %s: %v", sessionID, err)
		return
	}

	if err := s.sessionRepo.Update(ctx, sessionEntity); err != nil {
		s.logger.Errorf("Failed to update session %s in database: %v", sessionID, err)
		return
	}

	s.logger.Infof("Successfully updated session %s status from %s to %s", sessionID, currentStatus, status)
}

func (s *sessionManager) UpdateQRCode(sessionID string, qrCode string) {
	s.logger.Debugf("Updating QR code for session %s", sessionID)

	if s.sessionRepo == nil {
		s.logger.Warnf("No session repository available for session %s", sessionID)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sessionEntity, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		s.logger.Errorf("Failed to get session %s: %v", sessionID, err)
		return
	}

	if err := sessionEntity.SetQRCode(qrCode); err != nil {
		s.logger.Errorf("Failed to set QR code for session %s: %v", sessionID, err)
		return
	}

	if err := s.sessionRepo.Update(ctx, sessionEntity); err != nil {
		s.logger.Errorf("Failed to update session %s in database: %v", sessionID, err)
		return
	}

	s.logger.Infof("Successfully updated session %s QR code", sessionID)
}

func (s *sessionManager) GetSession(sessionID string) (*session.Session, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.sessionRepo.GetByID(ctx, sessionID)
}

// Device store utilities
func GetOrCreateDeviceStore(sessionID string, container *sqlstore.Container) *store.Device {
	return GetDeviceStoreForSession(sessionID, "", container)
}

func GetDeviceStoreForSession(sessionID, expectedDeviceJID string, container *sqlstore.Container) *store.Device {
	var deviceStore *store.Device

	if expectedDeviceJID != "" {
		// Try to get existing device store first
		jid, err := waTypes.ParseJID(expectedDeviceJID)
		if err != nil {
			fmt.Printf("Failed to parse expected JID %s: %v, creating new device\n", expectedDeviceJID, err)
		} else {
			ctx := context.Background()
			deviceStore, err = container.GetDevice(ctx, jid)
			if err != nil {
				fmt.Printf("Failed to get device store for expected JID %s: %v, creating new device\n", expectedDeviceJID, err)
			}
		}

		// If we couldn't get the existing device, create a new one
		if deviceStore == nil {
			fmt.Printf("Device store not found for expected JID %s, creating new device\n", expectedDeviceJID)
			deviceStore = container.NewDevice()
		}
	} else {
		// Create new device store
		deviceStore = container.NewDevice()
	}

	if deviceStore == nil {
		fmt.Printf("Failed to create device store for session %s\n", sessionID)
		return nil
	}

	return deviceStore
}

// Connection retry logic
type RetryConfig struct {
	MaxRetries    int
	RetryInterval time.Duration
}

func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:    5,
		RetryInterval: 30 * time.Second,
	}
}

func (c *connectionManager) ConnectWithRetry(client *whatsmeow.Client, sessionID string, config *RetryConfig) error {
	if config == nil {
		config = DefaultRetryConfig()
	}

	var lastErr error
	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			c.logger.Infof("Retry attempt %d/%d for session %s", attempt, config.MaxRetries, sessionID)
			time.Sleep(config.RetryInterval)
		}

		err := c.SafeConnect(client, sessionID)
		if err == nil {
			c.logger.Infof("Successfully connected session %s on attempt %d", sessionID, attempt+1)
			return nil
		}

		lastErr = err
		c.logger.Warnf("Connection attempt %d failed for session %s: %v", attempt+1, sessionID, err)
	}

	return fmt.Errorf("failed to connect after %d attempts: %w", config.MaxRetries+1, lastErr)
}
