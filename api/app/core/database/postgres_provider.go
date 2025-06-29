package database_core

import (
	app_core "base_lara_go_project/app/core/app"
)

// PostgresDatabaseProvider provides PostgreSQL database services
type PostgresDatabaseProvider struct {
	client *PostgresClient
}

// NewPostgresDatabaseProvider creates a new PostgreSQL database provider
func NewPostgresDatabaseProvider(client *PostgresClient) *PostgresDatabaseProvider {
	return &PostgresDatabaseProvider{
		client: client,
	}
}

// Connect establishes a connection to the database
func (p *PostgresDatabaseProvider) Connect() error {
	return p.client.Connect()
}

// Disconnect closes the database connection
func (p *PostgresDatabaseProvider) Disconnect() error {
	return p.client.Disconnect()
}

// Query executes a query and returns results
func (p *PostgresDatabaseProvider) Query(query string, args ...interface{}) ([]map[string]interface{}, error) {
	return p.client.Query(query, args...)
}

// Execute executes a query without returning results
func (p *PostgresDatabaseProvider) Execute(query string, args ...interface{}) (int64, error) {
	return p.client.Execute(query, args...)
}

// BeginTransaction begins a new transaction
func (p *PostgresDatabaseProvider) BeginTransaction() (app_core.TransactionInterface, error) {
	return p.client.BeginTransaction()
}

// Ping checks if the database is reachable
func (p *PostgresDatabaseProvider) Ping() error {
	return p.client.Ping()
}

// GetStats returns database statistics
func (p *PostgresDatabaseProvider) GetStats() map[string]interface{} {
	return p.client.GetStats()
}

// GetClient returns the underlying database client
func (p *PostgresDatabaseProvider) GetClient() app_core.DatabaseClientInterface {
	return p.client
}
