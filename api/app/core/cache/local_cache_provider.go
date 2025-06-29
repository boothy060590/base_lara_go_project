package cache_core

import (
	app_core "base_lara_go_project/app/core/app"
)

// LocalCacheProvider provides local cache services
type LocalCacheProvider struct {
	client *LocalCacheClient
}

// NewLocalCacheProvider creates a new local cache provider
func NewLocalCacheProvider(client *LocalCacheClient) *LocalCacheProvider {
	return &LocalCacheProvider{
		client: client,
	}
}

// Connect establishes a connection to the cache
func (p *LocalCacheProvider) Connect() error {
	return p.client.Connect()
}

// Disconnect closes the cache connection
func (p *LocalCacheProvider) Disconnect() error {
	return p.client.Disconnect()
}

// Get retrieves a value from cache
func (p *LocalCacheProvider) Get(key string) (interface{}, bool, error) {
	return p.client.Get(key)
}

// Set stores a value in cache
func (p *LocalCacheProvider) Set(key string, value interface{}, ttl ...int) error {
	return p.client.Set(key, value, ttl...)
}

// Delete removes a value from cache
func (p *LocalCacheProvider) Delete(key string) error {
	return p.client.Delete(key)
}

// Clear removes all values from cache
func (p *LocalCacheProvider) Clear() error {
	return p.client.Clear()
}

// Has checks if a key exists in cache
func (p *LocalCacheProvider) Has(key string) (bool, error) {
	return p.client.Has(key)
}

// Increment increments a numeric value
func (p *LocalCacheProvider) Increment(key string, value int) (int, error) {
	return p.client.Increment(key, value)
}

// Decrement decrements a numeric value
func (p *LocalCacheProvider) Decrement(key string, value int) (int, error) {
	return p.client.Decrement(key, value)
}

// GetStats returns cache statistics
func (p *LocalCacheProvider) GetStats() map[string]interface{} {
	return p.client.GetStats()
}

// GetClient returns the underlying cache client
func (p *LocalCacheProvider) GetClient() app_core.CacheClientInterface {
	return p.client
}
