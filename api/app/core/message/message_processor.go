package message_core

import (
	"fmt"
	"log"

	app_core "base_lara_go_project/app/core/app"

	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// MessageProcessorProvider implements the MessageProcessorService interface
type MessageProcessorProvider struct {
	jobDispatcher app_core.JobDispatcherService
	queueService  app_core.QueueService
}

// NewMessageProcessorProvider creates a new message processor provider
func NewMessageProcessorProvider(jobDispatcher app_core.JobDispatcherService, queueService app_core.QueueService) *MessageProcessorProvider {
	return &MessageProcessorProvider{
		jobDispatcher: jobDispatcher,
		queueService:  queueService,
	}
}

// ProcessMessage processes a single message from the queue
func (m *MessageProcessorProvider) ProcessMessage(message interface{}) error {
	sqsMessage, ok := message.(*types.Message)
	if !ok {
		return fmt.Errorf("message is not a valid SQS message")
	}

	if sqsMessage.Body == nil {
		return fmt.Errorf("message body is nil")
	}

	jobType := m.GetJobTypeFromMessage(message)
	queueName := m.GetQueueNameFromMessage(message)

	log.Printf("Processing message from queue %s with job type %s", queueName, jobType)

	// Process the job based on its type
	err := m.jobDispatcher.ProcessJobFromQueue([]byte(*sqsMessage.Body), jobType)
	if err != nil {
		log.Printf("Error processing job: %v", err)
		return err
	}

	// Delete the message from the queue after successful processing
	err = m.queueService.DeleteMessageFromQueue(*sqsMessage.ReceiptHandle, queueName)
	if err != nil {
		log.Printf("Error deleting message from queue: %v", err)
		return err
	}

	log.Printf("Successfully processed and deleted message from queue %s", queueName)
	return nil
}

// ProcessMessages processes multiple messages from the queue
func (m *MessageProcessorProvider) ProcessMessages(messages []interface{}) error {
	for _, message := range messages {
		err := m.ProcessMessage(message)
		if err != nil {
			log.Printf("Error processing message: %v", err)
			// Continue processing other messages even if one fails
			continue
		}
	}
	return nil
}

// GetJobTypeFromMessage extracts the job type from message attributes
func (m *MessageProcessorProvider) GetJobTypeFromMessage(message interface{}) string {
	sqsMessage, ok := message.(*types.Message)
	if !ok {
		return "default"
	}

	if sqsMessage.MessageAttributes == nil {
		return "default"
	}

	if jobTypeAttr, exists := sqsMessage.MessageAttributes["job_type"]; exists && jobTypeAttr.StringValue != nil {
		return *jobTypeAttr.StringValue
	}

	return "default"
}

// GetQueueNameFromMessage extracts the queue name from message attributes
func (m *MessageProcessorProvider) GetQueueNameFromMessage(message interface{}) string {
	sqsMessage, ok := message.(*types.Message)
	if !ok {
		return "default"
	}

	if sqsMessage.MessageAttributes == nil {
		return "default"
	}

	if queueAttr, exists := sqsMessage.MessageAttributes["queue"]; exists && queueAttr.StringValue != nil {
		return *queueAttr.StringValue
	}

	return "default"
}

// Global message processor service instance
var MessageProcessorServiceInstance app_core.MessageProcessorService

// SetMessageProcessorService sets the global message processor service
func SetMessageProcessorService(service app_core.MessageProcessorService) {
	MessageProcessorServiceInstance = service
}

// Helper functions for message processing operations
func ProcessMessage(message interface{}) error {
	return MessageProcessorServiceInstance.ProcessMessage(message)
}

func ProcessMessages(messages []interface{}) error {
	return MessageProcessorServiceInstance.ProcessMessages(messages)
}

func GetJobTypeFromMessage(message interface{}) string {
	return MessageProcessorServiceInstance.GetJobTypeFromMessage(message)
}

func GetQueueNameFromMessage(message interface{}) string {
	return MessageProcessorServiceInstance.GetQueueNameFromMessage(message)
}
