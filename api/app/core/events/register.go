package events_core

import (
	app_core "base_lara_go_project/app/core/app"
)

// RegisterEvent registers an event listener
func RegisterEvent(eventName string, handlerFactory func(app_core.EventInterface) app_core.ListenerInterface) {
	app_core.GlobalRegistry.RegisterListener(eventName, handlerFactory)
}
