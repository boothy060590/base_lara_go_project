package facades_core

import (
	app_core "base_lara_go_project/app/core/app"
	database_core "base_lara_go_project/app/core/database"
)

// DB provides a facade for database operations
type DB struct{}

// Table returns a database interface for the specified table
func (db *DB) Table(tableName string) app_core.DatabaseInterface {
	return database_core.DatabaseInstance.Table(tableName)
}

// Model returns a database interface for the specified model
func (db *DB) Model(value interface{}) app_core.DatabaseInterface {
	return database_core.DatabaseInstance.Model(value)
}

// Where adds a where clause to the query
func (db *DB) Where(query interface{}, args ...interface{}) app_core.DatabaseInterface {
	return database_core.DatabaseInstance.Where(query, args...)
}

// First finds the first record matching the conditions
func (db *DB) First(dest interface{}, conds ...interface{}) error {
	return database_core.DatabaseInstance.First(dest, conds...)
}

// Find finds all records matching the conditions
func (db *DB) Find(dest interface{}, conds ...interface{}) error {
	return database_core.DatabaseInstance.Find(dest, conds...)
}

// Create creates a new record
func (db *DB) Create(value interface{}) error {
	return database_core.DatabaseInstance.Create(value)
}

// Save saves the record
func (db *DB) Save(value interface{}) error {
	return database_core.DatabaseInstance.Save(value)
}

// Delete deletes the record
func (db *DB) Delete(value interface{}, conds ...interface{}) error {
	return database_core.DatabaseInstance.Delete(value, conds...)
}

// Transaction executes a function within a database transaction
func (db *DB) Transaction(fc func(tx app_core.DatabaseInterface) error) error {
	return database_core.DatabaseInstance.Transaction(fc)
}

// Raw executes a raw SQL query
func (db *DB) Raw(sql string, values ...interface{}) app_core.DatabaseInterface {
	return database_core.DatabaseInstance.Raw(sql, values...)
}

// Exec executes a SQL statement
func (db *DB) Exec(sql string, values ...interface{}) error {
	return database_core.DatabaseInstance.Exec(sql, values...)
}

// Preload preloads associations
func (db *DB) Preload(query string, args ...interface{}) app_core.DatabaseInterface {
	return database_core.DatabaseInstance.Preload(query, args...)
}

// Order adds an order clause to the query
func (db *DB) Order(value interface{}) app_core.DatabaseInterface {
	return database_core.DatabaseInstance.Order(value)
}

// Limit adds a limit clause to the query
func (db *DB) Limit(limit int) app_core.DatabaseInterface {
	return database_core.DatabaseInstance.Limit(limit)
}

// Offset adds an offset clause to the query
func (db *DB) Offset(offset int) app_core.DatabaseInterface {
	return database_core.DatabaseInstance.Offset(offset)
}

// GetDB returns the underlying database instance
func (db *DB) GetDB() interface{} {
	return database_core.DatabaseInstance.GetDB()
}

// Global DB instance
var Database = &DB{}
