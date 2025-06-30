package go_core

import (
	"context"

	"gorm.io/gorm"
)

// Repository defines a generic repository interface for any model type
type Repository[T any] interface {
	// Basic CRUD operations
	Find(id uint) (*T, error)
	FindBy(field string, value any) (*T, error)
	Create(model *T) error
	Update(model *T) error
	Delete(id uint) error

	// Query operations
	Where(conditions map[string]any) Query[T]
	WhereRaw(query string, args ...any) Query[T]

	// Transaction support
	Transaction(fn func(Repository[T]) error) error
	WithContext(ctx context.Context) Repository[T]

	// Utility operations
	Exists(id uint) (bool, error)
	Count() (int64, error)
	CountWhere(conditions map[string]any) (int64, error)
}

// Query defines a generic query builder interface
type Query[T any] interface {
	// Execution
	Get() ([]T, error)
	First() (*T, error)
	Paginate(page, perPage int) ([]T, int64, error)

	// Query building
	Where(field string, operator string, value any) Query[T]
	WhereIn(field string, values []any) Query[T]
	OrderBy(field string, direction string) Query[T]
	Limit(limit int) Query[T]
	Offset(offset int) Query[T]
	Preload(relation string) Query[T]

	// Context
	WithContext(ctx context.Context) Query[T]
}

// repository implements Repository[T] with GORM
type repository[T any] struct {
	db *gorm.DB
}

// NewRepository creates a new repository instance
func NewRepository[T any](db *gorm.DB) Repository[T] {
	return &repository[T]{db: db}
}

// Find retrieves a model by ID
func (r *repository[T]) Find(id uint) (*T, error) {
	var model T
	err := r.db.First(&model, id).Error
	if err != nil {
		return nil, err
	}
	return &model, nil
}

// FindBy retrieves a model by field and value
func (r *repository[T]) FindBy(field string, value any) (*T, error) {
	var model T
	err := r.db.Where(field+" = ?", value).First(&model).Error
	if err != nil {
		return nil, err
	}
	return &model, nil
}

// Create saves a new model
func (r *repository[T]) Create(model *T) error {
	return r.db.Create(model).Error
}

// Update saves an existing model
func (r *repository[T]) Update(model *T) error {
	return r.db.Save(model).Error
}

// Delete removes a model by ID
func (r *repository[T]) Delete(id uint) error {
	var model T
	return r.db.Delete(&model, id).Error
}

// Where creates a query with conditions
func (r *repository[T]) Where(conditions map[string]any) Query[T] {
	query := r.db
	for field, value := range conditions {
		query = query.Where(field+" = ?", value)
	}
	return &queryBuilder[T]{db: query}
}

// WhereRaw creates a query with raw SQL
func (r *repository[T]) WhereRaw(query string, args ...any) Query[T] {
	return &queryBuilder[T]{db: r.db.Where(query, args...)}
}

// Transaction executes a function within a database transaction
func (r *repository[T]) Transaction(fn func(Repository[T]) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		txRepo := &repository[T]{db: tx}
		return fn(txRepo)
	})
}

// WithContext returns a repository with context
func (r *repository[T]) WithContext(ctx context.Context) Repository[T] {
	return &repository[T]{db: r.db.WithContext(ctx)}
}

// Exists checks if a model exists by ID
func (r *repository[T]) Exists(id uint) (bool, error) {
	var count int64
	err := r.db.Model(new(T)).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// Count returns the total number of models
func (r *repository[T]) Count() (int64, error) {
	var count int64
	err := r.db.Model(new(T)).Count(&count).Error
	return count, err
}

// CountWhere returns the count with conditions
func (r *repository[T]) CountWhere(conditions map[string]any) (int64, error) {
	var count int64
	query := r.db.Model(new(T))
	for field, value := range conditions {
		query = query.Where(field+" = ?", value)
	}
	err := query.Count(&count).Error
	return count, err
}

// queryBuilder implements Query[T] with GORM
type queryBuilder[T any] struct {
	db *gorm.DB
}

// Get retrieves all matching models
func (q *queryBuilder[T]) Get() ([]T, error) {
	var models []T
	err := q.db.Find(&models).Error
	return models, err
}

// First retrieves the first matching model
func (q *queryBuilder[T]) First() (*T, error) {
	var model T
	err := q.db.First(&model).Error
	if err != nil {
		return nil, err
	}
	return &model, nil
}

// Paginate retrieves models with pagination
func (q *queryBuilder[T]) Paginate(page, perPage int) ([]T, int64, error) {
	var models []T
	var total int64

	// Count total
	err := q.db.Model(new(T)).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * perPage
	err = q.db.Offset(offset).Limit(perPage).Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	return models, total, nil
}

// Where adds a where clause
func (q *queryBuilder[T]) Where(field string, operator string, value any) Query[T] {
	return &queryBuilder[T]{db: q.db.Where(field+" "+operator+" ?", value)}
}

// WhereIn adds a where in clause
func (q *queryBuilder[T]) WhereIn(field string, values []any) Query[T] {
	return &queryBuilder[T]{db: q.db.Where(field+" IN ?", values)}
}

// OrderBy adds an order clause
func (q *queryBuilder[T]) OrderBy(field string, direction string) Query[T] {
	return &queryBuilder[T]{db: q.db.Order(field + " " + direction)}
}

// Limit adds a limit clause
func (q *queryBuilder[T]) Limit(limit int) Query[T] {
	return &queryBuilder[T]{db: q.db.Limit(limit)}
}

// Offset adds an offset clause
func (q *queryBuilder[T]) Offset(offset int) Query[T] {
	return &queryBuilder[T]{db: q.db.Offset(offset)}
}

// Preload adds a preload clause
func (q *queryBuilder[T]) Preload(relation string) Query[T] {
	return &queryBuilder[T]{db: q.db.Preload(relation)}
}

// WithContext returns a query with context
func (q *queryBuilder[T]) WithContext(ctx context.Context) Query[T] {
	return &queryBuilder[T]{db: q.db.WithContext(ctx)}
}
