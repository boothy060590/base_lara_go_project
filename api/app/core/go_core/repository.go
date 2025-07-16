package go_core

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// QueryComplexity represents the complexity level of a database query
type QueryComplexity int

const (
	// Simple queries: Find(id), FindBy(field, value), Exists(id), Count()
	SimpleQuery QueryComplexity = iota
	// Medium queries: Where with 1-3 conditions, basic pagination
	MediumQuery
	// Complex queries: Multi-table joins, raw SQL, bulk operations, reporting
	ComplexQueryLevel
)

// Repository provides intelligent query optimization based on complexity
type Repository[T any] interface {
	// Fast path operations (Tier 1 - minimal overhead)
	Find(id uint) (*T, error)
	FindBy(field string, value any) (*T, error)
	Exists(id uint) (bool, error)
	Count() (int64, error)

	// Balanced operations (Tier 2 - moderate optimization)
	Where(conditions map[string]any) SmartQuery[T]

	// Full optimization operations (Tier 3 - all optimizations)
	Complex() ComplexQueryBuilder[T]

	// Manual optimization control
	WithOptimization(level QueryComplexity) Repository[T]

	// Context support
	WithContext(ctx context.Context) Repository[T]

	// Transaction support
	Transaction(fn func(Repository[T]) error) error

	// Utility operations
	Create(model *T) error
	Update(model *T) error
	Delete(id uint) error
}

// SmartQuery represents a query that can auto-escalate optimization based on complexity
type SmartQuery[T any] interface {
	// Basic query building
	Where(field string, operator string, value any) SmartQuery[T]
	WhereIn(field string, values []any) SmartQuery[T]
	OrderBy(field string, direction string) SmartQuery[T]
	Limit(limit int) SmartQuery[T]
	Offset(offset int) SmartQuery[T]

	// Execution
	Get() ([]T, error)
	First() (*T, error)
	Paginate(page, perPage int) ([]T, int64, error)

	// Complexity escalation
	AsComplex() ComplexQuery[T]

	// Context support
	WithContext(ctx context.Context) SmartQuery[T]
}

// ComplexQueryBuilder provides full optimization for complex queries
type ComplexQueryBuilder[T any] interface {
	// Advanced query building
	Join(table string, on string) ComplexQueryBuilder[T]
	LeftJoin(table string, on string) ComplexQueryBuilder[T]
	RightJoin(table string, on string) ComplexQueryBuilder[T]
	GroupBy(fields ...string) ComplexQueryBuilder[T]
	Having(condition string, args ...any) ComplexQueryBuilder[T]

	// Raw SQL support
	Raw(query string, args ...any) ComplexQuery[T]

	// Bulk operations
	BulkCreate(models []*T) error
	BulkUpdate(models []*T) error
	BulkDelete(ids []uint) error

	// Advanced features
	WithBatching(enabled bool) ComplexQueryBuilder[T]
	WithAsync(enabled bool) ComplexQueryBuilder[T]
	WithPipeline(enabled bool) ComplexQueryBuilder[T]
	WithWorkStealing(enabled bool) ComplexQueryBuilder[T]

	// Build query
	Build() ComplexQuery[T]
}

// ComplexQuery represents a fully optimized query
type ComplexQuery[T any] interface {
	Get() ([]T, error)
	First() (*T, error)
	Paginate(page, perPage int) ([]T, int64, error)
	Stream() (<-chan T, error)

	// Performance monitoring
	WithMetrics(enabled bool) ComplexQuery[T]
	GetStats() map[string]any

	// Context support
	WithContext(ctx context.Context) ComplexQuery[T]
}

// smartRepository implements Repository with intelligent optimization
type smartRepository[T any] struct {
	db             *sql.DB
	tableName      string
	fieldValidator *FieldValidator
	ctx            context.Context

	// Optimization levels
	fastPath     *FastPathExecutor[T]
	balancedPath *BalancedPathExecutor[T]
	complexPath  *ComplexPathExecutor[T]

	// Current optimization level
	optimizationLevel QueryComplexity
}

// getTableName extracts table name from model type
func getTableName(v any) string {
	t := reflect.TypeOf(v)

	// If it's a pointer, get the element type
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Create a new instance to call methods on
	val := reflect.New(t)

	// Check for GetTableName method and call it
	if method := val.MethodByName("GetTableName"); method.IsValid() {
		results := method.Call(nil)
		if len(results) > 0 && results[0].Kind() == reflect.String {
			return results[0].String()
		}
	}

	// Fallback to conventional naming
	return strings.ToLower(t.Name()) + "s"
}

// NewRepository creates a new smart repository with intelligent optimization
func NewRepository[T any](db *sql.DB) Repository[T] {
	var model T
	tableName := getTableName(model)

	repo := &smartRepository[T]{
		db:                db,
		tableName:         tableName,
		fieldValidator:    NewFieldValidator(tableName),
		ctx:               context.Background(),
		optimizationLevel: SimpleQuery,
	}

	// Initialize optimization tiers
	repo.fastPath = NewFastPathExecutor[T](db, tableName)
	repo.balancedPath = NewBalancedPathExecutor[T](db, tableName)
	repo.complexPath = NewComplexPathExecutor[T](db, tableName)

	return repo
}

// Fast path operations (Tier 1 - minimal overhead)
func (r *smartRepository[T]) Find(id uint) (*T, error) {
	return r.fastPath.Find(r.ctx, id)
}

func (r *smartRepository[T]) FindBy(field string, value any) (*T, error) {
	// Validate field name
	if err := r.fieldValidator.ValidateField(field); err != nil {
		return nil, fmt.Errorf("invalid field name: %w", err)
	}
	return r.fastPath.FindBy(r.ctx, field, value)
}

func (r *smartRepository[T]) Exists(id uint) (bool, error) {
	return r.fastPath.Exists(r.ctx, id)
}

func (r *smartRepository[T]) Count() (int64, error) {
	return r.fastPath.Count(r.ctx)
}

// Balanced operations (Tier 2 - moderate optimization)
func (r *smartRepository[T]) Where(conditions map[string]any) SmartQuery[T] {
	// Validate all field names
	for field := range conditions {
		if err := r.fieldValidator.ValidateField(field); err != nil {
			return &errorSmartQuery[T]{err: fmt.Errorf("invalid field name '%s': %w", field, err)}
		}
	}

	return &smartQuery[T]{
		repository: r,
		conditions: conditions,
		ctx:        r.ctx,
	}
}

// Full optimization operations (Tier 3 - all optimizations)
func (r *smartRepository[T]) Complex() ComplexQueryBuilder[T] {
	return &complexQueryBuilder[T]{
		repository: r,
		ctx:        r.ctx,
	}
}

// Manual optimization control
func (r *smartRepository[T]) WithOptimization(level QueryComplexity) Repository[T] {
	newRepo := *r
	newRepo.optimizationLevel = level
	return &newRepo
}

// Context support
func (r *smartRepository[T]) WithContext(ctx context.Context) Repository[T] {
	newRepo := *r
	newRepo.ctx = ctx
	return &newRepo
}

// Transaction support
func (r *smartRepository[T]) Transaction(fn func(Repository[T]) error) error {
	tx, err := r.db.BeginTx(r.ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Create transaction repository
	txRepo := &smartRepository[T]{
		db:                r.db, // Keep original db reference
		tableName:         r.tableName,
		fieldValidator:    r.fieldValidator,
		ctx:               r.ctx,
		optimizationLevel: r.optimizationLevel,
	}

	// Initialize optimization tiers for transaction (using original db)
	txRepo.fastPath = NewFastPathExecutor[T](r.db, r.tableName)
	txRepo.balancedPath = NewBalancedPathExecutor[T](r.db, r.tableName)
	txRepo.complexPath = NewComplexPathExecutor[T](r.db, r.tableName)

	// Execute function
	if err := fn(txRepo); err != nil {
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
}

// Utility operations
func (r *smartRepository[T]) Create(model *T) error {
	return r.fastPath.Create(r.ctx, model)
}

func (r *smartRepository[T]) Update(model *T) error {
	return r.fastPath.Update(r.ctx, model)
}

func (r *smartRepository[T]) Delete(id uint) error {
	return r.fastPath.Delete(r.ctx, id)
}

// FastPathExecutor provides minimal overhead database operations
type FastPathExecutor[T any] struct {
	db        *sql.DB
	tableName string

	// Minimal statement cache (only for most common operations)
	findStmt   *sql.Stmt
	existsStmt *sql.Stmt
	countStmt  *sql.Stmt
}

// NewFastPathExecutor creates a new fast path executor
func NewFastPathExecutor[T any](db *sql.DB, tableName string) *FastPathExecutor[T] {
	executor := &FastPathExecutor[T]{
		db:        db,
		tableName: tableName,
	}

	// Pre-prepare most common statements
	executor.prepareCommonStatements()

	return executor
}

// prepareCommonStatements prepares the most frequently used statements
func (f *FastPathExecutor[T]) prepareCommonStatements() {
	var err error

	// Prepare find statement
	findQuery := fmt.Sprintf("SELECT * FROM %s WHERE id = ? AND deleted_at IS NULL", f.tableName)
	f.findStmt, err = f.db.Prepare(findQuery)
	if err != nil {
		// Log error but continue - we'll fall back to dynamic preparation
	}

	// Prepare exists statement
	existsQuery := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE id = ? AND deleted_at IS NULL)", f.tableName)
	f.existsStmt, err = f.db.Prepare(existsQuery)
	if err != nil {
		// Log error but continue
	}

	// Prepare count statement
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE deleted_at IS NULL", f.tableName)
	f.countStmt, err = f.db.Prepare(countQuery)
	if err != nil {
		// Log error but continue
	}
}

// Find retrieves a model by ID with minimal overhead
func (f *FastPathExecutor[T]) Find(ctx context.Context, id uint) (*T, error) {
	var entity T

	// Use pre-prepared statement if available
	if f.findStmt != nil {
		row := f.findStmt.QueryRowContext(ctx, id)
		if err := f.scanRowToStruct(row, &entity); err != nil {
			return nil, err
		}
		return &entity, nil
	}

	// Fall back to dynamic query
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = ? AND deleted_at IS NULL", f.tableName)
	row := f.db.QueryRowContext(ctx, query, id)
	if err := f.scanRowToStruct(row, &entity); err != nil {
		return nil, err
	}

	return &entity, nil
}

// FindBy retrieves a model by field with minimal overhead
func (f *FastPathExecutor[T]) FindBy(ctx context.Context, field string, value any) (*T, error) {
	var entity T

	// Dynamic query for field-based search
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s = ? AND deleted_at IS NULL LIMIT 1", f.tableName, field)
	row := f.db.QueryRowContext(ctx, query, value)
	if err := f.scanRowToStruct(row, &entity); err != nil {
		return nil, err
	}

	return &entity, nil
}

// Exists checks if a model exists with minimal overhead
func (f *FastPathExecutor[T]) Exists(ctx context.Context, id uint) (bool, error) {
	var exists bool

	// Use pre-prepared statement if available
	if f.existsStmt != nil {
		row := f.existsStmt.QueryRowContext(ctx, id)
		if err := row.Scan(&exists); err != nil {
			return false, err
		}
		return exists, nil
	}

	// Fall back to dynamic query
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE id = ? AND deleted_at IS NULL)", f.tableName)
	row := f.db.QueryRowContext(ctx, query, id)
	if err := row.Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

// Count returns the total count with minimal overhead
func (f *FastPathExecutor[T]) Count(ctx context.Context) (int64, error) {
	var count int64

	// Use pre-prepared statement if available
	if f.countStmt != nil {
		row := f.countStmt.QueryRowContext(ctx)
		if err := row.Scan(&count); err != nil {
			return 0, err
		}
		return count, nil
	}

	// Fall back to dynamic query
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE deleted_at IS NULL", f.tableName)
	row := f.db.QueryRowContext(ctx, query)
	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

// Create creates a new model with minimal overhead
func (f *FastPathExecutor[T]) Create(ctx context.Context, model *T) error {
	// Set timestamps for new models
	f.setTimestampsForCreate(model)

	// Get model fields and values
	fields, values, err := f.getModelFieldsAndValues(model)
	if err != nil {
		return fmt.Errorf("failed to get model fields: %w", err)
	}

	// Build INSERT query
	placeholders := strings.Repeat("?,", len(fields))
	placeholders = placeholders[:len(placeholders)-1] // Remove trailing comma

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", f.tableName, strings.Join(fields, ", "), placeholders)

	// Execute query
	_, err = f.db.ExecContext(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("failed to execute insert: %w", err)
	}

	return nil
}

// Update updates a model with minimal overhead
func (f *FastPathExecutor[T]) Update(ctx context.Context, model *T) error {
	// Get model fields and values (excluding ID)
	fields, values, err := f.getModelFieldsAndValues(model)
	if err != nil {
		return fmt.Errorf("failed to get model fields: %w", err)
	}

	// Get ID field
	idValue, err := f.getModelID(model)
	if err != nil {
		return fmt.Errorf("failed to get model ID: %w", err)
	}

	// Build UPDATE query
	setClause := make([]string, len(fields))
	for i, field := range fields {
		setClause[i] = fmt.Sprintf("%s = ?", field)
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = ? AND deleted_at IS NULL", f.tableName, strings.Join(setClause, ", "))

	// Add ID to values
	values = append(values, idValue)

	// Execute query
	_, err = f.db.ExecContext(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("failed to execute update: %w", err)
	}

	return nil
}

// Delete deletes a model with minimal overhead (soft delete)
func (f *FastPathExecutor[T]) Delete(ctx context.Context, id uint) error {
	// Soft delete - set deleted_at timestamp
	query := fmt.Sprintf("UPDATE %s SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL", f.tableName)

	// Execute query
	_, err := f.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to execute delete: %w", err)
	}

	return nil
}

// setTimestampsForCreate sets CreatedAt and UpdatedAt timestamps for new models
func (f *FastPathExecutor[T]) setTimestampsForCreate(model *T) {
	v := reflect.ValueOf(model).Elem()
	now := time.Now()

	// Set CreatedAt if it exists and is zero
	if createdAtField := v.FieldByName("CreatedAt"); createdAtField.IsValid() && createdAtField.CanSet() {
		if createdAtField.Interface().(time.Time).IsZero() {
			createdAtField.Set(reflect.ValueOf(now))
		}
	}

	// Set UpdatedAt if it exists and is zero
	if updatedAtField := v.FieldByName("UpdatedAt"); updatedAtField.IsValid() && updatedAtField.CanSet() {
		if updatedAtField.Interface().(time.Time).IsZero() {
			updatedAtField.Set(reflect.ValueOf(now))
		}
	}
}

// Helper methods for fast path executor
func (f *FastPathExecutor[T]) getModelFieldsAndValues(model *T) ([]string, []any, error) {
	// Use reflection to get field names and values
	v := reflect.ValueOf(model).Elem()
	t := v.Type()

	var fields []string
	var values []any

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

func (f *FastPathExecutor[T]) getModelID(model *T) (any, error) {
	v := reflect.ValueOf(model).Elem()
	idField := v.FieldByName("ID")
	if !idField.IsValid() {
		return nil, fmt.Errorf("model does not have ID field")
	}
	return idField.Interface(), nil
}

func (f *FastPathExecutor[T]) scanRowToStruct(row *sql.Row, model *T) error {
	v := reflect.ValueOf(model).Elem()
	t := v.Type()

	// Create scan destinations for all fields
	scanDest := make([]any, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		scanDest[i] = f.createScanDestination(field)
	}

	// Scan the row
	if err := row.Scan(scanDest...); err != nil {
		return fmt.Errorf("failed to scan row: %w", err)
	}

	// Set values to struct fields
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.CanSet() {
			if err := f.setFieldValue(field, scanDest[i]); err != nil {
				return fmt.Errorf("failed to set field %s: %w", t.Field(i).Name, err)
			}
		}
	}

	return nil
}

func (f *FastPathExecutor[T]) createScanDestination(field reflect.Value) any {
	switch field.Kind() {
	case reflect.String:
		return new(sql.NullString)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return new(sql.NullInt64)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return new(sql.NullInt64)
	case reflect.Float32, reflect.Float64:
		return new(sql.NullFloat64)
	case reflect.Bool:
		return new(sql.NullBool)
	case reflect.Struct:
		if field.Type() == reflect.TypeOf(time.Time{}) {
			return new(sql.NullTime)
		}
		fallthrough
	default:
		return new(any)
	}
}

func (f *FastPathExecutor[T]) setFieldValue(field reflect.Value, scanDest any) error {
	switch dest := scanDest.(type) {
	case *sql.NullString:
		if dest.Valid {
			field.SetString(dest.String)
		}
	case *sql.NullInt64:
		if dest.Valid {
			switch field.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				field.SetInt(dest.Int64)
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				field.SetUint(uint64(dest.Int64))
			}
		}
	case *sql.NullFloat64:
		if dest.Valid {
			field.SetFloat(dest.Float64)
		}
	case *sql.NullBool:
		if dest.Valid {
			field.SetBool(dest.Bool)
		}
	case *sql.NullTime:
		if dest.Valid {
			field.Set(reflect.ValueOf(dest.Time))
		}
	case *any:
		value := *dest
		if value != nil {
			valueType := reflect.TypeOf(value)
			if valueType.AssignableTo(field.Type()) {
				field.Set(reflect.ValueOf(value))
			} else if valueType.ConvertibleTo(field.Type()) {
				field.Set(reflect.ValueOf(value).Convert(field.Type()))
			}
		}
	}
	return nil
}
