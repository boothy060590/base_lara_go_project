package repositories

import (
	app_core "base_lara_go_project/app/core/go_core"
	"base_lara_go_project/app/models"
)

// ServiceRepository provides data access for services using canonical Repository[T] interface
type ServiceRepository struct {
	repository app_core.Repository[models.Service]
	cache      app_core.Cache[models.Service]
}

// NewServiceRepository creates a new service repository using canonical Repository[T]
func NewServiceRepository(repository app_core.Repository[models.Service], cache app_core.Cache[models.Service]) *ServiceRepository {
	return &ServiceRepository{
		repository: repository,
		cache:      cache,
	}
}

// Find retrieves a service by ID with automatic caching
func (r *ServiceRepository) Find(id uint) (*models.Service, error) {
	return r.repository.Find(id)
}

// FindByName retrieves a service by name
func (r *ServiceRepository) FindByName(name string) (*models.Service, error) {
	query := r.repository.Where(map[string]any{"name": name})
	services, err := query.Get()
	if err != nil {
		return nil, err
	}
	if len(services) == 0 {
		return nil, nil
	}
	return &services[0], nil
}

// Create creates a new service with automatic cache invalidation
func (r *ServiceRepository) Create(service *models.Service) error {
	return r.repository.Create(service)
}

// Update updates an existing service with automatic cache invalidation
func (r *ServiceRepository) Update(service *models.Service) error {
	return r.repository.Update(service)
}

// Delete deletes a service with automatic cache invalidation
func (r *ServiceRepository) Delete(id uint) error {
	return r.repository.Delete(id)
}

// FindAll retrieves all services with pagination
func (r *ServiceRepository) FindAll(page, perPage int) ([]models.Service, int64, error) {
	query := r.repository.Where(map[string]any{})
	return query.Paginate(page, perPage)
}

// Count returns the total number of services
func (r *ServiceRepository) Count() (int64, error) {
	return r.repository.Count()
}

// Exists checks if a service exists
func (r *ServiceRepository) Exists(id uint) (bool, error) {
	return r.repository.Exists(id)
}

// ExistsByName checks if a service exists by name
func (r *ServiceRepository) ExistsByName(name string) (bool, error) {
	service, err := r.FindByName(name)
	return service != nil, err
}
