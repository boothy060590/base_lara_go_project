package cache_core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	app_core "base_lara_go_project/app/core/app"
)

// LocalCacheClient provides file-based caching functionality
type LocalCacheClient struct {
	*BaseCacheClient
	data      map[string]*CacheItem
	mutex     sync.RWMutex
	cachePath string
}

// CacheItem represents a cached item with expiration
type CacheItem struct {
	Value      interface{}
	Expiration time.Time
}

// NewLocalCacheClient creates a new local cache client
func NewLocalCacheClient(config *app_core.ClientConfig) *LocalCacheClient {
	cachePath := "storage/cache"
	if configPath, ok := config.Options["cache_path"].(string); ok {
		cachePath = configPath
	}

	return &LocalCacheClient{
		BaseCacheClient: NewBaseCacheClient(config, "local"),
		data:            make(map[string]*CacheItem),
		cachePath:       cachePath,
	}
}

// Connect establishes the cache connection
func (c *LocalCacheClient) Connect() error {
	// Ensure cache directory exists
	if err := os.MkdirAll(c.cachePath, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %v", err)
	}

	// Load existing cache files
	if err := c.loadCacheFromDisk(); err != nil {
		return fmt.Errorf("failed to load cache from disk: %v", err)
	}

	return c.BaseClient.Connect()
}

// loadCacheFromDisk loads existing cache files from disk
func (c *LocalCacheClient) loadCacheFromDisk() error {
	entries, err := os.ReadDir(c.cachePath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".cache" {
			continue
		}

		filepath := filepath.Join(c.cachePath, entry.Name())
		data, err := os.ReadFile(filepath)
		if err != nil {
			continue // Skip corrupted files
		}

		var item CacheItem
		if err := json.Unmarshal(data, &item); err != nil {
			continue // Skip corrupted files
		}

		// Check if item has expired
		if time.Now().After(item.Expiration) {
			os.Remove(filepath) // Remove expired file
			continue
		}

		// Extract key from filename (remove .cache extension)
		key := entry.Name()[:len(entry.Name())-6] // Remove ".cache"
		c.data[key] = &item
	}

	return nil
}

// getCacheFilePath returns the file path for a cache key
func (c *LocalCacheClient) getCacheFilePath(key string) string {
	return filepath.Join(c.cachePath, key+".cache")
}

// writeCacheToDisk writes a cache item to disk
func (c *LocalCacheClient) writeCacheToDisk(key string, item *CacheItem) error {
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}

	filepath := c.getCacheFilePath(key)
	return os.WriteFile(filepath, data, 0644)
}

// removeCacheFromDisk removes a cache file from disk
func (c *LocalCacheClient) removeCacheFromDisk(key string) error {
	filepath := c.getCacheFilePath(key)
	return os.Remove(filepath)
}

// Disconnect closes the cache connection (no-op for local cache)
func (c *LocalCacheClient) Disconnect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Clear all data
	c.data = make(map[string]*CacheItem)
	return c.BaseClient.Disconnect()
}

// Get retrieves a value from cache
func (c *LocalCacheClient) Get(key string) (interface{}, bool, error) {
	if !c.IsConnected() {
		return nil, false, fmt.Errorf("cache client not connected")
	}

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	fullKey := c.BuildKey(key)
	item, exists := c.data[fullKey]

	if !exists {
		return nil, false, nil
	}

	// Check if item has expired
	if time.Now().After(item.Expiration) {
		// Remove expired item
		c.mutex.RUnlock()
		c.mutex.Lock()
		delete(c.data, fullKey)
		c.removeCacheFromDisk(fullKey) // Remove from disk
		c.mutex.Unlock()
		c.mutex.RLock()
		return nil, false, nil
	}

	return item.Value, true, nil
}

// Set stores a value in cache
func (c *LocalCacheClient) Set(key string, value interface{}, ttl ...int) error {
	if !c.IsConnected() {
		return fmt.Errorf("cache client not connected")
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	fullKey := c.BuildKey(key)

	// Determine TTL
	expirationTTL := c.GetTTL()
	if len(ttl) > 0 && ttl[0] > 0 {
		expirationTTL = ttl[0]
	}

	expiration := time.Now().Add(time.Duration(expirationTTL) * time.Second)

	item := &CacheItem{
		Value:      value,
		Expiration: expiration,
	}

	// Store in memory
	c.data[fullKey] = item

	// Store on disk
	return c.writeCacheToDisk(fullKey, item)
}

// Delete removes a value from cache
func (c *LocalCacheClient) Delete(key string) error {
	if !c.IsConnected() {
		return fmt.Errorf("cache client not connected")
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	fullKey := c.BuildKey(key)
	delete(c.data, fullKey)

	// Remove from disk
	return c.removeCacheFromDisk(fullKey)
}

// Clear removes all values from cache
func (c *LocalCacheClient) Clear() error {
	if !c.IsConnected() {
		return fmt.Errorf("cache client not connected")
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Clear memory
	c.data = make(map[string]*CacheItem)

	// Clear disk - remove all .cache files
	entries, err := os.ReadDir(c.cachePath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".cache" {
			filepath := filepath.Join(c.cachePath, entry.Name())
			os.Remove(filepath)
		}
	}

	return nil
}

// Has checks if a key exists in cache
func (c *LocalCacheClient) Has(key string) (bool, error) {
	if !c.IsConnected() {
		return false, fmt.Errorf("cache client not connected")
	}

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	fullKey := c.BuildKey(key)
	item, exists := c.data[fullKey]

	if !exists {
		return false, nil
	}

	// Check if item has expired
	if time.Now().After(item.Expiration) {
		// Remove expired item
		c.mutex.RUnlock()
		c.mutex.Lock()
		delete(c.data, fullKey)
		c.removeCacheFromDisk(fullKey) // Remove from disk
		c.mutex.Unlock()
		c.mutex.RLock()
		return false, nil
	}

	return true, nil
}

// Increment increments a numeric value
func (c *LocalCacheClient) Increment(key string, value int) (int, error) {
	if !c.IsConnected() {
		return 0, fmt.Errorf("cache client not connected")
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	fullKey := c.BuildKey(key)
	item, exists := c.data[fullKey]

	var currentValue int
	if exists && !time.Now().After(item.Expiration) {
		if numericValue, ok := item.Value.(int); ok {
			currentValue = numericValue
		}
	}

	newValue := currentValue + value
	expiration := time.Now().Add(time.Duration(c.GetTTL()) * time.Second)

	newItem := &CacheItem{
		Value:      newValue,
		Expiration: expiration,
	}

	// Store in memory
	c.data[fullKey] = newItem

	// Store on disk
	if err := c.writeCacheToDisk(fullKey, newItem); err != nil {
		return 0, err
	}

	return newValue, nil
}

// Decrement decrements a numeric value
func (c *LocalCacheClient) Decrement(key string, value int) (int, error) {
	return c.Increment(key, -value)
}

// GetStats returns cache statistics
func (c *LocalCacheClient) GetStats() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// Clean expired items first
	now := time.Now()
	expiredCount := 0
	for key, item := range c.data {
		if now.After(item.Expiration) {
			delete(c.data, key)
			expiredCount++
		}
	}

	return map[string]interface{}{
		"status":        "connected",
		"driver":        "local",
		"total_items":   len(c.data),
		"expired_items": expiredCount,
		"prefix":        c.GetPrefix(),
		"ttl":           c.GetTTL(),
	}
}

// Flush clears all cache entries
func (c *LocalCacheClient) Flush() error {
	return c.Clear()
}
