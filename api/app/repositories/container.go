package repositories

import (
	"base_lara_go_project/app/core"
	"base_lara_go_project/app/models/interfaces"
	"sync"

	"gorm.io/gorm"
)

// RepositoryContainer holds all registered repositories
type RepositoryContainer struct {
	repositories map[string]interface{}
	mutex        sync.RWMutex
}

// NewRepositoryContainer creates a new repository container
func NewRepositoryContainer() *RepositoryContainer {
	return &RepositoryContainer{
		repositories: make(map[string]interface{}),
	}
}

// Register registers a repository with a name
func (rc *RepositoryContainer) Register(name string, repository interface{}) {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()
	rc.repositories[name] = repository
}

// Get retrieves a repository by name
func (rc *RepositoryContainer) Get(name string) (interface{}, bool) {
	rc.mutex.RLock()
	defer rc.mutex.RUnlock()
	repository, exists := rc.repositories[name]
	return repository, exists
}

// GetUserRepository retrieves the user repository
func (rc *RepositoryContainer) GetUserRepository() (*UserRepository, bool) {
	if repo, exists := rc.Get("user"); exists {
		if userRepo, ok := repo.(*UserRepository); ok {
			return userRepo, true
		}
	}
	return nil, false
}

// Global repository container instance
var GlobalRepositoryContainer = NewRepositoryContainer()

// RegisterUserRepository registers the user repository with dependencies
func RegisterUserRepository(db *gorm.DB, cache core.CacheInterface) {
	userRepo := NewUserRepository(db, cache)
	GlobalRepositoryContainer.Register("user", userRepo)
}

// GetUserRepository is a global helper to get the user repository
func GetUserRepository() (*UserRepository, bool) {
	return GlobalRepositoryContainer.GetUserRepository()
}

// UserRepositoryInterface defines the interface for user repository operations
type UserRepositoryInterface interface {
	FindByID(id uint) (interfaces.UserInterface, error)
	FindByEmail(email string) (interfaces.UserInterface, error)
	Create(userData map[string]interface{}) (interfaces.UserInterface, error)
	Update(id uint, userData map[string]interface{}) (interfaces.UserInterface, error)
	Delete(id uint) error
}

// Ensure UserRepository implements UserRepositoryInterface
var _ UserRepositoryInterface = (*UserRepository)(nil)

// RegisterCategoryRepository registers the category repository
func RegisterCategoryRepository(db *gorm.DB) {
	categoryRepo := NewCategoryRepository(db)
	GlobalRepositoryContainer.Register("category", categoryRepo)
}

// GetCategoryRepository is a global helper to get the category repository
func GetCategoryRepository() (*CategoryRepository, bool) {
	if repo, exists := GlobalRepositoryContainer.Get("category"); exists {
		if categoryRepo, ok := repo.(*CategoryRepository); ok {
			return categoryRepo, true
		}
	}
	return nil, false
}

// RegisterServiceRepository registers the service repository
func RegisterServiceRepository(db *gorm.DB) {
	serviceRepo := NewServiceRepository(db)
	GlobalRepositoryContainer.Register("service", serviceRepo)
}

// GetServiceRepository is a global helper to get the service repository
func GetServiceRepository() (*ServiceRepository, bool) {
	if repo, exists := GlobalRepositoryContainer.Get("service"); exists {
		if serviceRepo, ok := repo.(*ServiceRepository); ok {
			return serviceRepo, true
		}
	}
	return nil, false
}

// RegisterRoleRepository registers the role repository
func RegisterRoleRepository(db *gorm.DB) {
	roleRepo := NewRoleRepository(db)
	GlobalRepositoryContainer.Register("role", roleRepo)
}

// GetRoleRepository is a global helper to get the role repository
func GetRoleRepository() (*RoleRepository, bool) {
	if repo, exists := GlobalRepositoryContainer.Get("role"); exists {
		if roleRepo, ok := repo.(*RoleRepository); ok {
			return roleRepo, true
		}
	}
	return nil, false
}

// RegisterPermissionRepository registers the permission repository
func RegisterPermissionRepository(db *gorm.DB) {
	permissionRepo := NewPermissionRepository(db)
	GlobalRepositoryContainer.Register("permission", permissionRepo)
}

// GetPermissionRepository is a global helper to get the permission repository
func GetPermissionRepository() (*PermissionRepository, bool) {
	if repo, exists := GlobalRepositoryContainer.Get("permission"); exists {
		if permissionRepo, ok := repo.(*PermissionRepository); ok {
			return permissionRepo, true
		}
	}
	return nil, false
}
