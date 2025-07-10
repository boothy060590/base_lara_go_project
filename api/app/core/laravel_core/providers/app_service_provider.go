package providers

import (
	"base_lara_go_project/app/core/go_core"
	app_core "base_lara_go_project/app/core/go_core"
	config_core "base_lara_go_project/app/core/laravel_core/config"
	logging_core "base_lara_go_project/app/core/laravel_core/logging"
	"fmt"
	"log"
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
	// Create config facade
	config := &config_core.ConfigFacade{}

	// Create optimized logging facade
	loggingFacade := logging_core.NewGenericLoggingFacade[map[string]interface{}](config)

	// Register handlers based on configuration
	if err := p.registerHandlers(loggingFacade); err != nil {
		return fmt.Errorf("failed to register logging handlers: %w", err)
	}

	// Register integrations if available
	p.registerIntegrations(loggingFacade, container)

	// Create monitoring system
	monitor := logging_core.NewLoggingMonitor[map[string]interface{}](config)
	monitor.Start()

	// Create metrics collector
	collector := logging_core.NewLoggingMetricsCollector(monitor)

	// Add health checks
	monitor.AddHealthCheck("logging_system", func() (bool, error) {
		// Check if logging system is healthy
		return true, nil
	})

	// Register the logging facade as a singleton
	container.Singleton("logger", func() (any, error) {
		return loggingFacade, nil
	})

	// Register monitoring components
	container.Singleton("logging.monitor", func() (any, error) {
		return monitor, nil
	})

	container.Singleton("logging.collector", func() (any, error) {
		return collector, nil
	})

	// Register individual logging methods for convenience
	container.Singleton("logging.debug", func() (any, error) {
		return loggingFacade.Debug, nil
	})
	container.Singleton("logging.info", func() (any, error) {
		return loggingFacade.Info, nil
	})
	container.Singleton("logging.error", func() (any, error) {
		return loggingFacade.Error, nil
	})

	return nil
}

func (p *LoggingServiceProvider) Provides() []string {
	return []string{"logging", "logger"}
}

// registerIntegrations registers cache, queue, and event integrations
func (p *LoggingServiceProvider) registerIntegrations(facade *logging_core.GenericLoggingFacade[map[string]interface{}], container *app_core.Container) {
	config := &config_core.ConfigFacade{}

	// Try to get cache instance
	if cacheInstance, err := container.Resolve("cache"); err == nil {
		if _, ok := cacheInstance.(go_core.Cache[map[string]interface{}]); ok {
			// Note: CacheLogHandler doesn't exist in the current implementation
			// This would be added when the cache integration is properly implemented
		}
	}

	// Try to get queue instance
	if queueInstance, err := container.Resolve("queue"); err == nil {
		if queue, ok := queueInstance.(go_core.Queue[map[string]interface{}]); ok {
			queueHandler := logging_core.NewQueueLogHandler(queue, config)
			facade.AddHandler("queue", queueHandler)
		}
	}

	// Try to get event dispatcher instance
	if eventInstance, err := container.Resolve("events"); err == nil {
		if eventDispatcher, ok := eventInstance.(go_core.EventDispatcher[map[string]interface{}]); ok {
			eventHandler := logging_core.NewEventLogHandler(eventDispatcher, config)
			facade.AddHandler("events", eventHandler)
		}
	}
}

// registerHandlers registers logging handlers based on configuration
func (p *LoggingServiceProvider) registerHandlers(facade *logging_core.GenericLoggingFacade[map[string]interface{}]) error {
	config := &config_core.ConfigFacade{}
	channels := config.Get("logging.channels").(map[string]interface{})
	defaultChannel := config.Get("logging.default").(string)

	// Get default channel configuration
	channelConfig, exists := channels[defaultChannel]
	if !exists {
		return fmt.Errorf("default channel '%s' not found", defaultChannel)
	}

	configMap := channelConfig.(map[string]interface{})
	driver := configMap["driver"].(string)

	switch driver {
	case "single":
		return p.registerSingleHandler(facade, configMap)
	case "daily":
		return p.registerDailyHandler(facade, configMap)
	case "stack":
		return p.registerStackHandler(facade, configMap)
	case "sentry":
		return p.registerSentryHandler(facade, configMap)
	case "slack":
		return p.registerSlackHandler(facade, configMap)
	case "null":
		return p.registerNullHandler(facade, configMap)
	default:
		return fmt.Errorf("unknown logging driver: %s", driver)
	}
}

// registerSingleHandler registers a single file handler
func (p *LoggingServiceProvider) registerSingleHandler(facade *logging_core.GenericLoggingFacade[map[string]interface{}], config map[string]interface{}) error {
	path := config["path"].(string)

	handler, err := logging_core.NewOptimizedFileHandler[map[string]interface{}](&config_core.ConfigFacade{}, path)
	if err != nil {
		return fmt.Errorf("failed to create single file handler: %w", err)
	}

	facade.AddHandler("single", handler)
	return nil
}

// registerDailyHandler registers a daily rotating file handler
func (p *LoggingServiceProvider) registerDailyHandler(facade *logging_core.GenericLoggingFacade[map[string]interface{}], config map[string]interface{}) error {
	path := config["path"].(string)
	// Remove .log extension for base path
	basePath := path[:len(path)-4] // Remove .log extension

	handler, err := logging_core.NewOptimizedDailyHandler[map[string]interface{}](&config_core.ConfigFacade{}, basePath)
	if err != nil {
		return fmt.Errorf("failed to create daily handler: %w", err)
	}

	facade.AddHandler("daily", handler)
	return nil
}

// registerStackHandler registers a stack handler
func (p *LoggingServiceProvider) registerStackHandler(facade *logging_core.GenericLoggingFacade[map[string]interface{}], config map[string]interface{}) error {
	stackChannels := config["channels"].([]interface{})
	handlers := make([]go_core.LogHandler[map[string]interface{}], 0, len(stackChannels))

	// Get channels config
	configFacade := &config_core.ConfigFacade{}
	channels := configFacade.Get("logging.channels").(map[string]interface{})

	// Create handlers for each channel in the stack
	for _, channelName := range stackChannels {
		channel := channelName.(string)
		channelConfig := channels[channel].(map[string]interface{})

		handler, err := p.createHandlerForChannel(channel, channelConfig)
		if err != nil {
			return fmt.Errorf("failed to create handler for channel %s: %w", channel, err)
		}

		handlers = append(handlers, handler)
	}

	stackHandler := logging_core.NewOptimizedStackHandler[map[string]interface{}](&config_core.ConfigFacade{}, handlers)
	facade.AddHandler("stack", stackHandler)
	return nil
}

// registerSentryHandler registers a Sentry handler
func (p *LoggingServiceProvider) registerSentryHandler(facade *logging_core.GenericLoggingFacade[map[string]interface{}], config map[string]interface{}) error {
	handler := logging_core.NewOptimizedSentryHandler[map[string]interface{}](&config_core.ConfigFacade{})
	facade.AddHandler("sentry", handler)
	return nil
}

// registerSlackHandler registers a Slack handler
func (p *LoggingServiceProvider) registerSlackHandler(facade *logging_core.GenericLoggingFacade[map[string]interface{}], config map[string]interface{}) error {
	handler := logging_core.NewOptimizedSlackHandler[map[string]interface{}](&config_core.ConfigFacade{})
	facade.AddHandler("slack", handler)
	return nil
}

// registerNullHandler registers a null handler
func (p *LoggingServiceProvider) registerNullHandler(facade *logging_core.GenericLoggingFacade[map[string]interface{}], config map[string]interface{}) error {
	handler := logging_core.NewOptimizedNullHandler[map[string]interface{}]()
	facade.AddHandler("null", handler)
	return nil
}

// createHandlerForChannel creates a handler for a specific channel
func (p *LoggingServiceProvider) createHandlerForChannel(channelName string, config map[string]interface{}) (go_core.LogHandler[map[string]interface{}], error) {
	driver := config["driver"].(string)

	switch driver {
	case "single":
		path := config["path"].(string)
		return logging_core.NewOptimizedFileHandler[map[string]interface{}](&config_core.ConfigFacade{}, path)
	case "daily":
		path := config["path"].(string)
		basePath := path[:len(path)-4] // Remove .log extension
		return logging_core.NewOptimizedDailyHandler[map[string]interface{}](&config_core.ConfigFacade{}, basePath)
	case "sentry":
		return logging_core.NewOptimizedSentryHandler[map[string]interface{}](&config_core.ConfigFacade{}), nil
	case "slack":
		return logging_core.NewOptimizedSlackHandler[map[string]interface{}](&config_core.ConfigFacade{}), nil
	case "null":
		return logging_core.NewOptimizedNullHandler[map[string]interface{}](), nil
	default:
		return nil, fmt.Errorf("unknown driver for channel %s: %s", channelName, driver)
	}
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
