package queue_core

import (
	app_core "base_lara_go_project/app/core/app"
)

// SQSQueueProvider provides SQS queue services
type SQSQueueProvider struct {
	client *SQSQueueClient
}

// NewSQSQueueProvider creates a new SQS queue provider
func NewSQSQueueProvider(client *SQSQueueClient) *SQSQueueProvider {
	return &SQSQueueProvider{
		client: client,
	}
}

// Connect establishes a connection to the queue
func (p *SQSQueueProvider) Connect() error {
	return p.client.Connect()
}

// Disconnect closes the queue connection
func (p *SQSQueueProvider) Disconnect() error {
	return p.client.Disconnect()
}

// Push adds a job to the queue
func (p *SQSQueueProvider) Push(queue string, job interface{}) error {
	return p.client.Push(queue, job)
}

// Pop retrieves a job from the queue
func (p *SQSQueueProvider) Pop(queue string) (interface{}, error) {
	return p.client.Pop(queue)
}

// Delete removes a job from the queue
func (p *SQSQueueProvider) Delete(queue string, job interface{}) error {
	return p.client.Delete(queue, job)
}

// Size returns the number of jobs in the queue
func (p *SQSQueueProvider) Size(queue string) (int, error) {
	return p.client.Size(queue)
}

// Clear clears all jobs from the queue
func (p *SQSQueueProvider) Clear(queue string) error {
	return p.client.Clear(queue)
}

// GetStats returns queue statistics
func (p *SQSQueueProvider) GetStats() map[string]interface{} {
	return p.client.GetStats()
}

// GetClient returns the underlying queue client
func (p *SQSQueueProvider) GetClient() app_core.QueueClientInterface {
	return p.client
}
