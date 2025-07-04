package core

// RegisterEvent registers an event listener
func RegisterEvent(eventName string, handlerFactory func(EventInterface) ListenerInterface) {
	GlobalRegistry.RegisterListener(eventName, handlerFactory)
}
