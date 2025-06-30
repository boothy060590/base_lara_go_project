package providers

import (
	app_core "base_lara_go_project/app/core/go_core"
)

// ValidationServiceProvider handles validation system registration
type ValidationServiceProvider struct {
	BaseServiceProvider
}

// Register registers the validation service and built-in rules
func (p *ValidationServiceProvider) Register(container *app_core.Container) error {
	// Register validation service (using go_core directly)
	container.Singleton("validation.service", func() (any, error) {
		return &app_core.Validator[map[string]any]{}, nil
	})

	// Register built-in validation rules
	container.Singleton("validation.rules", func() (any, error) {
		return map[string]app_core.ValidationRule{
			"required": app_core.Required(),
			"string":   app_core.String(),
			"max":      app_core.Max(255),
			"min":      app_core.Min(1),
		}, nil
	})

	// Register individual rule instances
	container.Singleton("validation.rule.required", func() (any, error) {
		return app_core.Required(), nil
	})

	container.Singleton("validation.rule.string", func() (any, error) {
		return app_core.String(), nil
	})

	container.Singleton("validation.rule.max", func() (any, error) {
		return app_core.Max(255), nil
	})

	container.Singleton("validation.rule.min", func() (any, error) {
		return app_core.Min(1), nil
	})

	return nil
}

// Boot boots the validation service provider
func (p *ValidationServiceProvider) Boot(container *app_core.Container) error {
	// TODO: Register custom validation rules from application
	return nil
}

// Provides returns the services this provider provides
func (p *ValidationServiceProvider) Provides() []string {
	return []string{"validation"}
}

// When returns the conditions when this provider should be loaded
func (p *ValidationServiceProvider) When() []string {
	return []string{}
}
