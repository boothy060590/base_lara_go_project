package providers

import (
	app_core "base_lara_go_project/app/core/go_core"
)

// ServiceProvider defines the interface for Laravel-style service providers
type ServiceProvider interface {
	// Register is called during the registration phase
	Register(container *app_core.Container) error

	// Boot is called after all providers have been registered
	Boot(container *app_core.Container) error

	// Provides returns the services this provider provides
	Provides() []string

	// When returns the conditions when this provider should be loaded
	When() []string
}

// BaseServiceProvider provides base functionality for service providers
type BaseServiceProvider struct {
	// Common fields can be added here if needed
}

// Register is the base implementation
func (p *BaseServiceProvider) Register(container *app_core.Container) error {
	return nil
}

// Boot is the base implementation
func (p *BaseServiceProvider) Boot(container *app_core.Container) error {
	return nil
}

// Provides returns the services this provider provides
func (p *BaseServiceProvider) Provides() []string {
	return []string{}
}

// When returns the conditions when this provider should be loaded
func (p *BaseServiceProvider) When() []string {
	return []string{}
}

// ProviderManager manages service providers
type ProviderManager struct {
	container *app_core.Container
	providers []ServiceProvider
}

// NewProviderManager creates a new provider manager
func NewProviderManager(container *app_core.Container) *ProviderManager {
	return &ProviderManager{
		container: container,
		providers: make([]ServiceProvider, 0),
	}
}

// Register registers a service provider
func (pm *ProviderManager) Register(provider ServiceProvider) error {
	pm.providers = append(pm.providers, provider)
	return provider.Register(pm.container)
}

// Boot boots all registered providers
func (pm *ProviderManager) Boot() error {
	for _, provider := range pm.providers {
		if err := provider.Boot(pm.container); err != nil {
			return err
		}
	}
	return nil
}

// GetProviders returns all registered providers
func (pm *ProviderManager) GetProviders() []ServiceProvider {
	return pm.providers
}
