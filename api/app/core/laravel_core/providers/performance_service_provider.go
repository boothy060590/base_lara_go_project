package providers

import (
	app_core "base_lara_go_project/app/core/go_core"
	facades_core "base_lara_go_project/app/core/laravel_core/facades"
)

// PerformanceServiceProvider integrates performance features into the framework
type PerformanceServiceProvider struct {
	BaseServiceProvider
}

// Register registers performance services with the container
func (p *PerformanceServiceProvider) Register(container *app_core.Container) error {
	// Register the performance facade with the container
	container.Singleton("performance", func() (any, error) {
		perf := app_core.NewPerformanceFacade()
		return perf, nil
	})

	// Register atomic counter for performance metrics
	container.Singleton("performance.counter", func() (any, error) {
		return app_core.NewAtomicCounter(), nil
	})

	// Register object pool for performance optimization
	container.Singleton("performance.pool", func() (any, error) {
		pool := app_core.NewObjectPool[any](100,
			func() any { return nil },
			func(obj any) any { return nil },
		)
		return pool, nil
	})

	// Register goroutine optimizer
	container.Singleton("performance.goroutine_optimizer", func() (any, error) {
		return app_core.NewGoroutineOptimizer(), nil
	})

	// Register dynamic optimizer
	container.Singleton("performance.dynamic_optimizer", func() (any, error) {
		return app_core.NewDynamicOptimizer(), nil
	})

	// Register optimization engine
	container.Singleton("performance.optimization_engine", func() (any, error) {
		return app_core.NewOptimizationEngine(), nil
	})

	// Set the container on the performance facade
	perfFacade := facades_core.Performance()
	perfFacade.SetContainer(container)

	return nil
}

// Boot initializes performance features
func (p *PerformanceServiceProvider) Boot(container *app_core.Container) error {
	// Get the goroutine optimizer and optimize GOMAXPROCS
	if optimizer, err := container.Resolve("performance.goroutine_optimizer"); err == nil {
		if goOptimizer, ok := optimizer.(*app_core.GoroutineOptimizer); ok {
			goOptimizer.OptimizeGOMAXPROCS()
		}
	}

	// Get the performance facade and initialize stats
	perf := facades_core.Performance()
	_ = perf.GetStats() // Initialize performance tracking

	// Log initial performance stats (you can integrate with your logging system)
	// log.Printf("Performance system initialized")

	return nil
}
