package providers

import (
	"base_lara_go_project/app/core"
	"encoding/json"
	"log"
)

// EventDispatcherProvider implements the EventDispatcher interface
type EventDispatcherProvider struct{}

// NewEventDispatcherProvider creates a new event dispatcher provider
func NewEventDispatcherProvider() *EventDispatcherProvider {
	return &EventDispatcherProvider{}
}

// DispatchAsync dispatches an event asynchronously via queue
func (d *EventDispatcherProvider) DispatchAsync(event core.EventInterface) error {
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

	err = SendMessageWithAttributes(string(jsonData), attributes)
	if err != nil {
		log.Printf("Error sending event to queue: %v", err)
		return err
	}

	log.Printf("Event %s dispatched successfully to queue", event.GetEventName())
	return nil
}

// DispatchSync dispatches an event synchronously
func (d *EventDispatcherProvider) DispatchSync(event core.EventInterface) error {
	return core.EventDispatcherInstance.DispatchSync(event)
}
