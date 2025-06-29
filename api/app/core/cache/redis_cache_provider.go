package cache_core

import (
	app_core "base_lara_go_project/app/core/app"
)

// RedisCacheProvider provides Redis cache services
type RedisCacheProvider struct {
	client *RedisCacheClient
}

// NewRedisCacheProvider creates a new Redis cache provider
func NewRedisCacheProvider(client *RedisCacheClient) *RedisCacheProvider {
	return &RedisCacheProvider{
		client: client,
	}
}

// Connect establishes a connection to the cache
func (p *RedisCacheProvider) Connect() error {
	return p.client.Connect()
}

// Disconnect closes the cache connection
func (p *RedisCacheProvider) Disconnect() error {
	return p.client.Disconnect()
}

// Get retrieves a value from cache
func (p *RedisCacheProvider) Get(key string) (interface{}, bool, error) {
	return p.client.Get(key)
}

// Set stores a value in cache
func (p *RedisCacheProvider) Set(key string, value interface{}, ttl ...int) error {
	return p.client.Set(key, value, ttl...)
}

// Delete removes a value from cache
func (p *RedisCacheProvider) Delete(key string) error {
	return p.client.Delete(key)
}

// Clear removes all values from cache
func (p *RedisCacheProvider) Clear() error {
	return p.client.Clear()
}

// Has checks if a key exists in cache
func (p *RedisCacheProvider) Has(key string) (bool, error) {
	return p.client.Has(key)
}

// Increment increments a numeric value
func (p *RedisCacheProvider) Increment(key string, value int) (int, error) {
	return p.client.Increment(key, value)
}

// Decrement decrements a numeric value
func (p *RedisCacheProvider) Decrement(key string, value int) (int, error) {
	return p.client.Decrement(key, value)
}

// GetStats returns cache statistics
func (p *RedisCacheProvider) GetStats() map[string]interface{} {
	return p.client.GetStats()
}

// GetClient returns the underlying cache client
func (p *RedisCacheProvider) GetClient() app_core.CacheClientInterface {
	return p.client
}
