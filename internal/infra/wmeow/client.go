package wmeow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"zpmeow/internal/domain/session"
	"zpmeow/internal/infra/logging"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waCompanionReg"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waTypes "go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type WameowClient struct {
	sessionID    string
	client       *whatsmeow.Client
	eventHandler EventHandler
	sessionRepo  session.Repository
	logger       logging.Logger
	waLogger     waLog.Logger

	mu           sync.RWMutex
	status       session.Status
	lastActivity time.Time

	qrCode       string
	qrCodeBase64 string
	qrLoopActive bool

	eventHandlerID uint32

	ctx           context.Context
	cancel        context.CancelFunc
	killChannel   chan bool
	qrStopChannel chan bool

	maxRetries    int
	retryCount    int
	retryInterval time.Duration

	sessionHelper    *SessionHelper
	qrHelper         *QRCodeHelper
	connectionHelper *ConnectionHelper
}

type EventHandler interface {
	HandleEvent(interface{})
}

func NewWameowClient(sessionID string, container *sqlstore.Container, waLogger waLog.Logger, eventHandler EventHandler, sessionRepo session.Repository) (*WameowClient, error) {
	return NewWameowClientWithDeviceJID(sessionID, "", container, waLogger, eventHandler, sessionRepo)
}

func NewWameowClientWithDeviceJID(sessionID, expectedDeviceJID string, container *sqlstore.Container, waLogger waLog.Logger, eventHandler EventHandler, sessionRepo session.Repository) (*WameowClient, error) {
	if waLogger == nil {
		waLogger = waLog.Noop
	}

	appLogger := logging.GetLogger().Sub("wameow-client").Sub(sessionID)

	deviceStore := GetDeviceStoreForSession(sessionID, expectedDeviceJID, container)
	if deviceStore == nil {
		return nil, fmt.Errorf("failed to create device store for session %s", sessionID)
	}

	store.DeviceProps.PlatformType = waCompanionReg.DeviceProps_UNKNOWN.Enum()
	osName := "zpmeow"
	store.DeviceProps.Os = &osName

	waClient := whatsmeow.NewClient(deviceStore, waLogger)
	if waClient == nil {
		return nil, fmt.Errorf("failed to create WhatsApp client for session %s", sessionID)
	}

	ctx, cancel := context.WithCancel(context.Background())

	sessionHelper := NewSessionHelper(sessionRepo, appLogger)
	qrHelper := NewQRCodeHelper(appLogger)
	connectionHelper := NewConnectionHelper(appLogger)

	client := &WameowClient{
		sessionID:        sessionID,
		client:           waClient,
		eventHandler:     eventHandler,
		sessionRepo:      sessionRepo,
		logger:           appLogger,
		waLogger:         waLogger,
		status:           session.StatusDisconnected,
		lastActivity:     time.Now(),
		ctx:              ctx,
		cancel:           cancel,
		killChannel:      make(chan bool, 1),
		qrStopChannel:    make(chan bool, 1),
		maxRetries:       5,
		retryCount:       0,
		retryInterval:    30 * time.Second,
		sessionHelper:    sessionHelper,
		qrHelper:         qrHelper,
		connectionHelper: connectionHelper,
	}

	if eventHandler != nil {
		client.eventHandlerID = waClient.AddEventHandler(eventHandler.HandleEvent)
	}

	return client, nil
}

func (c *WameowClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := ValidateClientAndStore(c.client, c.sessionID, c.logger); err != nil {
		return err
	}

	if c.client.IsConnected() {
		c.logger.Debugf("Client already connected for session %s", c.sessionID)
		return nil
	}

	c.setStatus(session.StatusConnecting)
	c.sessionHelper.UpdateSessionStatus(c.sessionID, session.StatusConnecting)

	go c.startClientLoop()

	return nil
}

func (c *WameowClient) Disconnect() error {
	c.logger.Infof("Disconnecting client for session %s", c.sessionID)
	c.logger.Debugf("WameowClient.Disconnect: Attempting to acquire lock for session %s", c.sessionID)

	c.mu.Lock()
	defer func() {
		c.logger.Debugf("WameowClient.Disconnect: Releasing lock for session %s", c.sessionID)
		c.mu.Unlock()
	}()

	c.logger.Debugf("WameowClient.Disconnect: Lock acquired, starting disconnect for session %s", c.sessionID)

	c.logger.Debugf("WameowClient.Disconnect: Stopping QR loop for session %s", c.sessionID)
	c.stopQRLoop()
	c.logger.Debugf("WameowClient.Disconnect: QR loop stopped for session %s", c.sessionID)

	c.logger.Debugf("WameowClient.Disconnect: Calling SafeDisconnect for session %s", c.sessionID)
	c.connectionHelper.SafeDisconnect(c.client, c.sessionID)
	c.logger.Debugf("WameowClient.Disconnect: SafeDisconnect completed for session %s", c.sessionID)

	c.logger.Debugf("WameowClient.Disconnect: Cancelling context for session %s", c.sessionID)
	if c.cancel != nil {
		c.cancel()
		c.logger.Debugf("WameowClient.Disconnect: Context cancelled for session %s", c.sessionID)
	} else {
		c.logger.Debugf("WameowClient.Disconnect: No context to cancel for session %s", c.sessionID)
	}

	c.logger.Debugf("WameowClient.Disconnect: Setting status to disconnected for session %s", c.sessionID)
	c.setStatus(session.StatusDisconnected)
	c.logger.Debugf("WameowClient.Disconnect: Status set to disconnected for session %s", c.sessionID)

	c.logger.Debugf("WameowClient.Disconnect: Updating session status in database for session %s", c.sessionID)
	go func() {
		c.sessionHelper.UpdateSessionStatus(c.sessionID, session.StatusDisconnected)
		c.logger.Debugf("WameowClient.Disconnect: Database update completed for session %s", c.sessionID)
	}()

	c.logger.Debugf("WameowClient.Disconnect: Completed successfully for session %s", c.sessionID)
	return nil
}

func (c *WameowClient) GetQRCode() (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.qrCode == "" {
		return "", fmt.Errorf("no QR code available")
	}

	return c.qrCode, nil
}

func (c *WameowClient) GetQRCodeBase64() (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.qrCodeBase64 == "" {
		return "", fmt.Errorf("no QR code image available")
	}

	return c.qrCodeBase64, nil
}

func (c *WameowClient) PairPhone(phoneNumber string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.logger.Infof("Pairing phone %s for session %s", phoneNumber, c.sessionID)

	if phoneNumber == "" {
		return "", fmt.Errorf("phone number cannot be empty")
	}

	code, err := c.client.PairPhone(context.Background(), phoneNumber, true, whatsmeow.PairClientChrome, "Chrome (Linux)")
	if err != nil {
		c.logger.Errorf("Failed to pair phone for session %s: %v", c.sessionID, err)
		return "", fmt.Errorf("failed to pair phone: %w", err)
	}

	c.logger.Infof("Pairing code generated for session %s", c.sessionID)
	return code, nil
}

func (c *WameowClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.client.IsConnected()
}

func (c *WameowClient) GetStatus() session.Status {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.status
}

func (c *WameowClient) GetClient() *whatsmeow.Client {
	return c.client
}

func (c *WameowClient) GetJID() waTypes.JID {
	if c.client.Store.ID == nil {
		return waTypes.EmptyJID
	}
	return *c.client.Store.ID
}

func (c *WameowClient) Logout() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.logger.Infof("Logging out session %s", c.sessionID)

	err := c.client.Logout(context.Background())
	if err != nil {
		c.logger.Errorf("Failed to logout session %s: %v", c.sessionID, err)
		return fmt.Errorf("failed to logout: %w", err)
	}

	if c.client.IsConnected() {
		c.client.Disconnect()
	}

	c.setStatus(session.StatusDisconnected)
	c.logger.Infof("Successfully logged out session %s", c.sessionID)
	return nil
}

func (c *WameowClient) Reconnect(ctx context.Context) error {
	c.logger.Infof("Attempting to reconnect session %s", c.sessionID)

	if err := c.Disconnect(); err != nil {
		c.logger.Warnf("Error during disconnect before reconnect for session %s: %v", c.sessionID, err)
	}

	time.Sleep(2 * time.Second)
	return c.Connect()
}

func (c *WameowClient) GetLastActivity() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastActivity
}

func (c *WameowClient) UpdateActivity() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastActivity = time.Now()
}

func (c *WameowClient) IsLoggedIn() bool {
	return c.client.Store.ID != nil
}

func (c *WameowClient) GetSessionID() string {
	return c.sessionID
}

func (c *WameowClient) setStatus(status session.Status) {
	c.status = status
	c.lastActivity = time.Now()
	if status == session.StatusConnected || status == session.StatusDisconnected ||
		status == session.StatusConnecting || status == session.StatusError {
		c.logger.Infof("Session %s status: %s", c.sessionID, status)
	}
}

func (c *WameowClient) startClientLoop() {
	defer func() {
		if r := recover(); r != nil {
			c.logger.Errorf("Client loop panic for session %s: %v", c.sessionID, r)
		}
	}()

	if !IsDeviceRegistered(c.client) {
		c.logger.Infof("Device not registered for session %s, starting QR code process", c.sessionID)
		c.handleNewDeviceRegistration()
	} else {
		c.logger.Infof("Device already registered for session %s, connecting directly", c.sessionID)
		c.handleExistingDeviceConnection()
	}
}

func (c *WameowClient) handleNewDeviceRegistration() {
	qrChan, err := c.client.GetQRChannel(context.Background())
	if err != nil {
		c.logger.Errorf("Failed to get QR channel for session %s: %v", c.sessionID, err)
		c.setStatus(session.StatusDisconnected)
		return
	}

	err = c.client.Connect()
	if err != nil {
		c.logger.Errorf("Failed to connect client for session %s: %v", c.sessionID, err)
		c.setStatus(session.StatusDisconnected)
		return
	}

	c.handleQRLoop(qrChan)
}

func (c *WameowClient) handleExistingDeviceConnection() {
	c.logger.Infof("Connecting existing device for session %s", c.sessionID)

	err := c.client.Connect()
	if err != nil {
		c.logger.Errorf("Failed to connect client for session %s: %v", c.sessionID, err)
		c.setStatus(session.StatusDisconnected)
		c.sessionHelper.UpdateSessionStatus(c.sessionID, session.StatusDisconnected)
		return
	}

	time.Sleep(2 * time.Second)

	if c.client.IsConnected() {
		c.logger.Infof("Successfully connected session %s", c.sessionID)
		c.setStatus(session.StatusConnected)
		c.sessionHelper.UpdateSessionStatus(c.sessionID, session.StatusConnected)
	} else {
		c.logger.Warnf("Connection attempt completed but client not connected for session %s", c.sessionID)
		c.setStatus(session.StatusDisconnected)
		c.sessionHelper.UpdateSessionStatus(c.sessionID, session.StatusDisconnected)
	}
}

func (c *WameowClient) handleQRLoop(qrChan <-chan whatsmeow.QRChannelItem) {
	c.mu.Lock()
	c.qrLoopActive = true
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		c.qrLoopActive = false
		c.mu.Unlock()
	}()

	for {
		select {
		case <-c.ctx.Done():
			c.logger.Infof("QR loop cancelled for session %s", c.sessionID)
			return

		case <-c.qrStopChannel:
			c.logger.Infof("QR loop stopped for session %s", c.sessionID)
			return

		case evt, ok := <-qrChan:
			if !ok {
				c.logger.Infof("QR channel closed for session %s", c.sessionID)
				c.setStatus(session.StatusDisconnected)
				c.sessionHelper.UpdateSessionQRCode(c.sessionID, "")
				return
			}

			switch evt.Event {
			case "code":
				c.mu.Lock()
				c.qrCode = evt.Code
				c.qrCodeBase64 = c.qrHelper.GenerateQRCodeImage(evt.Code)
				c.mu.Unlock()

				c.qrHelper.DisplayQRCodeInTerminal(evt.Code, c.sessionID)
				c.logger.Infof("QR code generated for session %s", c.sessionID)
				c.setStatus(session.StatusConnecting)

				c.sessionHelper.UpdateSessionQRCode(c.sessionID, evt.Code)

			case "success":
				c.logger.Infof("QR code scanned successfully for session %s", c.sessionID)
				c.setStatus(session.StatusConnected)

				go c.persistQRSuccess()
				return

			case "timeout":
				c.logger.Warnf("QR code timeout for session %s", c.sessionID)
				c.mu.Lock()
				c.qrCode = ""
				c.qrCodeBase64 = ""
				c.mu.Unlock()

				c.setStatus(session.StatusDisconnected)

				c.sessionHelper.UpdateSessionQRCode(c.sessionID, "")
				return

			default:
				c.logger.Infof("QR event: %s for session %s", evt.Event, c.sessionID)
			}
		}
	}
}

func (c *WameowClient) stopQRLoop() {
	c.logger.Debugf("WameowClient.stopQRLoop: Attempting to acquire lock for session %s", c.sessionID)

	if c.qrLoopActive {
		c.logger.Debugf("WameowClient.stopQRLoop: QR loop is active, sending stop signal for session %s", c.sessionID)
		select {
		case c.qrStopChannel <- true:
			c.logger.Debugf("WameowClient.stopQRLoop: Stop signal sent for session %s", c.sessionID)
		default:
			c.logger.Debugf("WameowClient.stopQRLoop: Stop channel full, skipping for session %s", c.sessionID)
		}
	} else {
		c.logger.Debugf("WameowClient.stopQRLoop: QR loop not active for session %s", c.sessionID)
	}
}

func (c *WameowClient) persistQRSuccess() {
	if c.sessionRepo == nil {
		c.logger.Warnf("No session repository available for session %s", c.sessionID)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sessionEntity, err := c.sessionRepo.GetByID(ctx, c.sessionID)
	if err != nil {
		c.logger.Errorf("Failed to get session %s from database: %v", c.sessionID, err)
		return
	}

	var deviceJID string
	if c.client != nil && c.client.Store.ID != nil {
		deviceJID = c.client.Store.ID.String()
	}

	if deviceJID != "" {
		if err := c.sessionRepo.ValidateDeviceUniqueness(ctx, c.sessionID, deviceJID); err != nil {
			c.logger.Errorf("Device uniqueness validation failed for session %s: %v", c.sessionID, err)
			return
		}

		c.logger.Infof("Assigning device JID %s to session %s", deviceJID, c.sessionID)
		err := sessionEntity.Authenticate(deviceJID)
		if err != nil {
			c.logger.Errorf("Failed to authenticate session %s: %v", c.sessionID, err)
			return
		}
	}

	err = sessionEntity.SetConnected()
	if err != nil {
		c.logger.Errorf("Failed to set session %s as connected: %v", c.sessionID, err)
		return
	}

	if err := c.sessionRepo.Update(ctx, sessionEntity); err != nil {
		c.logger.Errorf("Failed to update session %s in database after QR scan: %v", c.sessionID, err)
		return
	}

	c.logger.Infof("Successfully updated session %s in database after QR scan: JID=%s, Status=%s", c.sessionID, deviceJID, session.StatusConnected)
}
