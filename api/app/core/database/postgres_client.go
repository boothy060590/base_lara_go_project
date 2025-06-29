package database_core

import (
	"fmt"

	app_core "base_lara_go_project/app/core/app"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// PostgresClient provides PostgreSQL database functionality
type PostgresClient struct {
	*BaseDatabaseClient
	db *gorm.DB
}

// NewPostgresClient creates a new PostgreSQL client
func NewPostgresClient(config *app_core.ClientConfig) *PostgresClient {
	return &PostgresClient{
		BaseDatabaseClient: NewBaseDatabaseClient(config, "postgres"),
	}
}

// Connect establishes a connection to PostgreSQL
func (c *PostgresClient) Connect() error {
	// Build connection string
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.GetHost(),
		c.GetPort(),
		c.GetUsername(),
		c.GetPassword(),
		c.GetDatabaseName(),
	)

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %v", err)
	}

	c.db = db
	return c.BaseClient.Connect()
}

// Disconnect closes the PostgreSQL connection
func (c *PostgresClient) Disconnect() error {
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
func (c *PostgresClient) Query(query string, args ...interface{}) ([]map[string]interface{}, error) {
	if c.db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	var results []map[string]interface{}
	err := c.db.Raw(query, args...).Scan(&results).Error
	return results, err
}

// Execute executes a query without returning results
func (c *PostgresClient) Execute(query string, args ...interface{}) (int64, error) {
	if c.db == nil {
		return 0, fmt.Errorf("database not connected")
	}

	result := c.db.Exec(query, args...)
	return result.RowsAffected, result.Error
}

// BeginTransaction begins a new transaction
func (c *PostgresClient) BeginTransaction() (app_core.TransactionInterface, error) {
	if c.db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	tx := c.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &PostgresTransaction{tx: tx}, nil
}

// Ping checks if the database is reachable
func (c *PostgresClient) Ping() error {
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
func (c *PostgresClient) GetStats() map[string]interface{} {
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

// PostgresTransaction implements TransactionInterface for PostgreSQL
type PostgresTransaction struct {
	tx *gorm.DB
}

// Commit commits the transaction
func (t *PostgresTransaction) Commit() error {
	return t.tx.Commit().Error
}

// Rollback rolls back the transaction
func (t *PostgresTransaction) Rollback() error {
	return t.tx.Rollback().Error
}

// Query executes a query within the transaction
func (t *PostgresTransaction) Query(query string, args ...interface{}) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := t.tx.Raw(query, args...).Scan(&results).Error
	return results, err
}

// Execute executes a query within the transaction
func (t *PostgresTransaction) Execute(query string, args ...interface{}) (int64, error) {
	result := t.tx.Exec(query, args...)
	return result.RowsAffected, result.Error
}
