package cache_core

import (
	app_core "base_lara_go_project/app/core/app"
	client_core "base_lara_go_project/app/core/clients"
	"time"
)

// BaseCacheClient provides common functionality for all cache clients
type BaseCacheClient struct {
	*client_core.BaseClient
	config *app_core.ClientConfig
	prefix string
	ttl    int
}

// NewBaseCacheClient creates a new base cache client
func NewBaseCacheClient(config *app_core.ClientConfig, name string) *BaseCacheClient {
	prefix := "base_lara_go_cache_"
	if configPrefix, ok := config.Options["prefix"].(string); ok && configPrefix != "" {
		prefix = configPrefix
	}

	ttl := 3600 // 1 hour default
	if configTTL, ok := config.Options["ttl"].(int); ok && configTTL > 0 {
		ttl = configTTL
	}

	return &BaseCacheClient{
		BaseClient: client_core.NewBaseClient(config, name),
		config:     config,
		prefix:     prefix,
		ttl:        ttl,
	}
}

// GetPrefix returns the cache key prefix
func (c *BaseCacheClient) GetPrefix() string {
	return c.prefix
}

// GetTTL returns the default TTL in seconds
func (c *BaseCacheClient) GetTTL() int {
	return c.ttl
}

// GetConfig returns the cache configuration
func (c *BaseCacheClient) GetConfig() *app_core.ClientConfig {
	return c.config
}

// BuildKey builds a cache key with prefix
func (c *BaseCacheClient) BuildKey(key string) string {
	return c.prefix + key
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
