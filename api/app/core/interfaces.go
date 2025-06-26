package core

// EventInterface defines the interface for all events
type EventInterface interface {
	GetEventName() string
}

// ListenerInterface defines the interface for all listeners
type ListenerInterface interface {
	Handle(mailService interface{}) error
}

// JobInterface defines the interface for all jobs
type JobInterface interface {
	Handle() (any, error)
}
