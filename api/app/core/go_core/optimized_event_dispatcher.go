package go_core

import (
	"context"
	"sync"
)

// OptimizedEventDispatcher combines goroutine and context optimizations
type OptimizedEventDispatcher[T any] struct {
	eventManager     EventManagerInterface[T]
	goroutineManager *GoroutineManager[T]
	contextConfig    *ContextConfig
	mu               sync.RWMutex
}

// NewOptimizedEventDispatcher creates a unified optimized event dispatcher
func NewOptimizedEventDispatcher[T any](
	eventManager EventManagerInterface[T],
	goroutineManager *GoroutineManager[T],
	contextConfig *ContextConfig,
) *OptimizedEventDispatcher[T] {
	return &OptimizedEventDispatcher[T]{
		eventManager:     eventManager,
		goroutineManager: goroutineManager,
		contextConfig:    contextConfig,
	}
}

// Dispatch dispatches an event with context and goroutine optimizations
func (oed *OptimizedEventDispatcher[T]) Dispatch(event *Event[T]) error {
	oed.mu.RLock()
	defer oed.mu.RUnlock()

	// Use goroutine optimization for async dispatch
	if oed.goroutineManager != nil {
		return oed.dispatchWithGoroutine(event)
	}

	// Fallback to direct dispatch
	return oed.eventManager.Dispatch(event)
}

// DispatchAsync dispatches an event asynchronously with optimizations
func (oed *OptimizedEventDispatcher[T]) DispatchAsync(event *Event[T]) error {
	oed.mu.RLock()
	defer oed.mu.RUnlock()

	// Use goroutine manager for async dispatch
	if oed.goroutineManager != nil {
		return oed.dispatchWithGoroutine(event)
	}

	// Fallback to direct async dispatch
	return oed.eventManager.DispatchAsync(event)
}

// Listen registers an event listener with optimizations
func (oed *OptimizedEventDispatcher[T]) Listen(eventName string, listener EventListener[T]) error {
	oed.mu.Lock()
	defer oed.mu.Unlock()

	// Wrap listener with context and goroutine optimizations
	optimizedListener := oed.createOptimizedListener(listener)
	return oed.eventManager.Listen(eventName, optimizedListener)
}

// dispatchWithGoroutine dispatches using goroutine optimization
func (oed *OptimizedEventDispatcher[T]) dispatchWithGoroutine(event *Event[T]) error {
	// Create a job for the goroutine pool
	job := GoroutineJob[T]{
		Job: Job[T]{
			ID:   event.ID,
			Data: event.Data,
		},
		Timeout: oed.contextConfig.DefaultTimeout,
		Handler: func(ctx context.Context, job *GoroutineJob[T]) error {
			return oed.eventManager.Dispatch(event)
		},
	}

	// Submit to worker pool
	return oed.goroutineManager.GetWorkerPool().Submit(job)
}

// createOptimizedListener creates a listener with optimizations
func (oed *OptimizedEventDispatcher[T]) createOptimizedListener(listener EventListener[T]) EventListener[T] {
	return func(ctx context.Context, event *Event[T]) error {
		// Apply context timeout if not already set
		if oed.contextConfig.EnableDeadline {
			if _, ok := ctx.Deadline(); !ok {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, oed.contextConfig.DefaultTimeout)
				defer cancel()
			}
		}

		// Execute listener with optimizations
		return listener(ctx, event)
	}
}

// GetEventManager returns the underlying event manager
func (oed *OptimizedEventDispatcher[T]) GetEventManager() EventManagerInterface[T] {
	return oed.eventManager
}

// GetGoroutineManager returns the goroutine manager
func (oed *OptimizedEventDispatcher[T]) GetGoroutineManager() *GoroutineManager[T] {
	return oed.goroutineManager
}

// GetContextConfig returns the context configuration
func (oed *OptimizedEventDispatcher[T]) GetContextConfig() *ContextConfig {
	return oed.contextConfig
}

// Close closes the dispatcher and cleans up resources
func (oed *OptimizedEventDispatcher[T]) Close() error {
	oed.mu.Lock()
	defer oed.mu.Unlock()

	// Shutdown goroutine manager if available
	if oed.goroutineManager != nil {
		oed.goroutineManager.GetWorkerPool().Shutdown()
	}

	return nil
}
