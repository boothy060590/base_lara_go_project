package facades_core

import (
	"fmt"
	"time"
)

// CacheInterface defines the cache operations
type CacheInterface interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl ...time.Duration) error
	Delete(key string) error
	Has(key string) bool
	Flush() error
}

// RedisCacheDriver interface for increment/decrement operations
type RedisCacheDriver interface {
	Increment(key string, value ...int64) (int64, error)
	Decrement(key string, value ...int64) (int64, error)
}

// Global cache instance
var globalCacheInstance CacheInterface

// Cache provides a facade for cache operations
type Cache struct{}

// Get retrieves a value from cache
func (c *Cache) Get(key string) (interface{}, bool) {
	if globalCacheInstance == nil {
		return nil, false
	}
	return globalCacheInstance.Get(key)
}

// Set stores a value in cache
func (c *Cache) Set(key string, value interface{}, ttl ...time.Duration) error {
	if globalCacheInstance == nil {
		return fmt.Errorf("cache not initialized")
	}
	return globalCacheInstance.Set(key, value, ttl...)
}

// Delete removes a value from cache
func (c *Cache) Delete(key string) error {
	if globalCacheInstance == nil {
		return fmt.Errorf("cache not initialized")
	}
	return globalCacheInstance.Delete(key)
}

// Has checks if a key exists in cache
func (c *Cache) Has(key string) bool {
	if globalCacheInstance == nil {
		return false
	}
	return globalCacheInstance.Has(key)
}

// Flush clears all cache
func (c *Cache) Flush() error {
	if globalCacheInstance == nil {
		return fmt.Errorf("cache not initialized")
	}
	return globalCacheInstance.Flush()
}

// Remember retrieves a value from cache or stores the result of a callback
func (c *Cache) Remember(key string, ttl time.Duration, callback func() (interface{}, error)) (interface{}, error) {
	// Try to get from cache first
	if value, exists := c.Get(key); exists {
		return value, nil
	}

	// Execute callback if not in cache
	value, err := callback()
	if err != nil {
		return nil, err
	}

	// Store in cache
	if err := c.Set(key, value, ttl); err != nil {
		return value, err // Return value even if cache storage fails
	}

	return value, nil
}

// RememberForever retrieves a value from cache or stores the result of a callback forever
func (c *Cache) RememberForever(key string, callback func() (interface{}, error)) (interface{}, error) {
	// Try to get from cache first
	if value, exists := c.Get(key); exists {
		return value, nil
	}

	// Execute callback if not in cache
	value, err := callback()
	if err != nil {
		return nil, err
	}

	// Store in cache with very long TTL (effectively forever)
	if err := c.Set(key, value, 365*24*time.Hour); err != nil {
		return value, err // Return value even if cache storage fails
	}

	return value, nil
}

// Pull gets a value from cache and deletes it
func (c *Cache) Pull(key string) (interface{}, bool) {
	value, exists := c.Get(key)
	if exists {
		c.Delete(key)
	}
	return value, exists
}

// Add stores a value in cache only if it doesn't already exist
func (c *Cache) Add(key string, value interface{}, ttl ...time.Duration) bool {
	if c.Has(key) {
		return false
	}

	err := c.Set(key, value, ttl...)
	return err == nil
}

// Increment increments a numeric value in cache
func (c *Cache) Increment(key string, value ...int64) (int64, error) {
	// Check if the driver supports increment
	if redisDriver, ok := globalCacheInstance.(RedisCacheDriver); ok {
		return redisDriver.Increment(key, value...)
	}

	// Fallback for non-Redis drivers
	return 0, fmt.Errorf("increment not supported for this cache driver")
}

// Decrement decrements a numeric value in cache
func (c *Cache) Decrement(key string, value ...int64) (int64, error) {
	// Check if the driver supports decrement
	if redisDriver, ok := globalCacheInstance.(RedisCacheDriver); ok {
		return redisDriver.Decrement(key, value...)
	}

	// Fallback for non-Redis drivers
	return 0, fmt.Errorf("decrement not supported for this cache driver")
}

// Global cache instance
var CacheInstance = &Cache{}

// SetCache sets the global cache instance
func SetCache(cache CacheInterface) {
	globalCacheInstance = cache
}

// Helper functions for easy access
// Flush clears all cache
func Flush() error {
	return CacheInstance.Flush()
}

// Remember retrieves a value from cache or stores the result of a callback
func Remember(key string, ttl time.Duration, callback func() (interface{}, error)) (interface{}, error) {
	return CacheInstance.Remember(key, ttl, callback)
}

// RememberForever retrieves a value from cache or stores the result of a callback forever
func RememberForever(key string, callback func() (interface{}, error)) (interface{}, error) {
	return CacheInstance.RememberForever(key, callback)
}

// Pull gets a value from cache and deletes it
func Pull(key string) (interface{}, bool) {
	return CacheInstance.Pull(key)
}

// Add stores a value in cache only if it doesn't already exist
func Add(key string, value interface{}, ttl ...time.Duration) bool {
	return CacheInstance.Add(key, value, ttl...)
}

// Increment increments a numeric value in cache
func Increment(key string, value ...int64) (int64, error) {
	return CacheInstance.Increment(key, value...)
}

// Decrement decrements a numeric value in cache
func Decrement(key string, value ...int64) (int64, error) {
	return CacheInstance.Decrement(key, value...)
}
