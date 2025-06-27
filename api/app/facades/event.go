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

// Event dispatches an event synchronously
func Event(event interface{}) error {
	e, ok := event.(core.EventInterface)
	if !ok {
		return nil
	}
	return EventDispatcherInstance.DispatchSync(e)
}

// EventAsync dispatches an event asynchronously to the events queue from config
func EventAsync(event interface{}) error {
	e, ok := event.(core.EventInterface)
	if !ok {
		return nil
	}
	return EventDispatcherInstance.DispatchAsync(e)
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
