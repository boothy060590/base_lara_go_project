package observers_core

import (
	"log"

	"gorm.io/gorm"
)

// CacheInterface defines the interface for cache operations
type CacheInterface interface {
	Delete(key string) error
}

// CacheableModel defines the interface for cacheable models
type CacheableModel interface {
	GetCacheKey() string
}

// ModelObserver defines the interface for model observers
type ModelObserver interface {
	Created(tx interface{}) error
	Updated(tx interface{}) error
	Deleted(tx interface{}) error
	Saved(tx interface{}) error
}

// CacheableModelObserver provides automatic cache invalidation for cacheable models
type CacheableModelObserver struct {
	cacheService CacheInterface
}

// NewCacheableModelObserver creates a new cacheable model observer
func NewCacheableModelObserver(cacheService CacheInterface) *CacheableModelObserver {
	return &CacheableModelObserver{
		cacheService: cacheService,
	}
}

// Created handles model creation events
func (o *CacheableModelObserver) Created(tx interface{}) error {
	if gormTx, ok := tx.(*gorm.DB); ok {
		if cacheable, ok := gormTx.Statement.Model.(CacheableModel); ok {
			return o.invalidateCache(cacheable)
		}
	}
	return nil
}

// Updated handles model update events
func (o *CacheableModelObserver) Updated(tx interface{}) error {
	if gormTx, ok := tx.(*gorm.DB); ok {
		if cacheable, ok := gormTx.Statement.Model.(CacheableModel); ok {
			return o.invalidateCache(cacheable)
		}
	}
	return nil
}

// Deleted handles model deletion events
func (o *CacheableModelObserver) Deleted(tx interface{}) error {
	if gormTx, ok := tx.(*gorm.DB); ok {
		if cacheable, ok := gormTx.Statement.Model.(CacheableModel); ok {
			return o.invalidateCache(cacheable)
		}
	}
	return nil
}

// Saved handles model save events
func (o *CacheableModelObserver) Saved(tx interface{}) error {
	if gormTx, ok := tx.(*gorm.DB); ok {
		if cacheable, ok := gormTx.Statement.Model.(CacheableModel); ok {
			return o.invalidateCache(cacheable)
		}
	}
	return nil
}

// invalidateCache invalidates cache for a cacheable model
func (o *CacheableModelObserver) invalidateCache(cacheable CacheableModel) error {
	// Invalidate by cache key
	cacheKey := cacheable.GetCacheKey()
	if cacheKey != "" {
		return o.cacheService.Delete(cacheKey)
	}
	return nil
}

// RegisterModelObserver registers a model observer with GORM
func RegisterModelObserver(db *gorm.DB, model interface{}, observer ModelObserver) {
	// Register callbacks
	db.Callback().Create().After("gorm:create").Register("cache:invalidate", func(tx *gorm.DB) {
		if err := observer.Created(tx); err != nil {
			log.Printf("Error in model observer Created: %v", err)
		}
	})

	db.Callback().Update().After("gorm:update").Register("cache:invalidate", func(tx *gorm.DB) {
		if err := observer.Updated(tx); err != nil {
			log.Printf("Error in model observer Updated: %v", err)
		}
	})

	db.Callback().Delete().After("gorm:delete").Register("cache:invalidate", func(tx *gorm.DB) {
		if err := observer.Deleted(tx); err != nil {
			log.Printf("Error in model observer Deleted: %v", err)
		}
	})

	// Note: GORM doesn't have a direct Save callback, so we handle it in Create/Update
}

// RegisterCacheableModel registers a cacheable model with automatic cache invalidation
func RegisterCacheableModel(db *gorm.DB, model interface{}, cacheService CacheInterface) {
	observer := NewCacheableModelObserver(cacheService)
	RegisterModelObserver(db, model, observer)
}
