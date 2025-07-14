package repositories

import (
	app_core "base_lara_go_project/app/core/go_core"
	"base_lara_go_project/app/models"
)

// UserRepository provides data access for users using canonical Repository[T] interface
type UserRepository struct {
	repository app_core.Repository[models.User]
	cache      app_core.Cache[models.User]
}

// NewUserRepository creates a new user repository using canonical Repository[T]
func NewUserRepository(repository app_core.Repository[models.User], cache app_core.Cache[models.User]) *UserRepository {
	return &UserRepository{
		repository: repository,
		cache:      cache,
	}
}

// Find retrieves a user by ID with automatic caching
func (r *UserRepository) Find(id uint) (*models.User, error) {
	return r.repository.Find(id)
}

// FindByEmail retrieves a user by email
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	// Use the canonical repository Where method
	query := r.repository.Where(map[string]any{"email": email})
	users, err := query.Get()
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, nil
	}
	return &users[0], nil
}

// Create creates a new user with automatic cache invalidation
func (r *UserRepository) Create(user *models.User) error {
	return r.repository.Create(user)
}

// Update updates an existing user with automatic cache invalidation
func (r *UserRepository) Update(user *models.User) error {
	return r.repository.Update(user)
}

// Delete deletes a user with automatic cache invalidation
func (r *UserRepository) Delete(id uint) error {
	return r.repository.Delete(id)
}

// SoftDelete soft deletes a user
func (r *UserRepository) SoftDelete(id uint) error {
	// Use canonical repository Delete method (soft delete handled by model)
	return r.repository.Delete(id)
}

// FindAll retrieves all users with pagination
func (r *UserRepository) FindAll(page, perPage int) ([]models.User, int64, error) {
	// Use canonical repository Where method with pagination
	query := r.repository.Where(map[string]any{})
	return query.Paginate(page, perPage)
}

// FindByRole retrieves users by role
func (r *UserRepository) FindByRole(roleName string) ([]models.User, error) {
	// Use canonical repository Where method
	query := r.repository.Where(map[string]any{"role": roleName})
	return query.Get()
}

// FindActive retrieves active users
func (r *UserRepository) FindActive() ([]models.User, error) {
	// Use canonical repository Where method
	query := r.repository.Where(map[string]any{"status": "active"})
	return query.Get()
}

// Count returns the total number of users
func (r *UserRepository) Count() (int64, error) {
	return r.repository.Count()
}

// Exists checks if a user exists
func (r *UserRepository) Exists(id uint) (bool, error) {
	user, err := r.Find(id)
	return user != nil, err
}

// ExistsByEmail checks if a user exists by email
func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
	user, err := r.FindByEmail(email)
	return user != nil, err
}
