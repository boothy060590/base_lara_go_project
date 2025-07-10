package providers

import (
	app_core "base_lara_go_project/app/core/go_core"
	"context"
	"log"
)

// GoroutineServiceProvider integrates goroutine optimization with the event system
type GoroutineServiceProvider struct {
	BaseServiceProvider
}

// Register registers goroutine-optimized services
func (p *GoroutineServiceProvider) Register(container *app_core.Container) error {
	// Register goroutine manager for general operations
	container.Singleton("goroutine.manager", func() (any, error) {
		return app_core.NewGoroutineManager[any](nil), nil
	})

	// Register goroutine-aware event dispatcher that works with existing events
	container.Singleton("goroutine.event_dispatcher", func() (any, error) {
		// Create a new event bus for goroutine optimization
		eventBus := app_core.NewEventBus[any](nil, nil, nil)

		// Create goroutine-aware dispatcher
		goroutineManager := app_core.NewGoroutineManager[any](nil)
		return app_core.NewGoroutineAwareEventDispatcher[any](eventBus, goroutineManager), nil
	})

	// Register goroutine-aware job dispatcher
	container.Singleton("goroutine.job_dispatcher", func() (any, error) {
		// Get the existing queue
		queueInstance, err := container.Resolve("queue")
		if err != nil {
			// If no queue exists, create a new one
			queue := app_core.NewSyncQueue[any]()
			// Create job dispatcher
			jobDispatcher := app_core.NewJobDispatcher[any](queue, nil, nil, nil)

			// Create goroutine-aware dispatcher
			goroutineManager := app_core.NewGoroutineManager[any](nil)
			return app_core.NewGoroutineAwareJobDispatcher[any](jobDispatcher, goroutineManager), nil
		}

		// Use existing queue
		queue := queueInstance.(app_core.Queue[any])
		// Create job dispatcher
		jobDispatcher := app_core.NewJobDispatcher[any](queue, nil, nil, nil)

		// Create goroutine-aware dispatcher
		goroutineManager := app_core.NewGoroutineManager[any](nil)
		return app_core.NewGoroutineAwareJobDispatcher[any](jobDispatcher, goroutineManager), nil
	})

	// Register goroutine-aware repository factory
	container.Singleton("goroutine.repository_factory", func() (any, error) {
		return &GoroutineRepositoryFactory{
			container: container,
		}, nil
	})

	log.Printf("Goroutine services registered successfully")
	return nil
}

// Boot boots the goroutine service provider
func (p *GoroutineServiceProvider) Boot(container *app_core.Container) error {
	// Set up automatic goroutine optimization for existing listeners
	if err := p.setupGoroutineOptimization(container); err != nil {
		return err
	}

	log.Printf("Goroutine services booted successfully")
	return nil
}

// Provides returns the services this provider provides
func (p *GoroutineServiceProvider) Provides() []string {
	return []string{"goroutine.manager", "goroutine.event_dispatcher", "goroutine.job_dispatcher", "goroutine.repository_factory"}
}

// When returns the conditions when this provider should be loaded
func (p *GoroutineServiceProvider) When() []string {
	return []string{}
}

// setupGoroutineOptimization sets up automatic goroutine optimization for existing listeners
func (p *GoroutineServiceProvider) setupGoroutineOptimization(container *app_core.Container) error {
	// Get the goroutine-aware event dispatcher
	dispatcherInstance, err := container.Resolve("goroutine.event_dispatcher")
	if err != nil {
		return err
	}

	goroutineDispatcher := dispatcherInstance.(*app_core.GoroutineAwareEventDispatcher[any])

	// Get the existing event manager to register listeners
	eventManagerInstance, err := container.Resolve("event_manager")
	if err != nil {
		return err
	}

	eventManager := eventManagerInstance.(app_core.EventManagerInterface[any])

	// Register automatic goroutine optimization for all listeners
	// This will be called when listeners are registered in the ListenerServiceProvider
	container.Singleton("goroutine.listener_optimizer", func() (any, error) {
		return &GoroutineListenerOptimizer{
			eventManager:        eventManager,
			goroutineDispatcher: goroutineDispatcher,
		}, nil
	})

	return nil
}

// GoroutineRepositoryFactory creates goroutine-aware repositories
type GoroutineRepositoryFactory struct {
	container *app_core.Container
}

// Create creates a new goroutine-aware repository for a given model type
func (f *GoroutineRepositoryFactory) Create(repository app_core.Repository[any]) *app_core.GoroutineAwareRepository[any] {
	// Get the goroutine manager
	managerInstance, err := f.container.Resolve("goroutine.manager")
	if err != nil {
		// Create a new manager if not found
		manager := app_core.NewGoroutineManager[any](nil)
		return app_core.NewGoroutineAwareRepository(repository, manager)
	}

	// Use existing manager
	manager := managerInstance.(*app_core.GoroutineManager[any])
	return app_core.NewGoroutineAwareRepository(repository, manager)
}

// GoroutineListenerOptimizer automatically optimizes listeners with goroutines
type GoroutineListenerOptimizer struct {
	eventManager        app_core.EventManagerInterface[any]
	goroutineDispatcher *app_core.GoroutineAwareEventDispatcher[any]
}

// OptimizeListener wraps a listener with goroutine optimization
func (o *GoroutineListenerOptimizer) OptimizeListener(eventName string, listener app_core.EventListener[any]) {
	// Register the listener with automatic goroutine optimization
	o.eventManager.Listen(eventName, func(ctx context.Context, event *app_core.Event[any]) error {
		// Execute the listener in a goroutine
		return o.goroutineDispatcher.DispatchAsync(event)
	})
}

// OptimizeListenerStruct optimizes a listener struct with goroutine optimization
func (o *GoroutineListenerOptimizer) OptimizeListenerStruct(eventName string, listener interface{}) {
	// Create a listener function from the struct
	listenerFunc := func(ctx context.Context, event *app_core.Event[any]) error {
		if l, ok := listener.(interface {
			Handle(ctx context.Context, event *app_core.Event[any]) error
		}); ok {
			return l.Handle(ctx, event)
		}
		return nil
	}

	// Optimize it with goroutines
	o.OptimizeListener(eventName, listenerFunc)
}
