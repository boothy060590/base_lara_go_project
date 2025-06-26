package core

// EventDispatcher handles event dispatching
type EventDispatcher struct{}

// NewEventDispatcher creates a new event dispatcher
func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{}
}

// Register registers an event handler
func (d *EventDispatcher) Register(eventName string, handlerFactory func(EventInterface) ListenerInterface) {
	GlobalRegistry.RegisterListener(eventName, handlerFactory)
}

// DispatchSync dispatches an event to all its handlers (SYNCHRONOUS - immediate)
func (d *EventDispatcher) DispatchSync(event EventInterface) error {
	eventName := event.GetEventName()

	handlers := GlobalRegistry.GetListeners(eventName)
	for _, handlerFactory := range handlers {
		handler := handlerFactory(event)
		if err := handler.Handle(GetMailService()); err != nil {
			return err
		}
	}
	return nil
}

// MailServiceAdapter adapts the mail provider to the listener interface
type MailServiceAdapter struct{}

func (m *MailServiceAdapter) SendMail(to []string, subject, body string) error {
	// Import the mail provider function
	// This avoids import cycles by using a function pointer
	return SendMailFunc(to, subject, body)
}

// Global mail service instance
var mailService = &MailServiceAdapter{}

// GetMailService returns the mail service for listeners
func GetMailService() interface{} {
	return mailService
}

// SendMailFunc is a function pointer to avoid import cycles
var SendMailFunc func(to []string, subject, body string) error

// SetSendMailFunc sets the mail function
func SetSendMailFunc(fn func(to []string, subject, body string) error) {
	SendMailFunc = fn
}

// Global event dispatcher instance
var EventDispatcherInstance *EventDispatcher

// InitializeEventDispatcher initializes the event dispatcher
func InitializeEventDispatcher() {
	EventDispatcherInstance = NewEventDispatcher()
}
