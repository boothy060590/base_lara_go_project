package repositories

import (
	app_core "base_lara_go_project/app/core/go_core"
	"base_lara_go_project/app/models"
	"time"

	"gorm.io/gorm"
)

// UserRepository provides data access for users using the new generic model system
type UserRepository struct {
	model *app_core.BaseModel[models.User]
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB, cache app_core.Cache[models.User]) *UserRepository {
	config := app_core.ModelConfig{
		TableName: "users",
		Traits: app_core.ModelTraits{
			Cacheable:   true,
			SoftDeletes: true,
			HasRoles:    true,
			Timestamps:  true,
		},
		CacheTTL:    30 * time.Minute,
		CachePrefix: "user",
	}

	return &UserRepository{
		model: app_core.NewBaseModel[models.User](db, cache, config),
	}
}

// Find retrieves a user by ID with automatic caching
func (r *UserRepository) Find(id uint) (*models.User, error) {
	return r.model.Find(id)
}

// FindByEmail retrieves a user by email
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	// Use the generic Where method
	users, err := r.model.Where("email = ?", email).Get()
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
	return r.model.Create(user)
}

// Update updates an existing user with automatic cache invalidation
func (r *UserRepository) Update(user *models.User) error {
	return r.model.Update(user)
}

// Delete deletes a user with automatic cache invalidation
func (r *UserRepository) Delete(id uint) error {
	return r.model.Delete(id)
}

// SoftDelete soft deletes a user
func (r *UserRepository) SoftDelete(id uint) error {
	return r.model.SoftDelete(id)
}

// FindAll retrieves all users with pagination
func (r *UserRepository) FindAll(page, perPage int) ([]models.User, int64, error) {
	return r.model.Paginate(page, perPage)
}

// FindByRole retrieves users by role
func (r *UserRepository) FindByRole(roleName string) ([]models.User, error) {
	// This would need a custom query implementation
	// For now, return empty slice
	return []models.User{}, nil
}

// FindActive retrieves active users
func (r *UserRepository) FindActive() ([]models.User, error) {
	// This would need a custom query implementation
	// For now, return empty slice
	return []models.User{}, nil
}

// Count returns the total number of users
func (r *UserRepository) Count() (int64, error) {
	// TODO: Implement count method in BaseModel
	return 0, nil
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
