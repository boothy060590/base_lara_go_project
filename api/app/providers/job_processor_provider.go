package providers

import (
	"base_lara_go_project/app/core"
	"base_lara_go_project/app/jobs/processors"
)

func RegisterJobProcessors() {
	// Register mail job processor
	mailProcessor := processors.NewMailJobProcessor()
	core.RegisterJobProcessor(mailProcessor)

	// Register event job processor
	eventProcessor := processors.NewEventJobProcessor()
	core.RegisterJobProcessor(eventProcessor)

	// Register user job processor
	userProcessor := processors.NewUserJobProcessor()
	core.RegisterJobProcessor(userProcessor)
}
