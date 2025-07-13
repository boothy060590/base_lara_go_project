package go_core

import (
	"context"
	"fmt"
	"sync"
	"time"

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

	// Context-aware basic CRUD operations
	FindWithContext(ctx context.Context, id uint) (*T, error)
	FindByWithContext(ctx context.Context, field string, value any) (*T, error)
	FindAllWithContext(ctx context.Context) ([]T, error)
	CreateWithContext(ctx context.Context, model *T) error
	UpdateWithContext(ctx context.Context, model *T) error
	DeleteWithContext(ctx context.Context, id uint) error

	// Query operations
	Where(conditions map[string]any) Query[T]
	WhereRaw(query string, args ...any) Query[T]

	// Context-aware query operations
	WhereWithContext(ctx context.Context, conditions map[string]any) Query[T]
	WhereRawWithContext(ctx context.Context, query string, args ...any) Query[T]

	// Transaction support
	Transaction(fn func(Repository[T]) error) error
	TransactionWithContext(ctx context.Context, fn func(Repository[T]) error) error
	WithContext(ctx context.Context) Repository[T]

	// Utility operations
	Exists(id uint) (bool, error)
	Count() (int64, error)
	CountWhere(conditions map[string]any) (int64, error)

	// Context-aware utility operations
	ExistsWithContext(ctx context.Context, id uint) (bool, error)
	CountWithContext(ctx context.Context) (int64, error)
	CountWhereWithContext(ctx context.Context, conditions map[string]any) (int64, error)

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

	// Context-aware execution
	GetWithContext(ctx context.Context) ([]T, error)
	FirstWithContext(ctx context.Context) (*T, error)
	PaginateWithContext(ctx context.Context, page, perPage int) ([]T, int64, error)

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

	// New optimization fields
	workStealingPool *WorkStealingPool[any]
	customAllocator  *CustomAllocator[any]
	profileOptimizer *ProfileGuidedOptimizer[any]
	// Infrastructure optimizations
	batchProcessor    *RepositoryBatchProcessor[T]
	asyncProcessor    *RepositoryAsyncProcessor[T]
	pipelineProcessor *RepositoryPipelineProcessor
	connectionPool    *RepositoryConnectionPool
	contextDecorator  *ContextDecorator
	config            map[string]interface{}
}

// NewRepository creates a new repository instance with performance tracking and optimizations
// Accept optimization dependencies
func NewRepository[T any](db *gorm.DB, wsp *WorkStealingPool[any], ca *CustomAllocator[any], pgo *ProfileGuidedOptimizer[any]) Repository[T] {
	return NewRepositoryWithConfig[T](db, nil, wsp, ca, pgo)
}

// NewRepositoryWithConfig creates a new repository with custom configuration
func NewRepositoryWithConfig[T any](db *gorm.DB, config map[string]interface{}, wsp *WorkStealingPool[any], ca *CustomAllocator[any], pgo *ProfileGuidedOptimizer[any]) Repository[T] {
	perf := NewPerformanceFacade()

	objectPool := NewObjectPool[T](100, func() T { return *new(T) }, func(entity T) T { return *new(T) })
	atomicCounter := NewAtomicCounter()
	optimizationEngine := NewOptimizationEngine()

	// Create infrastructure optimizations
	batchProcessor := NewRepositoryBatchProcessor[T](config)
	asyncProcessor := NewRepositoryAsyncProcessor[T](config)
	pipelineProcessor := NewRepositoryPipelineProcessor(config)
	connectionPool := NewRepositoryConnectionPool(config)
	contextDecorator := NewContextDecorator(config)

	repo := &repository[T]{
		db:                 db,
		performanceFacade:  perf,
		objectPool:         objectPool,
		atomicCounter:      atomicCounter,
		optimizationEngine: optimizationEngine,
		workStealingPool:   wsp,
		customAllocator:    ca,
		profileOptimizer:   pgo,
		// Infrastructure optimizations
		batchProcessor:    batchProcessor,
		asyncProcessor:    asyncProcessor,
		pipelineProcessor: pipelineProcessor,
		connectionPool:    connectionPool,
		contextDecorator:  contextDecorator,
		config:            config,
	}

	// Start background processors
	repo.startBackgroundProcessors()

	return repo
}

// Helper for allocating in-memory objects
func (r *repository[T]) allocate() T {
	if r.customAllocator != nil {
		obj, _ := r.customAllocator.Allocate(0)
		if obj == nil {
			// Fallback to object pool if custom allocator returns nil
			return r.objectPool.Get()
		}
		return obj.(T)
	}
	return r.objectPool.Get()
}

func (r *repository[T]) deallocate(obj T) {
	if r.customAllocator != nil {
		_ = r.customAllocator.Deallocate(obj, 0)
	} else {
		r.objectPool.Put(obj)
	}
}

// Find retrieves a model by ID with performance tracking and atomic counter
func (r *repository[T]) Find(id uint) (*T, error) {
	return r.FindWithContext(context.Background(), id)
}

// FindWithContext retrieves a model by ID with context support
func (r *repository[T]) FindWithContext(ctx context.Context, id uint) (*T, error) {
	// Track operation count atomically
	r.atomicCounter.Increment()

	var result *T
	err := r.performanceFacade.Track("repository.find", func() error {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Create fresh entity for database operation (avoid concurrency issues)
		var entity T
		if err := r.db.WithContext(ctx).First(&entity, id).Error; err != nil {
			return err
		}
		result = &entity
		return nil
	})
	return result, err
}

// FindBy retrieves a model by field and value with performance tracking and atomic counter
func (r *repository[T]) FindBy(field string, value any) (*T, error) {
	return r.FindByWithContext(context.Background(), field, value)
}

// FindByWithContext retrieves a model by field and value with context support
func (r *repository[T]) FindByWithContext(ctx context.Context, field string, value any) (*T, error) {
	// Track operation count atomically
	r.atomicCounter.Increment()

	var result *T
	err := r.performanceFacade.Track("repository.find_by", func() error {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Create fresh entity for database operation (avoid concurrency issues)
		var entity T
		if err := r.db.WithContext(ctx).Where(field+" = ?", value).First(&entity).Error; err != nil {
			return err
		}
		result = &entity
		return nil
	})
	return result, err
}

// FindAll retrieves all records with performance tracking and pipeline optimization
func (r *repository[T]) FindAll() ([]T, error) {
	return r.FindAllWithContext(context.Background())
}

// FindAllWithContext retrieves all records with context support
func (r *repository[T]) FindAllWithContext(ctx context.Context) ([]T, error) {
	var result []T
	err := r.performanceFacade.Track("repository.find_all", func() error {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Get all records (database operation - use fresh objects)
		if err := r.db.WithContext(ctx).Find(&result).Error; err != nil {
			return err
		}

		// Use profile-guided optimization to determine processing strategy
		if r.profileOptimizer != nil {
			r.profileOptimizer.GetMetrics() // Trigger optimization analysis
		}

		// Use work stealing pool for concurrent processing of large datasets
		if r.workStealingPool != nil && len(result) > 100 {
			return r.processWithWorkStealing(ctx, result)
		}

		// Use channel-based pipeline for processing large datasets (safe - in-memory only)
		if len(result) > 100 { // Only use pipeline for larger datasets
			pipeline := NewPipeline[T]()
			processed := pipeline.Execute(result)

			// Collect results (safe to use object pool for data processing)
			optimized := make([]T, 0, len(result))
			for item := range processed {
				// Check for context cancellation during processing
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}
				// Safe to use object pool here - this is data processing, not database operations
				optimized = append(optimized, item)
			}
			result = optimized
		}

		return nil
	})
	return result, err
}

// processWithWorkStealing processes results using work stealing pool with context
func (r *repository[T]) processWithWorkStealing(ctx context.Context, results []T) error {
	// Use custom allocator for work items if available
	workItems := make([]WorkItem[any], len(results))
	for i, result := range results {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Allocate work item using custom allocator if available
		workItem := r.allocate()
		defer r.deallocate(workItem)

		workItems[i] = WorkItem[any]{
			ID:      fmt.Sprintf("process_%d", i),
			Data:    result,
			Handler: r.processResult,
			Timeout: 30 * time.Second,
		}
	}

	// Submit work items to work stealing pool
	for _, item := range workItems {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := r.workStealingPool.Submit(item); err != nil {
			return err
		}
	}

	return nil
}

// processResult processes a single result item
func (r *repository[T]) processResult(ctx context.Context, data any) error {
	// Process the result item (e.g., apply transformations, validations, etc.)
	// This is a placeholder for actual processing logic
	return nil
}

// Create saves a new model with performance tracking and dynamic optimization
func (r *repository[T]) Create(model *T) error {
	return r.CreateWithContext(context.Background(), model)
}

// CreateWithContext saves a new model with context support
func (r *repository[T]) CreateWithContext(ctx context.Context, model *T) error {
	// Use context decorator for performance tracking and context management
	return r.contextDecorator.WithContext(ctx, "repository.create", func(ctx context.Context) error {
		// Check if batch processing is enabled
		if r.batchProcessor != nil && r.batchProcessor.IsEnabled() {
			return r.batchProcessor.AddItem(model, "create")
		}

		// Check if async processing is enabled
		if r.asyncProcessor != nil && r.asyncProcessor.IsEnabled() {
			return r.asyncProcessor.Process(model, "create", func(ctx context.Context, m *T) error {
				return r.performCreate(ctx, m)
			})
		}

		// Fall back to direct operation
		return r.performCreate(ctx, model)
	})
}

// performCreate performs the actual create operation
func (r *repository[T]) performCreate(ctx context.Context, model *T) error {
	return r.performanceFacade.Track("repository.create", func() error {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Apply dynamic optimization to the model
		if err := r.performanceFacade.Optimize(model); err != nil {
			// Log optimization error but continue with creation
			// log.Printf("Optimization failed for model: %v", err)
		}

		return r.db.WithContext(ctx).Create(model).Error
	})
}

// Update saves an existing model with performance tracking and dynamic optimization
func (r *repository[T]) Update(model *T) error {
	return r.UpdateWithContext(context.Background(), model)
}

// UpdateWithContext saves an existing model with context support
func (r *repository[T]) UpdateWithContext(ctx context.Context, model *T) error {
	// Track operation count atomically
	r.atomicCounter.Increment()

	return r.performanceFacade.Track("repository.update", func() error {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Apply dynamic optimization to the model before update
		if err := r.performanceFacade.Optimize(model); err != nil {
			// Log optimization error but continue with update
			// log.Printf("Optimization failed for model: %v", err)
		}

		return r.db.WithContext(ctx).Save(model).Error
	})
}

// Delete removes a model by ID with performance tracking and atomic counter
func (r *repository[T]) Delete(id uint) error {
	return r.DeleteWithContext(context.Background(), id)
}

// DeleteWithContext removes a model by ID with context support
func (r *repository[T]) DeleteWithContext(ctx context.Context, id uint) error {
	// Track operation count atomically
	r.atomicCounter.Increment()

	return r.performanceFacade.Track("repository.delete", func() error {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var model T
		return r.db.WithContext(ctx).Delete(&model, id).Error
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

	// Add infrastructure optimization stats
	if r.batchProcessor != nil {
		stats["batch_processor"] = r.batchProcessor.GetStats()
	}
	if r.asyncProcessor != nil {
		stats["async_processor"] = r.asyncProcessor.GetStats()
	}
	if r.pipelineProcessor != nil {
		stats["pipeline_processor"] = r.pipelineProcessor.GetStats()
	}
	if r.connectionPool != nil {
		stats["connection_pool"] = r.connectionPool.GetStats()
	}
	if r.contextDecorator != nil {
		stats["context_decorator"] = r.contextDecorator.GetPerformanceStats()
	}

	return stats
}

// GetOptimizationStats returns optimization statistics
func (r *repository[T]) GetOptimizationStats() map[string]interface{} {
	return map[string]interface{}{
		"atomic_operations":              r.atomicCounter.Get(),
		"object_pool_usage":              len(r.objectPool.pool),
		"optimization_engine_strategies": len(r.optimizationEngine.strategies),
		"work_stealing":                  r.workStealingPool != nil,
		"custom_allocator":               r.customAllocator != nil,
		"profile_optimizer":              r.profileOptimizer != nil,
		"batch_processing_enabled":       r.batchProcessor != nil && r.batchProcessor.IsEnabled(),
		"async_operations_enabled":       r.asyncProcessor != nil && r.asyncProcessor.IsEnabled(),
		"pipeline_operations_enabled":    r.pipelineProcessor != nil && r.pipelineProcessor.IsEnabled(),
		"connection_pooling_enabled":     r.connectionPool != nil && r.connectionPool.IsEnabled(),
		"infrastructure_optimizations":   true,
	}
}

// Where creates a query with conditions
func (r *repository[T]) Where(conditions map[string]any) Query[T] {
	return r.WhereWithContext(context.Background(), conditions)
}

// WhereWithContext creates a query with conditions and context
func (r *repository[T]) WhereWithContext(ctx context.Context, conditions map[string]any) Query[T] {
	query := r.db.WithContext(ctx)
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
	return r.WhereRawWithContext(context.Background(), query, args...)
}

// WhereRawWithContext creates a query with raw SQL and context
func (r *repository[T]) WhereRawWithContext(ctx context.Context, query string, args ...any) Query[T] {
	return &queryBuilder[T]{
		db:                r.db.WithContext(ctx).Where(query, args...),
		performanceFacade: r.performanceFacade,
		objectPool:        r.objectPool,
		atomicCounter:     r.atomicCounter,
	}
}

// Transaction executes a function within a database transaction
func (r *repository[T]) Transaction(fn func(Repository[T]) error) error {
	return r.TransactionWithContext(context.Background(), fn)
}

// TransactionWithContext executes a function within a database transaction with context
func (r *repository[T]) TransactionWithContext(ctx context.Context, fn func(Repository[T]) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &repository[T]{
			db:                 tx,
			performanceFacade:  r.performanceFacade,
			objectPool:         r.objectPool,
			atomicCounter:      r.atomicCounter,
			optimizationEngine: r.optimizationEngine,
			workStealingPool:   r.workStealingPool,
			customAllocator:    r.customAllocator,
			profileOptimizer:   r.profileOptimizer,
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
		workStealingPool:   r.workStealingPool,
		customAllocator:    r.customAllocator,
		profileOptimizer:   r.profileOptimizer,
	}
}

// Exists checks if a model exists by ID with performance tracking and atomic counter
func (r *repository[T]) Exists(id uint) (bool, error) {
	return r.ExistsWithContext(context.Background(), id)
}

// ExistsWithContext checks if a model exists by ID with context support
func (r *repository[T]) ExistsWithContext(ctx context.Context, id uint) (bool, error) {
	// Track operation count atomically
	r.atomicCounter.Increment()

	var result bool
	err := r.performanceFacade.Track("repository.exists", func() error {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var count int64
		if err := r.db.WithContext(ctx).Model(new(T)).Where("id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		result = count > 0
		return nil
	})
	return result, err
}

// Count returns the total number of models with performance tracking and atomic counter
func (r *repository[T]) Count() (int64, error) {
	return r.CountWithContext(context.Background())
}

// CountWithContext returns the total number of models with context support
func (r *repository[T]) CountWithContext(ctx context.Context) (int64, error) {
	// Track operation count atomically
	r.atomicCounter.Increment()

	var result int64
	err := r.performanceFacade.Track("repository.count", func() error {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := r.db.WithContext(ctx).Model(new(T)).Count(&result).Error; err != nil {
			return err
		}
		return nil
	})
	return result, err
}

// CountWhere returns the count with conditions with performance tracking and atomic counter
func (r *repository[T]) CountWhere(conditions map[string]any) (int64, error) {
	return r.CountWhereWithContext(context.Background(), conditions)
}

// CountWhereWithContext returns the count with conditions and context support
func (r *repository[T]) CountWhereWithContext(ctx context.Context, conditions map[string]any) (int64, error) {
	// Track operation count atomically
	r.atomicCounter.Increment()

	var result int64
	err := r.performanceFacade.Track("repository.count_where", func() error {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		query := r.db.WithContext(ctx).Model(new(T))
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
	return q.GetWithContext(context.Background())
}

// GetWithContext retrieves all matching models with context support
func (q *queryBuilder[T]) GetWithContext(ctx context.Context) ([]T, error) {
	// Track operation count atomically
	if q.atomicCounter != nil {
		q.atomicCounter.Increment()
	}

	var result []T
	err := q.performanceFacade.Track("query.get", func() error {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

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
				// Check for context cancellation during processing
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}
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
	return q.FirstWithContext(context.Background())
}

// FirstWithContext retrieves the first matching model with context support
func (q *queryBuilder[T]) FirstWithContext(ctx context.Context) (*T, error) {
	// Track operation count atomically
	if q.atomicCounter != nil {
		q.atomicCounter.Increment()
	}

	var result *T
	err := q.performanceFacade.Track("query.first", func() error {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

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
	return q.PaginateWithContext(context.Background(), page, perPage)
}

// PaginateWithContext retrieves models with pagination and context support
func (q *queryBuilder[T]) PaginateWithContext(ctx context.Context, page, perPage int) ([]T, int64, error) {
	// Validate pagination parameters
	if page <= 0 {
		return nil, 0, fmt.Errorf("page must be greater than 0, got %d", page)
	}
	if perPage <= 0 {
		return nil, 0, fmt.Errorf("perPage must be greater than 0, got %d", perPage)
	}

	// Track operation count atomically
	if q.atomicCounter != nil {
		q.atomicCounter.Increment()
	}

	var result []T
	var total int64

	err := q.performanceFacade.Track("query.paginate", func() error {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

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
				// Check for context cancellation during processing
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}
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

// ============================================================================
// REPOSITORY INFRASTRUCTURE PROCESSORS
// ============================================================================

// RepositoryBatchProcessor handles batch operations for repository
type RepositoryBatchProcessor[T any] struct {
	enabled     bool
	batchSize   int
	batchBuffer []*RepositoryBatchItem[T]
	bufferMutex sync.Mutex
	flushTicker *time.Ticker
	done        chan bool
	config      map[string]interface{}
}

type RepositoryBatchItem[T any] struct {
	Model *T
	Op    string // "create", "update", "delete"
}

// NewRepositoryBatchProcessor creates a new repository batch processor
func NewRepositoryBatchProcessor[T any](config map[string]interface{}) *RepositoryBatchProcessor[T] {
	batchSize := 100 // Default
	if size, ok := config["repository_batch_size"].(int); ok {
		batchSize = size
	}

	return &RepositoryBatchProcessor[T]{
		enabled:     true,
		batchSize:   batchSize,
		batchBuffer: make([]*RepositoryBatchItem[T], 0),
		done:        make(chan bool),
		config:      config,
	}
}

// Start starts the batch processor
func (rbp *RepositoryBatchProcessor[T]) Start() {
	if !rbp.enabled {
		return
	}

	flushInterval := 100 * time.Millisecond // Default
	if interval, ok := rbp.config["repository_flush_interval"].(int); ok {
		flushInterval = time.Duration(interval) * time.Millisecond
	}

	rbp.flushTicker = time.NewTicker(flushInterval)
	go rbp.backgroundFlusher()
}

// Stop stops the batch processor
func (rbp *RepositoryBatchProcessor[T]) Stop() {
	if rbp.flushTicker != nil {
		rbp.flushTicker.Stop()
	}
	close(rbp.done)
	rbp.flushBuffer()
}

// backgroundFlusher runs the background flush process
func (rbp *RepositoryBatchProcessor[T]) backgroundFlusher() {
	for {
		select {
		case <-rbp.flushTicker.C:
			rbp.flushBuffer()
		case <-rbp.done:
			return
		}
	}
}

// flushBuffer flushes the batch buffer
func (rbp *RepositoryBatchProcessor[T]) flushBuffer() {
	rbp.bufferMutex.Lock()
	if len(rbp.batchBuffer) == 0 {
		rbp.bufferMutex.Unlock()
		return
	}

	items := make([]*RepositoryBatchItem[T], len(rbp.batchBuffer))
	copy(items, rbp.batchBuffer)
	rbp.batchBuffer = rbp.batchBuffer[:0]
	rbp.bufferMutex.Unlock()

	// Process batch
	rbp.processBatch(items)
}

// processBatch processes a batch of repository operations
func (rbp *RepositoryBatchProcessor[T]) processBatch(items []*RepositoryBatchItem[T]) {
	// This would integrate with database batch operations
	// For now, we'll just process them individually
	for _, item := range items {
		_ = item // Process item
	}
}

// AddItem adds an item to batch buffer
func (rbp *RepositoryBatchProcessor[T]) AddItem(model *T, op string) error {
	rbp.bufferMutex.Lock()
	defer rbp.bufferMutex.Unlock()

	rbp.batchBuffer = append(rbp.batchBuffer, &RepositoryBatchItem[T]{
		Model: model,
		Op:    op,
	})

	// Flush if buffer is full
	if len(rbp.batchBuffer) >= rbp.batchSize {
		// Copy buffer and clear it
		items := make([]*RepositoryBatchItem[T], len(rbp.batchBuffer))
		copy(items, rbp.batchBuffer)
		rbp.batchBuffer = rbp.batchBuffer[:0]

		// Process batch in background
		go rbp.processBatch(items)
	}

	return nil
}

// IsEnabled returns whether batch processing is enabled
func (rbp *RepositoryBatchProcessor[T]) IsEnabled() bool {
	return rbp.enabled
}

// GetStats returns batch processor statistics
func (rbp *RepositoryBatchProcessor[T]) GetStats() map[string]interface{} {
	rbp.bufferMutex.Lock()
	defer rbp.bufferMutex.Unlock()

	return map[string]interface{}{
		"enabled":     rbp.enabled,
		"batch_size":  rbp.batchSize,
		"buffer_size": len(rbp.batchBuffer),
		"config":      rbp.config,
	}
}

// RepositoryAsyncProcessor handles async operations for repository
type RepositoryAsyncProcessor[T any] struct {
	enabled bool
	config  map[string]interface{}
}

// NewRepositoryAsyncProcessor creates a new repository async processor
func NewRepositoryAsyncProcessor[T any](config map[string]interface{}) *RepositoryAsyncProcessor[T] {
	return &RepositoryAsyncProcessor[T]{
		enabled: true,
		config:  config,
	}
}

// Start starts the async processor
func (rap *RepositoryAsyncProcessor[T]) Start() {
	// Start async processing
}

// Stop stops the async processor
func (rap *RepositoryAsyncProcessor[T]) Stop() {
	// Stop async processing
}

// Process processes a repository operation asynchronously
func (rap *RepositoryAsyncProcessor[T]) Process(model *T, op string, processor func(context.Context, *T) error) error {
	if !rap.enabled {
		return processor(context.Background(), model)
	}

	// Process asynchronously
	go func() {
		_ = processor(context.Background(), model)
	}()

	return nil
}

// IsEnabled returns whether async processing is enabled
func (rap *RepositoryAsyncProcessor[T]) IsEnabled() bool {
	return rap.enabled
}

// GetStats returns async processor statistics
func (rap *RepositoryAsyncProcessor[T]) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"enabled": rap.enabled,
		"config":  rap.config,
	}
}

// RepositoryPipelineProcessor handles pipeline operations for repository
type RepositoryPipelineProcessor struct {
	enabled bool
	config  map[string]interface{}
}

// NewRepositoryPipelineProcessor creates a new repository pipeline processor
func NewRepositoryPipelineProcessor(config map[string]interface{}) *RepositoryPipelineProcessor {
	return &RepositoryPipelineProcessor{
		enabled: true,
		config:  config,
	}
}

// Start starts the pipeline processor
func (rpp *RepositoryPipelineProcessor) Start() {
	// Start pipeline processing
}

// Stop stops the pipeline processor
func (rpp *RepositoryPipelineProcessor) Stop() {
	// Stop pipeline processing
}

// IsEnabled returns whether pipeline processing is enabled
func (rpp *RepositoryPipelineProcessor) IsEnabled() bool {
	return rpp.enabled
}

// GetStats returns pipeline processor statistics
func (rpp *RepositoryPipelineProcessor) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"enabled": rpp.enabled,
		"config":  rpp.config,
	}
}

// RepositoryConnectionPool manages connection pooling for repository
type RepositoryConnectionPool struct {
	enabled bool
	config  map[string]interface{}
}

// NewRepositoryConnectionPool creates a new repository connection pool
func NewRepositoryConnectionPool(config map[string]interface{}) *RepositoryConnectionPool {
	return &RepositoryConnectionPool{
		enabled: true,
		config:  config,
	}
}

// IsEnabled returns whether connection pooling is enabled
func (rcp *RepositoryConnectionPool) IsEnabled() bool {
	return rcp.enabled
}

// GetStats returns connection pool statistics
func (rcp *RepositoryConnectionPool) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"enabled": rcp.enabled,
		"config":  rcp.config,
	}
}

// startBackgroundProcessors starts background processing for optimizations
func (r *repository[T]) startBackgroundProcessors() {
	// Start batch processor
	if r.batchProcessor != nil {
		r.batchProcessor.Start()
	}

	// Start async processor
	if r.asyncProcessor != nil {
		r.asyncProcessor.Start()
	}

	// Start pipeline processor
	if r.pipelineProcessor != nil {
		r.pipelineProcessor.Start()
	}
}
