package client_core

import (
	app_core "base_lara_go_project/app/core/app"
)

// Re-export interfaces and types from app_core for convenience
type ClientInterface = app_core.ClientInterface
type HTTPClientInterface = app_core.HTTPClientInterface
type LoggingClientInterface = app_core.LoggingClientInterface
type DatabaseClientInterface = app_core.DatabaseClientInterface
type TransactionInterface = app_core.TransactionInterface
type CacheClientInterface = app_core.CacheClientInterface
type QueueClientInterface = app_core.QueueClientInterface
type FileSystemClientInterface = app_core.FileSystemClientInterface
type MailClientInterface = app_core.MailClientInterface
type ClientConfig = app_core.ClientConfig

// BaseClient provides common functionality for all clients
type BaseClient struct {
	config    *ClientConfig
	connected bool
	name      string
}

// NewBaseClient creates a new base client
func NewBaseClient(config *ClientConfig, name string) *BaseClient {
	return &BaseClient{
		config: config,
		name:   name,
	}
}

// Connect implements ClientInterface
func (c *BaseClient) Connect() error {
	c.connected = true
	return nil
}

// Disconnect implements ClientInterface
func (c *BaseClient) Disconnect() error {
	c.connected = false
	return nil
}

// IsConnected implements ClientInterface
func (c *BaseClient) IsConnected() bool {
	return c.connected
}

// GetConfig implements ClientInterface
func (c *BaseClient) GetConfig() *ClientConfig {
	return c.config
}

// GetName implements ClientInterface
func (c *BaseClient) GetName() string {
	return c.name
}
