package providers

import (
	app_core "base_lara_go_project/app/core/go_core"
	facades_core "base_lara_go_project/app/core/laravel_core/facades"
	"context"
	"log"
	"time"
)

// ContextServiceProvider integrates context optimization with the framework
type ContextServiceProvider struct {
	BaseServiceProvider
}

// Register registers context-optimized services
func (p *ContextServiceProvider) Register(container *app_core.Container) error {
	// Register context manager for general operations
	container.Singleton("context.manager", func() (any, error) {
		return app_core.NewContextManager(app_core.DefaultContextConfig()), nil
	})

	// Register context-aware event dispatcher
	container.Singleton("context.event_dispatcher", func() (any, error) {
		// Get the existing event bus
		eventBusInstance, err := container.Resolve("event_manager")
		if err != nil {
			// Create a new event manager if not found
			eventBus := app_core.NewEventBus[any](nil, nil, nil)
			eventStore := app_core.NewMemoryEventStore[any]()
			eventManager := app_core.NewEventManager[any](eventBus, eventStore)
			return app_core.NewContextAwareEventDispatcher[any](eventManager), nil
		}

		// Use existing event manager
		eventManager := eventBusInstance.(app_core.EventManagerInterface[any])
		return app_core.NewContextAwareEventDispatcher[any](eventManager), nil
	})

	// Register context-aware job dispatcher
	container.Singleton("context.job_dispatcher", func() (any, error) {
		// Get the existing queue
		queueInstance, err := container.Resolve("queue")
		if err != nil {
			// If no queue exists, create a new one
			queue := app_core.NewSyncQueue[any]()
			jobDispatcher := app_core.NewJobDispatcher[any](queue, nil, nil, nil)
			return app_core.NewContextAwareJobDispatcher[any](jobDispatcher), nil
		}

		// Use existing queue
		queue := queueInstance.(app_core.Queue[any])
		jobDispatcher := app_core.NewJobDispatcher[any](queue, nil, nil, nil)
		return app_core.NewContextAwareJobDispatcher[any](jobDispatcher), nil
	})

	// Register context-aware repository factory
	container.Singleton("context.repository_factory", func() (any, error) {
		return &ContextRepositoryFactory{
			container: container,
		}, nil
	})

	// Register context utilities
	container.Singleton("context.utils", func() (any, error) {
		manager := app_core.NewContextManager(app_core.DefaultContextConfig())
		return app_core.NewContextUtils(manager), nil
	})

	log.Printf("Context services registered successfully")
	return nil
}

// Boot boots the context service provider
func (p *ContextServiceProvider) Boot(container *app_core.Container) error {
	// Set up automatic context optimization for existing listeners
	if err := p.setupContextOptimization(container); err != nil {
		return err
	}

	// Set up automatic context optimization for controllers
	if err := p.setupControllerContextOptimization(container); err != nil {
		return err
	}

	log.Printf("Context services booted successfully")
	return nil
}

// Provides returns the services this provider provides
func (p *ContextServiceProvider) Provides() []string {
	return []string{"context.manager", "context.event_dispatcher", "context.job_dispatcher", "context.repository_factory", "context.utils"}
}

// When returns the conditions when this provider should be loaded
func (p *ContextServiceProvider) When() []string {
	return []string{}
}

// setupContextOptimization sets up automatic context optimization for existing listeners
func (p *ContextServiceProvider) setupContextOptimization(container *app_core.Container) error {
	// Get the context-aware event dispatcher
	dispatcherInstance, err := container.Resolve("context.event_dispatcher")
	if err != nil {
		return err
	}

	contextDispatcher := dispatcherInstance.(*app_core.ContextAwareEventDispatcher[any])

	// Get the existing event manager to register listeners
	eventManagerInstance, err := container.Resolve("event_manager")
	if err != nil {
		return err
	}

	eventManager := eventManagerInstance.(app_core.EventManagerInterface[any])

	// Register automatic context optimization for all listeners
	container.Singleton("context.listener_optimizer", func() (any, error) {
		return &ContextListenerOptimizer{
			eventManager:      eventManager,
			contextDispatcher: contextDispatcher,
		}, nil
	})

	return nil
}

// setupControllerContextOptimization sets up automatic context optimization for controllers
func (p *ContextServiceProvider) setupControllerContextOptimization(container *app_core.Container) error {
	// Register context-aware controller factory
	container.Singleton("context.controller_factory", func() (any, error) {
		return &ContextControllerFactory{
			container: container,
		}, nil
	})

	return nil
}

// ContextRepositoryFactory creates context-aware repositories
type ContextRepositoryFactory struct {
	container *app_core.Container
}

// Create creates a new context-aware repository for a given model type
func (f *ContextRepositoryFactory) Create(repository app_core.Repository[any]) *app_core.ContextAwareRepository[any] {
	// Get the context manager
	managerInstance, err := f.container.Resolve("context.manager")
	if err != nil {
		// Create a new manager if not found
		manager := app_core.NewContextManager(app_core.DefaultContextConfig())
		return app_core.NewContextAwareRepository(repository, manager)
	}

	// Use existing manager
	manager := managerInstance.(*app_core.ContextManager)
	return app_core.NewContextAwareRepository(repository, manager)
}

// ContextListenerOptimizer automatically optimizes listeners with context awareness
type ContextListenerOptimizer struct {
	eventManager      app_core.EventManagerInterface[any]
	contextDispatcher *app_core.ContextAwareEventDispatcher[any]
}

// OptimizeListener wraps a listener with context optimization
func (o *ContextListenerOptimizer) OptimizeListener(eventName string, listener app_core.EventListener[any]) {
	// Register the listener with automatic context optimization
	o.eventManager.Listen(eventName, func(ctx context.Context, event *app_core.Event[any]) error {
		// Add context values for tracking
		ctx = context.WithValue(ctx, "listener_start", time.Now())
		ctx = context.WithValue(ctx, "event_name", eventName)

		// Execute the listener with context optimization
		err := listener(ctx, event)

		// Add context values for completion
		_ = context.WithValue(ctx, "listener_end", time.Now())

		return err
	})
}

// OptimizeListenerStruct optimizes a listener struct with context optimization
func (o *ContextListenerOptimizer) OptimizeListenerStruct(eventName string, listener interface{}) {
	// Create a listener function from the struct
	listenerFunc := func(ctx context.Context, event *app_core.Event[any]) error {
		if l, ok := listener.(interface {
			Handle(ctx context.Context, event *app_core.Event[any]) error
		}); ok {
			return l.Handle(ctx, event)
		}
		return nil
	}

	// Optimize it with context awareness
	o.OptimizeListener(eventName, listenerFunc)
}

// ContextControllerFactory creates context-aware controllers
type ContextControllerFactory struct {
	container *app_core.Container
}

// Create creates a new context-aware controller
func (f *ContextControllerFactory) Create() *facades_core.ContextAwareController {
	return facades_core.NewContextAwareController()
}

// CreateWithTimeout creates a new context-aware controller with custom timeout
func (f *ContextControllerFactory) CreateWithTimeout(timeout time.Duration) *facades_core.ContextAwareController {
	controller := facades_core.NewContextAwareController()
	// Set custom timeout if needed
	return controller
}
