package core

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// MessageProcessorService defines the interface for message processing operations
type MessageProcessorService interface {
	ProcessMessage(message *types.Message) error
	ProcessMessages(messages []types.Message) error
	GetJobTypeFromMessage(message *types.Message) string
	GetQueueNameFromMessage(message *types.Message) string
}

// MessageProcessorProvider implements the MessageProcessorService interface
type MessageProcessorProvider struct {
	// No specific config needed for message processing
}

// NewMessageProcessorProvider creates a new message processor provider
func NewMessageProcessorProvider() *MessageProcessorProvider {
	return &MessageProcessorProvider{}
}

// ProcessMessage processes a single message from the queue
func (m *MessageProcessorProvider) ProcessMessage(message *types.Message) error {
	if message.Body == nil {
		return fmt.Errorf("message body is nil")
	}

	jobType := m.GetJobTypeFromMessage(message)
	queueName := m.GetQueueNameFromMessage(message)

	log.Printf("Processing message from queue %s with job type %s", queueName, jobType)

	// Process the job based on its type
	err := ProcessJobFromQueue([]byte(*message.Body), jobType)
	if err != nil {
		log.Printf("Error processing job: %v", err)
		return err
	}

	// Delete the message from the queue after successful processing
	err = DeleteMessageFromQueue(*message.ReceiptHandle, queueName)
	if err != nil {
		log.Printf("Error deleting message from queue: %v", err)
		return err
	}

	log.Printf("Successfully processed and deleted message from queue %s", queueName)
	return nil
}

// ProcessMessages processes multiple messages from the queue
func (m *MessageProcessorProvider) ProcessMessages(messages []types.Message) error {
	for _, message := range messages {
		err := m.ProcessMessage(&message)
		if err != nil {
			log.Printf("Error processing message: %v", err)
			// Continue processing other messages even if one fails
			continue
		}
	}
	return nil
}

// GetJobTypeFromMessage extracts the job type from message attributes
func (m *MessageProcessorProvider) GetJobTypeFromMessage(message *types.Message) string {
	if message.MessageAttributes == nil {
		return "default"
	}

	if jobTypeAttr, exists := message.MessageAttributes["job_type"]; exists && jobTypeAttr.StringValue != nil {
		return *jobTypeAttr.StringValue
	}

	return "default"
}

// GetQueueNameFromMessage extracts the queue name from message attributes
func (m *MessageProcessorProvider) GetQueueNameFromMessage(message *types.Message) string {
	if message.MessageAttributes == nil {
		return "default"
	}

	if queueAttr, exists := message.MessageAttributes["queue"]; exists && queueAttr.StringValue != nil {
		return *queueAttr.StringValue
	}

	return "default"
}

// Global message processor service instance
var MessageProcessorServiceInstance MessageProcessorService

// SetMessageProcessorService sets the global message processor service
func SetMessageProcessorService(service MessageProcessorService) {
	MessageProcessorServiceInstance = service
}

// Helper functions for message processing operations
func ProcessMessage(message *types.Message) error {
	return MessageProcessorServiceInstance.ProcessMessage(message)
}

func ProcessMessages(messages []types.Message) error {
	return MessageProcessorServiceInstance.ProcessMessages(messages)
}

func GetJobTypeFromMessage(message *types.Message) string {
	return MessageProcessorServiceInstance.GetJobTypeFromMessage(message)
}

func GetQueueNameFromMessage(message *types.Message) string {
	return MessageProcessorServiceInstance.GetQueueNameFromMessage(message)
}
