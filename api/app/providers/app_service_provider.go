package providers

import (
	app_core "base_lara_go_project/app/core/go_core"
	laravel_providers "base_lara_go_project/app/core/laravel_core/providers"
	"log"
)

// AppServiceProvider extends the core AppServiceProvider
// Developers can add their own providers and custom logic here
type AppServiceProvider struct {
	laravel_providers.AppServiceProvider
}

// Register registers all core providers and application-specific providers
func (p *AppServiceProvider) Register(container *app_core.Container) error {
	// First, register all core providers (database, cache, events, etc.)
	if err := p.AppServiceProvider.Register(container); err != nil {
		return err
	}

	// Register application-specific providers
	appProviders := []laravel_providers.ServiceProvider{
		&ListenerServiceProvider{},
		&RepositoryServiceProvider{},
		&RouterServiceProvider{},
	}

	// Register all application providers
	for _, provider := range appProviders {
		if err := provider.Register(container); err != nil {
			log.Printf("Failed to register app provider %T: %v", provider, err)
			return err
		}
	}

	// Register additional services that aren't full providers yet
	// Validation is now handled by laravel_core ValidationServiceProvider
	// RunMigrations is now handled by laravel_core MigrationServiceProvider

	log.Printf("Application service provider registered successfully")
	return nil
}

// Boot boots all providers
func (p *AppServiceProvider) Boot(container *app_core.Container) error {
	// First, boot all core providers
	if err := p.AppServiceProvider.Boot(container); err != nil {
		return err
	}

	// Boot application-specific providers
	appProviders := []laravel_providers.ServiceProvider{
		&ListenerServiceProvider{},
		&RepositoryServiceProvider{},
		&RouterServiceProvider{},
	}

	for _, provider := range appProviders {
		if err := provider.Boot(container); err != nil {
			log.Printf("Failed to boot app provider %T: %v", provider, err)
			return err
		}
	}

	log.Printf("Application service provider booted successfully")
	return nil
}

// Provides returns the services this provider provides
func (p *AppServiceProvider) Provides() []string {
	services := p.AppServiceProvider.Provides()
	services = append(services, "app.listeners", "app.repositories", "app.routes")
	return services
}

// When returns the conditions when this provider should be loaded
func (p *AppServiceProvider) When() []string {
	return []string{}
}
