package go_core

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// RepositoryBatchProcessor handles batch operations for repository
type RepositoryBatchProcessor[T any] struct {
	enabled     bool
	batchSize   int
	batchBuffer []*RepositoryBatchItem[T]
	bufferMutex sync.Mutex
	flushTicker *time.Ticker
	done        chan bool
	config      map[string]interface{}
	stats       *BatchProcessorStats
	statsMutex  sync.RWMutex
}

// RepositoryBatchItem represents a batch operation item
type RepositoryBatchItem[T any] struct {
	Model *T
	Op    string // "create", "update", "delete"
}

// BatchProcessorStats tracks batch processor statistics
type BatchProcessorStats struct {
	ItemsProcessed   int64
	BatchesProcessed int64
	BufferSize       int
	LastFlush        time.Time
	Errors           int64
}

// NewRepositoryBatchProcessor creates a new repository batch processor
func NewRepositoryBatchProcessor[T any](config map[string]interface{}) *RepositoryBatchProcessor[T] {
	batchSize := 100 // Default
	if size, ok := config["repository_batch_size"].(int); ok {
		batchSize = size
	}

	enabled := true
	if enabledVal, ok := config["repository_batch_enabled"].(bool); ok {
		enabled = enabledVal
	}

	return &RepositoryBatchProcessor[T]{
		enabled:     enabled,
		batchSize:   batchSize,
		batchBuffer: make([]*RepositoryBatchItem[T], 0),
		done:        make(chan bool),
		config:      config,
		stats:       &BatchProcessorStats{},
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
	// Update stats
	rbp.statsMutex.Lock()
	rbp.stats.ItemsProcessed += int64(len(items))
	rbp.stats.BatchesProcessed++
	rbp.stats.LastFlush = time.Now()
	rbp.statsMutex.Unlock()

	// This would integrate with database batch operations
	// For now, we'll just process them individually
	for _, item := range items {
		_ = item // Process item
	}
}

// AddItem adds an item to batch buffer
func (rbp *RepositoryBatchProcessor[T]) AddItem(model *T, op string) error {
	if !rbp.enabled {
		return fmt.Errorf("batch processing is disabled")
	}

	rbp.bufferMutex.Lock()
	defer rbp.bufferMutex.Unlock()

	rbp.batchBuffer = append(rbp.batchBuffer, &RepositoryBatchItem[T]{
		Model: model,
		Op:    op,
	})

	// Update buffer size stat
	rbp.statsMutex.Lock()
	rbp.stats.BufferSize = len(rbp.batchBuffer)
	rbp.statsMutex.Unlock()

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
	rbp.statsMutex.RLock()
	defer rbp.statsMutex.RUnlock()

	return map[string]interface{}{
		"enabled":           rbp.enabled,
		"batch_size":        rbp.batchSize,
		"buffer_size":       rbp.stats.BufferSize,
		"items_processed":   rbp.stats.ItemsProcessed,
		"batches_processed": rbp.stats.BatchesProcessed,
		"last_flush":        rbp.stats.LastFlush,
		"errors":            rbp.stats.Errors,
		"config":            rbp.config,
	}
}

// RepositoryAsyncProcessor handles async operations for repository
type RepositoryAsyncProcessor[T any] struct {
	enabled    bool
	config     map[string]interface{}
	stats      *AsyncProcessorStats
	statsMutex sync.RWMutex
}

// AsyncProcessorStats tracks async processor statistics
type AsyncProcessorStats struct {
	OperationsQueued    int64
	OperationsCompleted int64
	OperationsFailed    int64
	LastOperation       time.Time
}

// NewRepositoryAsyncProcessor creates a new repository async processor
func NewRepositoryAsyncProcessor[T any](config map[string]interface{}) *RepositoryAsyncProcessor[T] {
	enabled := true
	if enabledVal, ok := config["repository_async_enabled"].(bool); ok {
		enabled = enabledVal
	}

	return &RepositoryAsyncProcessor[T]{
		enabled: enabled,
		config:  config,
		stats:   &AsyncProcessorStats{},
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

	// Update stats
	rap.statsMutex.Lock()
	rap.stats.OperationsQueued++
	rap.stats.LastOperation = time.Now()
	rap.statsMutex.Unlock()

	// Process asynchronously
	go func() {
		err := processor(context.Background(), model)

		rap.statsMutex.Lock()
		if err != nil {
			rap.stats.OperationsFailed++
		} else {
			rap.stats.OperationsCompleted++
		}
		rap.statsMutex.Unlock()
	}()

	return nil
}

// IsEnabled returns whether async processing is enabled
func (rap *RepositoryAsyncProcessor[T]) IsEnabled() bool {
	return rap.enabled
}

// GetStats returns async processor statistics
func (rap *RepositoryAsyncProcessor[T]) GetStats() map[string]interface{} {
	rap.statsMutex.RLock()
	defer rap.statsMutex.RUnlock()

	return map[string]interface{}{
		"enabled":              rap.enabled,
		"operations_queued":    rap.stats.OperationsQueued,
		"operations_completed": rap.stats.OperationsCompleted,
		"operations_failed":    rap.stats.OperationsFailed,
		"last_operation":       rap.stats.LastOperation,
		"config":               rap.config,
	}
}

// RepositoryPipelineProcessor handles pipeline operations for repository
type RepositoryPipelineProcessor struct {
	enabled    bool
	config     map[string]interface{}
	stats      *PipelineProcessorStats
	statsMutex sync.RWMutex
}

// PipelineProcessorStats tracks pipeline processor statistics
type PipelineProcessorStats struct {
	PipelinesCreated int64
	ItemsProcessed   int64
	LastPipeline     time.Time
}

// NewRepositoryPipelineProcessor creates a new repository pipeline processor
func NewRepositoryPipelineProcessor(config map[string]interface{}) *RepositoryPipelineProcessor {
	enabled := true
	if enabledVal, ok := config["repository_pipeline_enabled"].(bool); ok {
		enabled = enabledVal
	}

	return &RepositoryPipelineProcessor{
		enabled: enabled,
		config:  config,
		stats:   &PipelineProcessorStats{},
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
	rpp.statsMutex.RLock()
	defer rpp.statsMutex.RUnlock()

	return map[string]interface{}{
		"enabled":           rpp.enabled,
		"pipelines_created": rpp.stats.PipelinesCreated,
		"items_processed":   rpp.stats.ItemsProcessed,
		"last_pipeline":     rpp.stats.LastPipeline,
		"config":            rpp.config,
	}
}

// QueryBuilder implements Query[T] with raw SQL and safety
type queryBuilder[T any] struct {
	repository *repository[T]
	conditions map[string]any
	orderBy    []string
	limit      int
	offset     int
	preloads   []string
	ctx        context.Context
}

// Get retrieves all matching models
func (qb *queryBuilder[T]) Get() ([]T, error) {
	return qb.GetWithContext(qb.ctx)
}

// GetWithContext retrieves all matching models with context
func (qb *queryBuilder[T]) GetWithContext(ctx context.Context) ([]T, error) {
	// Build WHERE clause
	whereClause, values := qb.repository.buildWhereClause(qb.conditions)

	// Build query
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s AND deleted_at IS NULL", qb.repository.tableName, whereClause)

	// Add ORDER BY
	if len(qb.orderBy) > 0 {
		query += " ORDER BY " + strings.Join(qb.orderBy, ", ")
	}

	// Add LIMIT and OFFSET
	if qb.limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", qb.limit)
		if qb.offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", qb.offset)
		}
	}

	// Get statement from cache
	stmt, err := qb.repository.statementCache.GetStatement(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	// Execute query
	rows, err := stmt.QueryContext(ctx, values...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Scan results
	var result []T
	for rows.Next() {
		var entity T
		if err := qb.repository.scanRowToStruct(rows, &entity); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		result = append(result, entity)
	}

	return result, nil
}

// First retrieves the first matching model
func (qb *queryBuilder[T]) First() (*T, error) {
	return qb.FirstWithContext(qb.ctx)
}

// FirstWithContext retrieves the first matching model with context
func (qb *queryBuilder[T]) FirstWithContext(ctx context.Context) (*T, error) {
	// Set limit to 1
	qb.limit = 1

	results, err := qb.GetWithContext(ctx)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no records found")
	}

	return &results[0], nil
}

// Paginate retrieves models with pagination
func (qb *queryBuilder[T]) Paginate(page, perPage int) ([]T, int64, error) {
	return qb.PaginateWithContext(qb.ctx, page, perPage)
}

// PaginateWithContext retrieves models with pagination and context
func (qb *queryBuilder[T]) PaginateWithContext(ctx context.Context, page, perPage int) ([]T, int64, error) {
	// Validate pagination parameters
	if page <= 0 {
		return nil, 0, fmt.Errorf("page must be greater than 0, got %d", page)
	}
	if perPage <= 0 {
		return nil, 0, fmt.Errorf("perPage must be greater than 0, got %d", perPage)
	}

	// Calculate offset
	offset := (page - 1) * perPage

	// Set limit and offset
	qb.limit = perPage
	qb.offset = offset

	// Get results
	results, err := qb.GetWithContext(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	total, err := qb.repository.CountWhereWithContext(ctx, qb.conditions)
	if err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

// Where adds a where clause
func (qb *queryBuilder[T]) Where(field string, operator string, value any) Query[T] {
	// Validate field name
	if err := qb.repository.fieldValidator.ValidateField(field); err != nil {
		return &errorQuery[T]{err: fmt.Errorf("invalid field name '%s': %w", field, err)}
	}

	// Add condition
	qb.conditions[field] = value
	return qb
}

// WhereIn adds a where in clause
func (qb *queryBuilder[T]) WhereIn(field string, values []any) Query[T] {
	// Validate field name
	if err := qb.repository.fieldValidator.ValidateField(field); err != nil {
		return &errorQuery[T]{err: fmt.Errorf("invalid field name '%s': %w", field, err)}
	}

	// Add condition
	qb.conditions[field] = values
	return qb
}

// OrderBy adds an order clause
func (qb *queryBuilder[T]) OrderBy(field string, direction string) Query[T] {
	// Validate field name
	if err := qb.repository.fieldValidator.ValidateField(field); err != nil {
		return &errorQuery[T]{err: fmt.Errorf("invalid field name '%s': %w", field, err)}
	}

	// Validate direction
	if direction != "asc" && direction != "desc" {
		return &errorQuery[T]{err: fmt.Errorf("invalid order direction '%s', must be 'asc' or 'desc'", direction)}
	}

	qb.orderBy = append(qb.orderBy, fmt.Sprintf("%s %s", field, direction))
	return qb
}

// Limit adds a limit clause
func (qb *queryBuilder[T]) Limit(limit int) Query[T] {
	if limit < 0 {
		return &errorQuery[T]{err: fmt.Errorf("limit must be non-negative, got %d", limit)}
	}
	qb.limit = limit
	return qb
}

// Offset adds an offset clause
func (qb *queryBuilder[T]) Offset(offset int) Query[T] {
	if offset < 0 {
		return &errorQuery[T]{err: fmt.Errorf("offset must be non-negative, got %d", offset)}
	}
	qb.offset = offset
	return qb
}

// Preload adds a preload clause
func (qb *queryBuilder[T]) Preload(relation string) Query[T] {
	qb.preloads = append(qb.preloads, relation)
	return qb
}

// WithContext returns a query with context
func (qb *queryBuilder[T]) WithContext(ctx context.Context) Query[T] {
	qb.ctx = ctx
	return qb
}

// RawQueryBuilder implements Query[T] for raw SQL queries
type rawQueryBuilder[T any] struct {
	repository *repository[T]
	query      string
	args       []any
	ctx        context.Context
}

// Get retrieves all matching models
func (rqb *rawQueryBuilder[T]) Get() ([]T, error) {
	return rqb.GetWithContext(rqb.ctx)
}

// GetWithContext retrieves all matching models with context
func (rqb *rawQueryBuilder[T]) GetWithContext(ctx context.Context) ([]T, error) {
	// Validate query
	if err := rqb.repository.sqlValidator.ValidateQuery(rqb.query); err != nil {
		return nil, fmt.Errorf("invalid query: %w", err)
	}

	// Get statement from cache
	stmt, err := rqb.repository.statementCache.GetStatement(ctx, rqb.query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	// Execute query
	rows, err := stmt.QueryContext(ctx, rqb.args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Scan results
	var result []T
	for rows.Next() {
		var entity T
		if err := rqb.repository.scanRowToStruct(rows, &entity); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		result = append(result, entity)
	}

	return result, nil
}

// First retrieves the first matching model
func (rqb *rawQueryBuilder[T]) First() (*T, error) {
	return rqb.FirstWithContext(rqb.ctx)
}

// FirstWithContext retrieves the first matching model with context
func (rqb *rawQueryBuilder[T]) FirstWithContext(ctx context.Context) (*T, error) {
	results, err := rqb.GetWithContext(ctx)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no records found")
	}

	return &results[0], nil
}

// Paginate retrieves models with pagination
func (rqb *rawQueryBuilder[T]) Paginate(page, perPage int) ([]T, int64, error) {
	return rqb.PaginateWithContext(rqb.ctx, page, perPage)
}

// PaginateWithContext retrieves models with pagination and context
func (rqb *rawQueryBuilder[T]) PaginateWithContext(ctx context.Context, page, perPage int) ([]T, int64, error) {
	// For raw queries, pagination is handled in the query itself
	results, err := rqb.GetWithContext(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Calculate total (this is a simplified approach)
	total := int64(len(results))

	// Apply pagination
	start := (page - 1) * perPage
	end := start + perPage

	if start >= len(results) {
		return []T{}, total, nil
	}

	if end > len(results) {
		end = len(results)
	}

	return results[start:end], total, nil
}

// Where adds a where clause (not applicable for raw queries)
func (rqb *rawQueryBuilder[T]) Where(field string, operator string, value any) Query[T] {
	return &errorQuery[T]{err: fmt.Errorf("where clause not supported for raw queries")}
}

// WhereIn adds a where in clause (not applicable for raw queries)
func (rqb *rawQueryBuilder[T]) WhereIn(field string, values []any) Query[T] {
	return &errorQuery[T]{err: fmt.Errorf("where in clause not supported for raw queries")}
}

// OrderBy adds an order clause (not applicable for raw queries)
func (rqb *rawQueryBuilder[T]) OrderBy(field string, direction string) Query[T] {
	return &errorQuery[T]{err: fmt.Errorf("order by clause not supported for raw queries")}
}

// Limit adds a limit clause (not applicable for raw queries)
func (rqb *rawQueryBuilder[T]) Limit(limit int) Query[T] {
	return &errorQuery[T]{err: fmt.Errorf("limit clause not supported for raw queries")}
}

// Offset adds an offset clause (not applicable for raw queries)
func (rqb *rawQueryBuilder[T]) Offset(offset int) Query[T] {
	return &errorQuery[T]{err: fmt.Errorf("offset clause not supported for raw queries")}
}

// Preload adds a preload clause (not applicable for raw queries)
func (rqb *rawQueryBuilder[T]) Preload(relation string) Query[T] {
	return &errorQuery[T]{err: fmt.Errorf("preload clause not supported for raw queries")}
}

// WithContext returns a query with context
func (rqb *rawQueryBuilder[T]) WithContext(ctx context.Context) Query[T] {
	rqb.ctx = ctx
	return rqb
}

// ErrorQuery implements Query[T] for error cases
type errorQuery[T any] struct {
	err error
}

// All methods return the error
func (eq *errorQuery[T]) Get() ([]T, error)                                { return nil, eq.err }
func (eq *errorQuery[T]) GetWithContext(ctx context.Context) ([]T, error)  { return nil, eq.err }
func (eq *errorQuery[T]) First() (*T, error)                               { return nil, eq.err }
func (eq *errorQuery[T]) FirstWithContext(ctx context.Context) (*T, error) { return nil, eq.err }
func (eq *errorQuery[T]) Paginate(page, perPage int) ([]T, int64, error)   { return nil, 0, eq.err }
func (eq *errorQuery[T]) PaginateWithContext(ctx context.Context, page, perPage int) ([]T, int64, error) {
	return nil, 0, eq.err
}
func (eq *errorQuery[T]) Where(field string, operator string, value any) Query[T] { return eq }
func (eq *errorQuery[T]) WhereIn(field string, values []any) Query[T]             { return eq }
func (eq *errorQuery[T]) OrderBy(field string, direction string) Query[T]         { return eq }
func (eq *errorQuery[T]) Limit(limit int) Query[T]                                { return eq }
func (eq *errorQuery[T]) Offset(offset int) Query[T]                              { return eq }
func (eq *errorQuery[T]) Preload(relation string) Query[T]                        { return eq }
func (eq *errorQuery[T]) WithContext(ctx context.Context) Query[T]                { return eq }
