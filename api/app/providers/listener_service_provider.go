package providers

import (
	app_core "base_lara_go_project/app/core/go_core"
	laravel_providers "base_lara_go_project/app/core/laravel_core/providers"
	"base_lara_go_project/app/listeners"
)

// ListenerServiceProvider registers application listeners
type ListenerServiceProvider struct {
	laravel_providers.BaseServiceProvider
}

// Register registers all application listeners
func (p *ListenerServiceProvider) Register(container *app_core.Container) error {
	// Register listeners
	container.Singleton("listener.send_email_confirmation", func() (any, error) {
		return &listeners.SendEmailConfirmation{}, nil
	})

	// TODO: Register more listeners as needed
	// container.Singleton("listener.user_registered", func() (any, error) {
	//     return &listeners.UserRegistered{}, nil
	// })

	return nil
}
