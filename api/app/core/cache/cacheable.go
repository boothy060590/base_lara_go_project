package cache_core

import (
	"time"
)

// Cacheable interface defines methods for cacheable models
type Cacheable interface {
	// GetCacheKey returns the unique cache key for this model instance
	GetCacheKey() string

	// GetCacheTTL returns the TTL for this model's cache (optional, uses default if not implemented)
	GetCacheTTL() time.Duration

	// GetCacheData returns the data to be cached (usually the model itself)
	GetCacheData() interface{}

	// GetCacheTags returns cache tags for invalidation (optional)
	GetCacheTags() []string
}

// CacheableTrait provides default implementations for cacheable models
type CacheableTrait struct{}

// GetCacheTTL returns default TTL (1 hour)
func (t *CacheableTrait) GetCacheTTL() time.Duration {
	return time.Hour
}

// GetCacheData returns the trait itself (should be overridden)
func (t *CacheableTrait) GetCacheData() interface{} {
	return t
}

// GetCacheTags returns empty tags (should be overridden if needed)
func (t *CacheableTrait) GetCacheTags() []string {
	return []string{}
}

// CacheableModel interface for models that can be cached
type CacheableModel interface {
	Cacheable
	GetID() uint
	GetTableName() string
}

// CacheableModelTrait provides default implementations for cacheable models
type CacheableModelTrait struct {
	CacheableTrait
}

// GetCacheKey returns a default cache key based on table name and ID
func (t *CacheableModelTrait) GetCacheKey() string {
	// This should be overridden by the actual model
	return ""
}

// GetTableName returns the table name (should be overridden)
func (t *CacheableModelTrait) GetTableName() string {
	return ""
}

// GetID returns the model ID (should be overridden)
func (t *CacheableModelTrait) GetID() uint {
	return 0
}
