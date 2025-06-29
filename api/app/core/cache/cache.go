package cache_core

import (
	"time"

	app_core "base_lara_go_project/app/core/app"
)

// CacheProvider implements the core CacheInterface
type CacheProvider struct {
	driver app_core.CacheInterface
}

// NewCacheProvider creates a new cache provider
func NewCacheProvider(driver app_core.CacheInterface) *CacheProvider {
	return &CacheProvider{driver: driver}
}

// Get retrieves a value from cache
func (c *CacheProvider) Get(key string) (interface{}, bool) {
	return c.driver.Get(key)
}

// Set stores a value in cache
func (c *CacheProvider) Set(key string, value interface{}, ttl ...time.Duration) error {
	return c.driver.Set(key, value, ttl...)
}

// Delete removes a value from cache
func (c *CacheProvider) Delete(key string) error {
	return c.driver.Delete(key)
}

// Has checks if a key exists in cache
func (c *CacheProvider) Has(key string) bool {
	return c.driver.Has(key)
}

// Flush clears all cache
func (c *CacheProvider) Flush() error {
	return c.driver.Flush()
}

// Global cache instance
var CacheInstance app_core.CacheInterface

// Helper functions for cache operations (avoiding import cycles)

// Cache returns the global cache instance
func Cache() app_core.CacheInterface {
	return CacheInstance
}

// CacheGet retrieves a value from cache
func CacheGet(key string) (interface{}, bool) {
	return CacheInstance.Get(key)
}

// CacheSet stores a value in cache
func CacheSet(key string, value interface{}, ttl ...time.Duration) error {
	return CacheInstance.Set(key, value, ttl...)
}

// CacheDelete removes a value from cache
func CacheDelete(key string) error {
	return CacheInstance.Delete(key)
}

// CacheHas checks if a key exists in cache
func CacheHas(key string) bool {
	return CacheInstance.Has(key)
}

// CacheFlush clears all cache
func CacheFlush() error {
	return CacheInstance.Flush()
}
