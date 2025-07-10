package providers

import (
	app_core "base_lara_go_project/app/core/go_core"
	"base_lara_go_project/config"
	"log"
	"time"
)

// OptimizationServiceProvider integrates advanced optimizations with the framework
type OptimizationServiceProvider struct {
	BaseServiceProvider
}

// Register registers optimization services with the container
func (p *OptimizationServiceProvider) Register(container *app_core.Container) error {
	// Register work stealing pool
	if err := p.registerWorkStealingPool(container); err != nil {
		return err
	}

	// Register profile-guided optimizer
	if err := p.registerProfileGuidedOptimizer(container); err != nil {
		return err
	}

	// Register custom allocators
	if err := p.registerCustomAllocators(container); err != nil {
		return err
	}

	// Register optimization facade
	if err := p.registerOptimizationFacade(container); err != nil {
		return err
	}

	log.Printf("Optimization services registered successfully")
	return nil
}

// Boot boots the optimization service provider
func (p *OptimizationServiceProvider) Boot(container *app_core.Container) error {
	// Set up automatic optimization integration with core services
	if err := p.setupAutomaticOptimization(container); err != nil {
		return err
	}

	log.Printf("Optimization services booted successfully")
	return nil
}

// Provides returns the services this provider provides
func (p *OptimizationServiceProvider) Provides() []string {
	return []string{
		"optimization.work_stealing",
		"optimization.profile_guided",
		"optimization.custom_allocator",
		"optimization.facade",
	}
}

// When returns the conditions when this provider should be loaded
func (p *OptimizationServiceProvider) When() []string {
	return []string{}
}

// registerWorkStealingPool registers the work stealing pool
func (p *OptimizationServiceProvider) registerWorkStealingPool(container *app_core.Container) error {
	workStealingConfig := config.WorkStealingConfig()

	// Create work stealing pool configuration
	config := &app_core.WorkStealingConfig{
		NumWorkers:      workStealingConfig["workers"].(map[string]interface{})["num_workers"].(int),
		QueueSize:       workStealingConfig["workers"].(map[string]interface{})["queue_size"].(int),
		StealThreshold:  workStealingConfig["workers"].(map[string]interface{})["steal_threshold"].(int),
		StealBatchSize:  workStealingConfig["workers"].(map[string]interface{})["steal_batch_size"].(int),
		IdleTimeout:     time.Duration(workStealingConfig["workers"].(map[string]interface{})["idle_timeout"].(int)) * time.Millisecond,
		EnableMetrics:   workStealingConfig["optimizations"].(map[string]interface{})["enable_metrics"].(bool),
		EnableProfiling: workStealingConfig["optimizations"].(map[string]interface{})["enable_profiling"].(bool),
	}

	// Create work stealing pool
	pool := app_core.NewWorkStealingPool[app_core.WorkItem[any]](config)

	// Register as singleton
	container.Singleton("optimization.work_stealing", func() (any, error) {
		return pool, nil
	})

	// Register typed versions for common types
	container.Singleton("optimization.work_stealing.user", func() (any, error) {
		return app_core.NewWorkStealingPool[app_core.WorkItem[any]](config), nil
	})

	container.Singleton("optimization.work_stealing.job", func() (any, error) {
		return app_core.NewWorkStealingPool[app_core.WorkItem[any]](config), nil
	})

	container.Singleton("optimization.work_stealing.event", func() (any, error) {
		return app_core.NewWorkStealingPool[app_core.WorkItem[any]](config), nil
	})

	return nil
}

// registerProfileGuidedOptimizer registers the profile-guided optimizer
func (p *OptimizationServiceProvider) registerProfileGuidedOptimizer(container *app_core.Container) error {
	profileGuidedConfig := config.ProfileGuidedConfig()

	// Create profile-guided optimizer configuration
	config := &app_core.ProfileGuidedConfig{
		Enabled:              profileGuidedConfig["enabled"].(bool),
		SamplingInterval:     time.Duration(profileGuidedConfig["sampling"].(map[string]interface{})["interval"].(int)) * time.Second,
		OptimizationInterval: time.Duration(profileGuidedConfig["optimization"].(map[string]interface{})["interval"].(int)) * time.Second,
		MinSamples:           profileGuidedConfig["sampling"].(map[string]interface{})["min_samples"].(int),
		MaxOptimizations:     profileGuidedConfig["optimization"].(map[string]interface{})["max_optimizations"].(int),
		EnableAutoTuning:     profileGuidedConfig["optimization"].(map[string]interface{})["auto_tuning"].(bool),
		EnableMetrics:        profileGuidedConfig["optimizations"].(map[string]interface{})["enable_metrics"].(bool),
	}

	// Create profile-guided optimizer
	optimizer := app_core.NewProfileGuidedOptimizer[any](config)

	// Register as singleton
	container.Singleton("optimization.profile_guided", func() (any, error) {
		return optimizer, nil
	})

	// Register typed versions for common types
	container.Singleton("optimization.profile_guided.user", func() (any, error) {
		return app_core.NewProfileGuidedOptimizer[any](config), nil
	})

	container.Singleton("optimization.profile_guided.job", func() (any, error) {
		return app_core.NewProfileGuidedOptimizer[any](config), nil
	})

	container.Singleton("optimization.profile_guided.event", func() (any, error) {
		return app_core.NewProfileGuidedOptimizer[any](config), nil
	})

	return nil
}

// registerCustomAllocators registers the custom allocators
func (p *OptimizationServiceProvider) registerCustomAllocators(container *app_core.Container) error {
	customAllocatorsConfig := config.CustomAllocatorsConfig()

	// Create custom allocator configuration
	config := &app_core.CustomAllocatorConfig{
		Enabled:            customAllocatorsConfig["enabled"].(bool),
		PoolSize:           customAllocatorsConfig["pools"].(map[string]interface{})["size"].(int),
		MaxObjectSize:      customAllocatorsConfig["pools"].(map[string]interface{})["max_object_size"].(int),
		CleanupInterval:    time.Duration(customAllocatorsConfig["pools"].(map[string]interface{})["cleanup_interval"].(int)) * time.Second,
		EnableMetrics:      customAllocatorsConfig["optimizations"].(map[string]interface{})["enable_metrics"].(bool),
		EnableProfiling:    customAllocatorsConfig["optimizations"].(map[string]interface{})["enable_profiling"].(bool),
		AllocationStrategy: customAllocatorsConfig["strategies"].(map[string]interface{})["default"].(string),
	}

	// Create custom allocator
	allocator := app_core.NewCustomAllocator[any](config)

	// Register as singleton
	container.Singleton("optimization.custom_allocator", func() (any, error) {
		return allocator, nil
	})

	// Register typed versions for common types
	container.Singleton("optimization.custom_allocator.user", func() (any, error) {
		return app_core.NewCustomAllocator[any](config), nil
	})

	container.Singleton("optimization.custom_allocator.job", func() (any, error) {
		return app_core.NewCustomAllocator[any](config), nil
	})

	container.Singleton("optimization.custom_allocator.event", func() (any, error) {
		return app_core.NewCustomAllocator[any](config), nil
	})

	return nil
}

// registerOptimizationFacade registers the optimization facade
func (p *OptimizationServiceProvider) registerOptimizationFacade(container *app_core.Container) error {
	// Register optimization facade
	container.Singleton("optimization.facade", func() (any, error) {
		return &OptimizationFacade{
			container: container,
		}, nil
	})

	return nil
}

// setupAutomaticOptimization sets up automatic optimization integration
func (p *OptimizationServiceProvider) setupAutomaticOptimization(container *app_core.Container) error {
	// Set up automatic work stealing integration with existing services
	if err := p.setupWorkStealingIntegration(container); err != nil {
		return err
	}

	// Set up automatic profile-guided optimization integration
	if err := p.setupProfileGuidedIntegration(container); err != nil {
		return err
	}

	// Set up automatic custom allocator integration
	if err := p.setupCustomAllocatorIntegration(container); err != nil {
		return err
	}

	return nil
}

// setupWorkStealingIntegration sets up automatic work stealing integration
func (p *OptimizationServiceProvider) setupWorkStealingIntegration(container *app_core.Container) error {
	// Get work stealing pool
	poolInstance, err := container.Resolve("optimization.work_stealing")
	if err != nil {
		return err
	}

	pool := poolInstance.(*app_core.WorkStealingPool[app_core.WorkItem[any]])

	// Register work stealing integration for repositories
	container.Singleton("optimization.repository_work_stealing", func() (any, error) {
		return &RepositoryWorkStealingIntegration{
			pool: pool,
		}, nil
	})

	// Register work stealing integration for queues
	container.Singleton("optimization.queue_work_stealing", func() (any, error) {
		return &QueueWorkStealingIntegration{
			pool: pool,
		}, nil
	})

	// Register work stealing integration for events
	container.Singleton("optimization.event_work_stealing", func() (any, error) {
		return &EventWorkStealingIntegration{
			pool: pool,
		}, nil
	})

	return nil
}

// setupProfileGuidedIntegration sets up automatic profile-guided optimization integration
func (p *OptimizationServiceProvider) setupProfileGuidedIntegration(container *app_core.Container) error {
	// Get profile-guided optimizer
	optimizerInstance, err := container.Resolve("optimization.profile_guided")
	if err != nil {
		return err
	}

	optimizer := optimizerInstance.(*app_core.ProfileGuidedOptimizer[any])

	// Register profile-guided integration for repositories
	container.Singleton("optimization.repository_profile_guided", func() (any, error) {
		return &RepositoryProfileGuidedIntegration{
			optimizer: optimizer,
		}, nil
	})

	// Register profile-guided integration for queues
	container.Singleton("optimization.queue_profile_guided", func() (any, error) {
		return &QueueProfileGuidedIntegration{
			optimizer: optimizer,
		}, nil
	})

	// Register profile-guided integration for events
	container.Singleton("optimization.event_profile_guided", func() (any, error) {
		return &EventProfileGuidedIntegration{
			optimizer: optimizer,
		}, nil
	})

	return nil
}

// setupCustomAllocatorIntegration sets up automatic custom allocator integration
func (p *OptimizationServiceProvider) setupCustomAllocatorIntegration(container *app_core.Container) error {
	// Get custom allocator
	allocatorInstance, err := container.Resolve("optimization.custom_allocator")
	if err != nil {
		return err
	}

	allocator := allocatorInstance.(*app_core.CustomAllocator[any])

	// Register custom allocator integration for repositories
	container.Singleton("optimization.repository_custom_allocator", func() (any, error) {
		return &RepositoryCustomAllocatorIntegration{
			allocator: allocator,
		}, nil
	})

	// Register custom allocator integration for queues
	container.Singleton("optimization.queue_custom_allocator", func() (any, error) {
		return &QueueCustomAllocatorIntegration{
			allocator: allocator,
		}, nil
	})

	// Register custom allocator integration for events
	container.Singleton("optimization.event_custom_allocator", func() (any, error) {
		return &EventCustomAllocatorIntegration{
			allocator: allocator,
		}, nil
	})

	return nil
}

// OptimizationFacade provides a facade for optimization services
type OptimizationFacade struct {
	container *app_core.Container
}

// WorkStealing returns the work stealing pool
func (of *OptimizationFacade) WorkStealing() *app_core.WorkStealingPool[app_core.WorkItem[any]] {
	instance, err := of.container.Resolve("optimization.work_stealing")
	if err != nil {
		return nil
	}
	return instance.(*app_core.WorkStealingPool[app_core.WorkItem[any]])
}

// ProfileGuided returns the profile-guided optimizer
func (of *OptimizationFacade) ProfileGuided() *app_core.ProfileGuidedOptimizer[any] {
	instance, err := of.container.Resolve("optimization.profile_guided")
	if err != nil {
		return nil
	}
	return instance.(*app_core.ProfileGuidedOptimizer[any])
}

// CustomAllocator returns the custom allocator
func (of *OptimizationFacade) CustomAllocator() *app_core.CustomAllocator[any] {
	instance, err := of.container.Resolve("optimization.custom_allocator")
	if err != nil {
		return nil
	}
	return instance.(*app_core.CustomAllocator[any])
}

// Integration types for automatic optimization

type RepositoryWorkStealingIntegration struct {
	pool *app_core.WorkStealingPool[app_core.WorkItem[any]]
}

type QueueWorkStealingIntegration struct {
	pool *app_core.WorkStealingPool[app_core.WorkItem[any]]
}

type EventWorkStealingIntegration struct {
	pool *app_core.WorkStealingPool[app_core.WorkItem[any]]
}

type RepositoryProfileGuidedIntegration struct {
	optimizer *app_core.ProfileGuidedOptimizer[any]
}

type QueueProfileGuidedIntegration struct {
	optimizer *app_core.ProfileGuidedOptimizer[any]
}

type EventProfileGuidedIntegration struct {
	optimizer *app_core.ProfileGuidedOptimizer[any]
}

type RepositoryCustomAllocatorIntegration struct {
	allocator *app_core.CustomAllocator[any]
}

type QueueCustomAllocatorIntegration struct {
	allocator *app_core.CustomAllocator[any]
}

type EventCustomAllocatorIntegration struct {
	allocator *app_core.CustomAllocator[any]
}
