package listeners

import (
	app_core "base_lara_go_project/app/core/go_core"
	"context"
)

// BaseListener provides Laravel-style base functionality for all listeners
type BaseListener[T any] struct {
	// Common fields can be added here if needed
}

// Handle is the base implementation that should be overridden by specific listeners
func (l *BaseListener[T]) Handle(ctx context.Context, event *app_core.Event[T]) error {
	// Base implementation - should be overridden
	return nil
}

// ShouldQueue determines if the listener should be queued
func (l *BaseListener[T]) ShouldQueue() bool {
	return false
}

// GetQueueName returns the queue name for this listener
func (l *BaseListener[T]) GetQueueName() string {
	return "default"
}

// GetMaxAttempts returns the maximum number of attempts for this listener
func (l *BaseListener[T]) GetMaxAttempts() int {
	return 3
}

// GetRetryDelay returns the delay between retry attempts
func (l *BaseListener[T]) GetRetryDelay() int {
	return 60 // seconds
}

// ListenerFactory is a function type that creates listeners from events
type ListenerFactory[T any] func(event *app_core.Event[T]) app_core.EventListener[T]

// CreateListener creates a listener function from a listener struct
func CreateListener[T any](listener interface{}) app_core.EventListener[T] {
	return func(ctx context.Context, event *app_core.Event[T]) error {
		if l, ok := listener.(interface {
			Handle(ctx context.Context, event *app_core.Event[T]) error
		}); ok {
			return l.Handle(ctx, event)
		}
		return nil
	}
}
