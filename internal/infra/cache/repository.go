package cache

import (
	"context"
	"errors"
	"time"

	"zpmeow/internal/application/ports"
	"zpmeow/internal/domain/session"
	"zpmeow/internal/infra/logging"
)

type CachedSessionRepository struct {
	repo   session.Repository
	cache  ports.CacheService
	logger logging.Logger
}

func NewCachedSessionRepository(repo session.Repository, cache ports.CacheService) session.Repository {
	return &CachedSessionRepository{
		repo:   repo,
		cache:  cache,
		logger: logging.GetLogger().Sub("cached-repo"),
	}
}

func (c *CachedSessionRepository) GetByID(ctx context.Context, id string) (*session.Session, error) {
	if sess, err := c.cache.GetSession(ctx, id); err == nil {
		c.logger.With().Str("session_id", logging.TruncateID(id)).Logger().Debug("Cache HIT for session ID")
		return sess, nil
	}

	c.logger.With().Str("session_id", logging.TruncateID(id)).Logger().Debug("Cache MISS for session ID, fetching from database")

	sess, err := c.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if cacheErr := c.cache.SetSession(ctx, id, sess, 0); cacheErr != nil {
		c.logger.With().
			Str("session_id", logging.TruncateID(id)).
			Err(cacheErr).
			Logger().Warn("Failed to cache session")
	}

	return sess, nil
}

func (c *CachedSessionRepository) GetByName(ctx context.Context, name string) (*session.Session, error) {
	if sess, err := c.cache.GetSessionByName(ctx, name); err == nil {
		c.logger.With().Str("session_name", name).Logger().Debug("Cache HIT for session name")
		return sess, nil
	}

	c.logger.With().Str("session_name", name).Logger().Debug("Cache MISS for session name, fetching from database")

	sess, err := c.repo.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}

	if cacheErr := c.cache.SetSessionByName(ctx, name, sess, 0); cacheErr != nil {
		c.logger.Warnf("Failed to cache session by name %s: %v", name, cacheErr)
	}
	if cacheErr := c.cache.SetSession(ctx, sess.ID().Value(), sess, 0); cacheErr != nil {
		c.logger.Warnf("Failed to cache session by ID %s: %v", sess.ID().Value(), cacheErr)
	}

	return sess, nil
}

func (c *CachedSessionRepository) GetAll(ctx context.Context) ([]*session.Session, error) {
	sessions, err := c.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

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

func (c *CachedSessionRepository) Create(ctx context.Context, sess *session.Session) error {
	err := c.repo.Create(ctx, sess)
	if err != nil {
		return err
	}

	if cacheErr := c.cache.SetSession(ctx, sess.ID().Value(), sess, 0); cacheErr != nil {
		c.logger.Warnf("Failed to cache new session %s: %v", sess.ID().Value(), cacheErr)
	}

	return nil
}

func (c *CachedSessionRepository) CreateWithGeneratedID(ctx context.Context, sess *session.Session) (string, error) {
	id, err := c.repo.CreateWithGeneratedID(ctx, sess)
	if err != nil {
		return "", err
	}

	if updatedSess, getErr := c.repo.GetByID(ctx, id); getErr == nil {
		if cacheErr := c.cache.SetSession(ctx, id, updatedSess, 0); cacheErr != nil {
			c.logger.Warnf("Failed to cache new session with generated ID %s: %v", id, cacheErr)
		}
	}

	return id, nil
}

func (c *CachedSessionRepository) Update(ctx context.Context, sess *session.Session) error {
	err := c.repo.Update(ctx, sess)
	if err != nil {
		return err
	}

	if cacheErr := c.cache.SetSession(ctx, sess.ID().Value(), sess, 0); cacheErr != nil {
		c.logger.Warnf("Failed to update cached session %s: %v", sess.ID().Value(), cacheErr)
	}

	return nil
}

func (c *CachedSessionRepository) Delete(ctx context.Context, id string) error {
	sess, _ := c.repo.GetByID(ctx, id)

	err := c.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	if cacheErr := c.cache.DeleteSession(ctx, id); cacheErr != nil {
		c.logger.Warnf("Failed to delete session %s from cache: %v", id, cacheErr)
	}

	if sess != nil {
		if cacheErr := c.cache.DeleteSessionByName(ctx, sess.Name().Value()); cacheErr != nil {
			c.logger.Warnf("Failed to delete session by name %s from cache: %v", sess.Name().Value(), cacheErr)
		}
	}

	return nil
}

func (c *CachedSessionRepository) Exists(ctx context.Context, id string) (bool, error) {
	if _, err := c.cache.GetSession(ctx, id); err == nil {
		c.logger.Debugf("Cache HIT for session existence check: %s", id)
		return true, nil
	}

	return c.repo.Exists(ctx, id)
}

func (c *CachedSessionRepository) GetActive(ctx context.Context) ([]*session.Session, error) {
	return c.repo.GetActive(ctx)
}

func (c *CachedSessionRepository) GetInactive(ctx context.Context) ([]*session.Session, error) {
	return c.repo.GetInactive(ctx)
}

func (c *CachedSessionRepository) List(ctx context.Context, limit, offset int, status string) ([]*session.Session, int, error) {
	return c.repo.List(ctx, limit, offset, status)
}

func (c *CachedSessionRepository) GetByApiKey(ctx context.Context, apiKey string) (*session.Session, error) {
	return c.repo.GetByApiKey(ctx, apiKey)
}

func (c *CachedSessionRepository) ValidateDeviceUniqueness(ctx context.Context, sessionID, deviceJID string) error {
	return c.repo.ValidateDeviceUniqueness(ctx, sessionID, deviceJID)
}

func (c *CachedSessionRepository) GetByDeviceJID(ctx context.Context, deviceJID string) (*session.Session, error) {
	sess, err := c.repo.GetByDeviceJID(ctx, deviceJID)
	if err != nil {
		return nil, err
	}

	if cacheErr := c.cache.SetSession(ctx, sess.ID().Value(), sess, 0); cacheErr != nil {
		c.logger.Warnf("Failed to cache session %s: %v", sess.ID().Value(), cacheErr)
	}
	if cacheErr := c.cache.SetDeviceJID(ctx, sess.ID().Value(), deviceJID, 0); cacheErr != nil {
		c.logger.Warnf("Failed to cache device JID for session %s: %v", sess.ID().Value(), cacheErr)
	}

	return sess, nil
}

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
