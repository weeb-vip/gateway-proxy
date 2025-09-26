package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGraphQLCache_GenerateKey(t *testing.T) {
	cache := NewGraphQLCache(5 * time.Minute)

	key1 := cache.GenerateKey("user123", `{"query": "{ users { id name } }"}`)
	key2 := cache.GenerateKey("user123", `{"query": "{ users { id name } }"}`)
	key3 := cache.GenerateKey("user456", `{"query": "{ users { id name } }"}`)
	key4 := cache.GenerateKey("user123", `{"query": "{ posts { id title } }"}`)

	// Same user + same query should generate same key
	assert.Equal(t, key1, key2)

	// Different user + same query should generate different key
	assert.NotEqual(t, key1, key3)

	// Same user + different query should generate different key
	assert.NotEqual(t, key1, key4)
}

func TestGraphQLCache_SetAndGet(t *testing.T) {
	cache := NewGraphQLCache(5 * time.Minute)

	key := "test-key"
	response := []byte(`{"data": {"users": [{"id": "1", "name": "John"}]}}`)
	headers := map[string][]string{
		"Content-Type": {"application/json"},
		"X-Custom":     {"test-value"},
	}

	// Test cache miss
	entry, found := cache.Get(key)
	assert.False(t, found)
	assert.Nil(t, entry)

	// Set cache entry
	cache.Set(key, response, headers)

	// Test cache hit
	entry, found = cache.Get(key)
	assert.True(t, found)
	assert.NotNil(t, entry)
	assert.Equal(t, response, entry.Response)
	assert.Equal(t, headers, entry.Headers)
	assert.False(t, entry.IsExpired())
}

func TestGraphQLCache_Expiration(t *testing.T) {
	cache := NewGraphQLCache(100 * time.Millisecond)

	key := "test-key"
	response := []byte(`{"data": {"test": true}}`)
	headers := map[string][]string{"Content-Type": {"application/json"}}

	// Set cache entry
	cache.Set(key, response, headers)

	// Should be available immediately
	entry, found := cache.Get(key)
	assert.True(t, found)
	assert.False(t, entry.IsExpired())

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	entry, found = cache.Get(key)
	assert.False(t, found)
	assert.Nil(t, entry)
}

func TestGraphQLCache_Delete(t *testing.T) {
	cache := NewGraphQLCache(5 * time.Minute)

	key := "test-key"
	response := []byte(`{"data": {"test": true}}`)
	headers := map[string][]string{"Content-Type": {"application/json"}}

	// Set and verify
	cache.Set(key, response, headers)
	_, found := cache.Get(key)
	assert.True(t, found)

	// Delete and verify
	cache.Delete(key)
	_, found = cache.Get(key)
	assert.False(t, found)
}

func TestGraphQLCache_Clear(t *testing.T) {
	cache := NewGraphQLCache(5 * time.Minute)

	response := []byte(`{"data": {"test": true}}`)
	headers := map[string][]string{"Content-Type": {"application/json"}}

	// Set multiple entries
	cache.Set("key1", response, headers)
	cache.Set("key2", response, headers)
	cache.Set("key3", response, headers)

	// Verify all exist
	_, found1 := cache.Get("key1")
	_, found2 := cache.Get("key2")
	_, found3 := cache.Get("key3")
	assert.True(t, found1)
	assert.True(t, found2)
	assert.True(t, found3)

	// Clear and verify all gone
	cache.Clear()
	_, found1 = cache.Get("key1")
	_, found2 = cache.Get("key2")
	_, found3 = cache.Get("key3")
	assert.False(t, found1)
	assert.False(t, found2)
	assert.False(t, found3)
}

func TestGraphQLCache_Stats(t *testing.T) {
	cache := NewGraphQLCache(100 * time.Millisecond)

	response := []byte(`{"data": {"test": true}}`)
	headers := map[string][]string{"Content-Type": {"application/json"}}

	// Initially empty
	total, expired := cache.Stats()
	assert.Equal(t, 0, total)
	assert.Equal(t, 0, expired)

	// Add entries
	cache.Set("key1", response, headers)
	cache.Set("key2", response, headers)

	total, expired = cache.Stats()
	assert.Equal(t, 2, total)
	assert.Equal(t, 0, expired)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	total, expired = cache.Stats()
	assert.Equal(t, 2, total)
	assert.Equal(t, 2, expired)
}