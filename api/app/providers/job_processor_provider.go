package providers

import (
	jobs_core "base_lara_go_project/app/core/jobs"
	"base_lara_go_project/app/jobs/processors"
)

func RegisterJobProcessors() {
	// Register mail job processor
	mailProcessor := processors.NewMailJobProcessor()
	jobs_core.RegisterJobProcessor(mailProcessor)

	// Register event job processor
	eventProcessor := processors.NewEventJobProcessor()
	jobs_core.RegisterJobProcessor(eventProcessor)

	// Register user job processor
	userProcessor := processors.NewUserJobProcessor()
	jobs_core.RegisterJobProcessor(userProcessor)
}
