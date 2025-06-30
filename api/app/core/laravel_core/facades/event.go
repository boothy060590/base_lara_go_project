package facades_core

// EventInterface defines the interface for events
type EventInterface interface {
	GetName() string
	GetData() interface{}
}

// EventDispatcher defines the interface for dispatching events
type EventDispatcher interface {
	DispatchAsync(event EventInterface) error
	DispatchSync(event EventInterface) error
}

// Global event dispatcher instance
var EventDispatcherInstance EventDispatcher

// SetEventDispatcher sets the global event dispatcher
func SetEventDispatcher(dispatcher EventDispatcher) {
	EventDispatcherInstance = dispatcher
}

// Event dispatches an event synchronously
func Event(event interface{}) error {
	e, ok := event.(EventInterface)
	if !ok {
		return nil
	}
	return EventDispatcherInstance.DispatchSync(e)
}

// EventAsync dispatches an event asynchronously to the events queue from config
func EventAsync(event interface{}) error {
	e, ok := event.(EventInterface)
	if !ok {
		return nil
	}
	return EventDispatcherInstance.DispatchAsync(e)
}

// EventWithError dispatches an event asynchronously and returns any error
func EventWithError(event EventInterface) error {
	return EventDispatcherInstance.DispatchAsync(event)
}

// DispatchEvent dispatches an event asynchronously (via queue)
func DispatchEvent(event EventInterface) error {
	return EventDispatcherInstance.DispatchAsync(event)
}

// DispatchEventSync dispatches an event synchronously (immediate)
func DispatchEventSync(event EventInterface) error {
	return EventDispatcherInstance.DispatchSync(event)
}

// EventFacade provides a facade for event operations
type EventFacade struct{}

// Dispatch dispatches an event asynchronously
func (e *EventFacade) Dispatch(event EventInterface) error {
	// TODO: Implement using go_core event system
	return nil
}

// DispatchSync dispatches an event synchronously
func (e *EventFacade) DispatchSync(event EventInterface) error {
	// TODO: Implement using go_core event system
	return nil
}
