package database_core

import (
	app_core "base_lara_go_project/app/core/app"
	clients_core "base_lara_go_project/app/core/clients"
)

// Re-export interface from app_core for convenience
type DatabaseProviderServiceInterface = app_core.DatabaseProviderServiceInterface

// MySQLDatabaseProvider provides a database provider for MySQL
type MySQLDatabaseProvider struct {
	client clients_core.DatabaseClientInterface
}

func NewMySQLDatabaseProvider(client clients_core.DatabaseClientInterface) *MySQLDatabaseProvider {
	return &MySQLDatabaseProvider{client: client}
}

func (p *MySQLDatabaseProvider) Query(query string, args ...interface{}) ([]map[string]interface{}, error) {
	return p.client.Query(query, args...)
}

func (p *MySQLDatabaseProvider) Execute(query string, args ...interface{}) (int64, error) {
	return p.client.Execute(query, args...)
}

func (p *MySQLDatabaseProvider) BeginTransaction() (clients_core.TransactionInterface, error) {
	return p.client.BeginTransaction()
}

func (p *MySQLDatabaseProvider) Ping() error {
	return p.client.Ping()
}

func (p *MySQLDatabaseProvider) GetStats() map[string]interface{} {
	return p.client.GetStats()
}
