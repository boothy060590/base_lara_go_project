package providers

import (
	app_core "base_lara_go_project/app/core/go_core"
	"log"
)

// AppServiceProvider is the main application service provider
// It loads all core providers and can be extended by developers
type AppServiceProvider struct {
	BaseServiceProvider
}

// Register registers all core providers and any additional providers
func (p *AppServiceProvider) Register(container *app_core.Container) error {
	// Register core infrastructure providers
	coreProviders := []ServiceProvider{
		&CoreServiceProvider{},
		&DatabaseServiceProvider{},
		&CacheServiceProvider{},
		&EventServiceProvider{},
		&QueueServiceProvider{},
		&MailServiceProvider{},
		&LoggingServiceProvider{},
		&JobServiceProvider{},
		&MigrationServiceProvider{},
		&ValidationServiceProvider{},
	}

	// Register all core providers
	for _, provider := range coreProviders {
		if err := provider.Register(container); err != nil {
			log.Printf("Failed to register provider %T: %v", provider, err)
			return err
		}
	}

	// Call the developer's custom register method
	if err := p.registerCustomProviders(container); err != nil {
		return err
	}

	log.Printf("App service provider registered successfully")
	return nil
}

// Boot boots all providers
func (p *AppServiceProvider) Boot(container *app_core.Container) error {
	// Boot core providers
	coreProviders := []ServiceProvider{
		&CoreServiceProvider{},
		&DatabaseServiceProvider{},
		&CacheServiceProvider{},
		&EventServiceProvider{},
		&QueueServiceProvider{},
		&MailServiceProvider{},
		&LoggingServiceProvider{},
		&JobServiceProvider{},
		&MigrationServiceProvider{},
		&ValidationServiceProvider{},
	}

	for _, provider := range coreProviders {
		if err := provider.Boot(container); err != nil {
			log.Printf("Failed to boot provider %T: %v", provider, err)
			return err
		}
	}

	// Call the developer's custom boot method
	if err := p.bootCustomProviders(container); err != nil {
		return err
	}

	log.Printf("App service provider booted successfully")
	return nil
}

// Provides returns the services this provider provides
func (p *AppServiceProvider) Provides() []string {
	return []string{"app"}
}

// When returns the conditions when this provider should be loaded
func (p *AppServiceProvider) When() []string {
	return []string{}
}

// registerCustomProviders is called by developers to register additional providers
func (p *AppServiceProvider) registerCustomProviders(container *app_core.Container) error {
	// This method can be overridden by developers
	return nil
}

// bootCustomProviders is called by developers to boot additional providers
func (p *AppServiceProvider) bootCustomProviders(container *app_core.Container) error {
	// This method can be overridden by developers
	return nil
}

// Individual core service providers for better organization

// DatabaseServiceProvider handles database registration
type DatabaseServiceProvider struct {
	BaseServiceProvider
}

func (p *DatabaseServiceProvider) Register(container *app_core.Container) error {
	// TODO: Implement database registration
	container.Singleton("database", func() (any, error) {
		return nil, nil
	})
	container.Singleton("gorm.db", func() (any, error) {
		return nil, nil
	})
	return nil
}

func (p *DatabaseServiceProvider) Provides() []string {
	return []string{"database"}
}

// CacheServiceProvider handles cache registration
type CacheServiceProvider struct {
	BaseServiceProvider
}

func (p *CacheServiceProvider) Register(container *app_core.Container) error {
	// TODO: Implement cache registration
	container.Singleton("cache", func() (any, error) {
		return app_core.NewLocalCache[any](), nil
	})
	return nil
}

func (p *CacheServiceProvider) Provides() []string {
	return []string{"cache"}
}

// EventServiceProvider handles event system registration
type EventServiceProvider struct {
	BaseServiceProvider
}

func (p *EventServiceProvider) Register(container *app_core.Container) error {
	// TODO: Implement event system registration
	container.Singleton("event_manager", func() (any, error) {
		return app_core.NewEventManager[any](
			app_core.NewEventBus[any](),
			app_core.NewMemoryEventStore[any](),
		), nil
	})
	return nil
}

func (p *EventServiceProvider) Provides() []string {
	return []string{"events"}
}

// QueueServiceProvider handles queue system registration
type QueueServiceProvider struct {
	BaseServiceProvider
}

func (p *QueueServiceProvider) Register(container *app_core.Container) error {
	// TODO: Implement queue system registration
	container.Singleton("queue", func() (any, error) {
		return app_core.NewSyncQueue[any](), nil
	})
	return nil
}

func (p *QueueServiceProvider) Provides() []string {
	return []string{"queue"}
}

// MailServiceProvider handles mail system registration
type MailServiceProvider struct {
	BaseServiceProvider
}

func (p *MailServiceProvider) Register(container *app_core.Container) error {
	// TODO: Implement mail system registration
	container.Singleton("mail", func() (any, error) {
		return nil, nil
	})
	return nil
}

func (p *MailServiceProvider) Provides() []string {
	return []string{"mail"}
}

// LoggingServiceProvider handles logging system registration
type LoggingServiceProvider struct {
	BaseServiceProvider
}

func (p *LoggingServiceProvider) Register(container *app_core.Container) error {
	// TODO: Implement logging system registration
	container.Singleton("logger", func() (any, error) {
		return nil, nil
	})
	return nil
}

func (p *LoggingServiceProvider) Provides() []string {
	return []string{"logging"}
}

// JobServiceProvider handles job system registration
type JobServiceProvider struct {
	BaseServiceProvider
}

func (p *JobServiceProvider) Register(container *app_core.Container) error {
	// TODO: Implement job system registration
	queueInstance, err := container.Resolve("queue")
	if err != nil {
		return err
	}

	queue := queueInstance.(app_core.Queue[any])
	jobDispatcher := app_core.NewJobDispatcher[any](queue)

	container.Singleton("job.dispatcher", func() (any, error) {
		return jobDispatcher, nil
	})
	return nil
}

func (p *JobServiceProvider) Provides() []string {
	return []string{"jobs"}
}
