package providers

import (
	"base_lara_go_project/app/listeners"
)

// RegisterAppEvents registers all application-specific events and listeners
func RegisterAppEvents() {
	// Register listeners (they register themselves)
	listeners.RegisterSelf()
	// Add more event registrations here as needed
}
