package ports

import (
	"context"
	"fmt"
	"time"

	"zpmeow/internal/domain/session"
)

// CacheService defines the interface for caching operations
type CacheService interface {
	SessionCache
	QRCodeCache
	CredentialCache
	HealthChecker
}

// SessionCache handles session-related caching
type SessionCache interface {
	// GetSession retrieves a session from cache
	GetSession(ctx context.Context, sessionID string) (*session.Session, error)
	
	// SetSession stores a session in cache with TTL
	SetSession(ctx context.Context, sessionID string, sess *session.Session, ttl time.Duration) error
	
	// DeleteSession removes a session from cache
	DeleteSession(ctx context.Context, sessionID string) error
	
	// GetSessionByName retrieves a session by name from cache
	GetSessionByName(ctx context.Context, name string) (*session.Session, error)
	
	// SetSessionByName stores a session by name in cache
	SetSessionByName(ctx context.Context, name string, sess *session.Session, ttl time.Duration) error
	
	// DeleteSessionByName removes a session by name from cache
	DeleteSessionByName(ctx context.Context, name string) error
}

// QRCodeCache handles QR code caching with short TTL
type QRCodeCache interface {
	// GetQRCode retrieves QR code from cache
	GetQRCode(ctx context.Context, sessionID string) (string, error)
	
	// SetQRCode stores QR code in cache with short TTL (60 seconds)
	SetQRCode(ctx context.Context, sessionID string, qrCode string) error
	
	// DeleteQRCode removes QR code from cache
	DeleteQRCode(ctx context.Context, sessionID string) error
	
	// GetQRCodeBase64 retrieves base64 QR code from cache
	GetQRCodeBase64(ctx context.Context, sessionID string) (string, error)
	
	// SetQRCodeBase64 stores base64 QR code in cache
	SetQRCodeBase64(ctx context.Context, sessionID string, qrCodeBase64 string) error
}

// CredentialCache handles WhatsApp credentials caching
type CredentialCache interface {
	// GetDeviceJID retrieves device JID from cache
	GetDeviceJID(ctx context.Context, sessionID string) (string, error)
	
	// SetDeviceJID stores device JID in cache
	SetDeviceJID(ctx context.Context, sessionID string, deviceJID string, ttl time.Duration) error
	
	// DeleteDeviceJID removes device JID from cache
	DeleteDeviceJID(ctx context.Context, sessionID string) error
	
	// GetSessionStatus retrieves session status from cache
	GetSessionStatus(ctx context.Context, sessionID string) (session.Status, error)
	
	// SetSessionStatus stores session status in cache
	SetSessionStatus(ctx context.Context, sessionID string, status session.Status, ttl time.Duration) error
	
	// DeleteSessionStatus removes session status from cache
	DeleteSessionStatus(ctx context.Context, sessionID string) error
}

// HealthChecker provides cache health checking capabilities
type HealthChecker interface {
	// Ping checks if cache is available
	Ping(ctx context.Context) error
	
	// GetStats returns cache statistics
	GetStats(ctx context.Context) (CacheStats, error)
}

// CacheStats represents cache statistics
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

// CacheError represents cache-specific errors
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

// NewCacheError creates a new cache error
func NewCacheError(operation, key string, err error) *CacheError {
	return &CacheError{
		Operation: operation,
		Key:       key,
		Err:       err,
	}
}

// Cache key constants
const (
	SessionKeyPrefix       = "session:"
	SessionNameKeyPrefix   = "session_name:"
	QRCodeKeyPrefix        = "qr:"
	QRCodeBase64KeyPrefix  = "qr_base64:"
	DeviceJIDKeyPrefix     = "device_jid:"
	SessionStatusKeyPrefix = "session_status:"
)

// Default TTL values
const (
	DefaultSessionTTL    = 24 * time.Hour  // Sessions cached for 24 hours
	DefaultQRCodeTTL     = 60 * time.Second // QR codes cached for 60 seconds
	DefaultCredentialTTL = 6 * time.Hour   // Credentials cached for 6 hours
	DefaultStatusTTL     = 5 * time.Minute // Status cached for 5 minutes
)
