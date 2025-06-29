package logging_core

import (
	"fmt"
	"time"

	app_core "base_lara_go_project/app/core/app"
	clients_core "base_lara_go_project/app/core/clients"
)

// LoggingProviderFactory creates logging providers based on configuration
type LoggingProviderFactory struct {
	container *app_core.ServiceContainer
}

// NewLoggingProviderFactory creates a new logging provider factory
func NewLoggingProviderFactory(container *app_core.ServiceContainer) *LoggingProviderFactory {
	return &LoggingProviderFactory{
		container: container,
	}
}

// Create creates a logging provider based on configuration
func (f *LoggingProviderFactory) Create(config map[string]interface{}) (clients_core.LoggingClientInterface, error) {
	driver, ok := config["driver"].(string)
	if !ok {
		return nil, fmt.Errorf("driver not specified in logging configuration")
	}

	// Create client config
	clientConfig := f.buildClientConfig(config)

	// Create provider based on driver
	switch driver {
	case "local", "single":
		return f.createLocalProvider(clientConfig)
	case "daily":
		return f.createDailyProvider(clientConfig)
	case "sentry":
		return f.createSentryProvider(clientConfig)
	case "slack":
		return f.createSlackProvider(clientConfig)
	case "stderr":
		return f.createStderrProvider(clientConfig)
	case "emergency":
		return f.createEmergencyProvider(clientConfig)
	case "syslog":
		return f.createSyslogProvider(clientConfig)
	case "errorlog":
		return f.createErrorLogProvider(clientConfig)
	case "stack":
		return f.createStackProvider(clientConfig)
	case "null":
		return f.createNullProvider(clientConfig)
	default:
		return nil, fmt.Errorf("unknown logging driver: %s", driver)
	}
}

// buildClientConfig builds a client configuration from the config map
func (f *LoggingProviderFactory) buildClientConfig(config map[string]interface{}) *clients_core.ClientConfig {
	clientConfig := &clients_core.ClientConfig{
		Driver:  config["driver"].(string),
		Options: make(map[string]interface{}),
	}

	// Copy all config options
	for k, v := range config {
		clientConfig.Options[k] = v
	}

	// Set default timeout if not specified
	if _, exists := clientConfig.Options["timeout"]; !exists {
		clientConfig.Timeout = 30 * time.Second
	}

	// Set default retries if not specified
	if _, exists := clientConfig.Options["retries"]; !exists {
		clientConfig.Retries = 3
	}

	return clientConfig
}

// createLocalProvider creates a local file logging provider
func (f *LoggingProviderFactory) createLocalProvider(config *clients_core.ClientConfig) (clients_core.LoggingClientInterface, error) {
	provider := NewLocalLoggingProvider(config)

	// Connect the provider
	if err := provider.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect local logging provider: %w", err)
	}

	return provider, nil
}

// createSentryProvider creates a Sentry logging provider
func (f *LoggingProviderFactory) createSentryProvider(config *clients_core.ClientConfig) (clients_core.LoggingClientInterface, error) {
	provider := NewSentryLoggingProvider(config)

	// Connect the provider
	if err := provider.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect Sentry logging provider: %w", err)
	}

	return provider, nil
}

// createStackProvider creates a stack logging provider
func (f *LoggingProviderFactory) createStackProvider(config *clients_core.ClientConfig) (clients_core.LoggingClientInterface, error) {
	provider := NewStackLoggingProvider(config)

	// Connect the provider
	if err := provider.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect stack logging provider: %w", err)
	}

	return provider, nil
}

// createNullProvider creates a null logging provider
func (f *LoggingProviderFactory) createNullProvider(config *clients_core.ClientConfig) (clients_core.LoggingClientInterface, error) {
	provider := NewNullLoggingProvider(config)

	// Null provider doesn't need to connect
	return provider, nil
}

// createDailyProvider creates a daily file logging provider
func (f *LoggingProviderFactory) createDailyProvider(config *clients_core.ClientConfig) (clients_core.LoggingClientInterface, error) {
	// For now, use the local provider as daily is similar
	provider := NewLocalLoggingProvider(config)

	// Connect the provider
	if err := provider.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect daily logging provider: %w", err)
	}

	return provider, nil
}

// createSlackProvider creates a Slack logging provider
func (f *LoggingProviderFactory) createSlackProvider(config *clients_core.ClientConfig) (clients_core.LoggingClientInterface, error) {
	provider := NewSlackLoggingProvider(config)

	// Connect the provider
	if err := provider.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect Slack logging provider: %w", err)
	}

	return provider, nil
}

// createStderrProvider creates a stderr logging provider
func (f *LoggingProviderFactory) createStderrProvider(config *clients_core.ClientConfig) (clients_core.LoggingClientInterface, error) {
	provider := NewStderrLoggingProvider(config)

	// Stderr provider doesn't need to connect
	return provider, nil
}

// createEmergencyProvider creates an emergency logging provider
func (f *LoggingProviderFactory) createEmergencyProvider(config *clients_core.ClientConfig) (clients_core.LoggingClientInterface, error) {
	provider := NewEmergencyLoggingProvider(config)

	// Connect the provider
	if err := provider.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect emergency logging provider: %w", err)
	}

	return provider, nil
}

// createSyslogProvider creates a syslog provider
func (f *LoggingProviderFactory) createSyslogProvider(config *clients_core.ClientConfig) (clients_core.LoggingClientInterface, error) {
	// For now, use the local provider as syslog is similar
	provider := NewLocalLoggingProvider(config)

	// Connect the provider
	if err := provider.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect syslog provider: %w", err)
	}

	return provider, nil
}

// createErrorLogProvider creates an error log provider
func (f *LoggingProviderFactory) createErrorLogProvider(config *clients_core.ClientConfig) (clients_core.LoggingClientInterface, error) {
	// For now, use the local provider as errorlog is similar
	provider := NewLocalLoggingProvider(config)

	// Connect the provider
	if err := provider.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect error log provider: %w", err)
	}

	return provider, nil
}

// LoggingProviderRegistry manages logging provider bindings
type LoggingProviderRegistry struct {
	container *app_core.ServiceContainer
	factory   *LoggingProviderFactory
}

// NewLoggingProviderRegistry creates a new logging provider registry
func NewLoggingProviderRegistry(container *app_core.ServiceContainer) *LoggingProviderRegistry {
	return &LoggingProviderRegistry{
		container: container,
		factory:   NewLoggingProviderFactory(container),
	}
}

// Register registers a logging provider with the container
func (r *LoggingProviderRegistry) Register(name string, config map[string]interface{}) error {
	// Create the provider
	provider, err := r.factory.Create(config)
	if err != nil {
		return fmt.Errorf("failed to create logging provider %s: %w", name, err)
	}

	// Register as singleton
	r.container.Singleton("logging.provider."+name, provider)

	// If this is the default provider, also register it as the main provider
	if name == "default" {
		r.container.Singleton("logging.provider", provider)
	}

	return nil
}

// GetProvider gets a logging provider from the container
func (r *LoggingProviderRegistry) GetProvider(name string) (clients_core.LoggingClientInterface, error) {
	providerInterface, err := r.container.Resolve("logging.provider." + name)
	if err != nil {
		return nil, err
	}

	provider, ok := providerInterface.(clients_core.LoggingClientInterface)
	if !ok {
		return nil, fmt.Errorf("logging provider is not a LoggingClientInterface")
	}

	return provider, nil
}

// GetDefaultProvider gets the default logging provider
func (r *LoggingProviderRegistry) GetDefaultProvider() (clients_core.LoggingClientInterface, error) {
	providerInterface, err := r.container.Resolve("logging.provider")
	if err != nil {
		return nil, err
	}

	provider, ok := providerInterface.(clients_core.LoggingClientInterface)
	if !ok {
		return nil, fmt.Errorf("logging provider is not a LoggingClientInterface")
	}

	return provider, nil
}

// RegisterFromConfig registers a logging provider from configuration
func (f *LoggingProviderFactory) RegisterFromConfig(config map[string]interface{}) error {
	// Get default channel from config
	defaultChannel, ok := config["default"].(string)
	if !ok {
		return fmt.Errorf("default logging channel not set in config")
	}

	// Get channels configuration
	channels, ok := config["channels"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("logging channels not configured")
	}

	// Get the specific channel config
	channelConfig, ok := channels[defaultChannel].(map[string]interface{})
	if !ok {
		return fmt.Errorf("channel config for %s not found", defaultChannel)
	}

	// Create the provider
	provider, err := f.Create(channelConfig)
	if err != nil {
		return err
	}

	// Register as singleton
	f.container.Singleton("logging.provider", provider)
	return nil
}
