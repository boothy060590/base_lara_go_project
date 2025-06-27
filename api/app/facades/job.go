package facades

import (
	"base_lara_go_project/app/core"
	"base_lara_go_project/config"
)

// JobDispatcher defines the interface for dispatching jobs
type JobDispatcher interface {
	Dispatch(job core.JobInterface) error
	DispatchSync(job core.JobInterface) (any, error)
}

// Global job dispatcher instance
var JobDispatcherInstance JobDispatcher

// SetJobDispatcher sets the global job dispatcher
func SetJobDispatcher(dispatcher JobDispatcher) {
	JobDispatcherInstance = dispatcher
}

// Dispatch dispatches a job asynchronously (like Laravel's dispatch() helper)
func Dispatch(job core.JobInterface) error {
	return JobDispatcherInstance.Dispatch(job)
}

// DispatchSync dispatches a job synchronously and returns the result (like Laravel's dispatchSync() helper)
func DispatchSync(job core.JobInterface) (any, error) {
	return JobDispatcherInstance.DispatchSync(job)
}

// DispatchJobAsync dispatches a job asynchronously to the jobs queue from config
func DispatchJobAsync(job interface{}) error {
	queueConfig := config.QueueConfig()
	queues := queueConfig["queues"].(map[string]interface{})
	queueName := queues["jobs"].(string)
	return core.DispatchJob(job, queueName)
}
