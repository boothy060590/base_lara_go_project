package core

// EventListenerRegistry holds all registered event listeners
type EventListenerRegistry struct {
	listeners map[string][]func(EventInterface) ListenerInterface
}

// Global registry instance
var GlobalRegistry *EventListenerRegistry

// InitializeRegistry initializes the global registry
func InitializeRegistry() {
	GlobalRegistry = &EventListenerRegistry{
		listeners: make(map[string][]func(EventInterface) ListenerInterface),
	}
}

// RegisterListener registers a listener for an event
func (r *EventListenerRegistry) RegisterListener(eventName string, handlerFactory func(EventInterface) ListenerInterface) {
	r.listeners[eventName] = append(r.listeners[eventName], handlerFactory)
}

// GetListeners returns all listeners for an event
func (r *EventListenerRegistry) GetListeners(eventName string) []func(EventInterface) ListenerInterface {
	return r.listeners[eventName]
}
