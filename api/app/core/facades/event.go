package facades_core

import (
	app_core "base_lara_go_project/app/core/app"
	events_core "base_lara_go_project/app/core/events"
)

// EventDispatcher defines the interface for dispatching events
type EventDispatcher interface {
	DispatchAsync(event app_core.EventInterface) error
	DispatchSync(event app_core.EventInterface) error
}

// Global event dispatcher instance
var EventDispatcherInstance EventDispatcher

// SetEventDispatcher sets the global event dispatcher
func SetEventDispatcher(dispatcher EventDispatcher) {
	EventDispatcherInstance = dispatcher
}

// Event dispatches an event synchronously
func Event(event interface{}) error {
	e, ok := event.(app_core.EventInterface)
	if !ok {
		return nil
	}
	return EventDispatcherInstance.DispatchSync(e)
}

// EventAsync dispatches an event asynchronously to the events queue from config
func EventAsync(event interface{}) error {
	e, ok := event.(app_core.EventInterface)
	if !ok {
		return nil
	}
	return EventDispatcherInstance.DispatchAsync(e)
}

// EventWithError dispatches an event asynchronously and returns any error
func EventWithError(event app_core.EventInterface) error {
	return EventDispatcherInstance.DispatchAsync(event)
}

// DispatchEvent dispatches an event asynchronously (via queue)
func DispatchEvent(event app_core.EventInterface) error {
	return EventDispatcherInstance.DispatchAsync(event)
}

// DispatchEventSync dispatches an event synchronously (immediate)
func DispatchEventSync(event app_core.EventInterface) error {
	return EventDispatcherInstance.DispatchSync(event)
}

// EventFacade provides a facade for event operations
type EventFacade struct{}

// Dispatch dispatches an event asynchronously
func (e *EventFacade) Dispatch(event app_core.EventInterface) error {
	return events_core.DispatchEventAsync(event)
}

// DispatchSync dispatches an event synchronously
func (e *EventFacade) DispatchSync(event app_core.EventInterface) error {
	return events_core.DispatchEventSync(event)
}
