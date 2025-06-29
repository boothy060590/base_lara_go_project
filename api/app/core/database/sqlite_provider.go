package database_core

import (
	app_core "base_lara_go_project/app/core/app"
)

// SQLiteDatabaseProvider provides SQLite database services
type SQLiteDatabaseProvider struct {
	client *SQLiteClient
}

// NewSQLiteDatabaseProvider creates a new SQLite database provider
func NewSQLiteDatabaseProvider(client *SQLiteClient) *SQLiteDatabaseProvider {
	return &SQLiteDatabaseProvider{
		client: client,
	}
}

// Connect establishes a connection to the database
func (p *SQLiteDatabaseProvider) Connect() error {
	return p.client.Connect()
}

// Disconnect closes the database connection
func (p *SQLiteDatabaseProvider) Disconnect() error {
	return p.client.Disconnect()
}

// Query executes a query and returns results
func (p *SQLiteDatabaseProvider) Query(query string, args ...interface{}) ([]map[string]interface{}, error) {
	return p.client.Query(query, args...)
}

// Execute executes a query without returning results
func (p *SQLiteDatabaseProvider) Execute(query string, args ...interface{}) (int64, error) {
	return p.client.Execute(query, args...)
}

// BeginTransaction begins a new transaction
func (p *SQLiteDatabaseProvider) BeginTransaction() (app_core.TransactionInterface, error) {
	return p.client.BeginTransaction()
}

// Ping checks if the database is reachable
func (p *SQLiteDatabaseProvider) Ping() error {
	return p.client.Ping()
}

// GetStats returns database statistics
func (p *SQLiteDatabaseProvider) GetStats() map[string]interface{} {
	return p.client.GetStats()
}

// GetClient returns the underlying database client
func (p *SQLiteDatabaseProvider) GetClient() app_core.DatabaseClientInterface {
	return p.client
}
