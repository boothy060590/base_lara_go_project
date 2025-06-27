package providers

import (
	"base_lara_go_project/app/core"
)

func RegisterJobDispatcher() {
	// Create job dispatcher provider and set global instance
	jobDispatcherProvider := core.NewJobDispatcherProvider()
	core.SetJobDispatcherService(jobDispatcherProvider)
}
