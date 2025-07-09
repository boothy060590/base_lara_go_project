package go_core

import (
	"context"

	"gorm.io/gorm"
)

// Repository defines a generic repository interface for any model type
type Repository[T any] interface {
	// Basic CRUD operations
	Find(id uint) (*T, error)
	FindBy(field string, value any) (*T, error)
	FindAll() ([]T, error)
	Create(model *T) error
	Update(model *T) error
	Delete(id uint) error

	// Query operations
	Where(conditions map[string]any) Query[T]
	WhereRaw(query string, args ...any) Query[T]

	// Transaction support
	Transaction(fn func(Repository[T]) error) error
	WithContext(ctx context.Context) Repository[T]

	// Utility operations
	Exists(id uint) (bool, error)
	Count() (int64, error)
	CountWhere(conditions map[string]any) (int64, error)

	// Performance operations
	GetPerformanceStats() map[string]interface{}
	GetOptimizationStats() map[string]interface{}
}

// Query defines a generic query builder interface
type Query[T any] interface {
	// Execution
	Get() ([]T, error)
	First() (*T, error)
	Paginate(page, perPage int) ([]T, int64, error)

	// Query building
	Where(field string, operator string, value any) Query[T]
	WhereIn(field string, values []any) Query[T]
	OrderBy(field string, direction string) Query[T]
	Limit(limit int) Query[T]
	Offset(offset int) Query[T]
	Preload(relation string) Query[T]

	// Context
	WithContext(ctx context.Context) Query[T]
}

// repository implements Repository[T] with GORM and performance tracking
type repository[T any] struct {
	db                *gorm.DB
	performanceFacade *PerformanceFacade
	// Performance optimizations (safely used for non-database operations)
	objectPool         *ObjectPool[T] // Used for data processing, NOT database operations
	atomicCounter      *AtomicCounter // Safe for concurrent counting
	optimizationEngine *OptimizationEngine
}

// NewRepository creates a new repository instance with performance tracking and optimizations
func NewRepository[T any](db *gorm.DB) Repository[T] {
	perf := NewPerformanceFacade()

	// Create object pool for data processing operations (NOT database operations)
	// This is safe because we only use it for in-memory data transformations
	objectPool := NewObjectPool[T](100,
		func() T { return *new(T) },
		func(entity T) T { return *new(T) }, // Reset entity
	)

	// Create atomic counter for operations (safe for concurrent access)
	atomicCounter := NewAtomicCounter()

	// Create optimization engine
	optimizationEngine := NewOptimizationEngine()

	return &repository[T]{
		db:                 db,
		performanceFacade:  perf,
		objectPool:         objectPool,
		atomicCounter:      atomicCounter,
		optimizationEngine: optimizationEngine,
	}
}

// Find retrieves a model by ID with performance tracking and atomic counter
func (r *repository[T]) Find(id uint) (*T, error) {
	// Track operation count atomically
	r.atomicCounter.Increment()

	var result *T
	err := r.performanceFacade.Track("repository.find", func() error {
		// Create fresh entity for database operation (avoid concurrency issues)
		var entity T
		if err := r.db.First(&entity, id).Error; err != nil {
			return err
		}
		result = &entity
		return nil
	})
	return result, err
}

// FindBy retrieves a model by field and value with performance tracking and atomic counter
func (r *repository[T]) FindBy(field string, value any) (*T, error) {
	// Track operation count atomically
	r.atomicCounter.Increment()

	var result *T
	err := r.performanceFacade.Track("repository.find_by", func() error {
		// Create fresh entity for database operation (avoid concurrency issues)
		var entity T
		if err := r.db.Where(field+" = ?", value).First(&entity).Error; err != nil {
			return err
		}
		result = &entity
		return nil
	})
	return result, err
}

// FindAll retrieves all records with performance tracking and pipeline optimization
func (r *repository[T]) FindAll() ([]T, error) {
	var result []T
	err := r.performanceFacade.Track("repository.find_all", func() error {
		// Get all records (database operation - use fresh objects)
		if err := r.db.Find(&result).Error; err != nil {
			return err
		}

		// Use channel-based pipeline for processing large datasets (safe - in-memory only)
		if len(result) > 100 { // Only use pipeline for larger datasets
			pipeline := NewPipeline[T]()
			processed := pipeline.Execute(result)

			// Collect results (safe to use object pool for data processing)
			optimized := make([]T, 0, len(result))
			for item := range processed {
				// Safe to use object pool here - this is data processing, not database operations
				optimized = append(optimized, item)
			}
			result = optimized
		}

		return nil
	})
	return result, err
}

// Create saves a new model with performance tracking and dynamic optimization
func (r *repository[T]) Create(model *T) error {
	return r.performanceFacade.Track("repository.create", func() error {
		// Apply dynamic optimization to the model
		if err := r.performanceFacade.Optimize(model); err != nil {
			// Log optimization error but continue with creation
			// log.Printf("Optimization failed for model: %v", err)
		}

		return r.db.Create(model).Error
	})
}

// Update saves an existing model with performance tracking and dynamic optimization
func (r *repository[T]) Update(model *T) error {
	// Track operation count atomically
	r.atomicCounter.Increment()

	return r.performanceFacade.Track("repository.update", func() error {
		// Apply dynamic optimization to the model before update
		if err := r.performanceFacade.Optimize(model); err != nil {
			// Log optimization error but continue with update
			// log.Printf("Optimization failed for model: %v", err)
		}

		return r.db.Save(model).Error
	})
}

// Delete removes a model by ID with performance tracking and atomic counter
func (r *repository[T]) Delete(id uint) error {
	// Track operation count atomically
	r.atomicCounter.Increment()

	return r.performanceFacade.Track("repository.delete", func() error {
		var model T
		return r.db.Delete(&model, id).Error
	})
}

// GetPerformanceStats returns repository performance statistics
func (r *repository[T]) GetPerformanceStats() map[string]interface{} {
	stats := r.performanceFacade.GetStats()

	// Add repository-specific stats
	stats["repository"] = map[string]interface{}{
		"operations_count": r.atomicCounter.Get(),
		"object_pool_size": len(r.objectPool.pool),
	}

	return stats
}

// GetOptimizationStats returns optimization statistics
func (r *repository[T]) GetOptimizationStats() map[string]interface{} {
	return map[string]interface{}{
		"atomic_operations":              r.atomicCounter.Get(),
		"object_pool_usage":              len(r.objectPool.pool),
		"optimization_engine_strategies": len(r.optimizationEngine.strategies),
	}
}

// Where creates a query with conditions
func (r *repository[T]) Where(conditions map[string]any) Query[T] {
	query := r.db
	for field, value := range conditions {
		query = query.Where(field+" = ?", value)
	}
	return &queryBuilder[T]{
		db:                query,
		performanceFacade: r.performanceFacade,
		objectPool:        r.objectPool,
		atomicCounter:     r.atomicCounter,
	}
}

// WhereRaw creates a query with raw SQL
func (r *repository[T]) WhereRaw(query string, args ...any) Query[T] {
	return &queryBuilder[T]{
		db:                r.db.Where(query, args...),
		performanceFacade: r.performanceFacade,
		objectPool:        r.objectPool,
		atomicCounter:     r.atomicCounter,
	}
}

// Transaction executes a function within a database transaction
func (r *repository[T]) Transaction(fn func(Repository[T]) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		txRepo := &repository[T]{
			db:                 tx,
			performanceFacade:  r.performanceFacade,
			objectPool:         r.objectPool,
			atomicCounter:      r.atomicCounter,
			optimizationEngine: r.optimizationEngine,
		}
		return fn(txRepo)
	})
}

// WithContext returns a repository with context
func (r *repository[T]) WithContext(ctx context.Context) Repository[T] {
	return &repository[T]{
		db:                 r.db.WithContext(ctx),
		performanceFacade:  r.performanceFacade,
		objectPool:         r.objectPool,
		atomicCounter:      r.atomicCounter,
		optimizationEngine: r.optimizationEngine,
	}
}

// Exists checks if a model exists by ID with performance tracking and atomic counter
func (r *repository[T]) Exists(id uint) (bool, error) {
	// Track operation count atomically
	r.atomicCounter.Increment()

	var result bool
	err := r.performanceFacade.Track("repository.exists", func() error {
		var count int64
		if err := r.db.Model(new(T)).Where("id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		result = count > 0
		return nil
	})
	return result, err
}

// Count returns the total number of models with performance tracking and atomic counter
func (r *repository[T]) Count() (int64, error) {
	// Track operation count atomically
	r.atomicCounter.Increment()

	var result int64
	err := r.performanceFacade.Track("repository.count", func() error {
		if err := r.db.Model(new(T)).Count(&result).Error; err != nil {
			return err
		}
		return nil
	})
	return result, err
}

// CountWhere returns the count with conditions with performance tracking and atomic counter
func (r *repository[T]) CountWhere(conditions map[string]any) (int64, error) {
	// Track operation count atomically
	r.atomicCounter.Increment()

	var result int64
	err := r.performanceFacade.Track("repository.count_where", func() error {
		query := r.db.Model(new(T))
		for field, value := range conditions {
			query = query.Where(field+" = ?", value)
		}
		if err := query.Count(&result).Error; err != nil {
			return err
		}
		return nil
	})
	return result, err
}

// queryBuilder implements Query[T] with GORM and performance optimizations
type queryBuilder[T any] struct {
	db                *gorm.DB
	performanceFacade *PerformanceFacade
	objectPool        *ObjectPool[T]
	atomicCounter     *AtomicCounter
}

// Get retrieves all matching models with performance tracking and pipeline optimization
func (q *queryBuilder[T]) Get() ([]T, error) {
	// Track operation count atomically
	if q.atomicCounter != nil {
		q.atomicCounter.Increment()
	}

	var result []T
	err := q.performanceFacade.Track("query.get", func() error {
		if err := q.db.Find(&result).Error; err != nil {
			return err
		}

		// Use channel-based pipeline for processing large datasets
		if len(result) > 100 { // Only use pipeline for larger datasets
			pipeline := NewPipeline[T]()
			processed := pipeline.Execute(result)

			// Collect results
			optimized := make([]T, 0, len(result))
			for item := range processed {
				optimized = append(optimized, item)
			}
			result = optimized
		}

		return nil
	})
	return result, err
}

// First retrieves the first matching model with performance tracking and atomic counter
func (q *queryBuilder[T]) First() (*T, error) {
	// Track operation count atomically
	if q.atomicCounter != nil {
		q.atomicCounter.Increment()
	}

	var result *T
	err := q.performanceFacade.Track("query.first", func() error {
		// Create fresh entity for database operation (avoid concurrency issues)
		var entity T
		if err := q.db.First(&entity).Error; err != nil {
			return err
		}
		result = &entity
		return nil
	})
	return result, err
}

// Paginate retrieves models with pagination with performance tracking and pipeline optimization
func (q *queryBuilder[T]) Paginate(page, perPage int) ([]T, int64, error) {
	// Track operation count atomically
	if q.atomicCounter != nil {
		q.atomicCounter.Increment()
	}

	var result []T
	var total int64

	err := q.performanceFacade.Track("query.paginate", func() error {
		// Count total
		if err := q.db.Model(new(T)).Count(&total).Error; err != nil {
			return err
		}

		// Get paginated results
		offset := (page - 1) * perPage
		if err := q.db.Offset(offset).Limit(perPage).Find(&result).Error; err != nil {
			return err
		}

		// Use channel-based pipeline for processing large datasets
		if len(result) > 100 { // Only use pipeline for larger datasets
			pipeline := NewPipeline[T]()
			processed := pipeline.Execute(result)

			// Collect results
			optimized := make([]T, 0, len(result))
			for item := range processed {
				optimized = append(optimized, item)
			}
			result = optimized
		}

		return nil
	})

	return result, total, err
}

// Where adds a where clause
func (q *queryBuilder[T]) Where(field string, operator string, value any) Query[T] {
	return &queryBuilder[T]{
		db:                q.db.Where(field+" "+operator+" ?", value),
		performanceFacade: q.performanceFacade,
		objectPool:        q.objectPool,
		atomicCounter:     q.atomicCounter,
	}
}

// WhereIn adds a where in clause
func (q *queryBuilder[T]) WhereIn(field string, values []any) Query[T] {
	return &queryBuilder[T]{
		db:                q.db.Where(field+" IN ?", values),
		performanceFacade: q.performanceFacade,
		objectPool:        q.objectPool,
		atomicCounter:     q.atomicCounter,
	}
}

// OrderBy adds an order clause
func (q *queryBuilder[T]) OrderBy(field string, direction string) Query[T] {
	return &queryBuilder[T]{
		db:                q.db.Order(field + " " + direction),
		performanceFacade: q.performanceFacade,
		objectPool:        q.objectPool,
		atomicCounter:     q.atomicCounter,
	}
}

// Limit adds a limit clause
func (q *queryBuilder[T]) Limit(limit int) Query[T] {
	return &queryBuilder[T]{
		db:                q.db.Limit(limit),
		performanceFacade: q.performanceFacade,
		objectPool:        q.objectPool,
		atomicCounter:     q.atomicCounter,
	}
}

// Offset adds an offset clause
func (q *queryBuilder[T]) Offset(offset int) Query[T] {
	return &queryBuilder[T]{
		db:                q.db.Offset(offset),
		performanceFacade: q.performanceFacade,
		objectPool:        q.objectPool,
		atomicCounter:     q.atomicCounter,
	}
}

// Preload adds a preload clause
func (q *queryBuilder[T]) Preload(relation string) Query[T] {
	return &queryBuilder[T]{
		db:                q.db.Preload(relation),
		performanceFacade: q.performanceFacade,
		objectPool:        q.objectPool,
		atomicCounter:     q.atomicCounter,
	}
}

// WithContext returns a query with context
func (q *queryBuilder[T]) WithContext(ctx context.Context) Query[T] {
	return &queryBuilder[T]{
		db:                q.db.WithContext(ctx),
		performanceFacade: q.performanceFacade,
		objectPool:        q.objectPool,
		atomicCounter:     q.atomicCounter,
	}
}
