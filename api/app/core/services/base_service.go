package services_core

import (
	"context"
	"fmt"
	"reflect"

	app_core "base_lara_go_project/app/core/app"
)

// BaseService provides a generic implementation of BaseServiceInterface
type BaseService[T any] struct {
	repository app_core.RepositoryInterface
	cache      app_core.CacheInterface
}

// NewBaseService creates a new base service
func NewBaseService[T any](repository app_core.RepositoryInterface, cache app_core.CacheInterface) *BaseService[T] {
	return &BaseService[T]{
		repository: repository,
		cache:      cache,
	}
}

// Create creates a new entity
func (s *BaseService[T]) Create(data map[string]interface{}) (T, error) {
	// This would need to be implemented by specific services
	// as the repository interface doesn't support generic types
	var result T
	return result, fmt.Errorf("Create method not implemented in base service")
}

// CreateWithContext creates a new entity with context
func (s *BaseService[T]) CreateWithContext(ctx context.Context, data map[string]interface{}) (T, error) {
	return s.Create(data) // Repository doesn't support context yet
}

// FindByID finds an entity by ID
func (s *BaseService[T]) FindByID(id uint) (T, error) {
	// This would need to be implemented by specific services
	// as the repository interface doesn't support generic types
	var result T
	return result, fmt.Errorf("FindByID method not implemented in base service")
}

// FindByIDWithContext finds an entity by ID with context
func (s *BaseService[T]) FindByIDWithContext(ctx context.Context, id uint) (T, error) {
	return s.FindByID(id) // Repository doesn't support context yet
}

// FindByField finds an entity by field
func (s *BaseService[T]) FindByField(field string, value interface{}) (T, error) {
	// This would need to be implemented by specific services
	// as the repository interface doesn't support generic types
	var result T
	return result, fmt.Errorf("FindByField method not implemented in base service")
}

// FindByFieldWithContext finds an entity by field with context
func (s *BaseService[T]) FindByFieldWithContext(ctx context.Context, field string, value interface{}) (T, error) {
	return s.FindByField(field, value) // Repository doesn't support context yet
}

// All gets all entities
func (s *BaseService[T]) All() ([]T, error) {
	// This would need to be implemented by specific services
	// as the repository interface doesn't support generic types
	return nil, fmt.Errorf("All method not implemented in base service")
}

// AllWithContext gets all entities with context
func (s *BaseService[T]) AllWithContext(ctx context.Context) ([]T, error) {
	return s.All() // Repository doesn't support context yet
}

// Paginate gets paginated entities
func (s *BaseService[T]) Paginate(page, perPage int) ([]T, int64, error) {
	// This would need to be implemented by specific services
	// as the repository interface doesn't support generic types
	return nil, 0, fmt.Errorf("Paginate method not implemented in base service")
}

// PaginateWithContext gets paginated entities with context
func (s *BaseService[T]) PaginateWithContext(ctx context.Context, page, perPage int) ([]T, int64, error) {
	return s.Paginate(page, perPage) // Repository doesn't support context yet
}

// Update updates an entity
func (s *BaseService[T]) Update(id uint, data map[string]interface{}) (T, error) {
	// This would need to be implemented by specific services
	// as the repository interface doesn't support generic types
	var result T
	return result, fmt.Errorf("Update method not implemented in base service")
}

// UpdateWithContext updates an entity with context
func (s *BaseService[T]) UpdateWithContext(ctx context.Context, id uint, data map[string]interface{}) (T, error) {
	return s.Update(id, data) // Repository doesn't support context yet
}

// UpdateOrCreate updates or creates an entity
func (s *BaseService[T]) UpdateOrCreate(conditions map[string]interface{}, data map[string]interface{}) (T, error) {
	// This would need to be implemented by specific services
	// as the repository interface doesn't support generic types
	var result T
	return result, fmt.Errorf("UpdateOrCreate method not implemented in base service")
}

// UpdateOrCreateWithContext updates or creates an entity with context
func (s *BaseService[T]) UpdateOrCreateWithContext(ctx context.Context, conditions map[string]interface{}, data map[string]interface{}) (T, error) {
	return s.UpdateOrCreate(conditions, data) // Repository doesn't support context yet
}

// Delete deletes an entity
func (s *BaseService[T]) Delete(id uint) error {
	// This would need to be implemented by specific services
	// as the repository interface doesn't support generic types
	return fmt.Errorf("Delete method not implemented in base service")
}

// DeleteWithContext deletes an entity with context
func (s *BaseService[T]) DeleteWithContext(ctx context.Context, id uint) error {
	return s.Delete(id) // Repository doesn't support context yet
}

// DeleteWhere deletes entities by conditions
func (s *BaseService[T]) DeleteWhere(conditions map[string]interface{}) error {
	// This would need to be implemented by specific services
	// as the repository interface doesn't support generic types
	return fmt.Errorf("DeleteWhere method not implemented in base service")
}

// DeleteWhereWithContext deletes entities by conditions with context
func (s *BaseService[T]) DeleteWhereWithContext(ctx context.Context, conditions map[string]interface{}) error {
	return s.DeleteWhere(conditions) // Repository doesn't support context yet
}

// Exists checks if an entity exists
func (s *BaseService[T]) Exists(id uint) (bool, error) {
	// This would need to be implemented by specific services
	// as the repository interface doesn't support generic types
	return false, fmt.Errorf("Exists method not implemented in base service")
}

// ExistsWithContext checks if an entity exists with context
func (s *BaseService[T]) ExistsWithContext(ctx context.Context, id uint) (bool, error) {
	return s.Exists(id) // Repository doesn't support context yet
}

// Count counts all entities
func (s *BaseService[T]) Count() (int64, error) {
	// This would need to be implemented by specific services
	// as the repository interface doesn't support generic types
	return 0, fmt.Errorf("Count method not implemented in base service")
}

// CountWithContext counts all entities with context
func (s *BaseService[T]) CountWithContext(ctx context.Context) (int64, error) {
	return s.Count() // Repository doesn't support context yet
}

// CountWhere counts entities by conditions
func (s *BaseService[T]) CountWhere(conditions map[string]interface{}) (int64, error) {
	// This would need to be implemented by specific services
	// as the repository interface doesn't support generic types
	return 0, fmt.Errorf("CountWhere method not implemented in base service")
}

// CountWhereWithContext counts entities by conditions with context
func (s *BaseService[T]) CountWhereWithContext(ctx context.Context, conditions map[string]interface{}) (int64, error) {
	return s.CountWhere(conditions) // Repository doesn't support context yet
}

// GetCacheKey generates a cache key for an entity
func (s *BaseService[T]) GetCacheKey(id uint) string {
	return fmt.Sprintf("%s:%d", reflect.TypeOf(*new(T)).Name(), id)
}

// GetCacheKeyByField generates a cache key for an entity by field
func (s *BaseService[T]) GetCacheKeyByField(field string, value interface{}) string {
	return fmt.Sprintf("%s:%s:%v", reflect.TypeOf(*new(T)).Name(), field, value)
}
