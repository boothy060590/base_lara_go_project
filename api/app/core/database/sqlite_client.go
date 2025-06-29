package database_core

import (
	"fmt"
	"os"
	"path/filepath"

	app_core "base_lara_go_project/app/core/app"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SQLiteClient provides SQLite database functionality
type SQLiteClient struct {
	*BaseDatabaseClient
	db *gorm.DB
}

// NewSQLiteClient creates a new SQLite client
func NewSQLiteClient(config *app_core.ClientConfig) *SQLiteClient {
	return &SQLiteClient{
		BaseDatabaseClient: NewBaseDatabaseClient(config, "sqlite"),
	}
}

// Connect establishes a connection to SQLite
func (c *SQLiteClient) Connect() error {
	// Get database path from config
	databasePath := c.GetDatabaseName()
	if databasePath == "" {
		databasePath = "database.sqlite"
	}

	// Ensure directory exists
	dir := filepath.Dir(databasePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %v", err)
	}

	// Connect to database
	db, err := gorm.Open(sqlite.Open(databasePath), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to SQLite: %v", err)
	}

	c.db = db
	return c.BaseClient.Connect()
}

// Disconnect closes the SQLite connection
func (c *SQLiteClient) Disconnect() error {
	if c.db != nil {
		sqlDB, err := c.db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return c.BaseClient.Disconnect()
}

// Query executes a query and returns results
func (c *SQLiteClient) Query(query string, args ...interface{}) ([]map[string]interface{}, error) {
	if c.db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	var results []map[string]interface{}
	err := c.db.Raw(query, args...).Scan(&results).Error
	return results, err
}

// Execute executes a query without returning results
func (c *SQLiteClient) Execute(query string, args ...interface{}) (int64, error) {
	if c.db == nil {
		return 0, fmt.Errorf("database not connected")
	}

	result := c.db.Exec(query, args...)
	return result.RowsAffected, result.Error
}

// BeginTransaction begins a new transaction
func (c *SQLiteClient) BeginTransaction() (app_core.TransactionInterface, error) {
	if c.db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	tx := c.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &SQLiteTransaction{tx: tx}, nil
}

// Ping checks if the database is reachable
func (c *SQLiteClient) Ping() error {
	if c.db == nil {
		return fmt.Errorf("database not connected")
	}

	sqlDB, err := c.db.DB()
	if err != nil {
		return err
	}

	return sqlDB.Ping()
}

// GetStats returns database statistics
func (c *SQLiteClient) GetStats() map[string]interface{} {
	if c.db == nil {
		return map[string]interface{}{"status": "disconnected"}
	}

	sqlDB, err := c.db.DB()
	if err != nil {
		return map[string]interface{}{"status": "error", "error": err.Error()}
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"status":              "connected",
		"max_open_conns":      stats.MaxOpenConnections,
		"open_conns":          stats.OpenConnections,
		"in_use":              stats.InUse,
		"idle":                stats.Idle,
		"wait_count":          stats.WaitCount,
		"wait_duration":       stats.WaitDuration.String(),
		"max_idle_closed":     stats.MaxIdleClosed,
		"max_lifetime_closed": stats.MaxLifetimeClosed,
	}
}

// SQLiteTransaction implements TransactionInterface for SQLite
type SQLiteTransaction struct {
	tx *gorm.DB
}

// Commit commits the transaction
func (t *SQLiteTransaction) Commit() error {
	return t.tx.Commit().Error
}

// Rollback rolls back the transaction
func (t *SQLiteTransaction) Rollback() error {
	return t.tx.Rollback().Error
}

// Query executes a query within the transaction
func (t *SQLiteTransaction) Query(query string, args ...interface{}) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := t.tx.Raw(query, args...).Scan(&results).Error
	return results, err
}

// Execute executes a query within the transaction
func (t *SQLiteTransaction) Execute(query string, args ...interface{}) (int64, error) {
	result := t.tx.Exec(query, args...)
	return result.RowsAffected, result.Error
}
