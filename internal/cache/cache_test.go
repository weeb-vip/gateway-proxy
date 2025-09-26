package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/weeb-vip/gateway-proxy/config"
)

func getTestConfig() *config.Config {
	return &config.Config{
		RedisURL:      "redis://localhost:6379",
		RedisPassword: "",
		RedisDB:       1, // Use DB 1 for tests to avoid conflicts
	}
}

func TestGraphQLCache_GenerateKey(t *testing.T) {
	cfg := getTestConfig()
	cache, err := NewGraphQLCache(cfg, 5*time.Minute)
	require.NoError(t, err)
	defer cache.Close()

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

	// All keys should have the gql_cache prefix
	assert.Contains(t, key1, "gql_cache:")
}

func TestGraphQLCache_SetAndGet(t *testing.T) {
	cfg := getTestConfig()
	cache, err := NewGraphQLCache(cfg, 5*time.Minute)
	require.NoError(t, err)
	defer cache.Close()

	key := cache.GenerateKey("testuser", "testquery")
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
	assert.WithinDuration(t, time.Now(), entry.Timestamp, time.Second)
}

func TestGraphQLCache_Expiration(t *testing.T) {
	cfg := getTestConfig()
	cache, err := NewGraphQLCache(cfg, 100*time.Millisecond)
	require.NoError(t, err)
	defer cache.Close()

	key := cache.GenerateKey("testuser", "testquery")
	response := []byte(`{"data": {"test": true}}`)
	headers := map[string][]string{"Content-Type": {"application/json"}}

	// Set cache entry
	cache.Set(key, response, headers)

	// Should be available immediately
	entry, found := cache.Get(key)
	assert.True(t, found)
	assert.NotNil(t, entry)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired (Redis TTL will handle this)
	entry, found = cache.Get(key)
	assert.False(t, found)
	assert.Nil(t, entry)
}

func TestGraphQLCache_Delete(t *testing.T) {
	cfg := getTestConfig()
	cache, err := NewGraphQLCache(cfg, 5*time.Minute)
	require.NoError(t, err)
	defer cache.Close()

	key := cache.GenerateKey("testuser", "testquery")
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
	cfg := getTestConfig()
	cache, err := NewGraphQLCache(cfg, 5*time.Minute)
	require.NoError(t, err)
	defer cache.Close()

	response := []byte(`{"data": {"test": true}}`)
	headers := map[string][]string{"Content-Type": {"application/json"}}

	// Set multiple entries using proper keys
	key1 := cache.GenerateKey("user1", "query1")
	key2 := cache.GenerateKey("user2", "query2")
	key3 := cache.GenerateKey("user3", "query3")

	cache.Set(key1, response, headers)
	cache.Set(key2, response, headers)
	cache.Set(key3, response, headers)

	// Verify all exist
	_, found1 := cache.Get(key1)
	_, found2 := cache.Get(key2)
	_, found3 := cache.Get(key3)
	assert.True(t, found1)
	assert.True(t, found2)
	assert.True(t, found3)

	// Clear and verify all gone
	cache.Clear()
	_, found1 = cache.Get(key1)
	_, found2 = cache.Get(key2)
	_, found3 = cache.Get(key3)
	assert.False(t, found1)
	assert.False(t, found2)
	assert.False(t, found3)
}

func TestGraphQLCache_Stats(t *testing.T) {
	cfg := getTestConfig()
	cache, err := NewGraphQLCache(cfg, 5*time.Minute)
	require.NoError(t, err)
	defer cache.Close()

	// Clear any existing cache entries
	cache.Clear()

	response := []byte(`{"data": {"test": true}}`)
	headers := map[string][]string{"Content-Type": {"application/json"}}

	// Initially empty
	total, err := cache.Stats()
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)

	// Add entries
	key1 := cache.GenerateKey("user1", "query1")
	key2 := cache.GenerateKey("user2", "query2")
	cache.Set(key1, response, headers)
	cache.Set(key2, response, headers)

	total, err = cache.Stats()
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
}