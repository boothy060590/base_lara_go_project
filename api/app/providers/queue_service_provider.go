package providers

import (
	"fmt"

	app_core "base_lara_go_project/app/core/app"
	queue_core "base_lara_go_project/app/core/queue"
	"base_lara_go_project/config"
)

// QueueServiceProvider registers queue services with the container
type QueueServiceProvider struct{}

// NewQueueServiceProvider creates a new queue service provider
func NewQueueServiceProvider() *QueueServiceProvider {
	return &QueueServiceProvider{}
}

// Register registers queue services with the container
func (p *QueueServiceProvider) Register() error {
	// Create queue provider factory
	factory := queue_core.NewQueueProviderFactory(app_core.App)

	// Get queue configuration
	queueConfig := config.QueueConfig()

	// Register queue provider from config
	if err := factory.RegisterFromConfig(queueConfig); err != nil {
		return fmt.Errorf("failed to register queue provider: %w", err)
	}

	// Get the registered queue provider
	queueProvider, err := app_core.App.Resolve("queue.provider")
	if err != nil {
		return fmt.Errorf("queue provider not found in container: %w", err)
	}

	// Create queue service adapter
	queueService := queue_core.NewQueueServiceAdapter(queueProvider.(app_core.QueueProviderServiceInterface))
	app_core.App.Singleton("queue.service", queueService)

	return nil
}

// RegisterQueue registers the queue service provider
func RegisterQueue() error {
	provider := NewQueueServiceProvider()
	return provider.Register()
}
