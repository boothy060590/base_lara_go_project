package providers

import (
	"base_lara_go_project/app/core"
	authEvents "base_lara_go_project/app/events/auth"
	"base_lara_go_project/app/listeners"
)

// RegisterListeners registers all event listeners
func RegisterListeners() {
	// Register UserCreated event handlers
	core.EventDispatcherInstance.Register("UserCreated", func(event core.EventInterface) core.ListenerInterface {
		if userCreated, ok := event.(*authEvents.UserCreated); ok {
			return &listeners.SendEmailConfirmation{
				Event: *userCreated,
			}
		}
		return nil
	})
}
