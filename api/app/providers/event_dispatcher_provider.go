package providers

import (
	"base_lara_go_project/app/core"
)

func RegisterEventDispatcher() {
	// Create event dispatcher provider and set global instance
	eventDispatcherProvider := core.NewEventDispatcherProvider()
	core.SetEventDispatcherService(eventDispatcherProvider)
}
