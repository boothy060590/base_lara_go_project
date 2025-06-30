package processors

import (
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

	log.Printf("Processing event: %s", eventName)
	// TODO: Implement event processing using go_core event system
	return nil
}
