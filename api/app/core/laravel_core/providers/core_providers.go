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

	// Register optimization services
	if err := p.registerOptimizationServices(container); err != nil {
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

// registerCache registers the cache system with automatic context optimization
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

	// Create context-aware cache (automatic optimization)
	contextManager := app_core.NewContextManager(app_core.DefaultContextConfig())
	contextAwareCache := app_core.NewContextAwareCache[any](cache, contextManager)

	// Register both the original and context-aware versions
	container.Singleton("cache", func() (any, error) {
		return cache, nil
	})

	// Register context-aware version as the primary cache
	container.Singleton("cache.context_aware", func() (any, error) {
		return contextAwareCache, nil
	})

	// Register typed cache instances for all models
	container.Singleton("cache.user", func() (any, error) {
		userCache := app_core.NewLocalCache[any]()
		return app_core.NewContextAwareCache[any](userCache, contextManager), nil
	})

	container.Singleton("cache.session", func() (any, error) {
		sessionCache := app_core.NewLocalCache[any]()
		return app_core.NewContextAwareCache[any](sessionCache, contextManager), nil
	})

	return nil
}

// registerEventSystem registers the event system with automatic context optimization
func (p *CoreServiceProvider) registerEventSystem(container *app_core.Container) error {
	// Resolve optimization singletons
	wspInstance, _ := container.Resolve("optimization.work_stealing")
	caInstance, _ := container.Resolve("optimization.custom_allocator")
	pgoInstance, _ := container.Resolve("optimization.profile_guided")

	// Type assertions
	var wsp *app_core.WorkStealingPool[any]
	var ca *app_core.CustomAllocator[any]
	var pgo *app_core.ProfileGuidedOptimizer[any]

	if wspInstance != nil {
		wsp = wspInstance.(*app_core.WorkStealingPool[any])
	}
	if caInstance != nil {
		ca = caInstance.(*app_core.CustomAllocator[any])
	}
	if pgoInstance != nil {
		pgo = pgoInstance.(*app_core.ProfileGuidedOptimizer[any])
	}

	// Create event bus and store with optimizations
	eventBus := app_core.NewEventBus[any](wsp, ca, pgo)
	eventStore := app_core.NewMemoryEventStore[any]()

	// Create event manager
	eventManager := app_core.NewEventManager[any](eventBus, eventStore)

	// Get goroutine manager and context config for unified optimized dispatcher
	goroutineManagerInstance, _ := container.Resolve("goroutine.manager")
	contextConfigInstance, _ := container.Resolve("context.config")

	var gm *app_core.GoroutineManager[any]
	var cc *app_core.ContextConfig

	if goroutineManagerInstance != nil {
		gm = goroutineManagerInstance.(*app_core.GoroutineManager[any])
	}
	if contextConfigInstance != nil {
		cc = contextConfigInstance.(*app_core.ContextConfig)
	}

	// Create unified optimized event dispatcher
	optimizedEventDispatcher := app_core.NewOptimizedEventDispatcher[any](eventManager, gm, cc)

	// Register both the original and optimized versions
	container.Singleton("event_manager", func() (any, error) {
		return eventManager, nil
	})

	// Register optimized version as the primary event dispatcher
	container.Singleton("event_dispatcher", func() (any, error) {
		return optimizedEventDispatcher, nil
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
	// Resolve optimization singletons
	wspInstance, _ := container.Resolve("optimization.work_stealing")
	caInstance, _ := container.Resolve("optimization.custom_allocator")
	pgoInstance, _ := container.Resolve("optimization.profile_guided")

	// Type assertions
	var wsp *app_core.WorkStealingPool[any]
	var ca *app_core.CustomAllocator[any]
	var pgo *app_core.ProfileGuidedOptimizer[any]

	if wspInstance != nil {
		wsp = wspInstance.(*app_core.WorkStealingPool[any])
	}
	if caInstance != nil {
		ca = caInstance.(*app_core.CustomAllocator[any])
	}
	if pgoInstance != nil {
		pgo = pgoInstance.(*app_core.ProfileGuidedOptimizer[any])
	}

	// Create mailer with optimizations
	mailer := app_core.NewSMTPMailer[any]("localhost", 587, "", "", "noreply@example.com", wsp, ca, pgo)

	// Register mailer as singleton
	container.Singleton("mail", func() (any, error) {
		return mailer, nil
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

// registerJobSystem registers the job system with automatic context optimization
func (p *CoreServiceProvider) registerJobSystem(container *app_core.Container) error {
	// Get queue from container
	queueInstance, err := container.Resolve("queue")
	if err != nil {
		return err
	}

	queue := queueInstance.(app_core.Queue[any])

	// Resolve optimization singletons
	wspInstance, _ := container.Resolve("optimization.work_stealing")
	caInstance, _ := container.Resolve("optimization.custom_allocator")
	pgoInstance, _ := container.Resolve("optimization.profile_guided")

	// Type assertions
	var wsp *app_core.WorkStealingPool[any]
	var ca *app_core.CustomAllocator[any]
	var pgo *app_core.ProfileGuidedOptimizer[any]

	if wspInstance != nil {
		wsp = wspInstance.(*app_core.WorkStealingPool[any])
	}
	if caInstance != nil {
		ca = caInstance.(*app_core.CustomAllocator[any])
	}
	if pgoInstance != nil {
		pgo = pgoInstance.(*app_core.ProfileGuidedOptimizer[any])
	}

	// Create job dispatcher with optimizations
	jobDispatcher := app_core.NewJobDispatcher[any](queue, wsp, ca, pgo)

	// Create context-aware job dispatcher (automatic optimization)
	contextAwareJobDispatcher := app_core.NewContextAwareJobDispatcher[any](jobDispatcher)

	// Register both the original and context-aware versions
	container.Singleton("job.dispatcher", func() (any, error) {
		return jobDispatcher, nil
	})

	// Register context-aware version as the primary job dispatcher
	container.Singleton("job_dispatcher", func() (any, error) {
		return contextAwareJobDispatcher, nil
	})

	return nil
}

// registerOptimizationServices registers optimization services
func (p *CoreServiceProvider) registerOptimizationServices(container *app_core.Container) error {
	// Create optimization service provider
	optimizationProvider := &OptimizationServiceProvider{}

	// Register optimization services
	if err := optimizationProvider.Register(container); err != nil {
		return err
	}

	// Boot optimization services
	if err := optimizationProvider.Boot(container); err != nil {
		return err
	}

	return nil
}
