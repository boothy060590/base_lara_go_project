package client_core

// Re-export interfaces and types from app_core for convenience

// Define a minimal ClientInterface for client_core
// You can expand this as needed for your application

type ClientConfig struct {
	// Add fields as required
}

type ClientInterface interface {
	Connect() error
	Disconnect() error
	IsConnected() bool
	GetConfig() *ClientConfig
	GetName() string
}

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
