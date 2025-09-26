package cache

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"
)

type CacheEntry struct {
	Response  []byte
	Headers   map[string][]string
	Timestamp time.Time
	TTL       time.Duration
}

func (e *CacheEntry) IsExpired() bool {
	return time.Since(e.Timestamp) > e.TTL
}

type GraphQLCache struct {
	entries map[string]*CacheEntry
	mutex   sync.RWMutex
	ttl     time.Duration
}

func NewGraphQLCache(ttl time.Duration) *GraphQLCache {
	cache := &GraphQLCache{
		entries: make(map[string]*CacheEntry),
		ttl:     ttl,
	}

	// Start background cleanup goroutine
	go cache.cleanup()

	return cache
}

func (c *GraphQLCache) GenerateKey(userToken, requestBody string) string {
	hash := sha256.Sum256([]byte(userToken + "|" + requestBody))
	return fmt.Sprintf("%x", hash)
}

func (c *GraphQLCache) Get(key string) (*CacheEntry, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.entries[key]
	if !exists || entry.IsExpired() {
		return nil, false
	}

	return entry, true
}

func (c *GraphQLCache) Set(key string, response []byte, headers map[string][]string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.entries[key] = &CacheEntry{
		Response:  response,
		Headers:   headers,
		Timestamp: time.Now(),
		TTL:       c.ttl,
	}
}

func (c *GraphQLCache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.entries, key)
}

func (c *GraphQLCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.entries = make(map[string]*CacheEntry)
}

func (c *GraphQLCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mutex.Lock()
		for key, entry := range c.entries {
			if entry.IsExpired() {
				delete(c.entries, key)
			}
		}
		c.mutex.Unlock()
	}
}

func (c *GraphQLCache) Stats() (total, expired int) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	total = len(c.entries)
	for _, entry := range c.entries {
		if entry.IsExpired() {
			expired++
		}
	}

	return total, expired
}