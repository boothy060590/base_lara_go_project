package providers

import (
	facades_core "base_lara_go_project/app/core/facades"
	"base_lara_go_project/app/repositories"

	"gorm.io/gorm"
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
	dbInterface := facades_core.Database.GetDB()
	db := dbInterface.(*gorm.DB)
	cache := facades_core.CacheInstance

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
