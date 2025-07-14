package providers

import (
	app_core "base_lara_go_project/app/core/go_core"
	"context"
	"database/sql"
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
		// Get required optimization dependencies
		wsp, _ := container.Resolve("work_stealing_pool")
		ca, _ := container.Resolve("custom_allocator")
		pgo, _ := container.Resolve("profile_guided_optimizer")

		// Return canonical event bus with goroutine optimizations
		return app_core.NewEventBus[any](
			wsp.(*app_core.WorkStealingPool[any]),
			ca.(*app_core.CustomAllocator[any]),
			pgo.(*app_core.ProfileGuidedOptimizer[any]),
		), nil
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

			// JobDispatcher already has goroutine optimizations built-in
			return jobDispatcher, nil
		}

		// Use existing queue
		queue := queueInstance.(app_core.Queue[any])
		// Create job dispatcher
		jobDispatcher := app_core.NewJobDispatcher[any](queue, nil, nil, nil)

		// JobDispatcher already has goroutine optimizations built-in
		return jobDispatcher, nil
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

	goroutineDispatcher := dispatcherInstance.(app_core.EventDispatcher[any])

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

// Create creates a new goroutine-optimized repository
func (f *GoroutineRepositoryFactory) Create(db *sql.DB) app_core.Repository[any] {
	wsp, _ := f.container.Resolve("work_stealing_pool")
	ca, _ := f.container.Resolve("custom_allocator")
	pgo, _ := f.container.Resolve("profile_guided_optimizer")
	return app_core.NewRepository[any](db,
		wsp.(*app_core.WorkStealingPool[any]),
		ca.(*app_core.CustomAllocator[any]),
		pgo.(*app_core.ProfileGuidedOptimizer[any]),
	)
}

// GoroutineListenerOptimizer automatically optimizes listeners with goroutines
type GoroutineListenerOptimizer struct {
	eventManager        app_core.EventManagerInterface[any]
	goroutineDispatcher app_core.EventDispatcher[any]
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
