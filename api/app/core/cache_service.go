package core

import (
	"encoding/json"
	"fmt"
	"time"
)

// CacheService provides helper methods for caching
type CacheService struct{}

// NewCacheService creates a new cache service
func NewCacheService() *CacheService {
	return &CacheService{}
}

// Remember gets a value from cache or stores the result of a callback
func (s *CacheService) Remember(key string, ttl time.Duration, callback func() (interface{}, error)) (interface{}, error) {
	// Try to get from cache first
	if value, exists := CacheInstance.Get(key); exists {
		return value, nil
	}

	// If not in cache, execute callback
	value, err := callback()
	if err != nil {
		return nil, err
	}

	// Store in cache
	err = CacheInstance.Set(key, value, ttl)
	if err != nil {
		return nil, err
	}

	return value, nil
}

// RememberForever gets a value from cache or stores the result of a callback forever
func (s *CacheService) RememberForever(key string, callback func() (interface{}, error)) (interface{}, error) {
	return s.Remember(key, 0, callback) // 0 means no expiration
}

// CacheModel caches a cacheable model
func (s *CacheService) CacheModel(model Cacheable) error {
	cacheKey := model.GetCacheKey()
	if cacheKey == "" {
		return fmt.Errorf("cache key is empty for model")
	}

	ttl := model.GetCacheTTL()
	cacheData := model.GetCacheData()

	// Serialize to JSON for storage
	data, err := json.Marshal(cacheData)
	if err != nil {
		return err
	}

	return CacheInstance.Set(cacheKey, string(data), ttl)
}

// GetCachedModel retrieves a cached model
func (s *CacheService) GetCachedModel(cacheKey string, modelType interface{}) (interface{}, bool) {
	data, exists := CacheInstance.Get(cacheKey)
	if !exists {
		return nil, false
	}

	// Deserialize from JSON
	err := json.Unmarshal([]byte(data.(string)), modelType)
	if err != nil {
		return nil, false
	}

	return modelType, true
}

// GetCachedModelByID retrieves a cached model by ID using the base key
func (s *CacheService) GetCachedModelByID(baseKey string, id uint, model CacheModelInterface) (bool, error) {
	cacheKey := fmt.Sprintf("%s:%d:data", baseKey, id)

	data, exists := CacheInstance.Get(cacheKey)
	if !exists {
		return false, nil
	}

	// Deserialize from JSON
	var cacheData map[string]interface{}
	if jsonStr, ok := data.(string); ok {
		err := json.Unmarshal([]byte(jsonStr), &cacheData)
		if err != nil {
			return false, err
		}

		// Populate model from cache data
		err = model.FromCacheData(cacheData)
		if err != nil {
			return false, err
		}

		return true, nil
	}

	return false, fmt.Errorf("cached data is not a string")
}

// ForgetModel removes a cached model
func (s *CacheService) ForgetModel(model Cacheable) error {
	cacheKey := model.GetCacheKey()
	if cacheKey == "" {
		return fmt.Errorf("cache key is empty for model")
	}

	return CacheInstance.Delete(cacheKey)
}

// ForgetByKey removes a cached item by key
func (s *CacheService) ForgetByKey(key string) error {
	return CacheInstance.Delete(key)
}

// ForgetByTag removes cached items by tag
func (s *CacheService) ForgetByTag(tag string) error {
	tagKey := "tag:" + tag
	return CacheInstance.Delete(tagKey)
}

// Flush clears all cache
func (s *CacheService) Flush() error {
	return CacheInstance.Flush()
}

// Has checks if a key exists in cache
func (s *CacheService) Has(key string) bool {
	return CacheInstance.Has(key)
}

// Get retrieves a value from cache
func (s *CacheService) Get(key string) (interface{}, bool) {
	return CacheInstance.Get(key)
}

// Set stores a value in cache
func (s *CacheService) Set(key string, value interface{}, ttl ...time.Duration) error {
	return CacheInstance.Set(key, value, ttl...)
}

// Global cache service instance
var CacheServiceInstance = NewCacheService()

// Helper functions for easy access

// Remember gets a value from cache or stores the result of a callback
func Remember(key string, ttl time.Duration, callback func() (interface{}, error)) (interface{}, error) {
	return CacheServiceInstance.Remember(key, ttl, callback)
}

// RememberForever gets a value from cache or stores the result of a callback forever
func RememberForever(key string, callback func() (interface{}, error)) (interface{}, error) {
	return CacheServiceInstance.RememberForever(key, callback)
}

// CacheModel caches a cacheable model
func CacheModel(model Cacheable) error {
	return CacheServiceInstance.CacheModel(model)
}

// GetCachedModel retrieves a cached model
func GetCachedModel(cacheKey string, modelType interface{}) (interface{}, bool) {
	return CacheServiceInstance.GetCachedModel(cacheKey, modelType)
}

// ForgetModel removes a cached model
func ForgetModel(model Cacheable) error {
	return CacheServiceInstance.ForgetModel(model)
}

// ForgetByKey removes a cached item by key
func ForgetByKey(key string) error {
	return CacheServiceInstance.ForgetByKey(key)
}

// ForgetByTag removes cached items by tag
func ForgetByTag(tag string) error {
	return CacheServiceInstance.ForgetByTag(tag)
}

// GetCachedModelByID retrieves a cached model by ID using the base key
func GetCachedModelByID(baseKey string, id uint, model CacheModelInterface) (bool, error) {
	return CacheServiceInstance.GetCachedModelByID(baseKey, id, model)
}
