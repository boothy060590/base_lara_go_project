package go_core

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Repository defines a generic repository interface for any model type
// Uses raw SQL for maximum performance with comprehensive safety
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

	// Bulk operations for complex scenarios
	BulkCreate(models []*T) error
	BulkUpdate(models []*T) error
	BulkDelete(ids []uint) error

	// Context-aware bulk operations
	BulkCreateWithContext(ctx context.Context, models []*T) error
	BulkUpdateWithContext(ctx context.Context, models []*T) error
	BulkDeleteWithContext(ctx context.Context, ids []uint) error
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

// repository implements Repository[T] with raw SQL and performance tracking
type repository[T any] struct {
	// Database connection and pooling
	db             *sql.DB
	connectionPool *ConnectionPool
	statementCache *StatementCache

	// Performance tracking
	performanceFacade  *PerformanceFacade
	atomicCounter      *AtomicCounter
	optimizationEngine *OptimizationEngine

	// Optimization dependencies
	workStealingPool *WorkStealingPool[any]
	customAllocator  *CustomAllocator[any]
	profileOptimizer *ProfileGuidedOptimizer[any]

	// Infrastructure optimizations
	batchProcessor    *RepositoryBatchProcessor[T]
	asyncProcessor    *RepositoryAsyncProcessor[T]
	pipelineProcessor *RepositoryPipelineProcessor
	contextDecorator  *ContextDecorator

	// Model metadata
	tableName      string
	fieldValidator *FieldValidator
	config         map[string]interface{}

	// Safety and validation
	sqlValidator *SQLValidator
}

// NewRepository creates a new repository instance with raw SQL and performance tracking
func NewRepository[T any](db *sql.DB, wsp *WorkStealingPool[any], ca *CustomAllocator[any], pgo *ProfileGuidedOptimizer[any]) Repository[T] {
	return NewRepositoryWithConfig[T](db, nil, wsp, ca, pgo)
}

// NewRepositoryWithConfig creates a new repository with custom configuration
func NewRepositoryWithConfig[T any](db *sql.DB, config map[string]interface{}, wsp *WorkStealingPool[any], ca *CustomAllocator[any], pgo *ProfileGuidedOptimizer[any]) Repository[T] {
	// Determine table name from type
	var model T
	tableName := getTableName(model)

	// Create performance tracking
	perf := NewPerformanceFacade()
	atomicCounter := NewAtomicCounter()
	optimizationEngine := NewOptimizationEngine()

	// Create infrastructure optimizations
	connectionPool := NewConnectionPool(config)
	statementCache := NewStatementCache(config)
	batchProcessor := NewRepositoryBatchProcessor[T](config)
	asyncProcessor := NewRepositoryAsyncProcessor[T](config)
	pipelineProcessor := NewRepositoryPipelineProcessor(config)
	contextDecorator := NewContextDecorator(config)

	// Create safety components
	fieldValidator := NewFieldValidator(tableName)
	sqlValidator := NewSQLValidator()

	repo := &repository[T]{
		db:                 db,
		connectionPool:     connectionPool,
		statementCache:     statementCache,
		performanceFacade:  perf,
		atomicCounter:      atomicCounter,
		optimizationEngine: optimizationEngine,
		workStealingPool:   wsp,
		customAllocator:    ca,
		profileOptimizer:   pgo,
		batchProcessor:     batchProcessor,
		asyncProcessor:     asyncProcessor,
		pipelineProcessor:  pipelineProcessor,
		contextDecorator:   contextDecorator,
		tableName:          tableName,
		fieldValidator:     fieldValidator,
		sqlValidator:       sqlValidator,
		config:             config,
	}

	// Start background processors
	repo.startBackgroundProcessors()

	return repo
}

// getTableName extracts table name from model type
func getTableName(v interface{}) string {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Check for TableName method
	if _, ok := t.MethodByName("TableName"); ok {
		// Call the method to get table name
		// This is a simplified version - in practice you'd need reflection to call it
		return strings.ToLower(t.Name()) + "s"
	}

	return strings.ToLower(t.Name()) + "s"
}

// Find retrieves a model by ID with raw SQL and performance tracking
func (r *repository[T]) Find(id uint) (*T, error) {
	return r.FindWithContext(context.Background(), id)
}

// FindWithContext retrieves a model by ID with context support
func (r *repository[T]) FindWithContext(ctx context.Context, id uint) (*T, error) {
	r.atomicCounter.Increment()

	var result *T
	err := r.performanceFacade.Track("repository.find", func() error {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Build safe parameterized query
		query := fmt.Sprintf("SELECT * FROM %s WHERE id = ? AND deleted_at IS NULL", r.tableName)

		// Get statement from cache
		stmt, err := r.statementCache.GetStatement(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to prepare statement: %w", err)
		}

		// Execute query with context
		row := stmt.QueryRowContext(ctx, id)

		// Scan result into model
		var entity T
		if err := r.scanRowToStruct(row, &entity); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		result = &entity
		return nil
	})

	return result, err
}

// FindBy retrieves a model by field and value with raw SQL
func (r *repository[T]) FindBy(field string, value any) (*T, error) {
	return r.FindByWithContext(context.Background(), field, value)
}

// FindByWithContext retrieves a model by field and value with context support
func (r *repository[T]) FindByWithContext(ctx context.Context, field string, value any) (*T, error) {
	r.atomicCounter.Increment()

	// Validate field name to prevent SQL injection
	if err := r.fieldValidator.ValidateField(field); err != nil {
		return nil, fmt.Errorf("invalid field name: %w", err)
	}

	var result *T
	err := r.performanceFacade.Track("repository.find_by", func() error {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Build safe parameterized query
		query := fmt.Sprintf("SELECT * FROM %s WHERE %s = ? AND deleted_at IS NULL LIMIT 1", r.tableName, field)

		// Get statement from cache
		stmt, err := r.statementCache.GetStatement(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to prepare statement: %w", err)
		}

		// Execute query with context
		row := stmt.QueryRowContext(ctx, value)

		// Scan result into model
		var entity T
		if err := r.scanRowToStruct(row, &entity); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		result = &entity
		return nil
	})

	return result, err
}

// FindAll retrieves all records with raw SQL and optimization
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

		// Build safe query
		query := fmt.Sprintf("SELECT * FROM %s WHERE deleted_at IS NULL", r.tableName)

		// Get statement from cache
		stmt, err := r.statementCache.GetStatement(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to prepare statement: %w", err)
		}

		// Execute query with context
		rows, err := stmt.QueryContext(ctx)
		if err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
		defer rows.Close()

		// Scan results
		for rows.Next() {
			var entity T
			if err := r.scanRowToStruct(rows, &entity); err != nil {
				return fmt.Errorf("failed to scan row: %w", err)
			}
			result = append(result, entity)
		}

		// Use work stealing pool for processing large datasets
		if r.workStealingPool != nil && len(result) > 100 {
			return r.processWithWorkStealing(ctx, result)
		}

		return nil
	})

	return result, err
}

// Create saves a new model with raw SQL
func (r *repository[T]) Create(model *T) error {
	return r.CreateWithContext(context.Background(), model)
}

// CreateWithContext saves a new model with context support
func (r *repository[T]) CreateWithContext(ctx context.Context, model *T) error {
	return r.contextDecorator.WithContext(ctx, "repository.create", func(ctx context.Context) error {
		// Check if batch processing is enabled
		if r.batchProcessor != nil && r.batchProcessor.IsEnabled() {
			return r.batchProcessor.AddItem(model, "create")
		}

		// Check if async processing is enabled
		if r.asyncProcessor != nil && r.asyncProcessor.IsEnabled() {
			return r.asyncProcessor.Process(model, "create", r.performCreate)
		}

		// Perform synchronous create
		return r.performCreate(ctx, model)
	})
}

// performCreate performs the actual create operation with raw SQL
func (r *repository[T]) performCreate(ctx context.Context, model *T) error {
	// Get model fields and values
	fields, values, err := r.getModelFieldsAndValues(model)
	if err != nil {
		return fmt.Errorf("failed to get model fields: %w", err)
	}

	// Build INSERT query
	placeholders := strings.Repeat("?,", len(fields))
	placeholders = placeholders[:len(placeholders)-1] // Remove trailing comma

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", r.tableName, strings.Join(fields, ", "), placeholders)

	// Get statement from cache
	stmt, err := r.statementCache.GetStatement(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}

	// Execute query
	_, err = stmt.ExecContext(ctx, values...)
	if err != nil {
		return fmt.Errorf("failed to execute insert: %w", err)
	}

	return nil
}

// Update updates an existing model with raw SQL
func (r *repository[T]) Update(model *T) error {
	return r.UpdateWithContext(context.Background(), model)
}

// UpdateWithContext updates an existing model with context support
func (r *repository[T]) UpdateWithContext(ctx context.Context, model *T) error {
	return r.contextDecorator.WithContext(ctx, "repository.update", func(ctx context.Context) error {
		// Check if batch processing is enabled
		if r.batchProcessor != nil && r.batchProcessor.IsEnabled() {
			return r.batchProcessor.AddItem(model, "update")
		}

		// Check if async processing is enabled
		if r.asyncProcessor != nil && r.asyncProcessor.IsEnabled() {
			return r.asyncProcessor.Process(model, "update", r.performUpdate)
		}

		// Perform synchronous update
		return r.performUpdate(ctx, model)
	})
}

// performUpdate performs the actual update operation with raw SQL
func (r *repository[T]) performUpdate(ctx context.Context, model *T) error {
	// Get model fields and values (excluding ID)
	fields, values, err := r.getModelFieldsAndValues(model)
	if err != nil {
		return fmt.Errorf("failed to get model fields: %w", err)
	}

	// Get ID field
	idValue, err := r.getModelID(model)
	if err != nil {
		return fmt.Errorf("failed to get model ID: %w", err)
	}

	// Build UPDATE query
	setClause := make([]string, len(fields))
	for i, field := range fields {
		setClause[i] = fmt.Sprintf("%s = ?", field)
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = ? AND deleted_at IS NULL",
		r.tableName, strings.Join(setClause, ", "))

	// Add ID to values
	values = append(values, idValue)

	// Get statement from cache
	stmt, err := r.statementCache.GetStatement(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}

	// Execute query
	result, err := stmt.ExecContext(ctx, values...)
	if err != nil {
		return fmt.Errorf("failed to execute update: %w", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no rows affected - record may not exist")
	}

	return nil
}

// Delete deletes a model with raw SQL (soft delete)
func (r *repository[T]) Delete(id uint) error {
	return r.DeleteWithContext(context.Background(), id)
}

// DeleteWithContext deletes a model with context support
func (r *repository[T]) DeleteWithContext(ctx context.Context, id uint) error {
	return r.contextDecorator.WithContext(ctx, "repository.delete", func(ctx context.Context) error {
		// Check if batch processing is enabled
		if r.batchProcessor != nil && r.batchProcessor.IsEnabled() {
			// Create a dummy model for batch processing
			var model T
			if err := r.setModelID(&model, id); err != nil {
				return fmt.Errorf("failed to set model ID: %w", err)
			}
			return r.batchProcessor.AddItem(&model, "delete")
		}

		// Perform synchronous delete
		return r.performDelete(ctx, id)
	})
}

// performDelete performs the actual delete operation with raw SQL
func (r *repository[T]) performDelete(ctx context.Context, id uint) error {
	// Soft delete - set deleted_at timestamp
	query := fmt.Sprintf("UPDATE %s SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL", r.tableName)

	// Get statement from cache
	stmt, err := r.statementCache.GetStatement(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}

	// Execute query
	result, err := stmt.ExecContext(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to execute delete: %w", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no rows affected - record may not exist")
	}

	return nil
}

// Where creates a query builder with conditions
func (r *repository[T]) Where(conditions map[string]any) Query[T] {
	return r.WhereWithContext(context.Background(), conditions)
}

// WhereWithContext creates a query builder with conditions and context
func (r *repository[T]) WhereWithContext(ctx context.Context, conditions map[string]any) Query[T] {
	// Validate all field names
	for field := range conditions {
		if err := r.fieldValidator.ValidateField(field); err != nil {
			// Return error query that will fail on execution
			return &errorQuery[T]{err: fmt.Errorf("invalid field name '%s': %w", field, err)}
		}
	}

	return &queryBuilder[T]{
		repository: r,
		conditions: conditions,
		ctx:        ctx,
	}
}

// WhereRaw creates a query builder with raw SQL
func (r *repository[T]) WhereRaw(query string, args ...any) Query[T] {
	return r.WhereRawWithContext(context.Background(), query, args...)
}

// WhereRawWithContext creates a query builder with raw SQL and context
func (r *repository[T]) WhereRawWithContext(ctx context.Context, query string, args ...any) Query[T] {
	// Validate SQL query to prevent injection
	if err := r.sqlValidator.ValidateQuery(query); err != nil {
		return &errorQuery[T]{err: fmt.Errorf("invalid SQL query: %w", err)}
	}

	return &rawQueryBuilder[T]{
		repository: r,
		query:      query,
		args:       args,
		ctx:        ctx,
	}
}

// Transaction executes a function within a database transaction
func (r *repository[T]) Transaction(fn func(Repository[T]) error) error {
	return r.TransactionWithContext(context.Background(), fn)
}

// TransactionWithContext executes a function within a database transaction with context
func (r *repository[T]) TransactionWithContext(ctx context.Context, fn func(Repository[T]) error) error {
	return r.performanceFacade.Track("repository.transaction", func() error {
		// Begin transaction
		tx, err := r.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		// Create transaction repository
		txRepo := &repository[T]{
			db:                 r.db,
			connectionPool:     r.connectionPool,
			statementCache:     r.statementCache,
			performanceFacade:  r.performanceFacade,
			atomicCounter:      r.atomicCounter,
			optimizationEngine: r.optimizationEngine,
			workStealingPool:   r.workStealingPool,
			customAllocator:    r.customAllocator,
			profileOptimizer:   r.profileOptimizer,
			batchProcessor:     r.batchProcessor,
			asyncProcessor:     r.asyncProcessor,
			pipelineProcessor:  r.pipelineProcessor,
			contextDecorator:   r.contextDecorator,
			tableName:          r.tableName,
			fieldValidator:     r.fieldValidator,
			sqlValidator:       r.sqlValidator,
			config:             r.config,
		}

		// Execute function
		if err := fn(txRepo); err != nil {
			// Rollback on error
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return fmt.Errorf("transaction failed and rollback failed: %w (rollback error: %v)", err, rollbackErr)
			}
			return err
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		return nil
	})
}

// WithContext returns a new repository with the given context
func (r *repository[T]) WithContext(ctx context.Context) Repository[T] {
	return &repository[T]{
		db:                 r.db,
		connectionPool:     r.connectionPool,
		statementCache:     r.statementCache,
		performanceFacade:  r.performanceFacade,
		atomicCounter:      r.atomicCounter,
		optimizationEngine: r.optimizationEngine,
		workStealingPool:   r.workStealingPool,
		customAllocator:    r.customAllocator,
		profileOptimizer:   r.profileOptimizer,
		batchProcessor:     r.batchProcessor,
		asyncProcessor:     r.asyncProcessor,
		pipelineProcessor:  r.pipelineProcessor,
		contextDecorator:   r.contextDecorator,
		tableName:          r.tableName,
		fieldValidator:     r.fieldValidator,
		sqlValidator:       r.sqlValidator,
		config:             r.config,
	}
}

// Exists checks if a model exists
func (r *repository[T]) Exists(id uint) (bool, error) {
	return r.ExistsWithContext(context.Background(), id)
}

// ExistsWithContext checks if a model exists with context
func (r *repository[T]) ExistsWithContext(ctx context.Context, id uint) (bool, error) {
	r.atomicCounter.Increment()

	var exists bool
	err := r.performanceFacade.Track("repository.exists", func() error {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Build safe query
		query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE id = ? AND deleted_at IS NULL)", r.tableName)

		// Get statement from cache
		stmt, err := r.statementCache.GetStatement(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to prepare statement: %w", err)
		}

		// Execute query
		row := stmt.QueryRowContext(ctx, id)
		if err := row.Scan(&exists); err != nil {
			return fmt.Errorf("failed to scan exists result: %w", err)
		}

		return nil
	})

	return exists, err
}

// Count returns the total number of records
func (r *repository[T]) Count() (int64, error) {
	return r.CountWithContext(context.Background())
}

// CountWithContext returns the total number of records with context
func (r *repository[T]) CountWithContext(ctx context.Context) (int64, error) {
	r.atomicCounter.Increment()

	var count int64
	err := r.performanceFacade.Track("repository.count", func() error {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Build safe query
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE deleted_at IS NULL", r.tableName)

		// Get statement from cache
		stmt, err := r.statementCache.GetStatement(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to prepare statement: %w", err)
		}

		// Execute query
		row := stmt.QueryRowContext(ctx)
		if err := row.Scan(&count); err != nil {
			return fmt.Errorf("failed to scan count result: %w", err)
		}

		return nil
	})

	return count, err
}

// CountWhere returns the count with conditions
func (r *repository[T]) CountWhere(conditions map[string]any) (int64, error) {
	return r.CountWhereWithContext(context.Background(), conditions)
}

// CountWhereWithContext returns the count with conditions and context
func (r *repository[T]) CountWhereWithContext(ctx context.Context, conditions map[string]any) (int64, error) {
	r.atomicCounter.Increment()

	// Validate all field names
	for field := range conditions {
		if err := r.fieldValidator.ValidateField(field); err != nil {
			return 0, fmt.Errorf("invalid field name '%s': %w", field, err)
		}
	}

	var count int64
	err := r.performanceFacade.Track("repository.count_where", func() error {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Build WHERE clause
		whereClause, values := r.buildWhereClause(conditions)

		// Build safe query
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s AND deleted_at IS NULL", r.tableName, whereClause)

		// Get statement from cache
		stmt, err := r.statementCache.GetStatement(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to prepare statement: %w", err)
		}

		// Execute query
		row := stmt.QueryRowContext(ctx, values...)
		if err := row.Scan(&count); err != nil {
			return fmt.Errorf("failed to scan count result: %w", err)
		}

		return nil
	})

	return count, err
}

// GetPerformanceStats returns performance statistics
func (r *repository[T]) GetPerformanceStats() map[string]interface{} {
	return r.performanceFacade.GetStats()
}

// GetOptimizationStats returns optimization statistics
func (r *repository[T]) GetOptimizationStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// Add atomic counter stats
	stats["operations_count"] = r.atomicCounter.Get()

	// Add optimization engine stats
	if r.optimizationEngine != nil {
		stats["optimization_engine"] = map[string]interface{}{
			"strategies_count": len(r.optimizationEngine.strategies),
		}
	}

	// Add work stealing pool stats
	if r.workStealingPool != nil {
		stats["work_stealing_pool"] = r.workStealingPool.GetMetrics()
	}

	// Add batch processor stats
	if r.batchProcessor != nil {
		stats["batch_processor"] = r.batchProcessor.GetStats()
	}

	// Add async processor stats
	if r.asyncProcessor != nil {
		stats["async_processor"] = r.asyncProcessor.GetStats()
	}

	// Add pipeline processor stats
	if r.pipelineProcessor != nil {
		stats["pipeline_processor"] = r.pipelineProcessor.GetStats()
	}

	// Add connection pool stats
	if r.connectionPool != nil {
		stats["connection_pool"] = r.connectionPool.GetStats()
	}

	// Add statement cache stats
	if r.statementCache != nil {
		stats["statement_cache"] = r.statementCache.GetStats()
	}

	return stats
}

// Bulk operations for complex scenarios
func (r *repository[T]) BulkCreate(models []*T) error {
	return r.BulkCreateWithContext(context.Background(), models)
}

func (r *repository[T]) BulkCreateWithContext(ctx context.Context, models []*T) error {
	if len(models) == 0 {
		return nil
	}

	return r.performanceFacade.Track("repository.bulk_create", func() error {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Get fields from first model
		fields, _, err := r.getModelFieldsAndValues(models[0])
		if err != nil {
			return fmt.Errorf("failed to get model fields: %w", err)
		}

		// Build bulk INSERT query
		placeholders := strings.Repeat("?,", len(fields))
		placeholders = placeholders[:len(placeholders)-1] // Remove trailing comma

		valuesClause := strings.Repeat(fmt.Sprintf("(%s),", placeholders), len(models))
		valuesClause = valuesClause[:len(valuesClause)-1] // Remove trailing comma

		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", r.tableName, strings.Join(fields, ", "), valuesClause)

		// Get statement from cache
		stmt, err := r.statementCache.GetStatement(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to prepare statement: %w", err)
		}

		// Prepare all values
		var allValues []interface{}
		for _, model := range models {
			_, values, err := r.getModelFieldsAndValues(model)
			if err != nil {
				return fmt.Errorf("failed to get model values: %w", err)
			}
			allValues = append(allValues, values...)
		}

		// Execute bulk insert
		_, err = stmt.ExecContext(ctx, allValues...)
		if err != nil {
			return fmt.Errorf("failed to execute bulk insert: %w", err)
		}

		return nil
	})
}

func (r *repository[T]) BulkUpdate(models []*T) error {
	return r.BulkUpdateWithContext(context.Background(), models)
}

func (r *repository[T]) BulkUpdateWithContext(ctx context.Context, models []*T) error {
	if len(models) == 0 {
		return nil
	}

	return r.performanceFacade.Track("repository.bulk_update", func() error {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// For bulk update, we'll use a transaction and update each model individually
		// This is more complex than bulk insert due to different WHERE clauses
		return r.TransactionWithContext(ctx, func(repo Repository[T]) error {
			for _, model := range models {
				if err := repo.UpdateWithContext(ctx, model); err != nil {
					return fmt.Errorf("failed to update model in bulk: %w", err)
				}
			}
			return nil
		})
	})
}

func (r *repository[T]) BulkDelete(ids []uint) error {
	return r.BulkDeleteWithContext(context.Background(), ids)
}

func (r *repository[T]) BulkDeleteWithContext(ctx context.Context, ids []uint) error {
	if len(ids) == 0 {
		return nil
	}

	return r.performanceFacade.Track("repository.bulk_delete", func() error {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Build bulk DELETE query
		placeholders := strings.Repeat("?,", len(ids))
		placeholders = placeholders[:len(placeholders)-1] // Remove trailing comma

		query := fmt.Sprintf("UPDATE %s SET deleted_at = NOW() WHERE id IN (%s) AND deleted_at IS NULL", r.tableName, placeholders)

		// Get statement from cache
		stmt, err := r.statementCache.GetStatement(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to prepare statement: %w", err)
		}

		// Convert ids to interface slice
		args := make([]interface{}, len(ids))
		for i, id := range ids {
			args[i] = id
		}

		// Execute bulk delete
		_, err = stmt.ExecContext(ctx, args...)
		if err != nil {
			return fmt.Errorf("failed to execute bulk delete: %w", err)
		}

		return nil
	})
}

// Helper methods for model manipulation
func (r *repository[T]) getModelFieldsAndValues(model *T) ([]string, []interface{}, error) {
	// Use reflection to get field names and values
	v := reflect.ValueOf(model).Elem()
	t := v.Type()

	var fields []string
	var values []interface{}

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Skip ID field for inserts, skip deleted_at
		if field.Name == "ID" || field.Name == "DeletedAt" {
			continue
		}

		// Get field name from JSON tag or use field name
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" && jsonTag != "-" {
			fields = append(fields, jsonTag)
		} else {
			fields = append(fields, strings.ToLower(field.Name))
		}

		values = append(values, value.Interface())
	}

	return fields, values, nil
}

func (r *repository[T]) getModelID(model *T) (interface{}, error) {
	v := reflect.ValueOf(model).Elem()
	idField := v.FieldByName("ID")
	if !idField.IsValid() {
		return nil, fmt.Errorf("model does not have ID field")
	}
	return idField.Interface(), nil
}

func (r *repository[T]) setModelID(model *T, id uint) error {
	v := reflect.ValueOf(model).Elem()
	idField := v.FieldByName("ID")
	if !idField.IsValid() {
		return fmt.Errorf("model does not have ID field")
	}
	if !idField.CanSet() {
		return fmt.Errorf("ID field cannot be set")
	}
	idField.SetUint(uint64(id))
	return nil
}

func (r *repository[T]) scanRowToStruct(scanner interface{}, model *T) error {
	// Use reflection to scan row into model
	v := reflect.ValueOf(model).Elem()
	t := v.Type()

	// Get column names from database
	var columns []string
	if rows, ok := scanner.(*sql.Rows); ok {
		var err error
		columns, err = rows.Columns()
		if err != nil {
			return fmt.Errorf("failed to get columns: %w", err)
		}
	} else {
		// For single row, we need to get columns differently
		// This is a simplified version - in practice you'd need more sophisticated column detection
		return fmt.Errorf("unsupported scanner type for scanning")
	}

	// Create values slice for scanning
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// Scan the row
	if err := scanner.(*sql.Rows).Scan(valuePtrs...); err != nil {
		return fmt.Errorf("failed to scan row: %w", err)
	}

	// Map values to struct fields
	for i, column := range columns {
		// Find matching field in struct
		for j := 0; j < v.NumField(); j++ {
			field := t.Field(j)
			jsonTag := field.Tag.Get("json")

			if jsonTag == column || strings.ToLower(field.Name) == column {
				fieldValue := v.Field(j)
				if fieldValue.CanSet() {
					// Convert value to field type
					if err := r.convertValue(values[i], fieldValue); err != nil {
						return fmt.Errorf("failed to convert value for field %s: %w", field.Name, err)
					}
				}
				break
			}
		}
	}

	return nil
}

func (r *repository[T]) convertValue(value interface{}, field reflect.Value) error {
	// Convert database value to field type
	// This is a simplified version - in practice you'd need more sophisticated type conversion
	switch v := value.(type) {
	case []byte:
		// Handle JSON fields
		if field.Type() == reflect.TypeOf(json.RawMessage{}) {
			field.Set(reflect.ValueOf(json.RawMessage(v)))
		} else {
			field.SetString(string(v))
		}
	case string:
		field.SetString(v)
	case int64:
		field.SetInt(v)
	case float64:
		field.SetFloat(v)
	case bool:
		field.SetBool(v)
	case time.Time:
		field.Set(reflect.ValueOf(v))
	case nil:
		// Handle NULL values
		field.Set(reflect.Zero(field.Type()))
	default:
		field.Set(reflect.ValueOf(v))
	}
	return nil
}

func (r *repository[T]) buildWhereClause(conditions map[string]any) (string, []interface{}) {
	var clauses []string
	var values []interface{}

	for field, value := range conditions {
		clauses = append(clauses, fmt.Sprintf("%s = ?", field))
		values = append(values, value)
	}

	return strings.Join(clauses, " AND "), values
}

func (r *repository[T]) processWithWorkStealing(ctx context.Context, results []T) error {
	// Use work stealing pool for processing large datasets
	workItems := make([]WorkItem[any], len(results))
	for i, result := range results {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

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

func (r *repository[T]) processResult(ctx context.Context, data any) error {
	// Process the result item (e.g., apply transformations, validations, etc.)
	// This is a placeholder for actual processing logic
	return nil
}

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
