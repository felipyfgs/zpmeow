package cache

import (
	"context"
	"errors"
	"time"

	"zpmeow/internal/application/ports"
	"zpmeow/internal/domain/session"
	"zpmeow/internal/infra/logging"
)

// CachedSessionRepository wraps a session repository with caching capabilities
type CachedSessionRepository struct {
	repo   session.Repository
	cache  ports.CacheService
	logger logging.Logger
}

// NewCachedSessionRepository creates a new cached session repository
func NewCachedSessionRepository(repo session.Repository, cache ports.CacheService) session.Repository {
	return &CachedSessionRepository{
		repo:   repo,
		cache:  cache,
		logger: logging.GetLogger().Sub("cached-repo"),
	}
}

// GetByID retrieves a session by ID, first checking cache, then database
func (c *CachedSessionRepository) GetByID(ctx context.Context, id string) (*session.Session, error) {
	// Try cache first
	if sess, err := c.cache.GetSession(ctx, id); err == nil {
		c.logger.With().Str("session_id", logging.TruncateID(id)).Logger().Debug("Cache HIT for session ID")
		return sess, nil
	}

	c.logger.With().Str("session_id", logging.TruncateID(id)).Logger().Debug("Cache MISS for session ID, fetching from database")

	// Cache miss, get from database
	sess, err := c.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if cacheErr := c.cache.SetSession(ctx, id, sess, 0); cacheErr != nil {
		c.logger.With().
			Str("session_id", logging.TruncateID(id)).
			Err(cacheErr).
			Logger().Warn("Failed to cache session")
	}

	return sess, nil
}

// GetByName retrieves a session by name, first checking cache, then database
func (c *CachedSessionRepository) GetByName(ctx context.Context, name string) (*session.Session, error) {
	// Try cache first
	if sess, err := c.cache.GetSessionByName(ctx, name); err == nil {
		c.logger.With().Str("session_name", name).Logger().Debug("Cache HIT for session name")
		return sess, nil
	}

	c.logger.With().Str("session_name", name).Logger().Debug("Cache MISS for session name, fetching from database")

	// Cache miss, get from database
	sess, err := c.repo.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}

	// Cache the result by both name and ID
	if cacheErr := c.cache.SetSessionByName(ctx, name, sess, 0); cacheErr != nil {
		c.logger.Warnf("Failed to cache session by name %s: %v", name, cacheErr)
	}
	if cacheErr := c.cache.SetSession(ctx, sess.ID().Value(), sess, 0); cacheErr != nil {
		c.logger.Warnf("Failed to cache session by ID %s: %v", sess.ID().Value(), cacheErr)
	}

	return sess, nil
}

// GetAll retrieves all sessions from database (no caching for list operations)
func (c *CachedSessionRepository) GetAll(ctx context.Context) ([]*session.Session, error) {
	// For list operations, we don't cache the entire list
	// but we can cache individual sessions
	sessions, err := c.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	// Cache individual sessions in background
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		for _, sess := range sessions {
			if cacheErr := c.cache.SetSession(bgCtx, sess.ID().Value(), sess, 0); cacheErr != nil {
				c.logger.Warnf("Failed to cache session %s in background: %v", sess.ID().Value(), cacheErr)
			}
		}
	}()

	return sessions, nil
}

// Create creates a new session and caches it
func (c *CachedSessionRepository) Create(ctx context.Context, sess *session.Session) error {
	err := c.repo.Create(ctx, sess)
	if err != nil {
		return err
	}

	// Cache the new session
	if cacheErr := c.cache.SetSession(ctx, sess.ID().Value(), sess, 0); cacheErr != nil {
		c.logger.Warnf("Failed to cache new session %s: %v", sess.ID().Value(), cacheErr)
	}

	return nil
}

// CreateWithGeneratedID creates a session with generated ID and caches it
func (c *CachedSessionRepository) CreateWithGeneratedID(ctx context.Context, sess *session.Session) (string, error) {
	id, err := c.repo.CreateWithGeneratedID(ctx, sess)
	if err != nil {
		return "", err
	}

	// Get the session with the generated ID and cache it
	if updatedSess, getErr := c.repo.GetByID(ctx, id); getErr == nil {
		if cacheErr := c.cache.SetSession(ctx, id, updatedSess, 0); cacheErr != nil {
			c.logger.Warnf("Failed to cache new session with generated ID %s: %v", id, cacheErr)
		}
	}

	return id, nil
}

// Update updates a session and invalidates cache
func (c *CachedSessionRepository) Update(ctx context.Context, sess *session.Session) error {
	err := c.repo.Update(ctx, sess)
	if err != nil {
		return err
	}

	// Update cache with new data
	if cacheErr := c.cache.SetSession(ctx, sess.ID().Value(), sess, 0); cacheErr != nil {
		c.logger.Warnf("Failed to update cached session %s: %v", sess.ID().Value(), cacheErr)
	}

	return nil
}

// Delete deletes a session and removes it from cache
func (c *CachedSessionRepository) Delete(ctx context.Context, id string) error {
	// Get session before deletion to get name for cache invalidation
	sess, _ := c.repo.GetByID(ctx, id)

	err := c.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Remove from cache
	if cacheErr := c.cache.DeleteSession(ctx, id); cacheErr != nil {
		c.logger.Warnf("Failed to delete session %s from cache: %v", id, cacheErr)
	}

	// Also remove by name if we have the session
	if sess != nil {
		if cacheErr := c.cache.DeleteSessionByName(ctx, sess.Name().Value()); cacheErr != nil {
			c.logger.Warnf("Failed to delete session by name %s from cache: %v", sess.Name().Value(), cacheErr)
		}
	}

	return nil
}

// Exists checks if a session exists, first checking cache, then database
func (c *CachedSessionRepository) Exists(ctx context.Context, id string) (bool, error) {
	// Check cache first
	if _, err := c.cache.GetSession(ctx, id); err == nil {
		c.logger.Debugf("Cache HIT for session existence check: %s", id)
		return true, nil
	}

	// Check database
	return c.repo.Exists(ctx, id)
}

// GetActive retrieves active sessions from database (no caching for this operation)
func (c *CachedSessionRepository) GetActive(ctx context.Context) ([]*session.Session, error) {
	return c.repo.GetActive(ctx)
}

// GetInactive retrieves inactive sessions from database (no caching for this operation)
func (c *CachedSessionRepository) GetInactive(ctx context.Context) ([]*session.Session, error) {
	return c.repo.GetInactive(ctx)
}

// List retrieves sessions with pagination from database (no caching for list operations)
func (c *CachedSessionRepository) List(ctx context.Context, limit, offset int, status string) ([]*session.Session, int, error) {
	return c.repo.List(ctx, limit, offset, status)
}

// GetByApiKey retrieves a session by API key from database (no caching for security reasons)
func (c *CachedSessionRepository) GetByApiKey(ctx context.Context, apiKey string) (*session.Session, error) {
	return c.repo.GetByApiKey(ctx, apiKey)
}

// ValidateDeviceUniqueness validates device uniqueness in database
func (c *CachedSessionRepository) ValidateDeviceUniqueness(ctx context.Context, sessionID, deviceJID string) error {
	return c.repo.ValidateDeviceUniqueness(ctx, sessionID, deviceJID)
}

// GetByDeviceJID retrieves a session by device JID, with caching for device JID
func (c *CachedSessionRepository) GetByDeviceJID(ctx context.Context, deviceJID string) (*session.Session, error) {
	// For device JID lookups, we don't have a direct cache mapping
	// So we go to the database, but we can cache the result
	sess, err := c.repo.GetByDeviceJID(ctx, deviceJID)
	if err != nil {
		return nil, err
	}

	// Cache the session and the device JID mapping
	if cacheErr := c.cache.SetSession(ctx, sess.ID().Value(), sess, 0); cacheErr != nil {
		c.logger.Warnf("Failed to cache session %s: %v", sess.ID().Value(), cacheErr)
	}
	if cacheErr := c.cache.SetDeviceJID(ctx, sess.ID().Value(), deviceJID, 0); cacheErr != nil {
		c.logger.Warnf("Failed to cache device JID for session %s: %v", sess.ID().Value(), cacheErr)
	}

	return sess, nil
}

// ClearCache clears all cached data for a session
func (c *CachedSessionRepository) ClearCache(ctx context.Context, sessionID string) error {
	var errs []error

	if err := c.cache.DeleteSession(ctx, sessionID); err != nil {
		errs = append(errs, err)
	}
	if err := c.cache.DeleteDeviceJID(ctx, sessionID); err != nil {
		errs = append(errs, err)
	}
	if err := c.cache.DeleteSessionStatus(ctx, sessionID); err != nil {
		errs = append(errs, err)
	}
	if err := c.cache.DeleteQRCode(ctx, sessionID); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	c.logger.Debugf("Cleared all cache for session %s", sessionID)
	return nil
}
