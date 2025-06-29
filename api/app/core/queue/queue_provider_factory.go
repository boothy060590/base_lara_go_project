package queue_core

import (
	"fmt"

	app_core "base_lara_go_project/app/core/app"
	clients_core "base_lara_go_project/app/core/clients"
)

// QueueProviderFactory creates queue providers based on configuration
type QueueProviderFactory struct {
	container *app_core.ServiceContainer
}

// NewQueueProviderFactory creates a new queue provider factory
func NewQueueProviderFactory(container *app_core.ServiceContainer) *QueueProviderFactory {
	return &QueueProviderFactory{container: container}
}

var queueProviderMap = map[string]func(cfg *clients_core.ClientConfig) app_core.QueueProviderServiceInterface{
	"sync": func(cfg *clients_core.ClientConfig) app_core.QueueProviderServiceInterface {
		return NewSyncQueueProvider(NewSyncQueueClient(cfg))
	},
	"sqs": func(cfg *clients_core.ClientConfig) app_core.QueueProviderServiceInterface {
		return NewSQSQueueProvider(NewSQSQueueClient(cfg))
	},
}

// Create creates a queue provider based on the driver
func (f *QueueProviderFactory) Create(driver string, cfg *clients_core.ClientConfig) (app_core.QueueProviderServiceInterface, error) {
	constructor, ok := queueProviderMap[driver]
	if !ok {
		return nil, fmt.Errorf("unsupported queue driver: %s", driver)
	}
	return constructor(cfg), nil
}

// RegisterFromConfig registers a queue provider from configuration
func (f *QueueProviderFactory) RegisterFromConfig(config map[string]interface{}) error {
	// Get default connection from config
	defaultConnection, ok := config["default"].(string)
	if !ok {
		return fmt.Errorf("default queue connection not set in config")
	}

	// Get connections configuration
	connections, ok := config["connections"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("queue connections not configured")
	}

	// Get the specific connection config
	connectionConfig, ok := connections[defaultConnection].(map[string]interface{})
	if !ok {
		return fmt.Errorf("connection config for %s not found", defaultConnection)
	}

	// Build client config from connection config
	clientConfig := f.buildClientConfig(defaultConnection, connectionConfig)

	provider, err := f.Create(defaultConnection, clientConfig)
	if err != nil {
		return err
	}

	f.container.Singleton("queue.provider", provider)
	return nil
}

// buildClientConfig converts connection config to client config
func (f *QueueProviderFactory) buildClientConfig(driver string, config map[string]interface{}) *clients_core.ClientConfig {
	clientConfig := &clients_core.ClientConfig{
		Driver:  driver,
		Options: config,
	}

	// Set common fields
	if key, ok := config["key"].(string); ok {
		clientConfig.Username = key
	}
	if secret, ok := config["secret"].(string); ok {
		clientConfig.Password = secret
	}

	// Set driver-specific fields
	switch driver {
	case "sqs":
		if region, ok := config["region"].(string); ok {
			clientConfig.Options["region"] = region
		}
		if queue, ok := config["queue"].(string); ok {
			clientConfig.Options["queue"] = queue
		}
		if endpoint, ok := config["endpoint"].(string); ok {
			clientConfig.Options["endpoint"] = endpoint
		}
	case "sync":
		// Sync queue doesn't need additional configuration
	}

	return clientConfig
}
