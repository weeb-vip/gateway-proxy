package cache

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/weeb-vip/gateway-proxy/config"
	"github.com/weeb-vip/gateway-proxy/internal/logger"
)

type CacheEntry struct {
	Response  []byte            `json:"response"`
	Headers   map[string][]string `json:"headers"`
	Timestamp time.Time         `json:"timestamp"`
}

type GraphQLCache struct {
	client *redis.Client
	ttl    time.Duration
	ctx    context.Context
}

func NewGraphQLCache(cfg *config.Config, ttl time.Duration) (*GraphQLCache, error) {
	log := logger.Get()

	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Redis URL: %w", err)
	}

	// Override with explicit password and DB if provided
	if cfg.RedisPassword != "" {
		opt.Password = cfg.RedisPassword
	}
	opt.DB = cfg.RedisDB

	client := redis.NewClient(opt)

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Info().
		Str("redis_url", cfg.RedisURL).
		Int("redis_db", cfg.RedisDB).
		Msg("Connected to Redis")

	return &GraphQLCache{
		client: client,
		ttl:    ttl,
		ctx:    ctx,
	}, nil
}

func (c *GraphQLCache) GenerateKey(userToken, requestBody string) string {
	hash := sha256.Sum256([]byte(userToken + "|" + requestBody))
	return fmt.Sprintf("gql_cache:%x", hash)
}

func (c *GraphQLCache) Get(key string) (*CacheEntry, bool) {
	data, err := c.client.Get(c.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, false // Key doesn't exist
		}
		log := logger.Get()
		log.Error().Err(err).Str("key", key).Msg("Failed to get cache entry")
		return nil, false
	}

	var entry CacheEntry
	if err := json.Unmarshal([]byte(data), &entry); err != nil {
		log := logger.Get()
		log.Error().Err(err).Str("key", key).Msg("Failed to unmarshal cache entry")
		return nil, false
	}

	return &entry, true
}

func (c *GraphQLCache) Set(key string, response []byte, headers map[string][]string) {
	entry := CacheEntry{
		Response:  response,
		Headers:   headers,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		log := logger.Get()
		log.Error().Err(err).Str("key", key).Msg("Failed to marshal cache entry")
		return
	}

	if err := c.client.Set(c.ctx, key, data, c.ttl).Err(); err != nil {
		log := logger.Get()
		log.Error().Err(err).Str("key", key).Msg("Failed to set cache entry")
	}
}

func (c *GraphQLCache) Delete(key string) {
	if err := c.client.Del(c.ctx, key).Err(); err != nil {
		log := logger.Get()
		log.Error().Err(err).Str("key", key).Msg("Failed to delete cache entry")
	}
}

func (c *GraphQLCache) Clear() {
	// Delete all keys matching our pattern
	keys, err := c.client.Keys(c.ctx, "gql_cache:*").Result()
	if err != nil {
		log := logger.Get()
		log.Error().Err(err).Msg("Failed to get cache keys for clearing")
		return
	}

	if len(keys) > 0 {
		if err := c.client.Del(c.ctx, keys...).Err(); err != nil {
			log := logger.Get()
			log.Error().Err(err).Msg("Failed to clear cache")
		}
	}
}

func (c *GraphQLCache) Stats() (total int64, err error) {
	// Count keys matching our pattern
	keys, err := c.client.Keys(c.ctx, "gql_cache:*").Result()
	if err != nil {
		return 0, err
	}
	return int64(len(keys)), nil
}

func (c *GraphQLCache) Close() error {
	return c.client.Close()
}