package repositories

import (
	app_core "base_lara_go_project/app/core/go_core"
	"base_lara_go_project/app/models"
)

// CategoryRepository provides data access for categories using canonical Repository[T] interface
type CategoryRepository struct {
	repository app_core.Repository[models.Category]
	cache      app_core.Cache[models.Category]
}

// NewCategoryRepository creates a new category repository using canonical Repository[T]
func NewCategoryRepository(repository app_core.Repository[models.Category], cache app_core.Cache[models.Category]) *CategoryRepository {
	return &CategoryRepository{
		repository: repository,
		cache:      cache,
	}
}

// Find retrieves a category by ID with automatic caching
func (r *CategoryRepository) Find(id uint) (*models.Category, error) {
	return r.repository.Find(id)
}

// FindByName retrieves a category by name
func (r *CategoryRepository) FindByName(name string) (*models.Category, error) {
	query := r.repository.Where(map[string]any{"name": name})
	categories, err := query.Get()
	if err != nil {
		return nil, err
	}
	if len(categories) == 0 {
		return nil, nil
	}
	return &categories[0], nil
}

// Create creates a new category with automatic cache invalidation
func (r *CategoryRepository) Create(category *models.Category) error {
	return r.repository.Create(category)
}

// Update updates an existing category with automatic cache invalidation
func (r *CategoryRepository) Update(category *models.Category) error {
	return r.repository.Update(category)
}

// Delete deletes a category with automatic cache invalidation
func (r *CategoryRepository) Delete(id uint) error {
	return r.repository.Delete(id)
}

// FindAll retrieves all categories with pagination
func (r *CategoryRepository) FindAll(page, perPage int) ([]models.Category, int64, error) {
	query := r.repository.Where(map[string]any{})
	return query.Paginate(page, perPage)
}

// Count returns the total number of categories
func (r *CategoryRepository) Count() (int64, error) {
	return r.repository.Count()
}

// Exists checks if a category exists
func (r *CategoryRepository) Exists(id uint) (bool, error) {
	return r.repository.Exists(id)
}

// ExistsByName checks if a category exists by name
func (r *CategoryRepository) ExistsByName(name string) (bool, error) {
	category, err := r.FindByName(name)
	return category != nil, err
}
