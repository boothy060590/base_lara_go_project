package core

import (
	"log"

	"gorm.io/gorm"
)

// ModelObserver interface for observing model events
type ModelObserver interface {
	// Model events
	Created(tx *gorm.DB) error
	Updated(tx *gorm.DB) error
	Deleted(tx *gorm.DB) error
	Saved(tx *gorm.DB) error
}

// CacheableModelObserver provides automatic cache invalidation for cacheable models
type CacheableModelObserver struct{}

// Created handles cache invalidation when a model is created
func (o *CacheableModelObserver) Created(tx *gorm.DB) error {
	if cacheable, ok := tx.Statement.Model.(CacheableModel); ok {
		return o.invalidateCache(cacheable)
	}
	return nil
}

// Updated handles cache invalidation when a model is updated
func (o *CacheableModelObserver) Updated(tx *gorm.DB) error {
	if cacheable, ok := tx.Statement.Model.(CacheableModel); ok {
		return o.invalidateCache(cacheable)
	}
	return nil
}

// Deleted handles cache invalidation when a model is deleted
func (o *CacheableModelObserver) Deleted(tx *gorm.DB) error {
	if cacheable, ok := tx.Statement.Model.(CacheableModel); ok {
		return o.invalidateCache(cacheable)
	}
	return nil
}

// Saved handles cache invalidation when a model is saved (created or updated)
func (o *CacheableModelObserver) Saved(tx *gorm.DB) error {
	if cacheable, ok := tx.Statement.Model.(CacheableModel); ok {
		return o.invalidateCache(cacheable)
	}
	return nil
}

// invalidateCache invalidates cache for a cacheable model
func (o *CacheableModelObserver) invalidateCache(cacheable CacheableModel) error {
	// Invalidate by cache key
	cacheKey := cacheable.GetCacheKey()
	if cacheKey != "" {
		err := CacheInstance.Delete(cacheKey)
		if err != nil {
			log.Printf("Failed to invalidate cache for key %s: %v", cacheKey, err)
		}
	}

	// Invalidate by tags
	tags := cacheable.GetCacheTags()
	for _, tag := range tags {
		// For now, we'll use a simple tag-based invalidation
		// In a more sophisticated system, you might want to store tag-to-key mappings
		tagKey := "tag:" + tag
		err := CacheInstance.Delete(tagKey)
		if err != nil {
			log.Printf("Failed to invalidate cache tag %s: %v", tag, err)
		}
	}

	return nil
}

// RegisterModelObserver registers a model observer with GORM
func RegisterModelObserver(db *gorm.DB, model interface{}, observer ModelObserver) {
	// Register callbacks
	db.Callback().Create().After("gorm:create").Register("cache:invalidate", func(tx *gorm.DB) {
		if observer != nil {
			observer.Created(tx)
		}
	})

	db.Callback().Update().After("gorm:update").Register("cache:invalidate", func(tx *gorm.DB) {
		if observer != nil {
			observer.Updated(tx)
		}
	})

	db.Callback().Delete().After("gorm:delete").Register("cache:invalidate", func(tx *gorm.DB) {
		if observer != nil {
			observer.Deleted(tx)
		}
	})

	db.Callback().Create().After("gorm:create").Register("cache:invalidate_save", func(tx *gorm.DB) {
		if observer != nil {
			observer.Saved(tx)
		}
	})

	db.Callback().Update().After("gorm:update").Register("cache:invalidate_save", func(tx *gorm.DB) {
		if observer != nil {
			observer.Saved(tx)
		}
	})
}

// RegisterCacheableModel registers a cacheable model with automatic cache invalidation
func RegisterCacheableModel(db *gorm.DB, model interface{}) {
	observer := &CacheableModelObserver{}
	RegisterModelObserver(db, model, observer)
}
