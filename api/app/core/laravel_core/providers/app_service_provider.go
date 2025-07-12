package providers

import (
	"context"
	"fmt"
	"log"

	app_core "base_lara_go_project/app/core/go_core"
	facades_core "base_lara_go_project/app/core/laravel_core/facades"
	logging_core "base_lara_go_project/app/core/laravel_core/logging"
	"base_lara_go_project/config"
)

// AppServiceProvider is the main application service provider
// It loads all core providers and can be extended by developers
type AppServiceProvider struct {
	BaseServiceProvider
}

// Register registers all core providers and any additional providers
func (p *AppServiceProvider) Register(container *app_core.Container) error {
	// Register core infrastructure providers
	coreProviders := []ServiceProvider{
		&CoreServiceProvider{},
		&ConfigServiceProvider{},
		&DatabaseServiceProvider{},
		&CacheServiceProvider{},
		&EventServiceProvider{},
		&QueueServiceProvider{},
		&MailServiceProvider{},
		&LoggingServiceProvider{},
		&JobServiceProvider{},
		&MigrationServiceProvider{},
		&ValidationServiceProvider{},
		&GoroutineServiceProvider{},
		&ContextServiceProvider{},
	}

	// Register all core providers
	for _, provider := range coreProviders {
		if err := provider.Register(container); err != nil {
			log.Printf("Failed to register provider %T: %v", provider, err)
			return err
		}
	}

	// Call the developer's custom register method
	if err := p.registerCustomProviders(container); err != nil {
		return err
	}

	log.Printf("App service provider registered successfully")
	return nil
}

// Boot boots all providers
func (p *AppServiceProvider) Boot(container *app_core.Container) error {
	// Boot core providers
	coreProviders := []ServiceProvider{
		&CoreServiceProvider{},
		&DatabaseServiceProvider{},
		&CacheServiceProvider{},
		&EventServiceProvider{},
		&QueueServiceProvider{},
		&MailServiceProvider{},
		&LoggingServiceProvider{},
		&JobServiceProvider{},
		&MigrationServiceProvider{},
		&ValidationServiceProvider{},
		&GoroutineServiceProvider{},
	}

	for _, provider := range coreProviders {
		if err := provider.Boot(container); err != nil {
			log.Printf("Failed to boot provider %T: %v", provider, err)
			return err
		}
	}

	// Call the developer's custom boot method
	if err := p.bootCustomProviders(container); err != nil {
		return err
	}

	log.Printf("App service provider booted successfully")
	return nil
}

// Provides returns the services this provider provides
func (p *AppServiceProvider) Provides() []string {
	return []string{"app"}
}

// When returns the conditions when this provider should be loaded
func (p *AppServiceProvider) When() []string {
	return []string{}
}

// registerCustomProviders is called by developers to register additional providers
func (p *AppServiceProvider) registerCustomProviders(container *app_core.Container) error {
	// This method can be overridden by developers
	return nil
}

// bootCustomProviders is called by developers to boot additional providers
func (p *AppServiceProvider) bootCustomProviders(container *app_core.Container) error {
	// This method can be overridden by developers
	return nil
}

// Individual core service providers for better organization

// ConfigServiceProvider handles configuration system registration
type ConfigServiceProvider struct {
	BaseServiceProvider
}

func (p *ConfigServiceProvider) Register(container *app_core.Container) error {
	// Get the config loader
	configLoader := config.GetConfigLoader()

	// Register the config loader as a singleton
	container.Singleton("config.loader", func() (any, error) {
		return configLoader, nil
	})

	// Register the config facade as a singleton
	container.Singleton("config", func() (any, error) {
		return facades_core.Config(), nil
	})

	log.Printf("Config system initialized with %d configs: %v",
		len(configLoader.ListAvailableConfigs()),
		configLoader.ListAvailableConfigs())

	return nil
}

func (p *ConfigServiceProvider) Provides() []string {
	return []string{"config"}
}

// DatabaseServiceProvider handles database registration
type DatabaseServiceProvider struct {
	BaseServiceProvider
}

func (p *DatabaseServiceProvider) Register(container *app_core.Container) error {
	// TODO: Implement database registration
	container.Singleton("database", func() (any, error) {
		return nil, nil
	})
	container.Singleton("gorm.db", func() (any, error) {
		return nil, nil
	})
	return nil
}

func (p *DatabaseServiceProvider) Provides() []string {
	return []string{"database"}
}

// CacheServiceProvider handles cache registration
type CacheServiceProvider struct {
	BaseServiceProvider
}

func (p *CacheServiceProvider) Register(container *app_core.Container) error {
	// TODO: Implement cache registration
	container.Singleton("cache", func() (any, error) {
		return app_core.NewLocalCache[any](), nil
	})
	return nil
}

func (p *CacheServiceProvider) Provides() []string {
	return []string{"cache"}
}

// EventServiceProvider handles event system registration
type EventServiceProvider struct {
	BaseServiceProvider
}

func (p *EventServiceProvider) Register(container *app_core.Container) error {
	// TODO: Implement event system registration
	// Create event bus and store
	eventBus := app_core.NewEventBus[any](nil, nil, nil)
	eventStore := app_core.NewMemoryEventStore[any]()

	// Create event manager
	eventManager := app_core.NewEventManager[any](eventBus, eventStore)

	container.Singleton("event_manager", func() (any, error) {
		return eventManager, nil
	})
	return nil
}

func (p *EventServiceProvider) Provides() []string {
	return []string{"events"}
}

// QueueServiceProvider handles queue system registration
type QueueServiceProvider struct {
	BaseServiceProvider
}

func (p *QueueServiceProvider) Register(container *app_core.Container) error {
	// TODO: Implement queue system registration
	container.Singleton("queue", func() (any, error) {
		return app_core.NewSyncQueue[any](), nil
	})
	return nil
}

func (p *QueueServiceProvider) Provides() []string {
	return []string{"queue"}
}

// MailServiceProvider handles mail system registration
type MailServiceProvider struct {
	BaseServiceProvider
}

func (p *MailServiceProvider) Register(container *app_core.Container) error {
	// TODO: Implement mail system registration
	container.Singleton("mail", func() (any, error) {
		return nil, nil
	})
	return nil
}

func (p *MailServiceProvider) Provides() []string {
	return []string{"mail"}
}

// LoggingServiceProvider handles logging system registration
type LoggingServiceProvider struct {
	BaseServiceProvider
}

func (p *LoggingServiceProvider) Register(container *app_core.Container) error {
	// Create logging factory
	loggingFactory := logging_core.NewGenericLogFactory[map[string]interface{}]()

	// Create logger
	logger, err := loggingFactory.CreateLogger("default")
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	// Register handlers based on configuration
	if err := p.registerHandlers(logger); err != nil {
		return fmt.Errorf("failed to register logging handlers: %w", err)
	}

	// Register integrations if available
	p.registerIntegrations(logger, container)

	// Create monitoring system
	monitor := logging_core.NewLoggingMonitor[map[string]interface{}]()
	monitor.Start(context.Background())

	// Register the logger as a singleton
	container.Singleton("logger", func() (any, error) {
		return logger, nil
	})

	// Register monitoring components
	container.Singleton("logging.monitor", func() (any, error) {
		return monitor, nil
	})

	// Register individual logging methods for convenience
	container.Singleton("logging.debug", func() (any, error) {
		return logger.Debug, nil
	})
	container.Singleton("logging.info", func() (any, error) {
		return logger.Info, nil
	})
	container.Singleton("logging.error", func() (any, error) {
		return logger.Error, nil
	})

	return nil
}

func (p *LoggingServiceProvider) Provides() []string {
	return []string{"logging", "logger"}
}

// registerIntegrations registers cache, queue, and event integrations
func (p *LoggingServiceProvider) registerIntegrations(logger *app_core.Logger[map[string]interface{}], container *app_core.Container) {
	// Try to get cache instance
	if cacheInstance, err := container.Resolve("cache"); err == nil {
		if _, ok := cacheInstance.(app_core.Cache[map[string]interface{}]); ok {
			// Note: CacheLogHandler doesn't exist in the current implementation
			// This would be added when the cache integration is properly implemented
		}
	}

	// Try to get queue instance
	if queueInstance, err := container.Resolve("queue"); err == nil {
		if queue, ok := queueInstance.(app_core.Queue[map[string]interface{}]); ok {
			queueHandler := logging_core.NewQueueLogHandler(queue)
			logger.AddHandler("queue", queueHandler)
		}
	}
}

// registerHandlers registers logging handlers based on configuration
func (p *LoggingServiceProvider) registerHandlers(logger *app_core.Logger[map[string]interface{}]) error {
	// Get logging configuration
	loggingConfig := config.Get("logging").(map[string]interface{})
	channels := loggingConfig["channels"].(map[string]interface{})

	// Register handlers for each channel
	for channelName, channelConfig := range channels {
		configMap := channelConfig.(map[string]interface{})

		if err := p.registerSingleHandler(logger, configMap); err != nil {
			return fmt.Errorf("failed to register handler for channel %s: %w", channelName, err)
		}
	}

	return nil
}

// registerSingleHandler registers a single logging handler
func (p *LoggingServiceProvider) registerSingleHandler(logger *app_core.Logger[map[string]interface{}], config map[string]interface{}) error {
	driver := config["driver"].(string)

	switch driver {
	case "single":
		return p.registerSingleHandler(logger, config)
	case "daily":
		return p.registerDailyHandler(logger, config)
	case "stack":
		return p.registerStackHandler(logger, config)
	case "sentry":
		return p.registerSentryHandler(logger, config)
	case "slack":
		return p.registerSlackHandler(logger, config)
	case "null":
		return p.registerNullHandler(logger, config)
	default:
		return fmt.Errorf("unknown logging driver: %s", driver)
	}
}

// registerDailyHandler registers a daily logging handler
func (p *LoggingServiceProvider) registerDailyHandler(logger *app_core.Logger[map[string]interface{}], config map[string]interface{}) error {
	path := config["path"].(string)
	handler, err := logging_core.NewOptimizedDailyHandler[map[string]interface{}](path)
	if err != nil {
		return fmt.Errorf("failed to create daily handler: %w", err)
	}

	logger.AddHandler("daily", handler)
	return nil
}

// registerStackHandler registers a stack logging handler
func (p *LoggingServiceProvider) registerStackHandler(logger *app_core.Logger[map[string]interface{}], config map[string]interface{}) error {
	channels := config["channels"].([]interface{})
	var handlers []app_core.LogHandler[map[string]interface{}]

	for _, channelName := range channels {
		handler, err := p.createHandlerForChannel(channelName.(string), config)
		if err != nil {
			return fmt.Errorf("failed to create handler for channel %s: %w", channelName, err)
		}
		handlers = append(handlers, handler)
	}

	stackHandler := logging_core.NewOptimizedStackHandler(handlers)
	logger.AddHandler("stack", stackHandler)
	return nil
}

// registerSentryHandler registers a Sentry logging handler
func (p *LoggingServiceProvider) registerSentryHandler(logger *app_core.Logger[map[string]interface{}], config map[string]interface{}) error {
	handler := logging_core.NewOptimizedSentryHandler[map[string]interface{}]()
	logger.AddHandler("sentry", handler)
	return nil
}

// registerSlackHandler registers a Slack logging handler
func (p *LoggingServiceProvider) registerSlackHandler(logger *app_core.Logger[map[string]interface{}], config map[string]interface{}) error {
	handler := logging_core.NewOptimizedSlackHandler[map[string]interface{}]()
	logger.AddHandler("slack", handler)
	return nil
}

// registerNullHandler registers a null logging handler
func (p *LoggingServiceProvider) registerNullHandler(logger *app_core.Logger[map[string]interface{}], config map[string]interface{}) error {
	handler := logging_core.NewOptimizedNullHandler[map[string]interface{}]()
	logger.AddHandler("null", handler)
	return nil
}

// createHandlerForChannel creates a handler for a specific channel
func (p *LoggingServiceProvider) createHandlerForChannel(channelName string, config map[string]interface{}) (app_core.LogHandler[map[string]interface{}], error) {
	// This is a simplified implementation
	// In a real implementation, you'd create the appropriate handler based on the channel configuration
	path := fmt.Sprintf("storage/logs/%s.log", channelName)
	handler, err := logging_core.NewOptimizedFileHandler[map[string]interface{}](path)
	if err != nil {
		return nil, fmt.Errorf("failed to create file handler: %w", err)
	}

	return handler, nil
}

// JobServiceProvider handles job system registration
type JobServiceProvider struct {
	BaseServiceProvider
}

func (p *JobServiceProvider) Register(container *app_core.Container) error {
	// TODO: Implement job system registration
	queueInstance, err := container.Resolve("queue")
	if err != nil {
		return err
	}

	queue := queueInstance.(app_core.Queue[any])
	// Create job dispatcher
	jobDispatcher := app_core.NewJobDispatcher[any](queue, nil, nil, nil)

	container.Singleton("job.dispatcher", func() (any, error) {
		return jobDispatcher, nil
	})
	return nil
}

func (p *JobServiceProvider) Provides() []string {
	return []string{"jobs"}
}
