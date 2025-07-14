package repositories

import (
	app_core "base_lara_go_project/app/core/go_core"
	"base_lara_go_project/app/models"
)

// PermissionRepository provides data access for permissions using canonical Repository[T] interface
type PermissionRepository struct {
	repository app_core.Repository[models.Permission]
	cache      app_core.Cache[models.Permission]
}

// NewPermissionRepository creates a new permission repository using canonical Repository[T]
func NewPermissionRepository(repository app_core.Repository[models.Permission], cache app_core.Cache[models.Permission]) *PermissionRepository {
	return &PermissionRepository{
		repository: repository,
		cache:      cache,
	}
}

// Find retrieves a permission by ID with automatic caching
func (r *PermissionRepository) Find(id uint) (*models.Permission, error) {
	return r.repository.Find(id)
}

// FindByName retrieves a permission by name
func (r *PermissionRepository) FindByName(name string) (*models.Permission, error) {
	query := r.repository.Where(map[string]any{"name": name})
	permissions, err := query.Get()
	if err != nil {
		return nil, err
	}
	if len(permissions) == 0 {
		return nil, nil
	}
	return &permissions[0], nil
}

// Create creates a new permission with automatic cache invalidation
func (r *PermissionRepository) Create(permission *models.Permission) error {
	return r.repository.Create(permission)
}

// Update updates an existing permission with automatic cache invalidation
func (r *PermissionRepository) Update(permission *models.Permission) error {
	return r.repository.Update(permission)
}

// Delete deletes a permission with automatic cache invalidation
func (r *PermissionRepository) Delete(id uint) error {
	return r.repository.Delete(id)
}

// FindAll retrieves all permissions with pagination
func (r *PermissionRepository) FindAll(page, perPage int) ([]models.Permission, int64, error) {
	query := r.repository.Where(map[string]any{})
	return query.Paginate(page, perPage)
}

// Count returns the total number of permissions
func (r *PermissionRepository) Count() (int64, error) {
	return r.repository.Count()
}

// Exists checks if a permission exists
func (r *PermissionRepository) Exists(id uint) (bool, error) {
	return r.repository.Exists(id)
}

// ExistsByName checks if a permission exists by name
func (r *PermissionRepository) ExistsByName(name string) (bool, error) {
	permission, err := r.FindByName(name)
	return permission != nil, err
}
