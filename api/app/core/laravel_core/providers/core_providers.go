package providers

import (
	app_core "base_lara_go_project/app/core/go_core"
	"base_lara_go_project/config"
	"log"
)

// CoreServiceProvider automatically registers all core services
type CoreServiceProvider struct {
	BaseServiceProvider
}

// Register registers all core services automatically
func (p *CoreServiceProvider) Register(container *app_core.Container) error {
	// Register database
	if err := p.registerDatabase(container); err != nil {
		return err
	}

	// Register cache
	if err := p.registerCache(container); err != nil {
		return err
	}

	// Register event system
	if err := p.registerEventSystem(container); err != nil {
		return err
	}

	// Register queue system
	if err := p.registerQueueSystem(container); err != nil {
		return err
	}

	// Register mail system
	if err := p.registerMailSystem(container); err != nil {
		return err
	}

	// Register logging system
	if err := p.registerLoggingSystem(container); err != nil {
		return err
	}

	// Register job system
	if err := p.registerJobSystem(container); err != nil {
		return err
	}

	log.Printf("Core services registered successfully")
	return nil
}

// Boot boots all core services
func (p *CoreServiceProvider) Boot(container *app_core.Container) error {
	// Initialize core systems
	app_core.InitializeRegistry()
	app_core.InitializeEventDispatcher()

	log.Printf("Core services booted successfully")
	return nil
}

// Provides returns the services this provider provides
func (p *CoreServiceProvider) Provides() []string {
	return []string{
		"database", "cache", "events", "queue", "mail", "logging", "jobs",
	}
}

// registerDatabase registers the database system
func (p *CoreServiceProvider) registerDatabase(container *app_core.Container) error {
	// TODO: Implement database registration using existing database provider
	// For now, we'll use a placeholder
	container.Singleton("database", func() (any, error) {
		return nil, nil
	})

	container.Singleton("gorm.db", func() (any, error) {
		return nil, nil
	})

	return nil
}

// registerCache registers the cache system
func (p *CoreServiceProvider) registerCache(container *app_core.Container) error {
	cacheConfig := config.CacheConfig()

	// Create cache instance based on configuration
	var cache app_core.Cache[any]
	driver := cacheConfig["default"].(string)

	switch driver {
	case "redis":
		// TODO: Implement Redis cache
		cache = app_core.NewLocalCache[any]()
	case "file":
		// TODO: Implement file cache
		cache = app_core.NewLocalCache[any]()
	default:
		cache = app_core.NewLocalCache[any]()
	}

	// Register cache as singleton
	container.Singleton("cache", func() (any, error) {
		return cache, nil
	})

	// Register typed cache instances for all models
	container.Singleton("cache.user", func() (any, error) {
		return app_core.NewLocalCache[any](), nil
	})

	container.Singleton("cache.session", func() (any, error) {
		return app_core.NewLocalCache[any](), nil
	})

	return nil
}

// registerEventSystem registers the event system
func (p *CoreServiceProvider) registerEventSystem(container *app_core.Container) error {
	// Create event bus and store
	eventBus := app_core.NewEventBus[any]()
	eventStore := app_core.NewMemoryEventStore[any]()

	// Create event manager
	eventManager := app_core.NewEventManager[any](eventBus, eventStore)

	// Register event manager as singleton
	container.Singleton("event_manager", func() (any, error) {
		return eventManager, nil
	})

	return nil
}

// registerQueueSystem registers the queue system
func (p *CoreServiceProvider) registerQueueSystem(container *app_core.Container) error {
	queueConfig := config.QueueConfig()

	// Create queue based on configuration
	var queue app_core.Queue[any]
	driver := queueConfig["default"].(string)

	switch driver {
	case "redis":
		// TODO: Implement Redis queue
		queue = app_core.NewSyncQueue[any]()
	case "sync":
		queue = app_core.NewSyncQueue[any]()
	default:
		queue = app_core.NewSyncQueue[any]()
	}

	// Register queue as singleton
	container.Singleton("queue", func() (any, error) {
		return queue, nil
	})

	return nil
}

// registerMailSystem registers the mail system
func (p *CoreServiceProvider) registerMailSystem(container *app_core.Container) error {
	// TODO: Implement mail service registration
	// For now, we'll use a placeholder
	container.Singleton("mail", func() (any, error) {
		return nil, nil
	})

	return nil
}

// registerLoggingSystem registers the logging system
func (p *CoreServiceProvider) registerLoggingSystem(container *app_core.Container) error {
	// TODO: Implement logger registration
	// For now, we'll use a placeholder
	container.Singleton("logger", func() (any, error) {
		return nil, nil
	})

	return nil
}

// registerJobSystem registers the job system
func (p *CoreServiceProvider) registerJobSystem(container *app_core.Container) error {
	// Get queue from container
	queueInstance, err := container.Resolve("queue")
	if err != nil {
		return err
	}

	queue := queueInstance.(app_core.Queue[any])

	// Create job dispatcher
	jobDispatcher := app_core.NewJobDispatcher[any](queue)

	// Register job dispatcher as singleton
	container.Singleton("job.dispatcher", func() (any, error) {
		return jobDispatcher, nil
	})

	return nil
}
