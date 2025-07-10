package go_core

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Model represents a base model with common functionality
type Model[T any] struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// ModelTraits defines the traits a model can have
type ModelTraits struct {
	Cacheable     bool
	SoftDeletes   bool
	Authenticated bool
	HasRoles      bool
	Timestamps    bool
}

// ModelConfig defines model configuration
type ModelConfig struct {
	TableName   string
	Traits      ModelTraits
	CacheTTL    time.Duration
	CachePrefix string
	Relations   map[string]string
}

// BaseModel provides generic model functionality
type BaseModel[T any] struct {
	Model[T]
	config ModelConfig
	db     *gorm.DB
	cache  Cache[T]

	// Optimization fields
	workStealingPool any
	customAllocator  any
	profileOptimizer any
}

// NewBaseModel creates a new base model
func NewBaseModel[T any](db *gorm.DB, cache Cache[T], config ModelConfig, wsp any, ca any, pgo any) *BaseModel[T] {
	return &BaseModel[T]{
		config:           config,
		db:               db,
		cache:            cache,
		workStealingPool: wsp,
		customAllocator:  ca,
		profileOptimizer: pgo,
	}
}

// Find finds a model by ID with cache support
func (m *BaseModel[T]) Find(id uint) (*T, error) {
	if m.config.Traits.Cacheable {
		// Try cache first
		cacheKey := fmt.Sprintf("%s:%d", m.config.CachePrefix, id)
		if cached, err := m.cache.Get(cacheKey); err == nil && cached != nil {
			return cached, nil
		}

		// Fallback to database
		var model T
		if err := m.db.First(&model, id).Error; err != nil {
			return nil, err
		}

		// Cache the result
		m.cache.Set(cacheKey, &model, m.config.CacheTTL)
		return &model, nil
	}

	// No cache, direct database query
	var model T
	if err := m.db.First(&model, id).Error; err != nil {
		return nil, err
	}
	return &model, nil
}

// Create creates a new model with cache invalidation
func (m *BaseModel[T]) Create(model *T) error {
	if err := m.db.Create(model).Error; err != nil {
		return err
	}

	// Invalidate cache if cacheable
	if m.config.Traits.Cacheable {
		m.invalidateCache()
	}

	return nil
}

// Update updates a model with cache invalidation
func (m *BaseModel[T]) Update(model *T) error {
	if err := m.db.Save(model).Error; err != nil {
		return err
	}

	// Invalidate cache if cacheable
	if m.config.Traits.Cacheable {
		m.invalidateCache()
	}

	return nil
}

// Delete deletes a model with cache invalidation
func (m *BaseModel[T]) Delete(id uint) error {
	var model T
	if err := m.db.Delete(&model, id).Error; err != nil {
		return err
	}

	// Invalidate cache if cacheable
	if m.config.Traits.Cacheable {
		m.invalidateCache()
	}

	return nil
}

// SoftDelete soft deletes a model if trait is enabled
func (m *BaseModel[T]) SoftDelete(id uint) error {
	if !m.config.Traits.SoftDeletes {
		return m.Delete(id)
	}

	var model T
	if err := m.db.Delete(&model, id).Error; err != nil {
		return err
	}

	// Invalidate cache if cacheable
	if m.config.Traits.Cacheable {
		m.invalidateCache()
	}

	return nil
}

// invalidateCache invalidates the model's cache
func (m *BaseModel[T]) invalidateCache() {
	if m.cache != nil {
		// Clear all cache entries for this model
		pattern := fmt.Sprintf("%s:*", m.config.CachePrefix)
		m.cache.DeletePattern(pattern)
	}
}

// WithContext returns a model with context
func (m *BaseModel[T]) WithContext(ctx context.Context) *BaseModel[T] {
	newModel := *m
	newModel.db = m.db.WithContext(ctx)
	return &newModel
}

// Where adds a where clause
func (m *BaseModel[T]) Where(query interface{}, args ...interface{}) *BaseModel[T] {
	newModel := *m
	newModel.db = m.db.Where(query, args...)
	return &newModel
}

// Get retrieves all models
func (m *BaseModel[T]) Get() ([]T, error) {
	var models []T
	if err := m.db.Find(&models).Error; err != nil {
		return nil, err
	}
	return models, nil
}

// Paginate retrieves paginated models
func (m *BaseModel[T]) Paginate(page, perPage int) ([]T, int64, error) {
	var models []T
	var total int64

	// Count total
	if err := m.db.Model(new(T)).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * perPage
	if err := m.db.Offset(offset).Limit(perPage).Find(&models).Error; err != nil {
		return nil, 0, err
	}

	return models, total, nil
}
