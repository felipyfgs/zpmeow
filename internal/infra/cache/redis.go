package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"zpmeow/internal/application/ports"
	"zpmeow/internal/config"
	"zpmeow/internal/domain/session"
	"zpmeow/internal/infra/logging"

	"github.com/redis/go-redis/v9"
)

type RedisService struct {
	client *redis.Client
	config config.CacheConfigProvider
	logger logging.Logger
}

func NewRedisService(cfg config.CacheConfigProvider) (ports.CacheService, error) {
	logger := logging.GetLogger().Sub("cache")

	if !cfg.GetCacheEnabled() {
		logger.Info("Cache is disabled, returning no-op cache service")
		return NewNoOpCacheService(), nil
	}

	var opts *redis.Options
	var err error

	if cfg.GetRedisURL() != "" {
		opts, err = redis.ParseURL(cfg.GetRedisURL())
		if err != nil {
			return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
		}
	} else {
		opts = &redis.Options{
			Addr:     fmt.Sprintf("%s:%s", cfg.GetRedisHost(), cfg.GetRedisPort()),
			Password: cfg.GetRedisPassword(),
			DB:       cfg.GetRedisDB(),
		}
	}

	opts.PoolSize = cfg.GetPoolSize()
	opts.MinIdleConns = cfg.GetMinIdleConns()
	opts.MaxRetries = cfg.GetMaxRetries()
	opts.MinRetryBackoff = cfg.GetRetryDelay()
	opts.DialTimeout = cfg.GetDialTimeout()
	opts.ReadTimeout = cfg.GetReadTimeout()
	opts.WriteTimeout = cfg.GetWriteTimeout()

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Successfully connected to Redis cache")

	return &RedisService{
		client: client,
		config: cfg,
		logger: logger,
	}, nil
}

func (r *RedisService) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}


func (r *RedisService) GetSession(ctx context.Context, sessionID string) (*session.Session, error) {
	key := ports.SessionKeyPrefix + sessionID

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, ports.NewCacheError("get", key, fmt.Errorf("session not found in cache"))
		}
		return nil, ports.NewCacheError("get", key, err)
	}

	var sess session.Session
	if err := json.Unmarshal([]byte(data), &sess); err != nil {
		return nil, ports.NewCacheError("unmarshal", key, err)
	}

	return &sess, nil
}

func (r *RedisService) SetSession(ctx context.Context, sessionID string, sess *session.Session, ttl time.Duration) error {
	key := ports.SessionKeyPrefix + sessionID

	if ttl == 0 {
		ttl = r.config.GetSessionTTL()
	}

	data, err := json.Marshal(sess)
	if err != nil {
		return ports.NewCacheError("marshal", key, err)
	}

	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return ports.NewCacheError("set", key, err)
	}

	return nil
}

func (r *RedisService) DeleteSession(ctx context.Context, sessionID string) error {
	key := ports.SessionKeyPrefix + sessionID

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return ports.NewCacheError("delete", key, err)
	}

	r.logger.Debugf("Deleted session %s from cache", sessionID)
	return nil
}

func (r *RedisService) GetSessionByName(ctx context.Context, name string) (*session.Session, error) {
	key := ports.SessionNameKeyPrefix + name

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, ports.NewCacheError("get", key, fmt.Errorf("session not found in cache"))
		}
		return nil, ports.NewCacheError("get", key, err)
	}

	var sess session.Session
	if err := json.Unmarshal([]byte(data), &sess); err != nil {
		return nil, ports.NewCacheError("unmarshal", key, err)
	}

	return &sess, nil
}

func (r *RedisService) SetSessionByName(ctx context.Context, name string, sess *session.Session, ttl time.Duration) error {
	key := ports.SessionNameKeyPrefix + name

	if ttl == 0 {
		ttl = r.config.GetSessionTTL()
	}

	data, err := json.Marshal(sess)
	if err != nil {
		return ports.NewCacheError("marshal", key, err)
	}

	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return ports.NewCacheError("set", key, err)
	}

	return nil
}

func (r *RedisService) DeleteSessionByName(ctx context.Context, name string) error {
	key := ports.SessionNameKeyPrefix + name

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return ports.NewCacheError("delete", key, err)
	}

	r.logger.Debugf("Deleted session by name %s from cache", name)
	return nil
}


func (r *RedisService) GetQRCode(ctx context.Context, sessionID string) (string, error) {
	key := ports.QRCodeKeyPrefix + sessionID

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", ports.NewCacheError("get", key, fmt.Errorf("QR code not found in cache"))
		}
		return "", ports.NewCacheError("get", key, err)
	}

	return data, nil
}

func (r *RedisService) SetQRCode(ctx context.Context, sessionID string, qrCode string) error {
	key := ports.QRCodeKeyPrefix + sessionID
	ttl := r.config.GetQRCodeTTL()

	if err := r.client.Set(ctx, key, qrCode, ttl).Err(); err != nil {
		return ports.NewCacheError("set", key, err)
	}

	return nil
}

func (r *RedisService) DeleteQRCode(ctx context.Context, sessionID string) error {
	key := ports.QRCodeKeyPrefix + sessionID

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return ports.NewCacheError("delete", key, err)
	}

	r.logger.Debugf("Deleted QR code for session %s from cache", sessionID)
	return nil
}

func (r *RedisService) GetQRCodeBase64(ctx context.Context, sessionID string) (string, error) {
	key := ports.QRCodeBase64KeyPrefix + sessionID

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", ports.NewCacheError("get", key, fmt.Errorf("QR code base64 not found in cache"))
		}
		return "", ports.NewCacheError("get", key, err)
	}

	return data, nil
}

func (r *RedisService) SetQRCodeBase64(ctx context.Context, sessionID string, qrCodeBase64 string) error {
	key := ports.QRCodeBase64KeyPrefix + sessionID
	ttl := r.config.GetQRCodeTTL()

	if err := r.client.Set(ctx, key, qrCodeBase64, ttl).Err(); err != nil {
		return ports.NewCacheError("set", key, err)
	}

	return nil
}


func (r *RedisService) GetDeviceJID(ctx context.Context, sessionID string) (string, error) {
	key := ports.DeviceJIDKeyPrefix + sessionID

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", ports.NewCacheError("get", key, fmt.Errorf("device JID not found in cache"))
		}
		return "", ports.NewCacheError("get", key, err)
	}

	return data, nil
}

func (r *RedisService) SetDeviceJID(ctx context.Context, sessionID string, deviceJID string, ttl time.Duration) error {
	key := ports.DeviceJIDKeyPrefix + sessionID

	if ttl == 0 {
		ttl = r.config.GetCredentialTTL()
	}

	if err := r.client.Set(ctx, key, deviceJID, ttl).Err(); err != nil {
		return ports.NewCacheError("set", key, err)
	}

	return nil
}

func (r *RedisService) DeleteDeviceJID(ctx context.Context, sessionID string) error {
	key := ports.DeviceJIDKeyPrefix + sessionID

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return ports.NewCacheError("delete", key, err)
	}

	r.logger.Debugf("Deleted device JID for session %s from cache", sessionID)
	return nil
}

func (r *RedisService) GetSessionStatus(ctx context.Context, sessionID string) (session.Status, error) {
	key := ports.SessionStatusKeyPrefix + sessionID

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return session.StatusDisconnected, ports.NewCacheError("get", key, fmt.Errorf("session status not found in cache"))
		}
		return session.StatusDisconnected, ports.NewCacheError("get", key, err)
	}

	status := session.Status(data)
	return status, nil
}

func (r *RedisService) SetSessionStatus(ctx context.Context, sessionID string, status session.Status, ttl time.Duration) error {
	key := ports.SessionStatusKeyPrefix + sessionID

	if ttl == 0 {
		ttl = r.config.GetStatusTTL()
	}

	if err := r.client.Set(ctx, key, string(status), ttl).Err(); err != nil {
		return ports.NewCacheError("set", key, err)
	}

	return nil
}

func (r *RedisService) DeleteSessionStatus(ctx context.Context, sessionID string) error {
	key := ports.SessionStatusKeyPrefix + sessionID

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return ports.NewCacheError("delete", key, err)
	}

	r.logger.Debugf("Deleted session status for session %s from cache", sessionID)
	return nil
}


func (r *RedisService) Ping(ctx context.Context) error {
	if err := r.client.Ping(ctx).Err(); err != nil {
		return ports.NewCacheError("ping", "", err)
	}
	return nil
}

func (r *RedisService) GetStats(ctx context.Context) (ports.CacheStats, error) {
	stats := ports.CacheStats{
		Connected: true,
		Version:   "Redis",
	}

	dbSize, err := r.client.DBSize(ctx).Result()
	if err == nil {
		stats.TotalKeys = dbSize
	}

	return stats, nil
}
