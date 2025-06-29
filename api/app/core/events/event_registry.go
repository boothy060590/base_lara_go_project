package events_core

import (
	"fmt"

	app_core "base_lara_go_project/app/core/app"
)

type EventFactory func(data map[string]interface{}) (app_core.EventInterface, error)

var eventRegistry = map[string]EventFactory{}

func RegisterEventFactory(eventName string, factory EventFactory) {
	eventRegistry[eventName] = factory
}

func CreateEvent(eventName string, data map[string]interface{}) (app_core.EventInterface, error) {
	if factory, ok := eventRegistry[eventName]; ok {
		return factory(data)
	}
	return nil, fmt.Errorf("no factory registered for event: %s", eventName)
}
