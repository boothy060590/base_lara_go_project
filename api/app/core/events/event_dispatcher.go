package events_core

import (
	"base_lara_go_project/config"
	"encoding/json"
	"log"

	app_core "base_lara_go_project/app/core/app"
	queue_core "base_lara_go_project/app/core/queue"
)

// EventDispatcherService defines the interface for event dispatching operations
type EventDispatcherService interface {
	DispatchAsync(event app_core.EventInterface) error
	DispatchSync(event app_core.EventInterface) error
}

// EventDispatcherProvider implements the EventDispatcherService interface
type EventDispatcherProvider struct {
	// No specific config needed for event dispatching
}

// NewEventDispatcherProvider creates a new event dispatcher provider
func NewEventDispatcherProvider() *EventDispatcherProvider {
	return &EventDispatcherProvider{}
}

// DispatchAsync dispatches an event asynchronously via queue
func (d *EventDispatcherProvider) DispatchAsync(event app_core.EventInterface) error {
	// Queue the event for async processing
	eventData := map[string]interface{}{
		"job_type":  "event",
		"eventName": event.GetEventName(),
		"event":     event,
	}

	// Serialize event data to JSON
	jsonData, err := json.Marshal(eventData)
	if err != nil {
		log.Printf("Error marshaling event data: %v", err)
		return err
	}

	log.Printf("Dispatching event %s to queue: %s", event.GetEventName(), string(jsonData))

	attributes := map[string]string{
		"job_type": "event",
	}

	// Get the events queue name from config
	queueConfig := config.QueueConfig()
	queues := queueConfig["queues"].(map[string]interface{})
	eventsQueue := queues["events"].(string)

	err = queue_core.SendMessageToQueueWithAttributes(string(jsonData), attributes, eventsQueue)
	if err != nil {
		log.Printf("Error sending event to queue: %v", err)
		return err
	}

	log.Printf("Event %s dispatched successfully to queue %s", event.GetEventName(), eventsQueue)
	return nil
}

// DispatchSync dispatches an event synchronously
func (d *EventDispatcherProvider) DispatchSync(event app_core.EventInterface) error {
	return EventDispatcherInstance.DispatchSync(event)
}

// EventDispatcher handles event dispatching
type EventDispatcher struct{}

// NewEventDispatcher creates a new event dispatcher
func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{}
}

// Register registers an event handler
func (d *EventDispatcher) Register(eventName string, handlerFactory func(app_core.EventInterface) app_core.ListenerInterface) {
	app_core.GlobalRegistry.RegisterListener(eventName, handlerFactory)
}

// DispatchSync dispatches an event to all its handlers (SYNCHRONOUS - immediate)
func (d *EventDispatcher) DispatchSync(event app_core.EventInterface) error {
	eventName := event.GetEventName()

	handlers := app_core.GlobalRegistry.GetListeners(eventName)
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

// Global event dispatcher service instance
var EventDispatcherServiceInstance EventDispatcherService

// SetEventDispatcherService sets the global event dispatcher service
func SetEventDispatcherService(service EventDispatcherService) {
	EventDispatcherServiceInstance = service
}

// Helper functions for event dispatching operations
func DispatchEventAsync(event app_core.EventInterface) error {
	return EventDispatcherServiceInstance.DispatchAsync(event)
}

func DispatchEventSync(event app_core.EventInterface) error {
	return EventDispatcherServiceInstance.DispatchSync(event)
}

// InitializeEventDispatcher initializes the event dispatcher
func InitializeEventDispatcher() {
	EventDispatcherInstance = NewEventDispatcher()
}
