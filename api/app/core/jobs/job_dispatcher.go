package jobs_core

import (
	app_core "base_lara_go_project/app/core/app"
	"base_lara_go_project/config"
	"encoding/json"
	"fmt"
	"log"
)

// JobProcessor defines the interface for processing specific job types
type JobProcessor interface {
	CanProcess(jobType string) bool
	Process(jobData []byte) error
}

// JobDispatcherService defines the interface for job dispatching operations
type JobDispatcherService interface {
	Dispatch(job app_core.JobInterface) error
	DispatchSync(job app_core.JobInterface) (any, error)
	DispatchJob(job interface{}, queueName string) error
	DispatchJobWithAttributes(job interface{}, attributes map[string]string, queueName string) error
	ProcessJobFromQueue(jobData []byte, jobType string) error
	RegisterJobProcessor(processor JobProcessor)
}

// JobDispatcherProvider implements the JobDispatcherService interface
type JobDispatcherProvider struct {
	processors   []app_core.JobProcessor
	queueService app_core.QueueService
}

// NewJobDispatcherProvider creates a new job dispatcher provider
func NewJobDispatcherProvider(queueService app_core.QueueService) *JobDispatcherProvider {
	return &JobDispatcherProvider{
		processors:   make([]app_core.JobProcessor, 0),
		queueService: queueService,
	}
}

// RegisterJobProcessor registers a job processor for specific job types
func (j *JobDispatcherProvider) RegisterJobProcessor(processor app_core.JobProcessor) {
	j.processors = append(j.processors, processor)
}

// Dispatch dispatches a job asynchronously
func (j *JobDispatcherProvider) Dispatch(job app_core.JobInterface) error {
	// For now, we'll queue the job
	// In a full implementation, this would serialize the job and send to queue
	queueName := config.QueueConfig()["queues"].(map[string]interface{})["jobs"].(string)
	return j.DispatchJob(job, queueName)
}

// DispatchSync dispatches a job synchronously and returns the result
func (j *JobDispatcherProvider) DispatchSync(job app_core.JobInterface) (any, error) {
	return job.Handle()
}

// DispatchJob dispatches a job to a specific queue
func (j *JobDispatcherProvider) DispatchJob(job interface{}, queueName string) error {
	// Marshal job data
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job data: %v", err)
	}

	// Send to queue
	return j.queueService.SendMessageToQueue(string(jobData), queueName)
}

// DispatchJobWithAttributes dispatches a job with custom attributes to a specific queue
func (j *JobDispatcherProvider) DispatchJobWithAttributes(job interface{}, attributes map[string]string, queueName string) error {
	// Marshal job data
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job data: %v", err)
	}

	// Send to queue with attributes
	return j.queueService.SendMessageToQueueWithAttributes(string(jobData), attributes, queueName)
}

// ProcessJobFromQueue processes a job from the queue based on job type
func (j *JobDispatcherProvider) ProcessJobFromQueue(jobData []byte, jobType string) error {
	log.Printf("Processing job of type: %s", jobType)

	// Try to find a processor for this job type
	for _, processor := range j.processors {
		if processor.CanProcess(jobType) {
			return processor.Process(jobData)
		}
	}

	// If no processor found, return an error
	return fmt.Errorf("no processor found for job type: %s", jobType)
}

// Global job dispatcher service instance
var JobDispatcherServiceInstance app_core.JobDispatcherService

// SetJobDispatcherService sets the global job dispatcher service
func SetJobDispatcherService(service app_core.JobDispatcherService) {
	JobDispatcherServiceInstance = service
}

// Helper functions for job dispatching operations
func DispatchJob(job interface{}, queueName string) error {
	return JobDispatcherServiceInstance.DispatchJob(job, queueName)
}

func DispatchJobWithAttributes(job interface{}, attributes map[string]string, queueName string) error {
	return JobDispatcherServiceInstance.DispatchJobWithAttributes(job, attributes, queueName)
}

func ProcessJobFromQueue(jobData []byte, jobType string) error {
	return JobDispatcherServiceInstance.ProcessJobFromQueue(jobData, jobType)
}

// RegisterJobProcessor registers a job processor with the global job dispatcher
func RegisterJobProcessor(processor app_core.JobProcessor) {
	if dispatcher, ok := JobDispatcherServiceInstance.(*JobDispatcherProvider); ok {
		dispatcher.RegisterJobProcessor(processor)
	}
}
