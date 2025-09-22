package ports

import (
	"context"
	"fmt"
	"time"

	"zpmeow/internal/domain/session"
)

type CacheService interface {
	SessionCache
	QRCodeCache
	CredentialCache
	HealthChecker
}

type SessionCache interface {
	GetSession(ctx context.Context, sessionID string) (*session.Session, error)

	SetSession(ctx context.Context, sessionID string, sess *session.Session, ttl time.Duration) error

	DeleteSession(ctx context.Context, sessionID string) error

	GetSessionByName(ctx context.Context, name string) (*session.Session, error)

	SetSessionByName(ctx context.Context, name string, sess *session.Session, ttl time.Duration) error

	DeleteSessionByName(ctx context.Context, name string) error
}

type QRCodeCache interface {
	GetQRCode(ctx context.Context, sessionID string) (string, error)

	SetQRCode(ctx context.Context, sessionID string, qrCode string) error

	DeleteQRCode(ctx context.Context, sessionID string) error

	GetQRCodeBase64(ctx context.Context, sessionID string) (string, error)

	SetQRCodeBase64(ctx context.Context, sessionID string, qrCodeBase64 string) error
}

type CredentialCache interface {
	GetDeviceJID(ctx context.Context, sessionID string) (string, error)

	SetDeviceJID(ctx context.Context, sessionID string, deviceJID string, ttl time.Duration) error

	DeleteDeviceJID(ctx context.Context, sessionID string) error

	GetSessionStatus(ctx context.Context, sessionID string) (session.Status, error)

	SetSessionStatus(ctx context.Context, sessionID string, status session.Status, ttl time.Duration) error

	DeleteSessionStatus(ctx context.Context, sessionID string) error
}

type HealthChecker interface {
	Ping(ctx context.Context) error

	GetStats(ctx context.Context) (CacheStats, error)
}

type CacheStats struct {
	Connected     bool   `json:"connected"`
	TotalKeys     int64  `json:"total_keys"`
	UsedMemory    string `json:"used_memory"`
	HitRate       string `json:"hit_rate"`
	MissRate      string `json:"miss_rate"`
	Uptime        string `json:"uptime"`
	Version       string `json:"version"`
	LastError     string `json:"last_error,omitempty"`
	LastErrorTime string `json:"last_error_time,omitempty"`
}

type CacheError struct {
	Operation string
	Key       string
	Err       error
}

func (e *CacheError) Error() string {
	if e.Key != "" {
		return fmt.Sprintf("cache %s operation failed for key '%s': %v", e.Operation, e.Key, e.Err)
	}
	return fmt.Sprintf("cache %s operation failed: %v", e.Operation, e.Err)
}

func (e *CacheError) Unwrap() error {
	return e.Err
}

func NewCacheError(operation, key string, err error) *CacheError {
	return &CacheError{
		Operation: operation,
		Key:       key,
		Err:       err,
	}
}

const (
	SessionKeyPrefix       = "session:"
	SessionNameKeyPrefix   = "session_name:"
	QRCodeKeyPrefix        = "qr:"
	QRCodeBase64KeyPrefix  = "qr_base64:"
	DeviceJIDKeyPrefix     = "device_jid:"
	SessionStatusKeyPrefix = "session_status:"
)

const (
	DefaultSessionTTL    = 24 * time.Hour   // Sessions cached for 24 hours
	DefaultQRCodeTTL     = 60 * time.Second // QR codes cached for 60 seconds
	DefaultCredentialTTL = 6 * time.Hour    // Credentials cached for 6 hours
	DefaultStatusTTL     = 5 * time.Minute  // Status cached for 5 minutes
)
