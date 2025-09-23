package wmeow

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// CacheItem represents a cached item with expiration
type CacheItem struct {
	Value     interface{}
	ExpiresAt time.Time
}

// IsExpired checks if the cache item has expired
func (c *CacheItem) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// MemoryCache provides in-memory caching with TTL support
type MemoryCache struct {
	items map[string]*CacheItem
	mutex sync.RWMutex
	ttl   time.Duration
}

// NewMemoryCache creates a new memory cache with default TTL
func NewMemoryCache(ttl time.Duration) *MemoryCache {
	cache := &MemoryCache{
		items: make(map[string]*CacheItem),
		ttl:   ttl,
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Set stores a value in cache with default TTL
func (c *MemoryCache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, c.ttl)
}

// SetWithTTL stores a value in cache with custom TTL
func (c *MemoryCache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items[key] = &CacheItem{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// Get retrieves a value from cache
func (c *MemoryCache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	if item.IsExpired() {
		// Remove expired item
		delete(c.items, key)
		return nil, false
	}

	return item.Value, true
}

// Delete removes a value from cache
func (c *MemoryCache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.items, key)
}

// Clear removes all values from cache
func (c *MemoryCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items = make(map[string]*CacheItem)
}

// Size returns the number of items in cache
func (c *MemoryCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.items)
}

// cleanup removes expired items periodically
func (c *MemoryCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mutex.Lock()
		for key, item := range c.items {
			if item.IsExpired() {
				delete(c.items, key)
			}
		}
		c.mutex.Unlock()
	}
}

// CacheKeyBuilder provides utilities for building cache keys
type CacheKeyBuilder struct{}

// NewCacheKeyBuilder creates a new cache key builder
func NewCacheKeyBuilder() *CacheKeyBuilder {
	return &CacheKeyBuilder{}
}

// SessionKey builds cache key for session
func (b *CacheKeyBuilder) SessionKey(sessionID string) string {
	return fmt.Sprintf(CacheKeySession, sessionID)
}

// ContactKey builds cache key for contact
func (b *CacheKeyBuilder) ContactKey(sessionID, contactJID string) string {
	return fmt.Sprintf(CacheKeyContact, sessionID, contactJID)
}

// GroupKey builds cache key for group
func (b *CacheKeyBuilder) GroupKey(sessionID, groupJID string) string {
	return fmt.Sprintf(CacheKeyGroup, sessionID, groupJID)
}

// MediaKey builds cache key for media
func (b *CacheKeyBuilder) MediaKey(sessionID, messageID string) string {
	return fmt.Sprintf(CacheKeyMedia, sessionID, messageID)
}

// QRCodeKey builds cache key for QR code
func (b *CacheKeyBuilder) QRCodeKey(sessionID string) string {
	return fmt.Sprintf(CacheKeyQRCode, sessionID)
}

// PresenceKey builds cache key for presence
func (b *CacheKeyBuilder) PresenceKey(sessionID, contactJID string) string {
	return fmt.Sprintf(CacheKeyPresence, sessionID, contactJID)
}

// CacheManager provides high-level cache management
type CacheManager struct {
	cache      *MemoryCache
	keyBuilder *CacheKeyBuilder
}

// NewCacheManager creates a new cache manager
func NewCacheManager(ttl time.Duration) *CacheManager {
	return &CacheManager{
		cache:      NewMemoryCache(ttl),
		keyBuilder: NewCacheKeyBuilder(),
	}
}

// SetSession caches session data
func (m *CacheManager) SetSession(sessionID string, data interface{}) {
	key := m.keyBuilder.SessionKey(sessionID)
	m.cache.Set(key, data)
}

// GetSession retrieves session data from cache
func (m *CacheManager) GetSession(sessionID string) (interface{}, bool) {
	key := m.keyBuilder.SessionKey(sessionID)
	return m.cache.Get(key)
}

// SetContact caches contact data
func (m *CacheManager) SetContact(sessionID, contactJID string, data interface{}) {
	key := m.keyBuilder.ContactKey(sessionID, contactJID)
	m.cache.Set(key, data)
}

// GetContact retrieves contact data from cache
func (m *CacheManager) GetContact(sessionID, contactJID string) (interface{}, bool) {
	key := m.keyBuilder.ContactKey(sessionID, contactJID)
	return m.cache.Get(key)
}

// SetGroup caches group data
func (m *CacheManager) SetGroup(sessionID, groupJID string, data interface{}) {
	key := m.keyBuilder.GroupKey(sessionID, groupJID)
	m.cache.Set(key, data)
}

// GetGroup retrieves group data from cache
func (m *CacheManager) GetGroup(sessionID, groupJID string) (interface{}, bool) {
	key := m.keyBuilder.GroupKey(sessionID, groupJID)
	return m.cache.Get(key)
}

// SetMedia caches media data
func (m *CacheManager) SetMedia(sessionID, messageID string, data interface{}) {
	key := m.keyBuilder.MediaKey(sessionID, messageID)
	// Use longer TTL for media (1 hour)
	m.cache.SetWithTTL(key, data, 1*time.Hour)
}

// GetMedia retrieves media data from cache
func (m *CacheManager) GetMedia(sessionID, messageID string) (interface{}, bool) {
	key := m.keyBuilder.MediaKey(sessionID, messageID)
	return m.cache.Get(key)
}

// SetQRCode caches QR code data
func (m *CacheManager) SetQRCode(sessionID string, data interface{}) {
	key := m.keyBuilder.QRCodeKey(sessionID)
	// Use shorter TTL for QR codes (2 minutes)
	m.cache.SetWithTTL(key, data, 2*time.Minute)
}

// GetQRCode retrieves QR code data from cache
func (m *CacheManager) GetQRCode(sessionID string) (interface{}, bool) {
	key := m.keyBuilder.QRCodeKey(sessionID)
	return m.cache.Get(key)
}

// SetPresence caches presence data
func (m *CacheManager) SetPresence(sessionID, contactJID string, data interface{}) {
	key := m.keyBuilder.PresenceKey(sessionID, contactJID)
	// Use shorter TTL for presence (30 seconds)
	m.cache.SetWithTTL(key, data, 30*time.Second)
}

// GetPresence retrieves presence data from cache
func (m *CacheManager) GetPresence(sessionID, contactJID string) (interface{}, bool) {
	key := m.keyBuilder.PresenceKey(sessionID, contactJID)
	return m.cache.Get(key)
}

// ClearSession removes all cached data for a session
func (m *CacheManager) ClearSession(sessionID string) {
	// This is a simple implementation - in production you might want
	// to track keys by session for more efficient cleanup
	sessionPrefix := sessionID + ":"

	// Get all keys and delete those that start with session prefix
	m.cache.mutex.Lock()
	defer m.cache.mutex.Unlock()

	for key := range m.cache.items {
		if strings.Contains(key, sessionPrefix) {
			delete(m.cache.items, key)
		}
	}
}

// GetStats returns cache statistics
func (m *CacheManager) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"size": m.cache.Size(),
		"ttl":  m.cache.ttl.String(),
	}
}
