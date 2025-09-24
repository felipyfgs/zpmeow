package cache

import (
	"context"
	"fmt"
	"time"

	"zpmeow/internal/application/ports"
	"zpmeow/internal/domain/session"
)

type NoOpCacheService struct{}

func NewNoOpCacheService() ports.CacheManager {
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

func (n *NoOpCacheService) SetQRCode(ctx context.Context, sessionID string, qrCode string, ttl time.Duration) error {
	return nil
}

func (n *NoOpCacheService) DeleteQRCode(ctx context.Context, sessionID string) error {
	return nil
}

func (n *NoOpCacheService) GetQRCodeBase64(ctx context.Context, sessionID string) (string, error) {
	return "", ports.NewCacheError("get", "qr_base64:"+sessionID, fmt.Errorf("cache is disabled"))
}

func (n *NoOpCacheService) SetQRCodeBase64(ctx context.Context, sessionID string, qrCodeBase64 string, ttl time.Duration) error {
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

func (n *NoOpCacheService) Clear(ctx context.Context) error {
	return ports.NewCacheError("clear", "all", fmt.Errorf("cache disabled"))
}

// Generic cache methods
func (n *NoOpCacheService) Get(ctx context.Context, key string) (interface{}, error) {
	return nil, ports.NewCacheError("get", key, fmt.Errorf("cache disabled"))
}

func (n *NoOpCacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return ports.NewCacheError("set", key, fmt.Errorf("cache disabled"))
}

func (n *NoOpCacheService) Delete(ctx context.Context, key string) error {
	return ports.NewCacheError("delete", key, fmt.Errorf("cache disabled"))
}

func (n *NoOpCacheService) GetStats(ctx context.Context) (*ports.CacheStats, error) {
	return &ports.CacheStats{
		Hits:        0,
		Misses:      0,
		Keys:        0,
		Memory:      0,
		Connections: 0,
	}, nil
}
