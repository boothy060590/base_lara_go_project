package providers

import (
	app_core "base_lara_go_project/app/core/go_core"
	laravel_providers "base_lara_go_project/app/core/laravel_core/providers"
	"base_lara_go_project/app/models"
	"base_lara_go_project/app/repositories"
	"database/sql"
	"log"
)

// RepositoryServiceProvider registers application repositories using canonical Repository[T] interface
type RepositoryServiceProvider struct {
	laravel_providers.BaseServiceProvider
}

// Register registers all application repositories using canonical constructors
func (p *RepositoryServiceProvider) Register(container *app_core.Container) error {
	// Resolve optimization singletons
	wsp, _ := container.Resolve("optimization.work_stealing")
	ca, _ := container.Resolve("optimization.custom_allocator")
	pgo, _ := container.Resolve("optimization.profile_guided")

	// Get database connection from container
	dbInstance, err := container.Resolve("sql.db")
	if err != nil {
		log.Printf("Database connection not found: %v", err)
		return err
	}

	// Type assert to *sql.DB - handle nil gracefully
	db, ok := dbInstance.(*sql.DB)
	if !ok || db == nil {
		log.Printf("Database connection is not *sql.DB or is nil: %T", dbInstance)
		log.Printf("Skipping repository registration - database not configured")
		return nil // Skip repository registration instead of failing
	}

	// Register repositories with canonical Repository[T] interface
	container.Singleton("repository.user", func() (any, error) {
		// Get or create cache for users
		var cache app_core.Cache[models.User]
		cacheInstance, err := container.Resolve("cache.user")
		if err != nil {
			log.Printf("User cache not found, creating local cache: %v", err)
			cache = app_core.NewLocalCache[models.User]()
		} else {
			cache = cacheInstance.(app_core.Cache[models.User])
		}

		repo := app_core.NewRepository[models.User](
			db,
			wsp.(*app_core.WorkStealingPool[any]),
			ca.(*app_core.CustomAllocator[any]),
			pgo.(*app_core.ProfileGuidedOptimizer[any]),
		)

		return repositories.NewUserRepository(repo, cache), nil
	})

	container.Singleton("repository.role", func() (any, error) {
		var cache app_core.Cache[models.Role]
		cacheInstance, err := container.Resolve("cache.role")
		if err != nil {
			cache = app_core.NewLocalCache[models.Role]()
		} else {
			cache = cacheInstance.(app_core.Cache[models.Role])
		}

		repo := app_core.NewRepository[models.Role](
			db,
			wsp.(*app_core.WorkStealingPool[any]),
			ca.(*app_core.CustomAllocator[any]),
			pgo.(*app_core.ProfileGuidedOptimizer[any]),
		)

		return repositories.NewRoleRepository(repo, cache), nil
	})

	container.Singleton("repository.permission", func() (any, error) {
		var cache app_core.Cache[models.Permission]
		cacheInstance, err := container.Resolve("cache.permission")
		if err != nil {
			cache = app_core.NewLocalCache[models.Permission]()
		} else {
			cache = cacheInstance.(app_core.Cache[models.Permission])
		}

		repo := app_core.NewRepository[models.Permission](
			db,
			wsp.(*app_core.WorkStealingPool[any]),
			ca.(*app_core.CustomAllocator[any]),
			pgo.(*app_core.ProfileGuidedOptimizer[any]),
		)

		return repositories.NewPermissionRepository(repo, cache), nil
	})

	container.Singleton("repository.category", func() (any, error) {
		var cache app_core.Cache[models.Category]
		cacheInstance, err := container.Resolve("cache.category")
		if err != nil {
			cache = app_core.NewLocalCache[models.Category]()
		} else {
			cache = cacheInstance.(app_core.Cache[models.Category])
		}

		repo := app_core.NewRepository[models.Category](
			db,
			wsp.(*app_core.WorkStealingPool[any]),
			ca.(*app_core.CustomAllocator[any]),
			pgo.(*app_core.ProfileGuidedOptimizer[any]),
		)

		return repositories.NewCategoryRepository(repo, cache), nil
	})

	container.Singleton("repository.service", func() (any, error) {
		var cache app_core.Cache[models.Service]
		cacheInstance, err := container.Resolve("cache.service")
		if err != nil {
			cache = app_core.NewLocalCache[models.Service]()
		} else {
			cache = cacheInstance.(app_core.Cache[models.Service])
		}

		repo := app_core.NewRepository[models.Service](
			db,
			wsp.(*app_core.WorkStealingPool[any]),
			ca.(*app_core.CustomAllocator[any]),
			pgo.(*app_core.ProfileGuidedOptimizer[any]),
		)

		return repositories.NewServiceRepository(repo, cache), nil
	})

	return nil
}

// Boot boots the repository service provider
func (p *RepositoryServiceProvider) Boot(container *app_core.Container) error {
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
