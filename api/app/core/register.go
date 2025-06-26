package core

// RegisterEvent registers an event listener
func RegisterEvent(eventName string, handlerFactory func(EventInterface) ListenerInterface) {
	GlobalRegistry.RegisterListener(eventName, handlerFactory)
}

// RegisterUserCreatedEvent is a convenience function for registering UserCreated listeners
func RegisterUserCreatedEvent(handlerFactory func(EventInterface) ListenerInterface) {
	RegisterEvent("UserCreated", handlerFactory)
}
