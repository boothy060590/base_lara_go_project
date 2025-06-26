package core

import (
	"fmt"
)

type EventFactory func(data map[string]interface{}) (EventInterface, error)

var eventRegistry = map[string]EventFactory{}

func RegisterEventFactory(eventName string, factory EventFactory) {
	eventRegistry[eventName] = factory
}

func CreateEvent(eventName string, data map[string]interface{}) (EventInterface, error) {
	if factory, ok := eventRegistry[eventName]; ok {
		return factory(data)
	}
	return nil, fmt.Errorf("no factory registered for event: %s", eventName)
}
