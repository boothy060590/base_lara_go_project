package facades_core

import (
	"sync"
	"time"

	app_core "base_lara_go_project/app/core/go_core"
)

// ============================================================================
// PERFORMANCE FACADE
// ============================================================================

// PerformanceFacade provides Laravel-style access to performance features
type PerformanceFacade struct {
	container *app_core.Container
	mu        sync.RWMutex
}

var (
	performanceInstance *PerformanceFacade
	performanceOnce     sync.Once
)

// Performance returns the singleton performance facade instance
func Performance() *PerformanceFacade {
	performanceOnce.Do(func() {
		performanceInstance = &PerformanceFacade{}
	})
	return performanceInstance
}

// SetContainer sets the container for dependency injection
func (pf *PerformanceFacade) SetContainer(container *app_core.Container) {
	pf.mu.Lock()
	defer pf.mu.Unlock()
	pf.container = container
}

// GetContainer returns the current container
func (pf *PerformanceFacade) GetContainer() *app_core.Container {
	pf.mu.RLock()
	defer pf.mu.RUnlock()
	return pf.container
}

// Track tracks performance of a function
func (pf *PerformanceFacade) Track(name string, fn func() error) error {
	container := pf.GetContainer()
	if container != nil {
		return container.TrackPerformance(name, fn)
	}

	// Fallback to direct performance facade if no container
	perf := app_core.NewPerformanceFacade()
	return perf.Track(name, fn)
}

// Optimize optimizes an object
func (pf *PerformanceFacade) Optimize(obj interface{}) error {
	container := pf.GetContainer()
	if container != nil {
		return container.OptimizeObject(obj)
	}

	// Fallback to direct performance facade if no container
	perf := app_core.NewPerformanceFacade()
	return perf.Optimize(obj)
}

// GetStats returns performance statistics
func (pf *PerformanceFacade) GetStats() map[string]interface{} {
	container := pf.GetContainer()
	if container != nil {
		return container.GetPerformanceStats()
	}

	// Fallback to direct performance facade if no container
	perf := app_core.NewPerformanceFacade()
	return perf.GetStats()
}

// CreatePipeline creates a new processing pipeline
func (pf *PerformanceFacade) CreatePipeline() *app_core.Pipeline[any] {
	perf := app_core.NewPerformanceFacade()
	return perf.CreatePipeline()
}

// ============================================================================
// PERFORMANCE UTILITY METHODS
// ============================================================================

// Time measures the execution time of a function
func (pf *PerformanceFacade) Time(name string, fn func() error) (time.Duration, error) {
	start := time.Now()
	err := pf.Track(name, fn)
	duration := time.Since(start)
	return duration, err
}

// Benchmark benchmarks a function multiple times
func (pf *PerformanceFacade) Benchmark(name string, iterations int, fn func() error) (map[string]interface{}, error) {
	var totalDuration time.Duration
	var minDuration time.Duration
	var maxDuration time.Duration
	var errors int

	for i := 0; i < iterations; i++ {
		duration, err := pf.Time(name, fn)
		totalDuration += duration

		if err != nil {
			errors++
		}

		if i == 0 || duration < minDuration {
			minDuration = duration
		}
		if i == 0 || duration > maxDuration {
			maxDuration = duration
		}
	}

	avgDuration := totalDuration / time.Duration(iterations)
	errorRate := float64(errors) / float64(iterations) * 100

	return map[string]interface{}{
		"iterations":     iterations,
		"total_duration": totalDuration,
		"avg_duration":   avgDuration,
		"min_duration":   minDuration,
		"max_duration":   maxDuration,
		"errors":         errors,
		"error_rate":     errorRate,
	}, nil
}

// MemoryUsage returns current memory usage statistics
func (pf *PerformanceFacade) MemoryUsage() map[string]interface{} {
	stats := pf.GetStats()
	if goroutineStats, ok := stats["num_goroutines"]; ok {
		return map[string]interface{}{
			"goroutines": goroutineStats,
			"memory":     stats,
		}
	}
	return stats
}

// ============================================================================
// LARAVEL-STYLE PERFORMANCE HELPERS
// ============================================================================

// Profile profiles a function and returns detailed metrics
func (pf *PerformanceFacade) Profile(name string, fn func() error) map[string]interface{} {
	duration, err := pf.Time(name, fn)

	metrics := map[string]interface{}{
		"name":     name,
		"duration": duration,
		"success":  err == nil,
	}

	if err != nil {
		metrics["error"] = err.Error()
	}

	return metrics
}

// Monitor monitors a function and logs performance data
func (pf *PerformanceFacade) Monitor(name string, fn func() error) error {
	metrics := pf.Profile(name, fn)

	// Log performance data (you can integrate with your logging system)
	// log.Printf("Performance: %s took %v", name, metrics["duration"])

	if err, ok := metrics["error"]; ok {
		return err.(error)
	}
	return nil
}
