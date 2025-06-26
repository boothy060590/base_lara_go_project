package providers

import (
	"base_lara_go_project/app/core"
)

// JobDispatcherProvider implements the JobDispatcher interface
type JobDispatcherProvider struct{}

// NewJobDispatcherProvider creates a new job dispatcher provider
func NewJobDispatcherProvider() *JobDispatcherProvider {
	return &JobDispatcherProvider{}
}

// Dispatch dispatches a job asynchronously
func (d *JobDispatcherProvider) Dispatch(job core.JobInterface) error {
	// For now, we'll queue the job
	// In a full implementation, this would serialize the job and send to queue
	return SendMessage("job")
}

// DispatchSync dispatches a job synchronously and returns the result
func (d *JobDispatcherProvider) DispatchSync(job core.JobInterface) (any, error) {
	return job.Handle()
}
