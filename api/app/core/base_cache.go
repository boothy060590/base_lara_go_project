package core

import (
	"time"
)

// CacheInterface defines the core cache operations
type CacheInterface interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl ...time.Duration) error
	Delete(key string) error
	Has(key string) bool
	Flush() error
}

// BaseCacheProvider provides common functionality for all cache drivers
type BaseCacheProvider struct {
	prefix string
	ttl    time.Duration
}

// NewBaseCacheProvider creates a new base cache provider
func NewBaseCacheProvider(prefix string, ttl time.Duration) *BaseCacheProvider {
	return &BaseCacheProvider{
		prefix: prefix,
		ttl:    ttl,
	}
}

// GetPrefix returns the cache prefix
func (b *BaseCacheProvider) GetPrefix() string {
	return b.prefix
}

// GetTTL returns the default TTL
func (b *BaseCacheProvider) GetTTL() time.Duration {
	return b.ttl
}

// GetFullKey returns the full key with prefix
func (b *BaseCacheProvider) GetFullKey(key string) string {
	return b.prefix + key
}

// GetEffectiveTTL returns the effective TTL (default or provided)
func (b *BaseCacheProvider) GetEffectiveTTL(ttl ...time.Duration) time.Duration {
	if len(ttl) > 0 {
		return ttl[0]
	}
	return b.ttl
}
