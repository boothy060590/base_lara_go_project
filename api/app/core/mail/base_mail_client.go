package mail_core

import (
	app_core "base_lara_go_project/app/core/app"
	client_core "base_lara_go_project/app/core/clients"
)

// BaseMailClient provides common functionality for all mail clients
type BaseMailClient struct {
	*client_core.BaseClient
	config   *app_core.ClientConfig
	from     string
	fromName string
}

// NewBaseMailClient creates a new base mail client
func NewBaseMailClient(config *app_core.ClientConfig, name string) *BaseMailClient {
	from := "no-reply@example.com"
	if configFrom, ok := config.Options["from"].(string); ok && configFrom != "" {
		from = configFrom
	}

	fromName := "Base Laravel Go Project"
	if configFromName, ok := config.Options["from_name"].(string); ok && configFromName != "" {
		fromName = configFromName
	}

	return &BaseMailClient{
		BaseClient: client_core.NewBaseClient(config, name),
		config:     config,
		from:       from,
		fromName:   fromName,
	}
}

// GetFrom returns the from email address
func (c *BaseMailClient) GetFrom() string {
	return c.from
}

// GetFromName returns the from name
func (c *BaseMailClient) GetFromName() string {
	return c.fromName
}

// GetConfig returns the mail configuration
func (c *BaseMailClient) GetConfig() *app_core.ClientConfig {
	return c.config
}
