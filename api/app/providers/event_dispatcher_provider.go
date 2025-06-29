package providers

import (
	events_core "base_lara_go_project/app/core/events"
)

func RegisterEventDispatcher() {
	// Create event dispatcher provider and set global instance
	eventDispatcherProvider := events_core.NewEventDispatcherProvider()
	events_core.SetEventDispatcherService(eventDispatcherProvider)
}
