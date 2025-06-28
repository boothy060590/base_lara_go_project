package providers

import (
	"base_lara_go_project/app/facades"
	"base_lara_go_project/app/repositories"
)

// RepositoryServiceProvider handles repository registration
type RepositoryServiceProvider struct{}

// NewRepositoryServiceProvider creates a new repository service provider
func NewRepositoryServiceProvider() *RepositoryServiceProvider {
	return &RepositoryServiceProvider{}
}

// Register registers all repositories with their dependencies
func (p *RepositoryServiceProvider) Register() {
	// Get database and cache instances
	db := facades.Database.GetDB()
	cache := facades.CacheInstance

	// Register user repository
	repositories.RegisterUserRepository(db, cache)
	// Register other repositories
	repositories.RegisterCategoryRepository(db)
	repositories.RegisterServiceRepository(db)
	repositories.RegisterRoleRepository(db)
	repositories.RegisterPermissionRepository(db)
}

// Boot performs any bootstrapping after registration
func (p *RepositoryServiceProvider) Boot() {
	// Any bootstrapping logic can go here
}

// RegisterRepository registers the repository service provider
func RegisterRepository() {
	provider := NewRepositoryServiceProvider()
	provider.Register()
	provider.Boot()
}
