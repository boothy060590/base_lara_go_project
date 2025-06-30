package go_core

import (
	"fmt"
)

// Global container instance for Laravel-style compatibility
var App = NewContainer()

// InitializeRegistry initializes the global registry (Laravel-style compatibility)
func InitializeRegistry() {
	// The container is already initialized when created
	// This function exists for Laravel-style compatibility
}

// EventDispatcherServiceInstance provides Laravel-style global event dispatcher
var EventDispatcherServiceInstance EventManagerInterface[any]

// JobDispatcherServiceInstance provides Laravel-style global job dispatcher
var JobDispatcherServiceInstance Queue[any]

// CacheInstance provides Laravel-style global cache instance
var CacheInstance Cache[any]

// QueueServiceInstance provides Laravel-style global queue instance
var QueueServiceInstance Queue[any]

// InitializeEventDispatcher initializes the global event dispatcher
func InitializeEventDispatcher() {
	// Create basic event dispatcher and store
	dispatcher := NewEventBus[any]()
	store := NewMemoryEventStore[any]()
	EventDispatcherServiceInstance = NewEventManager[any](dispatcher, store)
}

// SetSendMailFunc sets the mail function for event dispatcher (Laravel-style compatibility)
func SetSendMailFunc(mailFunc func(to []string, subject, body string) error) {
	// This would be implemented when we have mail functionality
	// For now, it's a placeholder for Laravel-style compatibility
}

// SendMail provides Laravel-style mail sending function
func SendMail(to []string, subject, body string) error {
	// TODO: Implement actual mail sending
	fmt.Printf("Sending mail to %v: %s - %s\n", to, subject, body)
	return nil
}

// MessageProcessorService interface for Laravel-style compatibility
type MessageProcessorService interface {
	Process(messageType string, data map[string]interface{}) error
}

// Basic message processor implementation
type BasicMessageProcessor struct{}

func (m *BasicMessageProcessor) Process(messageType string, data map[string]interface{}) error {
	// TODO: Implement actual message processing
	fmt.Printf("Processing message type %s with data %+v\n", messageType, data)
	return nil
}

// Initialize global instances
func init() {
	// Initialize global instances for Laravel-style compatibility
	InitializeEventDispatcher()

	// Create basic instances
	JobDispatcherServiceInstance = NewSyncQueue[any]()
	CacheInstance = NewLocalCache[any]()
	QueueServiceInstance = NewSyncQueue[any]()

	// Register basic message processor
	App.Singleton("message_processor", func() (any, error) {
		return &BasicMessageProcessor{}, nil
	})
}
