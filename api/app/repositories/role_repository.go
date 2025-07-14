package repositories

import (
	app_core "base_lara_go_project/app/core/go_core"
	"base_lara_go_project/app/models"
)

// RoleRepository provides data access for roles using canonical Repository[T] interface
type RoleRepository struct {
	repository app_core.Repository[models.Role]
	cache      app_core.Cache[models.Role]
}

// NewRoleRepository creates a new role repository using canonical Repository[T]
func NewRoleRepository(repository app_core.Repository[models.Role], cache app_core.Cache[models.Role]) *RoleRepository {
	return &RoleRepository{
		repository: repository,
		cache:      cache,
	}
}

// Find retrieves a role by ID with automatic caching
func (r *RoleRepository) Find(id uint) (*models.Role, error) {
	return r.repository.Find(id)
}

// FindByName retrieves a role by name
func (r *RoleRepository) FindByName(name string) (*models.Role, error) {
	query := r.repository.Where(map[string]any{"name": name})
	roles, err := query.Get()
	if err != nil {
		return nil, err
	}
	if len(roles) == 0 {
		return nil, nil
	}
	return &roles[0], nil
}

// Create creates a new role with automatic cache invalidation
func (r *RoleRepository) Create(role *models.Role) error {
	return r.repository.Create(role)
}

// Update updates an existing role with automatic cache invalidation
func (r *RoleRepository) Update(role *models.Role) error {
	return r.repository.Update(role)
}

// Delete deletes a role with automatic cache invalidation
func (r *RoleRepository) Delete(id uint) error {
	return r.repository.Delete(id)
}

// FindAll retrieves all roles with pagination
func (r *RoleRepository) FindAll(page, perPage int) ([]models.Role, int64, error) {
	query := r.repository.Where(map[string]any{})
	return query.Paginate(page, perPage)
}

// Count returns the total number of roles
func (r *RoleRepository) Count() (int64, error) {
	return r.repository.Count()
}

// Exists checks if a role exists
func (r *RoleRepository) Exists(id uint) (bool, error) {
	return r.repository.Exists(id)
}

// ExistsByName checks if a role exists by name
func (r *RoleRepository) ExistsByName(name string) (bool, error) {
	role, err := r.FindByName(name)
	return role != nil, err
}
