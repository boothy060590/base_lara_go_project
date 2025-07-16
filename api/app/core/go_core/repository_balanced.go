package go_core

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// BalancedPathExecutor provides moderate optimization for medium complexity queries
type BalancedPathExecutor[T any] struct {
	db             *sql.DB
	tableName      string
	statementCache *StatementCache
	connectionPool *ConnectionPool
}

// NewBalancedPathExecutor creates a new balanced path executor
func NewBalancedPathExecutor[T any](db *sql.DB, tableName string) *BalancedPathExecutor[T] {
	// Create lightweight statement cache (not the heavy one from original)
	config := map[string]any{
		"cache_size":   100,   // Smaller cache
		"cache_ttl":    300,   // 5 minutes
		"enable_stats": false, // No stats overhead
	}

	return &BalancedPathExecutor[T]{
		db:             db,
		tableName:      tableName,
		statementCache: NewStatementCache(config),
		connectionPool: NewConnectionPool(config),
	}
}

// smartQuery implements SmartQuery[T] with balanced optimization
type smartQuery[T any] struct {
	repository *smartRepository[T]
	conditions map[string]any
	orderBy    []string
	limit      int
	offset     int
	ctx        context.Context
}

// Where adds a where clause
func (sq *smartQuery[T]) Where(field string, operator string, value any) SmartQuery[T] {
	// Auto-escalate to complex if query becomes too complex
	if len(sq.conditions) >= 3 || len(sq.orderBy) >= 2 {
		// For complex queries, we need to return an error or handle differently
		// since ComplexQuery doesn't have Where method
		return &errorSmartQuery[T]{err: fmt.Errorf("query too complex, use Complex() builder directly")}
	}

	// Validate field name
	if err := sq.repository.fieldValidator.ValidateField(field); err != nil {
		return &errorSmartQuery[T]{err: fmt.Errorf("invalid field name '%s': %w", field, err)}
	}

	// Add condition with operator support
	conditionKey := fmt.Sprintf("%s_%s", field, operator)
	sq.conditions[conditionKey] = value
	return sq
}

// WhereIn adds a where in clause
func (sq *smartQuery[T]) WhereIn(field string, values []any) SmartQuery[T] {
	// Auto-escalate to complex for IN queries with many values
	if len(values) > 10 {
		return &errorSmartQuery[T]{err: fmt.Errorf("query too complex, use Complex() builder directly")}
	}

	// Validate field name
	if err := sq.repository.fieldValidator.ValidateField(field); err != nil {
		return &errorSmartQuery[T]{err: fmt.Errorf("invalid field name '%s': %w", field, err)}
	}

	sq.conditions[field+"_in"] = values
	return sq
}

// OrderBy adds an order clause
func (sq *smartQuery[T]) OrderBy(field string, direction string) SmartQuery[T] {
	// Auto-escalate to complex for multiple order by clauses
	if len(sq.orderBy) >= 2 {
		return &errorSmartQuery[T]{err: fmt.Errorf("query too complex, use Complex() builder directly")}
	}

	// Validate field name
	if err := sq.repository.fieldValidator.ValidateField(field); err != nil {
		return &errorSmartQuery[T]{err: fmt.Errorf("invalid field name '%s': %w", field, err)}
	}

	// Validate direction
	if direction != "asc" && direction != "desc" {
		return &errorSmartQuery[T]{err: fmt.Errorf("invalid order direction '%s'", direction)}
	}

	sq.orderBy = append(sq.orderBy, fmt.Sprintf("%s %s", field, direction))
	return sq
}

// Limit adds a limit clause
func (sq *smartQuery[T]) Limit(limit int) SmartQuery[T] {
	if limit < 0 {
		return &errorSmartQuery[T]{err: fmt.Errorf("limit must be non-negative")}
	}
	sq.limit = limit
	return sq
}

// Offset adds an offset clause
func (sq *smartQuery[T]) Offset(offset int) SmartQuery[T] {
	if offset < 0 {
		return &errorSmartQuery[T]{err: fmt.Errorf("offset must be non-negative")}
	}
	sq.offset = offset
	return sq
}

// Get retrieves all matching models
func (sq *smartQuery[T]) Get() ([]T, error) {
	return sq.repository.balancedPath.ExecuteQuery(sq.ctx, sq.conditions, sq.orderBy, sq.limit, sq.offset)
}

// First retrieves the first matching model
func (sq *smartQuery[T]) First() (*T, error) {
	// Set limit to 1 for efficiency
	sq.limit = 1

	results, err := sq.Get()
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no records found")
	}

	return &results[0], nil
}

// Paginate retrieves models with pagination
func (sq *smartQuery[T]) Paginate(page, perPage int) ([]T, int64, error) {
	// Validate pagination parameters
	if page <= 0 {
		return nil, 0, fmt.Errorf("page must be greater than 0")
	}
	if perPage <= 0 {
		return nil, 0, fmt.Errorf("perPage must be greater than 0")
	}

	// Calculate offset
	offset := (page - 1) * perPage

	// Set limit and offset
	sq.limit = perPage
	sq.offset = offset

	// Get results
	results, err := sq.Get()
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	total, err := sq.repository.balancedPath.ExecuteCount(sq.ctx, sq.conditions)
	if err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

// AsComplex escalates to complex query optimization
func (sq *smartQuery[T]) AsComplex() ComplexQuery[T] {
	// Create a complex query builder and apply existing conditions
	builder := sq.repository.Complex()

	// Since ComplexQuery doesn't have Where/WhereIn/OrderBy methods,
	// we need to build the query directly
	return builder.Build()
}

// WithContext returns a query with context
func (sq *smartQuery[T]) WithContext(ctx context.Context) SmartQuery[T] {
	sq.ctx = ctx
	return sq
}

// ExecuteQuery executes a query with balanced optimization
func (bp *BalancedPathExecutor[T]) ExecuteQuery(ctx context.Context, conditions map[string]any, orderBy []string, limit, offset int) ([]T, error) {
	// Build WHERE clause
	whereClause, values := bp.buildWhereClause(conditions)

	// Build query - handle empty conditions properly
	var query string
	if len(conditions) == 0 {
		query = fmt.Sprintf("SELECT * FROM %s WHERE deleted_at IS NULL", bp.tableName)
	} else {
		query = fmt.Sprintf("SELECT * FROM %s WHERE %s AND deleted_at IS NULL", bp.tableName, whereClause)
	}

	// Add ORDER BY
	if len(orderBy) > 0 {
		query += " ORDER BY " + strings.Join(orderBy, ", ")
	}

	// Add LIMIT and OFFSET
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
		if offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", offset)
		}
	}

	fmt.Printf("[DEBUG] Preparing query: %s\n", query)
	fmt.Printf("[DEBUG] Query args: %v\n", values)

	// Get statement from cache
	stmt, err := bp.statementCache.GetStatement(ctx, query)
	if err != nil || stmt == nil {
		// Fall back to direct preparation
		stmt, err = bp.db.PrepareContext(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare statement: %w", err)
		}
		defer stmt.Close()
	}

	fmt.Printf("[DEBUG] Executing query: %s with args: %v\n", query, values)

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
		if err := bp.scanRowToStruct(rows, &entity); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		result = append(result, entity)
	}

	return result, nil
}

// ExecuteCount executes a count query with balanced optimization
func (bp *BalancedPathExecutor[T]) ExecuteCount(ctx context.Context, conditions map[string]any) (int64, error) {
	// Build WHERE clause
	whereClause, values := bp.buildWhereClause(conditions)

	// Build count query - handle empty conditions properly
	var query string
	if len(conditions) == 0 {
		query = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE deleted_at IS NULL", bp.tableName)
	} else {
		query = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s AND deleted_at IS NULL", bp.tableName, whereClause)
	}

	// Get statement from cache
	stmt, err := bp.statementCache.GetStatement(ctx, query)
	if err != nil || stmt == nil {
		// Fall back to direct preparation
		stmt, err = bp.db.PrepareContext(ctx, query)
		if err != nil {
			return 0, fmt.Errorf("failed to prepare statement: %w", err)
		}
		defer stmt.Close()
	}

	// Execute query
	var count int64
	row := stmt.QueryRowContext(ctx, values...)
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to scan count: %w", err)
	}

	return count, nil
}

// buildWhereClause builds WHERE clause from conditions
func (bp *BalancedPathExecutor[T]) buildWhereClause(conditions map[string]any) (string, []any) {
	if len(conditions) == 0 {
		return "", nil
	}

	var clauses []string
	var values []any

	for conditionKey, value := range conditions {
		// Parse condition key (field_operator)
		parts := strings.Split(conditionKey, "_")
		if len(parts) < 2 {
			// Simple equality
			clauses = append(clauses, fmt.Sprintf("%s = ?", conditionKey))
			values = append(values, value)
			continue
		}

		field := parts[0]
		operator := parts[1]

		switch operator {
		case "in":
			// Handle IN operator
			if valueSlice, ok := value.([]any); ok {
				placeholders := strings.Repeat("?,", len(valueSlice))
				placeholders = placeholders[:len(placeholders)-1] // Remove trailing comma
				clauses = append(clauses, fmt.Sprintf("%s IN (%s)", field, placeholders))
				values = append(values, valueSlice...)
			}
		case "gt":
			clauses = append(clauses, fmt.Sprintf("%s > ?", field))
			values = append(values, value)
		case "gte":
			clauses = append(clauses, fmt.Sprintf("%s >= ?", field))
			values = append(values, value)
		case "lt":
			clauses = append(clauses, fmt.Sprintf("%s < ?", field))
			values = append(values, value)
		case "lte":
			clauses = append(clauses, fmt.Sprintf("%s <= ?", field))
			values = append(values, value)
		case "like":
			clauses = append(clauses, fmt.Sprintf("%s LIKE ?", field))
			values = append(values, value)
		case "ne":
			clauses = append(clauses, fmt.Sprintf("%s != ?", field))
			values = append(values, value)
		default:
			// Default to equality
			clauses = append(clauses, fmt.Sprintf("%s = ?", conditionKey))
			values = append(values, value)
		}
	}

	return strings.Join(clauses, " AND "), values
}

// scanRowToStruct scans a row into a struct
func (bp *BalancedPathExecutor[T]) scanRowToStruct(rows *sql.Rows, model *T) error {
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns: %w", err)
	}

	v := reflect.ValueOf(model).Elem()
	t := v.Type()

	// Create a map of field names to field indices
	fieldMap := make(map[string]int)
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		// Use db tag if available, otherwise use lowercase field name
		dbTag := field.Tag.Get("db")
		if dbTag != "" && dbTag != "-" {
			fieldMap[dbTag] = i
		} else {
			fieldMap[strings.ToLower(field.Name)] = i
		}
	}

	// Create scan destinations based on column order
	scanDest := make([]any, len(columns))
	for i, column := range columns {
		if fieldIndex, exists := fieldMap[column]; exists {
			field := v.Field(fieldIndex)
			scanDest[i] = bp.createScanDestination(field)
		} else {
			// For columns that don't have a corresponding field, use interface{}
			scanDest[i] = new(any)
		}
	}

	// Scan the row
	if err := rows.Scan(scanDest...); err != nil {
		return fmt.Errorf("failed to scan row: %w", err)
	}

	// Set values to struct fields
	for i, column := range columns {
		if fieldIndex, exists := fieldMap[column]; exists {
			field := v.Field(fieldIndex)
			if field.CanSet() {
				if err := bp.setFieldValue(field, scanDest[i]); err != nil {
					return fmt.Errorf("failed to set field %s: %w", t.Field(fieldIndex).Name, err)
				}
			}
		}
	}

	return nil
}

// createScanDestination creates appropriate scan destination for a field
func (bp *BalancedPathExecutor[T]) createScanDestination(field reflect.Value) any {
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

// setFieldValue sets a field value from scan destination
func (bp *BalancedPathExecutor[T]) setFieldValue(field reflect.Value, scanDest any) error {
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

// errorSmartQuery implements SmartQuery[T] for error cases
type errorSmartQuery[T any] struct {
	err error
}

// All methods return the error
func (esq *errorSmartQuery[T]) Where(field string, operator string, value any) SmartQuery[T] {
	return esq
}
func (esq *errorSmartQuery[T]) WhereIn(field string, values []any) SmartQuery[T]     { return esq }
func (esq *errorSmartQuery[T]) OrderBy(field string, direction string) SmartQuery[T] { return esq }
func (esq *errorSmartQuery[T]) Limit(limit int) SmartQuery[T]                        { return esq }
func (esq *errorSmartQuery[T]) Offset(offset int) SmartQuery[T]                      { return esq }
func (esq *errorSmartQuery[T]) Get() ([]T, error)                                    { return nil, esq.err }
func (esq *errorSmartQuery[T]) First() (*T, error)                                   { return nil, esq.err }
func (esq *errorSmartQuery[T]) Paginate(page, perPage int) ([]T, int64, error) {
	return nil, 0, esq.err
}
func (esq *errorSmartQuery[T]) AsComplex() ComplexQuery[T] {
	return &errorComplexQuery[T]{err: esq.err}
}
func (esq *errorSmartQuery[T]) WithContext(ctx context.Context) SmartQuery[T] { return esq }
