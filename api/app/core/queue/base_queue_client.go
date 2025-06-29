package queue_core

import (
	app_core "base_lara_go_project/app/core/app"
	client_core "base_lara_go_project/app/core/clients"
)

// BaseQueueClient provides common functionality for all queue clients
type BaseQueueClient struct {
	*client_core.BaseClient
	config *app_core.ClientConfig
}

// NewBaseQueueClient creates a new base queue client
func NewBaseQueueClient(config *app_core.ClientConfig, name string) *BaseQueueClient {
	return &BaseQueueClient{
		BaseClient: client_core.NewBaseClient(config, name),
		config:     config,
	}
}

// GetConfig returns the queue configuration
func (c *BaseQueueClient) GetConfig() *app_core.ClientConfig {
	return c.config
}
