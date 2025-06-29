package listeners

import (
	app_core "base_lara_go_project/app/core/app"
)

// BaseListener provides a base structure for all listeners
type BaseListener struct {
	// Common fields can be added here if needed
}

// Handle is the base implementation - should be overridden by specific listeners
func (l *BaseListener) Handle(mailService interface{}) error {
	// Base implementation - should be overridden
	return nil
}

// ListenerFactory is a function type that creates listeners from events
type ListenerFactory func(event interface{}) app_core.ListenerInterface
