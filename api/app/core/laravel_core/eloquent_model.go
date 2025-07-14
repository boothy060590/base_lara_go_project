package laravel_core

import (
	"context"
	"database/sql"
	"reflect"
	"strings"
	"time"

	app_core "base_lara_go_project/app/core/go_core"
)

// EloquentModel provides Laravel-style model functionality with raw SQL
type EloquentModel[T any] struct {
	repository  app_core.Repository[T]
	tableName   string
	fillable    []string
	hidden      []string
	timestamps  bool
	softDeletes bool
}

// NewEloquentModel creates a new Eloquent model
func NewEloquentModel[T any](db *sql.DB, wsp *app_core.WorkStealingPool[any], ca *app_core.CustomAllocator[any], pgo *app_core.ProfileGuidedOptimizer[any]) *EloquentModel[T] {
	repo := app_core.NewRepository[T](db, wsp, ca, pgo)

	return &EloquentModel[T]{
		repository:  repo,
		tableName:   getTableName[T](),
		fillable:    []string{},
		hidden:      []string{},
		timestamps:  true,
		softDeletes: true,
	}
}

// getTableName extracts table name from model type
func getTableName[T any]() string {
	var model T
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Check for TableName method
	if _, ok := t.MethodByName("TableName"); ok {
		// Call the method to get table name
		// This is a simplified version - in practice you'd need reflection to call it
		return strings.ToLower(t.Name()) + "s"
	}

	return strings.ToLower(t.Name()) + "s"
}

// TableName sets the table name for the model
func (m *EloquentModel[T]) TableName(name string) *EloquentModel[T] {
	m.tableName = name
	return m
}

// Fillable sets the fillable fields
func (m *EloquentModel[T]) Fillable(fields []string) *EloquentModel[T] {
	m.fillable = fields
	return m
}

// Hidden sets the hidden fields
func (m *EloquentModel[T]) Hidden(fields []string) *EloquentModel[T] {
	m.hidden = fields
	return m
}

// Timestamps enables/disables timestamps
func (m *EloquentModel[T]) Timestamps(enabled bool) *EloquentModel[T] {
	m.timestamps = enabled
	return m
}

// SoftDeletes enables/disables soft deletes
func (m *EloquentModel[T]) SoftDeletes(enabled bool) *EloquentModel[T] {
	m.softDeletes = enabled
	return m
}

// Find finds a model by ID
func (m *EloquentModel[T]) Find(id uint) (*T, error) {
	return m.repository.Find(id)
}

// FindWithContext finds a model by ID with context
func (m *EloquentModel[T]) FindWithContext(ctx context.Context, id uint) (*T, error) {
	return m.repository.FindWithContext(ctx, id)
}

// FindBy finds a model by field and value
func (m *EloquentModel[T]) FindBy(field string, value interface{}) (*T, error) {
	return m.repository.FindBy(field, value)
}

// FindByWithContext finds a model by field and value with context
func (m *EloquentModel[T]) FindByWithContext(ctx context.Context, field string, value interface{}) (*T, error) {
	return m.repository.FindByWithContext(ctx, field, value)
}

// All gets all models
func (m *EloquentModel[T]) All() ([]T, error) {
	return m.repository.FindAll()
}

// AllWithContext gets all models with context
func (m *EloquentModel[T]) AllWithContext(ctx context.Context) ([]T, error) {
	return m.repository.FindAllWithContext(ctx)
}

// Create creates a new model
func (m *EloquentModel[T]) Create(model *T) error {
	// Add timestamps if enabled
	if m.timestamps {
		m.setTimestamps(model, true)
	}

	return m.repository.Create(model)
}

// CreateWithContext creates a new model with context
func (m *EloquentModel[T]) CreateWithContext(ctx context.Context, model *T) error {
	// Add timestamps if enabled
	if m.timestamps {
		m.setTimestamps(model, true)
	}

	return m.repository.CreateWithContext(ctx, model)
}

// Update updates a model
func (m *EloquentModel[T]) Update(model *T) error {
	// Update timestamps if enabled
	if m.timestamps {
		m.setTimestamps(model, false)
	}

	return m.repository.Update(model)
}

// UpdateWithContext updates a model with context
func (m *EloquentModel[T]) UpdateWithContext(ctx context.Context, model *T) error {
	// Update timestamps if enabled
	if m.timestamps {
		m.setTimestamps(model, false)
	}

	return m.repository.UpdateWithContext(ctx, model)
}

// Delete deletes a model
func (m *EloquentModel[T]) Delete(id uint) error {
	return m.repository.Delete(id)
}

// DeleteWithContext deletes a model with context
func (m *EloquentModel[T]) DeleteWithContext(ctx context.Context, id uint) error {
	return m.repository.DeleteWithContext(ctx, id)
}

// Where creates a query with conditions
func (m *EloquentModel[T]) Where(conditions map[string]interface{}) app_core.Query[T] {
	return m.repository.Where(conditions)
}

// WhereWithContext creates a query with conditions and context
func (m *EloquentModel[T]) WhereWithContext(ctx context.Context, conditions map[string]interface{}) app_core.Query[T] {
	return m.repository.WhereWithContext(ctx, conditions)
}

// WhereRaw creates a query with raw SQL
func (m *EloquentModel[T]) WhereRaw(query string, args ...interface{}) app_core.Query[T] {
	return m.repository.WhereRaw(query, args...)
}

// WhereRawWithContext creates a query with raw SQL and context
func (m *EloquentModel[T]) WhereRawWithContext(ctx context.Context, query string, args ...interface{}) app_core.Query[T] {
	return m.repository.WhereRawWithContext(ctx, query, args...)
}

// Transaction executes a function within a transaction
func (m *EloquentModel[T]) Transaction(fn func(app_core.Repository[T]) error) error {
	return m.repository.Transaction(fn)
}

// TransactionWithContext executes a function within a transaction with context
func (m *EloquentModel[T]) TransactionWithContext(ctx context.Context, fn func(app_core.Repository[T]) error) error {
	return m.repository.TransactionWithContext(ctx, fn)
}

// WithContext returns a model with context
func (m *EloquentModel[T]) WithContext(ctx context.Context) *EloquentModel[T] {
	newModel := *m
	newModel.repository = m.repository.WithContext(ctx)
	return &newModel
}

// Exists checks if a model exists
func (m *EloquentModel[T]) Exists(id uint) (bool, error) {
	return m.repository.Exists(id)
}

// ExistsWithContext checks if a model exists with context
func (m *EloquentModel[T]) ExistsWithContext(ctx context.Context, id uint) (bool, error) {
	return m.repository.ExistsWithContext(ctx, id)
}

// Count returns the total number of models
func (m *EloquentModel[T]) Count() (int64, error) {
	return m.repository.Count()
}

// CountWithContext returns the total number of models with context
func (m *EloquentModel[T]) CountWithContext(ctx context.Context) (int64, error) {
	return m.repository.CountWithContext(ctx)
}

// CountWhere returns the count with conditions
func (m *EloquentModel[T]) CountWhere(conditions map[string]interface{}) (int64, error) {
	return m.repository.CountWhere(conditions)
}

// CountWhereWithContext returns the count with conditions and context
func (m *EloquentModel[T]) CountWhereWithContext(ctx context.Context, conditions map[string]interface{}) (int64, error) {
	return m.repository.CountWhereWithContext(ctx, conditions)
}

// BulkCreate creates multiple models
func (m *EloquentModel[T]) BulkCreate(models []*T) error {
	// Add timestamps if enabled
	if m.timestamps {
		for _, model := range models {
			m.setTimestamps(model, true)
		}
	}

	return m.repository.BulkCreate(models)
}

// BulkCreateWithContext creates multiple models with context
func (m *EloquentModel[T]) BulkCreateWithContext(ctx context.Context, models []*T) error {
	// Add timestamps if enabled
	if m.timestamps {
		for _, model := range models {
			m.setTimestamps(model, true)
		}
	}

	return m.repository.BulkCreateWithContext(ctx, models)
}

// BulkUpdate updates multiple models
func (m *EloquentModel[T]) BulkUpdate(models []*T) error {
	// Update timestamps if enabled
	if m.timestamps {
		for _, model := range models {
			m.setTimestamps(model, false)
		}
	}

	return m.repository.BulkUpdate(models)
}

// BulkUpdateWithContext updates multiple models with context
func (m *EloquentModel[T]) BulkUpdateWithContext(ctx context.Context, models []*T) error {
	// Update timestamps if enabled
	if m.timestamps {
		for _, model := range models {
			m.setTimestamps(model, false)
		}
	}

	return m.repository.BulkUpdateWithContext(ctx, models)
}

// BulkDelete deletes multiple models
func (m *EloquentModel[T]) BulkDelete(ids []uint) error {
	return m.repository.BulkDelete(ids)
}

// BulkDeleteWithContext deletes multiple models with context
func (m *EloquentModel[T]) BulkDeleteWithContext(ctx context.Context, ids []uint) error {
	return m.repository.BulkDeleteWithContext(ctx, ids)
}

// GetPerformanceStats returns performance statistics
func (m *EloquentModel[T]) GetPerformanceStats() map[string]interface{} {
	return m.repository.GetPerformanceStats()
}

// GetOptimizationStats returns optimization statistics
func (m *EloquentModel[T]) GetOptimizationStats() map[string]interface{} {
	return m.repository.GetOptimizationStats()
}

// setTimestamps sets created_at and updated_at timestamps
func (m *EloquentModel[T]) setTimestamps(model *T, isCreate bool) {
	v := reflect.ValueOf(model).Elem()

	now := time.Now()

	// Set created_at for new models
	if isCreate {
		if field := v.FieldByName("CreatedAt"); field.IsValid() && field.CanSet() {
			field.Set(reflect.ValueOf(now))
		}
	}

	// Set updated_at for all models
	if field := v.FieldByName("UpdatedAt"); field.IsValid() && field.CanSet() {
		field.Set(reflect.ValueOf(now))
	}
}

// GetRepository returns the underlying repository
func (m *EloquentModel[T]) GetRepository() app_core.Repository[T] {
	return m.repository
}

// GetTableName returns the table name
func (m *EloquentModel[T]) GetTableName() string {
	return m.tableName
}

// GetFillable returns the fillable fields
func (m *EloquentModel[T]) GetFillable() []string {
	return m.fillable
}

// GetHidden returns the hidden fields
func (m *EloquentModel[T]) GetHidden() []string {
	return m.hidden
}

// IsTimestampsEnabled returns whether timestamps are enabled
func (m *EloquentModel[T]) IsTimestampsEnabled() bool {
	return m.timestamps
}

// IsSoftDeletesEnabled returns whether soft deletes are enabled
func (m *EloquentModel[T]) IsSoftDeletesEnabled() bool {
	return m.softDeletes
}
