package cache_core

import (
	"context"
	"fmt"
	"strconv"
	"time"

	app_core "base_lara_go_project/app/core/app"

	"github.com/redis/go-redis/v9"
)

// RedisCacheClient provides Redis caching functionality
type RedisCacheClient struct {
	*BaseCacheClient
	client *redis.Client
	ctx    context.Context
}

// NewRedisCacheClient creates a new Redis cache client
func NewRedisCacheClient(config *app_core.ClientConfig) *RedisCacheClient {
	return &RedisCacheClient{
		BaseCacheClient: NewBaseCacheClient(config, "redis"),
		ctx:             context.Background(),
	}
}

// Connect establishes a connection to Redis
func (c *RedisCacheClient) Connect() error {
	// Get Redis configuration from options
	host := "localhost"
	if configHost, ok := c.config.Options["host"].(string); ok {
		host = configHost
	}

	port := 6379
	if configPort, ok := c.config.Options["port"].(int); ok {
		port = configPort
	}

	password := ""
	if configPassword, ok := c.config.Options["password"].(string); ok {
		password = configPassword
	}

	// Get database number
	database := 0
	if dbStr, ok := c.config.Options["database"].(string); ok {
		if db, err := strconv.Atoi(dbStr); err == nil {
			database = db
		}
	}

	// Create Redis client
	c.client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       database,
	})

	// Test connection
	if err := c.client.Ping(c.ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return c.BaseClient.Connect()
}

// Disconnect closes the Redis connection
func (c *RedisCacheClient) Disconnect() error {
	if c.client != nil {
		if err := c.client.Close(); err != nil {
			return err
		}
	}
	return c.BaseClient.Disconnect()
}

// Get retrieves a value from Redis
func (c *RedisCacheClient) Get(key string) (interface{}, bool, error) {
	if !c.IsConnected() {
		return nil, false, fmt.Errorf("cache client not connected")
	}

	fullKey := c.BuildKey(key)
	result, err := c.client.Get(c.ctx, fullKey).Result()

	if err == redis.Nil {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, err
	}

	return result, true, nil
}

// Set stores a value in Redis
func (c *RedisCacheClient) Set(key string, value interface{}, ttl ...int) error {
	if !c.IsConnected() {
		return fmt.Errorf("cache client not connected")
	}

	fullKey := c.BuildKey(key)

	// Determine TTL
	expirationTTL := c.GetTTL()
	if len(ttl) > 0 && ttl[0] > 0 {
		expirationTTL = ttl[0]
	}

	// Convert value to string
	valueStr := fmt.Sprintf("%v", value)

	// Set with expiration
	err := c.client.Set(c.ctx, fullKey, valueStr, time.Duration(expirationTTL)*time.Second).Err()
	return err
}

// Delete removes a value from Redis
func (c *RedisCacheClient) Delete(key string) error {
	if !c.IsConnected() {
		return fmt.Errorf("cache client not connected")
	}

	fullKey := c.BuildKey(key)
	return c.client.Del(c.ctx, fullKey).Err()
}

// Clear removes all values from Redis (use pattern matching)
func (c *RedisCacheClient) Clear() error {
	if !c.IsConnected() {
		return fmt.Errorf("cache client not connected")
	}

	// Get all keys with prefix
	pattern := c.GetPrefix() + "*"
	keys, err := c.client.Keys(c.ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return c.client.Del(c.ctx, keys...).Err()
	}

	return nil
}

// Has checks if a key exists in Redis
func (c *RedisCacheClient) Has(key string) (bool, error) {
	if !c.IsConnected() {
		return false, fmt.Errorf("cache client not connected")
	}

	fullKey := c.BuildKey(key)
	result, err := c.client.Exists(c.ctx, fullKey).Result()
	return result > 0, err
}

// Increment increments a numeric value in Redis
func (c *RedisCacheClient) Increment(key string, value int) (int, error) {
	if !c.IsConnected() {
		return 0, fmt.Errorf("cache client not connected")
	}

	fullKey := c.BuildKey(key)
	result, err := c.client.IncrBy(c.ctx, fullKey, int64(value)).Result()
	return int(result), err
}

// Decrement decrements a numeric value in Redis
func (c *RedisCacheClient) Decrement(key string, value int) (int, error) {
	if !c.IsConnected() {
		return 0, fmt.Errorf("cache client not connected")
	}

	fullKey := c.BuildKey(key)
	result, err := c.client.DecrBy(c.ctx, fullKey, int64(value)).Result()
	return int(result), err
}

// GetStats returns Redis cache statistics
func (c *RedisCacheClient) GetStats() map[string]interface{} {
	if !c.IsConnected() {
		return map[string]interface{}{"status": "disconnected"}
	}

	info, err := c.client.Info(c.ctx).Result()
	if err != nil {
		return map[string]interface{}{"status": "error", "error": err.Error()}
	}

	// Parse basic info
	stats := map[string]interface{}{
		"status": "connected",
		"driver": "redis",
		"prefix": c.GetPrefix(),
		"ttl":    c.GetTTL(),
		"info":   info,
	}

	// Get database size
	if dbSize, err := c.client.DBSize(c.ctx).Result(); err == nil {
		stats["db_size"] = dbSize
	}

	return stats
}

// Flush clears all cache entries
func (c *RedisCacheClient) Flush() error {
	return c.Clear()
}
