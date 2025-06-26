package providers

import (
	"encoding/json"
	"fmt"
	"log"

	"base_lara_go_project/app/core"

	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// MessageProcessorProvider implements the MessageProcessor interface
type MessageProcessorProvider struct{}

// NewMessageProcessorProvider creates a new message processor provider
func NewMessageProcessorProvider() *MessageProcessorProvider {
	return &MessageProcessorProvider{}
}

// ProcessMessages processes messages from the queue
func (m *MessageProcessorProvider) ProcessMessages() error {
	result, err := ReceiveMessage()
	if err != nil {
		return fmt.Errorf("error receiving messages: %v", err)
	}

	log.Printf("Received %d messages from queue", len(result.Messages))

	for _, message := range result.Messages {
		log.Printf("Processing message: %s", *message.Body)
		if err := m.processMessage(message); err != nil {
			log.Printf("Error processing message: %v", err)
			continue
		}

		// Delete the message after successful processing
		if err := DeleteMessage(*message.ReceiptHandle); err != nil {
			log.Printf("Error deleting message: %v", err)
		}
	}

	return nil
}

// processMessage processes a single message
func (m *MessageProcessorProvider) processMessage(message types.Message) error {
	if message.Body == nil {
		return fmt.Errorf("message body is nil")
	}

	log.Printf("Message attributes: %+v", message.MessageAttributes)

	// First try to get job_type from message attributes
	var jobType string
	if jobTypeAttr, exists := message.MessageAttributes["job_type"]; exists {
		log.Printf("Found job_type attribute: %+v", jobTypeAttr)
		if jobTypeAttr.StringValue != nil {
			jobType = *jobTypeAttr.StringValue
			log.Printf("Job type from attributes: %s", jobType)
		}
	}

	// If no job_type in attributes, try to get it from message body
	if jobType == "" {
		var messageData map[string]interface{}
		if err := json.Unmarshal([]byte(*message.Body), &messageData); err == nil {
			if bodyJobType, ok := messageData["job_type"].(string); ok {
				jobType = bodyJobType
				log.Printf("Job type from body: %s", jobType)
			}
		}
	}

	// Process based on job type
	switch jobType {
	case "send_mail":
		return ProcessMailJobFromQueue([]byte(*message.Body))
	case "event":
		return m.processEventJob([]byte(*message.Body))
	case "job":
		return m.processGenericJob([]byte(*message.Body))
	default:
		log.Printf("No job_type found, processing as generic job")
		return m.processGenericJob([]byte(*message.Body))
	}
}

// processEventJob processes an event job from the queue
func (m *MessageProcessorProvider) processEventJob(jobData []byte) error {
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

// processGenericJob processes a generic job from the queue
func (m *MessageProcessorProvider) processGenericJob(jobData []byte) error {
	var job map[string]interface{}
	if err := json.Unmarshal(jobData, &job); err != nil {
		return fmt.Errorf("failed to unmarshal job: %v", err)
	}

	log.Printf("Processed generic job: %v", job)
	return nil
}
