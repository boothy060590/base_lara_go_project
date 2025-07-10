package providers

import (
	app_core "base_lara_go_project/app/core/go_core"
	laravel_providers "base_lara_go_project/app/core/laravel_core/providers"
	"base_lara_go_project/app/models"
	"base_lara_go_project/app/repositories"
	"log"
	"time"

	"gorm.io/gorm"
)

// RepositoryServiceProvider registers application repositories
type RepositoryServiceProvider struct {
	laravel_providers.BaseServiceProvider
}

// Register registers all application repositories
func (p *RepositoryServiceProvider) Register(container *app_core.Container) error {
	// Resolve optimization singletons
	wsp, _ := container.Resolve("optimization.work_stealing")
	ca, _ := container.Resolve("optimization.custom_allocator")
	pgo, _ := container.Resolve("optimization.profile_guided")

	// Register repositories with dependency injection
	container.Singleton("repository.user", func() (any, error) {
		// Get database instance from container
		dbInstance, err := container.Resolve("gorm.db")
		if err != nil {
			log.Printf("Database not found, creating repository with nil database: %v", err)
			// Create a temporary cache for now
			cache := app_core.NewLocalCache[models.User]()
			return repositories.NewUserRepository(nil, cache, wsp, ca, pgo), nil
		}

		db := dbInstance.(*gorm.DB)

		// Get or create cache for users
		var cache app_core.Cache[models.User]
		cacheInstance, err := container.Resolve("cache.user")
		if err != nil {
			log.Printf("User cache not found, creating local cache: %v", err)
			cache = app_core.NewLocalCache[models.User]()
		} else {
			cache = cacheInstance.(app_core.Cache[models.User])
		}

		return repositories.NewUserRepository(db, cache, wsp, ca, pgo), nil
	})

	container.Singleton("repository.role", func() (any, error) {
		dbInstance, err := container.Resolve("gorm.db")
		if err != nil {
			log.Printf("Database not found, creating repository with nil database: %v", err)
			cache := app_core.NewLocalCache[models.Role]()
			return repositories.NewRoleRepository(nil, cache, wsp, ca, pgo), nil
		}
		db := dbInstance.(*gorm.DB)
		var cache app_core.Cache[models.Role]
		cacheInstance, err := container.Resolve("cache.role")
		if err != nil {
			cache = app_core.NewLocalCache[models.Role]()
		} else {
			cache = cacheInstance.(app_core.Cache[models.Role])
		}
		return repositories.NewRoleRepository(db, cache, wsp, ca, pgo), nil
	})

	container.Singleton("repository.permission", func() (any, error) {
		dbInstance, err := container.Resolve("gorm.db")
		if err != nil {
			log.Printf("Database not found, creating repository with nil database: %v", err)
			cache := app_core.NewLocalCache[models.Permission]()
			return repositories.NewPermissionRepository(nil, cache, wsp, ca, pgo), nil
		}
		db := dbInstance.(*gorm.DB)
		var cache app_core.Cache[models.Permission]
		cacheInstance, err := container.Resolve("cache.permission")
		if err != nil {
			cache = app_core.NewLocalCache[models.Permission]()
		} else {
			cache = cacheInstance.(app_core.Cache[models.Permission])
		}
		return repositories.NewPermissionRepository(db, cache, wsp, ca, pgo), nil
	})

	container.Singleton("repository.category", func() (any, error) {
		dbInstance, err := container.Resolve("gorm.db")
		if err != nil {
			log.Printf("Database not found, creating repository with nil database: %v", err)
			cache := app_core.NewLocalCache[models.Category]()
			return repositories.NewCategoryRepository(nil, cache, wsp, ca, pgo), nil
		}
		db := dbInstance.(*gorm.DB)
		var cache app_core.Cache[models.Category]
		cacheInstance, err := container.Resolve("cache.category")
		if err != nil {
			cache = app_core.NewLocalCache[models.Category]()
		} else {
			cache = cacheInstance.(app_core.Cache[models.Category])
		}
		return repositories.NewCategoryRepository(db, cache, wsp, ca, pgo), nil
	})

	container.Singleton("repository.service", func() (any, error) {
		dbInstance, err := container.Resolve("gorm.db")
		if err != nil {
			log.Printf("Database not found, creating repository with nil database: %v", err)
			cache := app_core.NewLocalCache[models.Service]()
			config := app_core.ModelConfig{
				TableName: "services",
				Traits: app_core.ModelTraits{
					Cacheable:   true,
					SoftDeletes: true,
					Timestamps:  true,
				},
				CacheTTL:    30 * time.Minute,
				CachePrefix: "service",
			}
			return app_core.NewBaseModel[models.Service](nil, cache, config, wsp, ca, pgo), nil
		}

		db := dbInstance.(*gorm.DB)

		var cache app_core.Cache[models.Service]
		cacheInstance, err := container.Resolve("cache.service")
		if err != nil {
			cache = app_core.NewLocalCache[models.Service]()
		} else {
			cache = cacheInstance.(app_core.Cache[models.Service])
		}

		config := app_core.ModelConfig{
			TableName: "services",
			Traits: app_core.ModelTraits{
				Cacheable:   true,
				SoftDeletes: true,
				Timestamps:  true,
			},
			CacheTTL:    30 * time.Minute,
			CachePrefix: "service",
		}
		return app_core.NewBaseModel[models.Service](db, cache, config, wsp, ca, pgo), nil
	})

	return nil
}

// Boot boots the repository service provider
func (p *RepositoryServiceProvider) Boot(container *app_core.Container) error {
	// TODO: Inject database instances into repositories if they were created with nil
	return nil
}

// Provides returns the services this provider provides
func (p *RepositoryServiceProvider) Provides() []string {
	return []string{"repositories"}
}

// When returns the conditions when this provider should be loaded
func (p *RepositoryServiceProvider) When() []string {
	return []string{}
}
