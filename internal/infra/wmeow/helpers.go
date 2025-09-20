package wmeow

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"sync"
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

type SessionHelper struct {
	sessionRepo session.Repository
	logger      logging.Logger
}

func NewSessionHelper(sessionRepo session.Repository, logger logging.Logger) *SessionHelper {
	return &SessionHelper{
		sessionRepo: sessionRepo,
		logger:      logger,
	}
}

func (h *SessionHelper) UpdateSessionStatus(sessionID string, status session.Status) {
	h.logger.Debugf("SessionHelper.UpdateSessionStatus: Starting update for session %s to status %s", sessionID, status)

	if h.sessionRepo == nil {
		h.logger.Warnf("No session repository available for session %s", sessionID)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	h.logger.Debugf("SessionHelper.UpdateSessionStatus: Getting session %s from database", sessionID)
	sessionEntity, err := h.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		h.logger.Errorf("Failed to get session %s from database: %v", sessionID, err)
		return
	}

	h.logger.Debugf("SessionHelper.UpdateSessionStatus: Current status for session %s is %s, target status is %s", sessionID, sessionEntity.Status(), status)

	switch status {
	case session.StatusConnected:
		if sessionEntity.IsConnected() {
			h.logger.Debugf("SessionHelper.UpdateSessionStatus: Session %s already connected, no-op", sessionID)
		} else if sessionEntity.IsConnecting() {
			h.logger.Debugf("SessionHelper.UpdateSessionStatus: Setting session %s as connected", sessionID)
			if err := sessionEntity.SetConnected(); err != nil {
				h.logger.Errorf("Failed to set session %s as connected: %v", sessionID, err)
				return
			}
		} else {
			h.logger.Warnf("Skipping SetConnected for session %s: current status=%s", sessionID, sessionEntity.Status())
			return
		}
	case session.StatusDisconnected:
		h.logger.Debugf("SessionHelper.UpdateSessionStatus: Disconnecting session %s", sessionID)
		err := sessionEntity.Disconnect("status update")
		if err != nil {
			h.logger.Errorf("Failed to disconnect session %s: %v", sessionID, err)
			return
		}
	case session.StatusConnecting:
		if sessionEntity.IsConnected() || sessionEntity.IsConnecting() {
			h.logger.Debugf("SessionHelper.UpdateSessionStatus: Session %s already connected/connecting, no-op", sessionID)
		} else {
			h.logger.Debugf("SessionHelper.UpdateSessionStatus: Setting session %s as connecting", sessionID)
			if err := sessionEntity.Connect(); err != nil {
				h.logger.Errorf("Failed to set session %s as connecting: %v", sessionID, err)
				return
			}
		}
	}

	h.logger.Debugf("SessionHelper.UpdateSessionStatus: Updating session %s in database", sessionID)
	if err := h.sessionRepo.Update(ctx, sessionEntity); err != nil {
		h.logger.Errorf("Failed to update session %s status to %s in database: %v", sessionID, status, err)
		return
	}

	h.logger.Infof("Successfully updated session %s status to %s in database", sessionID, status)
	h.logger.Debugf("SessionHelper.UpdateSessionStatus: Completed successfully for session %s", sessionID)
}

func (h *SessionHelper) UpdateSessionQRCode(sessionID string, qrCode string) {
	if h.sessionRepo == nil {
		h.logger.Warnf("No session repository available for session %s", sessionID)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sessionEntity, err := h.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		h.logger.Errorf("Failed to get session %s from database: %v", sessionID, err)
		return
	}

	if err := sessionEntity.SetQRCode(qrCode); err != nil {
		h.logger.Errorf("Failed to set QR code for session %s: %v", sessionID, err)
		return
	}

	if err := h.sessionRepo.Update(ctx, sessionEntity); err != nil {
		h.logger.Errorf("Failed to update session %s QR code in database: %v", sessionID, err)
		return
	}

	h.logger.Infof("Successfully updated session %s QR code in database", sessionID)
}

func ValidateClientAndStore(client *whatsmeow.Client, sessionID string, logger logging.Logger) error {
	if client == nil {
		return fmt.Errorf("WhatsApp client is nil for session %s", sessionID)
	}

	if client.Store == nil {
		return fmt.Errorf("WhatsApp client store is nil for session %s", sessionID)
	}

	return nil
}

func GetOrCreateDeviceStore(sessionID string, container *sqlstore.Container) *store.Device {
	return GetDeviceStoreForSession(sessionID, "", container)
}

func GetDeviceStoreForSession(sessionID, expectedDeviceJID string, container *sqlstore.Container) *store.Device {
	var deviceStore *store.Device
	var err error

	if expectedDeviceJID != "" {
		jid, ok := parseJID(expectedDeviceJID)
		if ok {
			deviceStore, err = container.GetDevice(context.Background(), jid)
			if err != nil {
				fmt.Printf("Failed to get device for JID %s: %v\n", expectedDeviceJID, err)
				deviceStore = container.NewDevice()
			} else {
				fmt.Printf("Successfully retrieved existing device for JID %s\n", expectedDeviceJID)
			}
		} else {
			fmt.Printf("Failed to parse JID %s, creating new device\n", expectedDeviceJID)
			deviceStore = container.NewDevice()
		}
	} else {
		fmt.Printf("No device JID provided for session %s, creating new device\n", sessionID)
		deviceStore = container.NewDevice()
	}

	if deviceStore == nil {
		fmt.Printf("Device store is nil, creating fallback device\n")
		deviceStore = container.NewDevice()
	}

	return deviceStore
}

func parseJID(arg string) (waTypes.JID, bool) {
	if arg[0] == '+' {
		arg = arg[1:]
	}
	if !strings.ContainsRune(arg, '@') {
		return waTypes.NewJID(arg, waTypes.DefaultUserServer), true
	} else {
		recipient, err := waTypes.ParseJID(arg)
		if err != nil {
			fmt.Printf("Invalid JID: %v\n", err)
			return recipient, false
		} else if recipient.User == "" {
			fmt.Printf("Invalid JID no server specified\n")
			return recipient, false
		}
		return recipient, true
	}
}

type QRCodeHelper struct {
	logger logging.Logger
}

func NewQRCodeHelper(logger logging.Logger) *QRCodeHelper {
	return &QRCodeHelper{
		logger: logger,
	}
}

func (h *QRCodeHelper) GenerateQRCodeImage(qrText string) string {
	qrPNG, err := qrcode.Encode(qrText, qrcode.Medium, 256)
	if err != nil {
		h.logger.Errorf("Failed to generate QR code image: %v", err)
		return ""
	}

	base64Str := base64.StdEncoding.EncodeToString(qrPNG)
	return "data:image/png;base64," + base64Str
}

func (h *QRCodeHelper) DisplayQRCodeInTerminal(qrCode, sessionID string) {
	fmt.Printf("\n=== QR Code for Session %s ===\n", sessionID)
	qrterminal.GenerateHalfBlock(qrCode, qrterminal.L, os.Stdout)
	fmt.Printf("QR Code String: %s\n", qrCode)
	fmt.Printf("=== End QR Code ===\n\n")
}

type ConnectionHelper struct {
	logger logging.Logger
}

func NewConnectionHelper(logger logging.Logger) *ConnectionHelper {
	return &ConnectionHelper{
		logger: logger,
	}
}

func (h *ConnectionHelper) SafeConnect(client *whatsmeow.Client, sessionID string) error {
	if err := ValidateClientAndStore(client, sessionID, h.logger); err != nil {
		return err
	}

	if client.IsConnected() {
		h.logger.Debugf("Client already connected for session %s", sessionID)
		return nil
	}

	h.logger.Infof("Connecting client for session %s", sessionID)
	return client.Connect()
}

func (h *ConnectionHelper) SafeDisconnect(client *whatsmeow.Client, sessionID string) {
	h.logger.Debugf("ConnectionHelper.SafeDisconnect: Starting for session %s", sessionID)

	if client == nil {
		h.logger.Debugf("ConnectionHelper.SafeDisconnect: Client is nil for session %s", sessionID)
		return
	}

	if !client.IsConnected() {
		h.logger.Debugf("ConnectionHelper.SafeDisconnect: Client not connected for session %s", sessionID)
		return
	}

	h.logger.Infof("Disconnecting client for session %s", sessionID)
	h.logger.Debugf("ConnectionHelper.SafeDisconnect: Calling client.Disconnect() for session %s", sessionID)
	client.Disconnect()
	h.logger.Debugf("ConnectionHelper.SafeDisconnect: Completed for session %s", sessionID)
}

func IsDeviceRegistered(client *whatsmeow.Client) bool {
	return client != nil && client.Store != nil && client.Store.ID != nil
}

type StatusHelper struct {
	mu     *sync.RWMutex
	status session.Status
	logger logging.Logger
}

func NewStatusHelper(logger logging.Logger) *StatusHelper {
	return &StatusHelper{
		mu:     &sync.RWMutex{},
		status: session.StatusDisconnected,
		logger: logger,
	}
}

func (h *StatusHelper) SetStatus(status session.Status, sessionID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.status = status
	if status == session.StatusConnected || status == session.StatusDisconnected ||
		status == session.StatusConnecting || status == session.StatusError {
		h.logger.Infof("Session %s status: %s", sessionID, status)
	}
}

func (h *StatusHelper) GetStatus() session.Status {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.status
}
