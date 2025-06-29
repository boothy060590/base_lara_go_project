package providers

import (
	app_core "base_lara_go_project/app/core/app"
	message_core "base_lara_go_project/app/core/message"
)

func RegisterMessageProcessor() {
	// Get required services from container
	jobDispatcherService, err := app_core.App.Resolve("job.dispatcher.service")
	if err != nil {
		panic("Job dispatcher service not found in container")
	}

	queueService, err := app_core.App.Resolve("queue.service")
	if err != nil {
		panic("Queue service not found in container")
	}

	// Create message processor provider and set global instance
	messageProcessorProvider := message_core.NewMessageProcessorProvider(
		jobDispatcherService.(app_core.JobDispatcherService),
		queueService.(app_core.QueueService),
	)
	message_core.SetMessageProcessorService(messageProcessorProvider)
}
