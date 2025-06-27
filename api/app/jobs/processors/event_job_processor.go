package processors

import (
	"base_lara_go_project/app/core"
	"encoding/json"
	"fmt"
	"log"
)

// EventJobProcessor handles event job processing
type EventJobProcessor struct{}

// NewEventJobProcessor creates a new event job processor
func NewEventJobProcessor() *EventJobProcessor {
	return &EventJobProcessor{}
}

// CanProcess checks if this processor can handle the given job type
func (e *EventJobProcessor) CanProcess(jobType string) bool {
	return jobType == "event"
}

// Process processes an event job
func (e *EventJobProcessor) Process(jobData []byte) error {
	var eventData map[string]interface{}
	if err := json.Unmarshal(jobData, &eventData); err != nil {
		return fmt.Errorf("failed to unmarshal event data: %v", err)
	}

	eventName, ok := eventData["eventName"].(string)
	if !ok {
		return fmt.Errorf("invalid event name in job data")
	}

	eventPayload, ok := eventData["event"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid event payload in job data")
	}

	log.Printf("Processing event: %s", eventName)
	event, err := core.CreateEvent(eventName, eventPayload)
	if err != nil {
		return fmt.Errorf("failed to create event: %v", err)
	}

	return core.EventDispatcherInstance.DispatchSync(event)
}
