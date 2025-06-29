package observers_core

import (
	"log"

	app_core "base_lara_go_project/app/core/app"

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
type CacheableModelObserver struct {
	cacheService app_core.CacheInterface
}

// NewCacheableModelObserver creates a new cacheable model observer
func NewCacheableModelObserver(cacheService app_core.CacheInterface) *CacheableModelObserver {
	return &CacheableModelObserver{
		cacheService: cacheService,
	}
}

// Created handles cache invalidation when a model is created
func (o *CacheableModelObserver) Created(tx interface{}) error {
	if gormTx, ok := tx.(*gorm.DB); ok {
		if cacheable, ok := gormTx.Statement.Model.(app_core.CacheableModel); ok {
			return o.invalidateCache(cacheable)
		}
	}
	return nil
}

// Updated handles cache invalidation when a model is updated
func (o *CacheableModelObserver) Updated(tx interface{}) error {
	if gormTx, ok := tx.(*gorm.DB); ok {
		if cacheable, ok := gormTx.Statement.Model.(app_core.CacheableModel); ok {
			return o.invalidateCache(cacheable)
		}
	}
	return nil
}

// Deleted handles cache invalidation when a model is deleted
func (o *CacheableModelObserver) Deleted(tx interface{}) error {
	if gormTx, ok := tx.(*gorm.DB); ok {
		if cacheable, ok := gormTx.Statement.Model.(app_core.CacheableModel); ok {
			return o.invalidateCache(cacheable)
		}
	}
	return nil
}

// Saved handles cache invalidation when a model is saved (created or updated)
func (o *CacheableModelObserver) Saved(tx interface{}) error {
	if gormTx, ok := tx.(*gorm.DB); ok {
		if cacheable, ok := gormTx.Statement.Model.(app_core.CacheableModel); ok {
			return o.invalidateCache(cacheable)
		}
	}
	return nil
}

// invalidateCache invalidates cache for a cacheable model
func (o *CacheableModelObserver) invalidateCache(cacheable app_core.CacheableModel) error {
	// Invalidate by cache key
	cacheKey := cacheable.GetCacheKey()
	if cacheKey != "" {
		err := o.cacheService.Delete(cacheKey)
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
		err := o.cacheService.Delete(tagKey)
		if err != nil {
			log.Printf("Failed to invalidate cache tag %s: %v", tag, err)
		}
	}

	return nil
}

// RegisterModelObserver registers a model observer with GORM
func RegisterModelObserver(db *gorm.DB, model interface{}, observer app_core.ModelObserver) {
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
func RegisterCacheableModel(db *gorm.DB, model interface{}, cacheService app_core.CacheInterface) {
	observer := NewCacheableModelObserver(cacheService)
	RegisterModelObserver(db, model, observer)
}
