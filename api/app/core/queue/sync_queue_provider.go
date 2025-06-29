package queue_core

import (
	app_core "base_lara_go_project/app/core/app"
)

// SyncQueueProvider provides synchronous queue services
type SyncQueueProvider struct {
	client *SyncQueueClient
}

// NewSyncQueueProvider creates a new synchronous queue provider
func NewSyncQueueProvider(client *SyncQueueClient) *SyncQueueProvider {
	return &SyncQueueProvider{
		client: client,
	}
}

// Connect establishes a connection to the queue
func (p *SyncQueueProvider) Connect() error {
	return p.client.Connect()
}

// Disconnect closes the queue connection
func (p *SyncQueueProvider) Disconnect() error {
	return p.client.Disconnect()
}

// Push adds a job to the queue
func (p *SyncQueueProvider) Push(queue string, job interface{}) error {
	return p.client.Push(queue, job)
}

// Pop retrieves a job from the queue
func (p *SyncQueueProvider) Pop(queue string) (interface{}, error) {
	return p.client.Pop(queue)
}

// Delete removes a job from the queue
func (p *SyncQueueProvider) Delete(queue string, job interface{}) error {
	return p.client.Delete(queue, job)
}

// Size returns the number of jobs in the queue
func (p *SyncQueueProvider) Size(queue string) (int, error) {
	return p.client.Size(queue)
}

// Clear clears all jobs from the queue
func (p *SyncQueueProvider) Clear(queue string) error {
	return p.client.Clear(queue)
}

// GetStats returns queue statistics
func (p *SyncQueueProvider) GetStats() map[string]interface{} {
	return p.client.GetStats()
}

// GetClient returns the underlying queue client
func (p *SyncQueueProvider) GetClient() app_core.QueueClientInterface {
	return p.client
}
