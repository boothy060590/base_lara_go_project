package facades

import "base_lara_go_project/app/core"

// DB facade provides Laravel-style database operations
type DB struct{}

// Table starts a new query for a table
func (db *DB) Table(tableName string) core.DatabaseInterface {
	return core.DatabaseInstance.Table(tableName)
}

// Model starts a new query for a model
func (db *DB) Model(value interface{}) core.DatabaseInterface {
	return core.DatabaseInstance.Model(value)
}

// Where adds a where clause to the query
func (db *DB) Where(query interface{}, args ...interface{}) core.DatabaseInterface {
	return core.DatabaseInstance.Where(query, args...)
}

// First retrieves the first record
func (db *DB) First(dest interface{}, conds ...interface{}) error {
	return core.DatabaseInstance.First(dest, conds...)
}

// Find retrieves all records
func (db *DB) Find(dest interface{}, conds ...interface{}) error {
	return core.DatabaseInstance.Find(dest, conds...)
}

// Create creates a new record
func (db *DB) Create(value interface{}) error {
	return core.DatabaseInstance.Create(value)
}

// Save saves a record
func (db *DB) Save(value interface{}) error {
	return core.DatabaseInstance.Save(value)
}

// Delete deletes a record
func (db *DB) Delete(value interface{}, conds ...interface{}) error {
	return core.DatabaseInstance.Delete(value, conds...)
}

// Transaction executes a function within a database transaction
func (db *DB) Transaction(fc func(tx core.DatabaseInterface) error) error {
	return core.DatabaseInstance.Transaction(fc)
}

// Raw executes a raw SQL query
func (db *DB) Raw(sql string, values ...interface{}) core.DatabaseInterface {
	return core.DatabaseInstance.Raw(sql, values...)
}

// Exec executes a raw SQL statement
func (db *DB) Exec(sql string, values ...interface{}) error {
	return core.DatabaseInstance.Exec(sql, values...)
}

// Preload preloads associations
func (db *DB) Preload(query string, args ...interface{}) core.DatabaseInterface {
	return core.DatabaseInstance.Preload(query, args...)
}

// Order adds an order clause
func (db *DB) Order(value interface{}) core.DatabaseInterface {
	return core.DatabaseInstance.Order(value)
}

// Limit adds a limit clause
func (db *DB) Limit(limit int) core.DatabaseInterface {
	return core.DatabaseInstance.Limit(limit)
}

// Offset adds an offset clause
func (db *DB) Offset(offset int) core.DatabaseInterface {
	return core.DatabaseInstance.Offset(offset)
}

// Global DB instance
var Database = &DB{}
