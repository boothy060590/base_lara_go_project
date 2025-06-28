package core

import (
	"encoding/json"
	"fmt"
	"time"
)

// CachedModel provides caching functionality for models
type CachedModel struct {
	BaseModelData
}

// NewCachedModel creates a new cached model
func NewCachedModel() *CachedModel {
	return &CachedModel{
		BaseModelData: *NewBaseModel(),
	}
}

// CacheModelInterface defines the interface for cacheable models
type CacheModelInterface interface {
	BaseModelInterface
	GetBaseKey() string
	GetCacheKey() string
	GetCacheTTL() time.Duration
	GetCacheData() interface{}
	GetCacheTags() []string
	FromCacheData(data map[string]interface{}) error
}

// GetBaseKey returns the base key for this model type (e.g., "users", "categories")
func (c *CachedModel) GetBaseKey() string {
	// This should be overridden by the embedding struct
	// Return empty string as default to indicate it needs to be implemented
	return ""
}

// GetCacheKey returns the cache key for this model
func (c *CachedModel) GetCacheKey() string {
	baseKey := c.GetBaseKey()
	id := c.GetID()
	if baseKey == "" || id == 0 {
		return ""
	}
	return fmt.Sprintf("%s:%d:data", baseKey, id)
}

// GetCacheTTL returns the TTL for this model's cache
func (c *CachedModel) GetCacheTTL() time.Duration {
	return time.Hour
}

// GetCacheData returns the data to be cached
func (c *CachedModel) GetCacheData() interface{} {
	return c.GetData()
}

// GetCacheTags returns cache tags for invalidation
func (c *CachedModel) GetCacheTags() []string {
	return []string{
		c.GetTableName(),
		fmt.Sprintf("%s:%d", c.GetTableName(), c.GetID()),
	}
}

// FromCacheData populates the model from cached data
func (c *CachedModel) FromCacheData(data map[string]interface{}) error {
	c.Fill(data)
	return nil
}

// GetTableName returns the table name (must be implemented by embedding struct)
func (c *CachedModel) GetTableName() string {
	// This should be overridden by the embedding struct
	// Return empty string as default to indicate it needs to be implemented
	return ""
}

// GetCreatedAt returns the created at time
func (c *CachedModel) GetCreatedAt() time.Time {
	if createdAt, ok := c.Get("created_at").(time.Time); ok {
		return createdAt
	}
	return time.Time{}
}

// GetUpdatedAt returns the updated at time
func (c *CachedModel) GetUpdatedAt() time.Time {
	if updatedAt, ok := c.Get("updated_at").(time.Time); ok {
		return updatedAt
	}
	return time.Time{}
}

// GetDeletedAt returns the deleted at time
func (c *CachedModel) GetDeletedAt() *time.Time {
	if deletedAt, ok := c.Get("deleted_at").(*time.Time); ok {
		return deletedAt
	}
	return nil
}

// StoreInCache stores the model in cache
func (c *CachedModel) StoreInCache() error {
	return CacheModel(c)
}

// GetFromCache retrieves the model from cache
func (c *CachedModel) GetFromCache() (bool, error) {
	cacheKey := c.GetCacheKey()
	if cacheKey == "" {
		return false, fmt.Errorf("cache key is empty")
	}

	data, exists := CacheInstance.Get(cacheKey)
	if !exists {
		return false, nil
	}

	// Deserialize from JSON
	var cacheData map[string]interface{}
	err := json.Unmarshal([]byte(data.(string)), &cacheData)
	if err != nil {
		return false, err
	}

	// Populate model from cache data
	err = c.FromCacheData(cacheData)
	if err != nil {
		return false, err
	}

	return true, nil
}

// Forget removes the model from cache
func (c *CachedModel) Forget() error {
	return ForgetModel(c)
}
