package repositories_core

import (
	"fmt"
	"reflect"

	app_core "base_lara_go_project/app/core/app"
)

// Repository provides base repository functionality
type Repository struct {
	modelType reflect.Type
	model     app_core.ModelInterface
}

// NewRepository creates a new repository
func NewRepository(model app_core.ModelInterface) *Repository {
	return &Repository{
		modelType: reflect.TypeOf(model),
		model:     model,
	}
}

// Find finds a model by ID
func (r *Repository) Find(id uint) (app_core.ModelInterface, error) {
	// This should be overridden by specific repository implementations
	return nil, fmt.Errorf("Find method not implemented")
}

// FindBy finds a model by field and value
func (r *Repository) FindBy(field string, value interface{}) (app_core.ModelInterface, error) {
	// This should be overridden by specific repository implementations
	return nil, fmt.Errorf("FindBy method not implemented")
}

// Create creates a new model
func (r *Repository) Create(model app_core.ModelInterface) error {
	// This should be overridden by specific repository implementations
	return fmt.Errorf("Create method not implemented")
}

// Update updates an existing model
func (r *Repository) Update(model app_core.ModelInterface) error {
	// This should be overridden by specific repository implementations
	return fmt.Errorf("Update method not implemented")
}

// Delete deletes a model
func (r *Repository) Delete(model app_core.ModelInterface) error {
	// This should be overridden by specific repository implementations
	return fmt.Errorf("Delete method not implemented")
}

// All retrieves all models
func (r *Repository) All() ([]app_core.ModelInterface, error) {
	// This should be overridden by specific repository implementations
	return nil, fmt.Errorf("All method not implemented")
}

// Where adds a where clause to the query
func (r *Repository) Where(query interface{}, args ...interface{}) app_core.RepositoryInterface {
	// This should be overridden by specific repository implementations
	return r
}

// First retrieves the first model from the query
func (r *Repository) First(dest interface{}, conds ...interface{}) error {
	// This should be overridden by specific repository implementations
	return fmt.Errorf("First method not implemented")
}

// Get retrieves all models from the query
func (r *Repository) Get() ([]app_core.ModelInterface, error) {
	// This should be overridden by specific repository implementations
	return nil, fmt.Errorf("Get method not implemented")
}

// GetModelType returns the model type
func (r *Repository) GetModelType() reflect.Type {
	return r.modelType
}

// GetModel returns the model instance
func (r *Repository) GetModel() app_core.ModelInterface {
	return r.model
}

// RepositoryContainer holds registered repositories
type RepositoryContainer struct {
	repositories map[string]app_core.RepositoryInterface
}

// NewRepositoryContainer creates a new repository container
func NewRepositoryContainer() *RepositoryContainer {
	return &RepositoryContainer{
		repositories: make(map[string]app_core.RepositoryInterface),
	}
}

// Register registers a repository
func (c *RepositoryContainer) Register(name string, repository app_core.RepositoryInterface) {
	c.repositories[name] = repository
}

// Get retrieves a repository by name
func (c *RepositoryContainer) Get(name string) (app_core.RepositoryInterface, error) {
	repository, exists := c.repositories[name]
	if !exists {
		return nil, fmt.Errorf("repository %s not found", name)
	}
	return repository, nil
}

// Has checks if a repository exists
func (c *RepositoryContainer) Has(name string) bool {
	_, exists := c.repositories[name]
	return exists
}

// Global repository container
var RepositoryContainerInstance = NewRepositoryContainer()

// RegisterRepository registers a repository globally
func RegisterRepository(name string, repository app_core.RepositoryInterface) {
	RepositoryContainerInstance.Register(name, repository)
}

// GetRepository retrieves a repository globally
func GetRepository(name string) (app_core.RepositoryInterface, error) {
	return RepositoryContainerInstance.Get(name)
}

// HasRepository checks if a repository exists globally
func HasRepository(name string) bool {
	return RepositoryContainerInstance.Has(name)
}

// FindAll retrieves all models (alias for All)
func (r *Repository) FindAll() ([]app_core.ModelInterface, error) {
	return r.All()
}

// Save saves a value
func (r *Repository) Save(value interface{}) error {
	// This should be overridden by specific repository implementations
	return fmt.Errorf("Save method not implemented")
}

// Transaction executes a function within a transaction
func (r *Repository) Transaction(fc func(tx app_core.RepositoryInterface) error) error {
	// This should be overridden by specific repository implementations
	return fmt.Errorf("Transaction method not implemented")
}

// Raw executes a raw SQL query
func (r *Repository) Raw(sql string, values ...interface{}) app_core.RepositoryInterface {
	// This should be overridden by specific repository implementations
	return r
}

// Exec executes a SQL statement
func (r *Repository) Exec(sql string, values ...interface{}) error {
	// This should be overridden by specific repository implementations
	return fmt.Errorf("Exec method not implemented")
}

// Preload preloads associations
func (r *Repository) Preload(query string, args ...interface{}) app_core.RepositoryInterface {
	// This should be overridden by specific repository implementations
	return r
}

// Order adds an order clause to the query
func (r *Repository) Order(value interface{}) app_core.RepositoryInterface {
	// This should be overridden by specific repository implementations
	return r
}

// Limit adds a limit clause to the query
func (r *Repository) Limit(limit int) app_core.RepositoryInterface {
	// This should be overridden by specific repository implementations
	return r
}

// Offset adds an offset clause to the query
func (r *Repository) Offset(offset int) app_core.RepositoryInterface {
	// This should be overridden by specific repository implementations
	return r
}

// GetDB returns the underlying database instance
func (r *Repository) GetDB() interface{} {
	// This should be overridden by specific repository implementations
	return nil
}
