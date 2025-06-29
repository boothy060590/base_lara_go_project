package database_core

import (
	app_core "base_lara_go_project/app/core/app"
	client_core "base_lara_go_project/app/core/clients"
)

// BaseDatabaseClient provides common functionality for all database clients
type BaseDatabaseClient struct {
	*client_core.BaseClient
	config *app_core.ClientConfig
}

// NewBaseDatabaseClient creates a new base database client
func NewBaseDatabaseClient(config *app_core.ClientConfig, name string) *BaseDatabaseClient {
	return &BaseDatabaseClient{
		BaseClient: client_core.NewBaseClient(config, name),
		config:     config,
	}
}

// GetDatabaseConfig returns the database-specific configuration
func (c *BaseDatabaseClient) GetDatabaseConfig() *app_core.ClientConfig {
	return c.config
}

// GetDatabaseName returns the database name
func (c *BaseDatabaseClient) GetDatabaseName() string {
	return c.config.Database
}

// GetHost returns the database host
func (c *BaseDatabaseClient) GetHost() string {
	return c.config.Host
}

// GetPort returns the database port
func (c *BaseDatabaseClient) GetPort() int {
	return c.config.Port
}

// GetUsername returns the database username
func (c *BaseDatabaseClient) GetUsername() string {
	return c.config.Username
}

// GetPassword returns the database password
func (c *BaseDatabaseClient) GetPassword() string {
	return c.config.Password
}

// Connect implements ClientInterface
func (c *BaseDatabaseClient) Connect() error {
	// Base implementation - override in specific clients
	return c.BaseClient.Connect()
}

// Disconnect implements ClientInterface
func (c *BaseDatabaseClient) Disconnect() error {
	// Base implementation - override in specific clients
	return c.BaseClient.Disconnect()
}

// IsConnected implements ClientInterface
func (c *BaseDatabaseClient) IsConnected() bool {
	return c.BaseClient.IsConnected()
}

// GetConfig returns the client configuration
func (c *BaseDatabaseClient) GetConfig() *app_core.ClientConfig {
	return c.config
}

// GetName implements ClientInterface
func (c *BaseDatabaseClient) GetName() string {
	return c.BaseClient.GetName()
}
