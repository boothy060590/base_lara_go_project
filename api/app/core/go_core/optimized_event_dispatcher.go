package go_core

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"
)

// OptimizedEventDispatcher combines goroutine and context optimizations
type listenerPair[T any] struct {
	original EventListener[T]
	wrapped  EventListener[T]
}

type OptimizedEventDispatcher[T any] struct {
	eventManager     EventManagerInterface[T]
	goroutineManager *GoroutineManager[T]
	contextConfig    *ContextConfig
	mu               sync.RWMutex
	listenerPairs    []listenerPair[T] // slice of (original, wrapped) pairs
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
		listenerPairs:    make([]listenerPair[T], 0),
	}
}

// Dispatch dispatches an event with context and goroutine optimizations
func (oed *OptimizedEventDispatcher[T]) Dispatch(event *Event[T]) error {
	oed.mu.RLock()
	defer oed.mu.RUnlock()

	// Check if event manager is nil
	if oed.eventManager == nil {
		return fmt.Errorf("event manager is nil")
	}

	// For synchronous dispatch, use the event manager directly
	// This ensures errors are propagated back to the caller
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
	oed.listenerPairs = append(oed.listenerPairs, listenerPair[T]{original: listener, wrapped: optimizedListener})
	return oed.eventManager.Listen(eventName, optimizedListener)
}

// RemoveListener removes an event listener
func (oed *OptimizedEventDispatcher[T]) RemoveListener(eventName string, listener EventListener[T]) error {
	oed.mu.Lock()
	defer oed.mu.Unlock()

	var wrapped EventListener[T]
	idx := -1
	for i, pair := range oed.listenerPairs {
		if reflect.ValueOf(pair.original).Pointer() == reflect.ValueOf(listener).Pointer() {
			wrapped = pair.wrapped
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Errorf("listener not found for removal")
	}
	// Remove the pair from the slice
	oed.listenerPairs = append(oed.listenerPairs[:idx], oed.listenerPairs[idx+1:]...)

	if eventManager, ok := oed.eventManager.(*EventManager[T]); ok {
		return eventManager.dispatcher.RemoveListener(eventName, wrapped)
	}

	return fmt.Errorf("event manager does not support RemoveListener")
}

// Handle handles an event directly
func (oed *OptimizedEventDispatcher[T]) Handle(event *Event[T]) error {
	oed.mu.RLock()
	defer oed.mu.RUnlock()

	// Get the underlying event bus from the event manager
	if eventManager, ok := oed.eventManager.(*EventManager[T]); ok {
		return eventManager.dispatcher.Handle(event)
	}

	return fmt.Errorf("event manager does not support Handle")
}

// HasListeners checks if there are listeners for an event
func (oed *OptimizedEventDispatcher[T]) HasListeners(eventName string) bool {
	oed.mu.RLock()
	defer oed.mu.RUnlock()

	return oed.eventManager.HasListeners(eventName)
}

// GetListenerCount returns the number of listeners for an event
func (oed *OptimizedEventDispatcher[T]) GetListenerCount(eventName string) int {
	oed.mu.RLock()
	defer oed.mu.RUnlock()

	return oed.eventManager.GetListenerCount(eventName)
}

// WithContext creates a context-aware dispatcher
func (oed *OptimizedEventDispatcher[T]) WithContext(ctx context.Context) EventDispatcher[T] {
	oed.mu.RLock()
	defer oed.mu.RUnlock()

	// Create a new dispatcher with the given context
	return &OptimizedEventDispatcher[T]{
		eventManager:     oed.eventManager,
		goroutineManager: oed.goroutineManager,
		contextConfig:    oed.contextConfig,
	}
}

// GetPerformanceStats returns performance statistics
func (oed *OptimizedEventDispatcher[T]) GetPerformanceStats() map[string]interface{} {
	oed.mu.RLock()
	defer oed.mu.RUnlock()

	stats := make(map[string]interface{})

	if oed.goroutineManager != nil {
		pool := oed.goroutineManager.GetWorkerPool()
		stats["active_workers"] = pool.GetActiveWorkerCount()
		stats["total_workers"] = len(pool.workers)
		stats["queue_size"] = pool.QueueLength()
	}

	return stats
}

// GetOptimizationStats returns optimization statistics
func (oed *OptimizedEventDispatcher[T]) GetOptimizationStats() map[string]interface{} {
	oed.mu.RLock()
	defer oed.mu.RUnlock()

	stats := make(map[string]interface{})

	if oed.contextConfig != nil {
		stats["context_timeout_enabled"] = oed.contextConfig.EnableDeadline
		stats["default_timeout"] = oed.contextConfig.DefaultTimeout
		stats["context_optimization"] = oed.contextConfig.EnableCancellation
	}

	if oed.goroutineManager != nil {
		stats["goroutine_optimization"] = true
		stats["worker_pool_size"] = len(oed.goroutineManager.GetWorkerPool().workers)
	} else {
		stats["goroutine_optimization"] = false
	}

	return stats
}

// dispatchWithGoroutine dispatches using goroutine optimization
func (oed *OptimizedEventDispatcher[T]) dispatchWithGoroutine(event *Event[T]) error {
	if oed.goroutineManager == nil {
		if oed.eventManager == nil {
			return fmt.Errorf("event manager is nil")
		}
		return oed.eventManager.Dispatch(event)
	}

	// Get timeout from context config or use default
	timeout := 30 * time.Second
	if oed.contextConfig != nil {
		timeout = oed.contextConfig.DefaultTimeout
	}

	// Create a job for the goroutine pool
	job := GoroutineJob[T]{
		Job: Job[T]{
			ID:   event.ID,
			Data: event.Data,
		},
		Timeout: timeout,
		Handler: func(ctx context.Context, job *GoroutineJob[T]) error {
			if oed.eventManager == nil {
				return fmt.Errorf("event manager is nil")
			}
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
		if oed.contextConfig != nil && oed.contextConfig.EnableDeadline {
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
