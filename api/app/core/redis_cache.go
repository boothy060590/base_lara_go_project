package core

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisCacheDriver implements Redis caching
type RedisCacheDriver struct {
	*BaseCacheProvider
	client *redis.Client
}

// NewRedisCacheDriver creates a new Redis cache driver
func NewRedisCacheDriver(client *redis.Client, prefix string, ttl time.Duration) *RedisCacheDriver {
	return &RedisCacheDriver{
		BaseCacheProvider: NewBaseCacheProvider(prefix, ttl),
		client:            client,
	}
}

// Get retrieves a value from Redis cache
func (d *RedisCacheDriver) Get(key string) (interface{}, bool) {
	fullKey := d.GetFullKey(key)
	ctx := context.Background()

	val, err := d.client.Get(ctx, fullKey).Result()
	if err != nil {
		return nil, false
	}

	return val, true
}

// Set stores a value in Redis cache
func (d *RedisCacheDriver) Set(key string, value interface{}, ttl ...time.Duration) error {
	fullKey := d.GetFullKey(key)
	ctx := context.Background()

	duration := d.GetEffectiveTTL(ttl...)

	return d.client.Set(ctx, fullKey, value, duration).Err()
}

// Delete removes a value from Redis cache
func (d *RedisCacheDriver) Delete(key string) error {
	fullKey := d.GetFullKey(key)
	ctx := context.Background()
	return d.client.Del(ctx, fullKey).Err()
}

// Has checks if a key exists in Redis cache
func (d *RedisCacheDriver) Has(key string) bool {
	fullKey := d.GetFullKey(key)
	ctx := context.Background()

	_, err := d.client.Get(ctx, fullKey).Result()
	return err == nil
}

// Flush clears all Redis cache
func (d *RedisCacheDriver) Flush() error {
	ctx := context.Background()
	return d.client.FlushDB(ctx).Err()
}

// Increment increments a numeric value in Redis cache
func (d *RedisCacheDriver) Increment(key string, value ...int64) (int64, error) {
	fullKey := d.GetFullKey(key)
	ctx := context.Background()

	if len(value) > 0 {
		return d.client.IncrBy(ctx, fullKey, value[0]).Result()
	}
	return d.client.Incr(ctx, fullKey).Result()
}

// Decrement decrements a numeric value in Redis cache
func (d *RedisCacheDriver) Decrement(key string, value ...int64) (int64, error) {
	fullKey := d.GetFullKey(key)
	ctx := context.Background()

	if len(value) > 0 {
		return d.client.DecrBy(ctx, fullKey, value[0]).Result()
	}
	return d.client.Decr(ctx, fullKey).Result()
}
