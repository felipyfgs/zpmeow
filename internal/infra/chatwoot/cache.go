package chatwoot

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"zpmeow/internal/application/ports"
)

// CacheManager implementa a interface ChatwootCacheManager
type CacheManager struct {
	contactCache      *cache
	conversationCache *cache
	mutex             sync.RWMutex
}

// cache representa um cache thread-safe com TTL
type cache struct {
	data   map[string]*cacheItem
	mutex  sync.RWMutex
	ticker *time.Ticker
	done   chan bool
}

// cacheItem representa um item no cache
type cacheItem struct {
	value     interface{}
	expiresAt time.Time
}

// NewCacheManager cria um novo gerenciador de cache
func NewCacheManager() ports.ChatwootCacheManager {
	return &CacheManager{
		contactCache:      newCache(5 * time.Minute),  // TTL padrão de 5 minutos
		conversationCache: newCache(10 * time.Minute), // TTL padrão de 10 minutos
	}
}

// newCache cria um novo cache com TTL especificado
func newCache(defaultTTL time.Duration) *cache {
	c := &cache{
		data:   make(map[string]*cacheItem),
		ticker: time.NewTicker(defaultTTL),
		done:   make(chan bool),
	}

	// Inicia limpeza automática
	go c.cleanup()

	return c
}

// Contact cache operations
func (cm *CacheManager) GetContact(phoneNumber string) (*ports.ContactResponse, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	if item := cm.contactCache.get(phoneNumber); item != nil {
		if contact, ok := item.(*ports.ContactResponse); ok {
			return contact, true
		}
	}
	return nil, false
}

func (cm *CacheManager) SetContact(phoneNumber string, contact *ports.ContactResponse, ttl time.Duration) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.contactCache.set(phoneNumber, contact, ttl)
}

func (cm *CacheManager) DeleteContact(phoneNumber string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.contactCache.delete(phoneNumber)
}

// Conversation cache operations
func (cm *CacheManager) GetConversation(contactID int) (*ports.ConversationResponse, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	key := conversationKey(contactID)
	if item := cm.conversationCache.get(key); item != nil {
		if conversation, ok := item.(*ports.ConversationResponse); ok {
			return conversation, true
		}
	}
	return nil, false
}

func (cm *CacheManager) SetConversation(contactID int, conversation *ports.ConversationResponse, ttl time.Duration) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	key := conversationKey(contactID)
	cm.conversationCache.set(key, conversation, ttl)
}

func (cm *CacheManager) DeleteConversation(contactID int) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	key := conversationKey(contactID)
	cm.conversationCache.delete(key)
}

// General cache operations
func (cm *CacheManager) Clear() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.contactCache.clear()
	cm.conversationCache.clear()
}

func (cm *CacheManager) Size() int {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	return cm.contactCache.size() + cm.conversationCache.size()
}

func (cm *CacheManager) Close() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.contactCache.close()
	cm.conversationCache.close()
}

// Internal cache methods
func (c *cache) get(key string) interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.data[key]
	if !exists {
		return nil
	}

	// Verifica se o item expirou
	if time.Now().After(item.expiresAt) {
		delete(c.data, key)
		return nil
	}

	return item.value
}

func (c *cache) set(key string, value interface{}, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data[key] = &cacheItem{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
}

func (c *cache) delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.data, key)
}

func (c *cache) clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data = make(map[string]*cacheItem)
}

func (c *cache) size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.data)
}

func (c *cache) close() {
	if c.ticker != nil {
		c.ticker.Stop()
	}
	select {
	case c.done <- true:
	default:
	}
}

// cleanup remove itens expirados do cache
func (c *cache) cleanup() {
	for {
		select {
		case <-c.ticker.C:
			c.removeExpiredItems()
		case <-c.done:
			return
		}
	}
}

// removeExpiredItems remove itens expirados
func (c *cache) removeExpiredItems() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	for key, item := range c.data {
		if now.After(item.expiresAt) {
			delete(c.data, key)
		}
	}
}

// Utility functions
func conversationKey(contactID int) string {
	return fmt.Sprintf("conversation:%s", strconv.Itoa(contactID))
}

// CacheStats fornece estatísticas do cache
type CacheStats struct {
	ContactCacheSize      int `json:"contact_cache_size"`
	ConversationCacheSize int `json:"conversation_cache_size"`
	TotalSize             int `json:"total_size"`
}

// GetStats retorna estatísticas do cache
func (cm *CacheManager) GetStats() *CacheStats {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	contactSize := cm.contactCache.size()
	conversationSize := cm.conversationCache.size()

	return &CacheStats{
		ContactCacheSize:      contactSize,
		ConversationCacheSize: conversationSize,
		TotalSize:             contactSize + conversationSize,
	}
}
