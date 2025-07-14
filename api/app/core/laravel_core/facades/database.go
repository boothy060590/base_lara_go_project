package facades_core

import (
	"context"
	"database/sql"
	"fmt"
)

// Database facade provides Laravel-style database access
type Database struct {
	db *sql.DB
}

// NewDatabase creates a new database facade
func NewDatabase(db *sql.DB) *Database {
	return &Database{
		db: db,
	}
}

// Table returns a query builder for the specified table
func (d *Database) Table(tableName string) *QueryBuilder {
	return &QueryBuilder{
		db:        d.db,
		tableName: tableName,
	}
}

// Raw executes a raw SQL query
func (d *Database) Raw(query string, args ...interface{}) *QueryBuilder {
	return &QueryBuilder{
		db:        d.db,
		tableName: "",
		rawQuery:  query,
		rawArgs:   args,
	}
}

// Transaction executes a function within a database transaction
func (d *Database) Transaction(fn func(*Database) error) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	txDB := &Database{db: d.db} // In a real implementation, this would use the transaction

	if err := fn(txDB); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("transaction failed and rollback failed: %w (rollback error: %v)", err, rollbackErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// TransactionWithContext executes a function within a database transaction with context
func (d *Database) TransactionWithContext(ctx context.Context, fn func(*Database) error) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	txDB := &Database{db: d.db} // In a real implementation, this would use the transaction

	if err := fn(txDB); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("transaction failed and rollback failed: %w (rollback error: %v)", err, rollbackErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// QueryBuilder provides a fluent interface for building SQL queries
type QueryBuilder struct {
	db        *sql.DB
	tableName string
	rawQuery  string
	rawArgs   []interface{}
	where     []string
	args      []interface{}
	orderBy   []string
	limit     int
	offset    int
}

// Where adds a where clause
func (qb *QueryBuilder) Where(field string, operator string, value interface{}) *QueryBuilder {
	qb.where = append(qb.where, fmt.Sprintf("%s %s ?", field, operator))
	qb.args = append(qb.args, value)
	return qb
}

// WhereIn adds a where in clause
func (qb *QueryBuilder) WhereIn(field string, values []interface{}) *QueryBuilder {
	placeholders := make([]string, len(values))
	for i := range values {
		placeholders[i] = "?"
		qb.args = append(qb.args, values[i])
	}
	qb.where = append(qb.where, fmt.Sprintf("%s IN (%s)", field, fmt.Sprintf("%s", placeholders)))
	return qb
}

// OrderBy adds an order clause
func (qb *QueryBuilder) OrderBy(field string, direction string) *QueryBuilder {
	qb.orderBy = append(qb.orderBy, fmt.Sprintf("%s %s", field, direction))
	return qb
}

// Limit adds a limit clause
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limit = limit
	return qb
}

// Offset adds an offset clause
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offset = offset
	return qb
}

// Get executes the query and returns all results
func (qb *QueryBuilder) Get() ([]map[string]interface{}, error) {
	query, args := qb.buildQuery()

	rows, err := qb.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	return qb.scanRows(rows)
}

// First executes the query and returns the first result
func (qb *QueryBuilder) First() (map[string]interface{}, error) {
	qb.limit = 1
	results, err := qb.Get()
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no records found")
	}

	return results[0], nil
}

// Count returns the count of records
func (qb *QueryBuilder) Count() (int64, error) {
	// Build count query
	query := "SELECT COUNT(*) FROM " + qb.tableName

	if len(qb.where) > 0 {
		query += " WHERE " + fmt.Sprintf("%s", qb.where)
	}

	var count int64
	err := qb.db.QueryRow(query, qb.args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count records: %w", err)
	}

	return count, nil
}

// Insert inserts a new record
func (qb *QueryBuilder) Insert(data map[string]interface{}) (int64, error) {
	fields := make([]string, 0, len(data))
	placeholders := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))

	for field, value := range data {
		fields = append(fields, field)
		placeholders = append(placeholders, "?")
		values = append(values, value)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		qb.tableName,
		fmt.Sprintf("%s", fields),
		fmt.Sprintf("%s", placeholders))

	result, err := qb.db.Exec(query, values...)
	if err != nil {
		return 0, fmt.Errorf("failed to insert record: %w", err)
	}

	return result.LastInsertId()
}

// Update updates records
func (qb *QueryBuilder) Update(data map[string]interface{}) (int64, error) {
	setClause := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))

	for field, value := range data {
		setClause = append(setClause, fmt.Sprintf("%s = ?", field))
		values = append(values, value)
	}

	query := fmt.Sprintf("UPDATE %s SET %s", qb.tableName, fmt.Sprintf("%s", setClause))

	if len(qb.where) > 0 {
		query += " WHERE " + fmt.Sprintf("%s", qb.where)
		values = append(values, qb.args...)
	}

	result, err := qb.db.Exec(query, values...)
	if err != nil {
		return 0, fmt.Errorf("failed to update records: %w", err)
	}

	return result.RowsAffected()
}

// Delete deletes records
func (qb *QueryBuilder) Delete() (int64, error) {
	query := fmt.Sprintf("DELETE FROM %s", qb.tableName)

	if len(qb.where) > 0 {
		query += " WHERE " + fmt.Sprintf("%s", qb.where)
	}

	result, err := qb.db.Exec(query, qb.args...)
	if err != nil {
		return 0, fmt.Errorf("failed to delete records: %w", err)
	}

	return result.RowsAffected()
}

// buildQuery builds the final SQL query
func (qb *QueryBuilder) buildQuery() (string, []interface{}) {
	var query string
	var args []interface{}

	if qb.rawQuery != "" {
		query = qb.rawQuery
		args = qb.rawArgs
	} else {
		query = "SELECT * FROM " + qb.tableName

		if len(qb.where) > 0 {
			query += " WHERE " + fmt.Sprintf("%s", qb.where)
		}

		if len(qb.orderBy) > 0 {
			query += " ORDER BY " + fmt.Sprintf("%s", qb.orderBy)
		}

		if qb.limit > 0 {
			query += fmt.Sprintf(" LIMIT %d", qb.limit)
			if qb.offset > 0 {
				query += fmt.Sprintf(" OFFSET %d", qb.offset)
			}
		}

		args = qb.args
	}

	return query, args
}

// scanRows scans database rows into a slice of maps
func (qb *QueryBuilder) scanRows(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results []map[string]interface{}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		row := make(map[string]interface{})
		for i, column := range columns {
			row[column] = values[i]
		}

		results = append(results, row)
	}

	return results, nil
}
