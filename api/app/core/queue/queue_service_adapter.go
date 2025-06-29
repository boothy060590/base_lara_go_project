package queue_core

import (
	"encoding/json"
	"fmt"

	app_core "base_lara_go_project/app/core/app"
)

// QueueServiceAdapter adapts a QueueProviderServiceInterface to QueueService
type QueueServiceAdapter struct {
	provider app_core.QueueProviderServiceInterface
}

// NewQueueServiceAdapter creates a new queue service adapter
func NewQueueServiceAdapter(provider app_core.QueueProviderServiceInterface) *QueueServiceAdapter {
	return &QueueServiceAdapter{
		provider: provider,
	}
}

// SendMessage sends a message to the default queue
func (a *QueueServiceAdapter) SendMessage(messageBody string) error {
	return a.provider.Push("default", messageBody)
}

// SendMessageToQueue sends a message to a specific queue
func (a *QueueServiceAdapter) SendMessageToQueue(messageBody string, queueName string) error {
	return a.provider.Push(queueName, messageBody)
}

// SendMessageWithAttributes sends a message with attributes to the default queue
func (a *QueueServiceAdapter) SendMessageWithAttributes(messageBody string, attributes map[string]string) error {
	// Create a message with attributes
	message := map[string]interface{}{
		"body":       messageBody,
		"attributes": attributes,
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return a.provider.Push("default", string(messageBytes))
}

// SendMessageToQueueWithAttributes sends a message with attributes to a specific queue
func (a *QueueServiceAdapter) SendMessageToQueueWithAttributes(messageBody string, attributes map[string]string, queueName string) error {
	// Create a message with attributes
	message := map[string]interface{}{
		"body":       messageBody,
		"attributes": attributes,
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return a.provider.Push(queueName, string(messageBytes))
}

// ReceiveMessage receives a message from the default queue
func (a *QueueServiceAdapter) ReceiveMessage() (interface{}, error) {
	return a.provider.Pop("default")
}

// ReceiveMessageFromQueue receives a message from a specific queue
func (a *QueueServiceAdapter) ReceiveMessageFromQueue(queueName string) (interface{}, error) {
	return a.provider.Pop(queueName)
}

// DeleteMessage deletes a message from the default queue
func (a *QueueServiceAdapter) DeleteMessage(receiptHandle string) error {
	return a.provider.Delete("default", receiptHandle)
}

// DeleteMessageFromQueue deletes a message from a specific queue
func (a *QueueServiceAdapter) DeleteMessageFromQueue(receiptHandle string, queueName string) error {
	return a.provider.Delete(queueName, receiptHandle)
}
