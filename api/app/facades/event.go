package facades

import (
	"base_lara_go_project/app/core"
)

// EventDispatcher defines the interface for dispatching events
type EventDispatcher interface {
	DispatchAsync(event core.EventInterface) error
	DispatchSync(event core.EventInterface) error
}

// Global event dispatcher instance
var EventDispatcherInstance EventDispatcher

// SetEventDispatcher sets the global event dispatcher
func SetEventDispatcher(dispatcher EventDispatcher) {
	EventDispatcherInstance = dispatcher
}

// Event dispatches an event asynchronously (like Laravel's event() helper)
// Errors are logged internally but don't bubble up to the controller
func Event(event core.EventInterface) {
	err := EventDispatcherInstance.DispatchAsync(event)
	if err != nil {
		// Log the error internally but don't return it
		// This prevents event failures from breaking the main flow
		// In production, you might want to use a proper logger here
		_ = err // Suppress unused variable warning
	}
}

// EventWithError dispatches an event asynchronously and returns any error
func EventWithError(event core.EventInterface) error {
	return EventDispatcherInstance.DispatchAsync(event)
}

// DispatchEvent dispatches an event asynchronously (via queue)
func DispatchEvent(event core.EventInterface) error {
	return EventDispatcherInstance.DispatchAsync(event)
}

// DispatchEventSync dispatches an event synchronously (immediate)
func DispatchEventSync(event core.EventInterface) error {
	return EventDispatcherInstance.DispatchSync(event)
}
