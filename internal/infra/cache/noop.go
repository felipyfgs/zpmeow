package cache

import (
	"context"
	"fmt"
	"time"

	"zpmeow/internal/application/ports"
	"zpmeow/internal/domain/session"
)

type NoOpCacheService struct{}

func NewNoOpCacheService() ports.CacheService {
	return &NoOpCacheService{}
}

func (n *NoOpCacheService) GetSession(ctx context.Context, sessionID string) (*session.Session, error) {
	return nil, ports.NewCacheError("get", "session:"+sessionID, fmt.Errorf("cache is disabled"))
}

func (n *NoOpCacheService) SetSession(ctx context.Context, sessionID string, sess *session.Session, ttl time.Duration) error {
	return nil
}

func (n *NoOpCacheService) DeleteSession(ctx context.Context, sessionID string) error {
	return nil
}

func (n *NoOpCacheService) GetSessionByName(ctx context.Context, name string) (*session.Session, error) {
	return nil, ports.NewCacheError("get", "session_name:"+name, fmt.Errorf("cache is disabled"))
}

func (n *NoOpCacheService) SetSessionByName(ctx context.Context, name string, sess *session.Session, ttl time.Duration) error {
	return nil
}

func (n *NoOpCacheService) DeleteSessionByName(ctx context.Context, name string) error {
	return nil
}

func (n *NoOpCacheService) GetQRCode(ctx context.Context, sessionID string) (string, error) {
	return "", ports.NewCacheError("get", "qr:"+sessionID, fmt.Errorf("cache is disabled"))
}

func (n *NoOpCacheService) SetQRCode(ctx context.Context, sessionID string, qrCode string) error {
	return nil
}

func (n *NoOpCacheService) DeleteQRCode(ctx context.Context, sessionID string) error {
	return nil
}

func (n *NoOpCacheService) GetQRCodeBase64(ctx context.Context, sessionID string) (string, error) {
	return "", ports.NewCacheError("get", "qr_base64:"+sessionID, fmt.Errorf("cache is disabled"))
}

func (n *NoOpCacheService) SetQRCodeBase64(ctx context.Context, sessionID string, qrCodeBase64 string) error {
	return nil
}

func (n *NoOpCacheService) GetDeviceJID(ctx context.Context, sessionID string) (string, error) {
	return "", ports.NewCacheError("get", "device_jid:"+sessionID, fmt.Errorf("cache is disabled"))
}

func (n *NoOpCacheService) SetDeviceJID(ctx context.Context, sessionID string, deviceJID string, ttl time.Duration) error {
	return nil
}

func (n *NoOpCacheService) DeleteDeviceJID(ctx context.Context, sessionID string) error {
	return nil
}

func (n *NoOpCacheService) GetSessionStatus(ctx context.Context, sessionID string) (session.Status, error) {
	return session.StatusDisconnected, ports.NewCacheError("get", "session_status:"+sessionID, fmt.Errorf("cache is disabled"))
}

func (n *NoOpCacheService) SetSessionStatus(ctx context.Context, sessionID string, status session.Status, ttl time.Duration) error {
	return nil
}

func (n *NoOpCacheService) DeleteSessionStatus(ctx context.Context, sessionID string) error {
	return nil
}

func (n *NoOpCacheService) Ping(ctx context.Context) error {
	return fmt.Errorf("cache is disabled")
}

func (n *NoOpCacheService) GetStats(ctx context.Context) (ports.CacheStats, error) {
	return ports.CacheStats{
		Connected:     false,
		TotalKeys:     0,
		UsedMemory:    "0B",
		HitRate:       "0%",
		MissRate:      "100%",
		Uptime:        "0s",
		Version:       "No-Op Cache",
		LastError:     "Cache is disabled",
		LastErrorTime: time.Now().Format(time.RFC3339),
	}, nil
}
