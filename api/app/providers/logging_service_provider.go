package providers

import (
	"fmt"

	app_core "base_lara_go_project/app/core/app"
	logging_core "base_lara_go_project/app/core/logging"
	"base_lara_go_project/config"
)

// LoggingServiceProvider registers logging services with the container
type LoggingServiceProvider struct{}

// NewLoggingServiceProvider creates a new logging service provider
func NewLoggingServiceProvider() *LoggingServiceProvider {
	return &LoggingServiceProvider{}
}

// Register registers logging services with the container
func (p *LoggingServiceProvider) Register() error {
	// Create logging provider factory
	factory := logging_core.NewLoggingProviderFactory(app_core.App)

	// Get logging configuration
	loggingConfig := config.LoggingConfig()

	// Register logging provider from config
	if err := factory.RegisterFromConfig(loggingConfig); err != nil {
		return fmt.Errorf("failed to register logging provider: %w", err)
	}

	return nil
}

// RegisterLogging registers the logging service provider
func RegisterLogging() error {
	provider := NewLoggingServiceProvider()
	return provider.Register()
}
