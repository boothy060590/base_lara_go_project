package cache_core

import (
	"sync"
	"time"
)

// cacheItem represents a cached item with expiration
type cacheItem struct {
	value      interface{}
	expiration time.Time
}

// ArrayCacheDriver implements in-memory caching
type ArrayCacheDriver struct {
	*BaseCacheProvider
	store map[string]cacheItem
	mutex sync.RWMutex
}

// NewArrayCacheDriver creates a new array cache driver
func NewArrayCacheDriver(prefix string, ttl time.Duration) *ArrayCacheDriver {
	return &ArrayCacheDriver{
		BaseCacheProvider: NewBaseCacheProvider(prefix, ttl),
		store:             make(map[string]cacheItem),
	}
}

// Get retrieves a value from array cache
func (d *ArrayCacheDriver) Get(key string) (interface{}, bool) {
	fullKey := d.GetFullKey(key)

	d.mutex.RLock()
	defer d.mutex.RUnlock()

	item, exists := d.store[fullKey]
	if !exists {
		return nil, false
	}

	if time.Now().After(item.expiration) {
		// Clean up expired item
		d.mutex.RUnlock()
		d.mutex.Lock()
		delete(d.store, fullKey)
		d.mutex.Unlock()
		d.mutex.RLock()
		return nil, false
	}

	return item.value, true
}

// Set stores a value in array cache
func (d *ArrayCacheDriver) Set(key string, value interface{}, ttl ...time.Duration) error {
	fullKey := d.GetFullKey(key)
	duration := d.GetEffectiveTTL(ttl...)

	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.store[fullKey] = cacheItem{
		value:      value,
		expiration: time.Now().Add(duration),
	}
	return nil
}

// Delete removes a value from array cache
func (d *ArrayCacheDriver) Delete(key string) error {
	fullKey := d.GetFullKey(key)

	d.mutex.Lock()
	defer d.mutex.Unlock()

	delete(d.store, fullKey)
	return nil
}

// Has checks if a key exists in array cache
func (d *ArrayCacheDriver) Has(key string) bool {
	_, exists := d.Get(key)
	return exists
}

// Flush clears all array cache
func (d *ArrayCacheDriver) Flush() error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.store = make(map[string]cacheItem)
	return nil
}

// GetStats returns cache statistics
func (d *ArrayCacheDriver) GetStats() map[string]interface{} {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	expired := 0
	valid := 0

	now := time.Now()
	for _, item := range d.store {
		if now.After(item.expiration) {
			expired++
		} else {
			valid++
		}
	}

	return map[string]interface{}{
		"total_items":   len(d.store),
		"valid_items":   valid,
		"expired_items": expired,
	}
}
