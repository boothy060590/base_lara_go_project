package providers

import (
	app_core "base_lara_go_project/app/core/app"
	jobs_core "base_lara_go_project/app/core/jobs"
)

func RegisterJobDispatcher() {
	// Get queue service from container
	queueService, err := app_core.App.Resolve("queue.service")
	if err != nil {
		panic("Queue service not found in container")
	}

	// Create job dispatcher provider and set global instance
	jobDispatcherProvider := jobs_core.NewJobDispatcherProvider(queueService.(app_core.QueueService))
	jobs_core.SetJobDispatcherService(jobDispatcherProvider)
}
