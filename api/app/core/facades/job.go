package facades_core

import (
	app_core "base_lara_go_project/app/core/app"
	jobs_core "base_lara_go_project/app/core/jobs"
	"base_lara_go_project/config"
)

// JobDispatcher defines the interface for dispatching jobs
type JobDispatcher interface {
	Dispatch(job app_core.JobInterface) error
	DispatchSync(job app_core.JobInterface) (any, error)
}

// Global job dispatcher instance
var JobDispatcherInstance JobDispatcher

// SetJobDispatcher sets the global job dispatcher
func SetJobDispatcher(dispatcher JobDispatcher) {
	JobDispatcherInstance = dispatcher
}

// Dispatch dispatches a job asynchronously (like Laravel's dispatch() helper)
func Dispatch(job app_core.JobInterface) error {
	return JobDispatcherInstance.Dispatch(job)
}

// DispatchSync dispatches a job synchronously and returns the result (like Laravel's dispatchSync() helper)
func DispatchSync(job app_core.JobInterface) (any, error) {
	return JobDispatcherInstance.DispatchSync(job)
}

// DispatchJobAsync dispatches a job asynchronously to the jobs queue from config
func DispatchJobAsync(job interface{}) error {
	queueConfig := config.QueueConfig()
	queues := queueConfig["queues"].(map[string]interface{})
	queueName := queues["jobs"].(string)
	return jobs_core.DispatchJob(job, queueName)
}

// JobDispatcherFacade provides a facade for job dispatching
type JobDispatcherFacade struct{}

// Dispatch dispatches a job asynchronously
func (j *JobDispatcherFacade) Dispatch(job app_core.JobInterface) error {
	return jobs_core.DispatchJob(job, "default")
}

// DispatchSync dispatches a job synchronously
func (j *JobDispatcherFacade) DispatchSync(job app_core.JobInterface) (any, error) {
	return jobs_core.JobDispatcherServiceInstance.DispatchSync(job)
}

// DispatchToQueue dispatches a job to a specific queue
func (j *JobDispatcherFacade) DispatchToQueue(job app_core.JobInterface, queueName string) error {
	return jobs_core.DispatchJob(job, queueName)
}
