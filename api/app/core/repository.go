package core

import (
	"fmt"
	"reflect"
)

// RepositoryInterface defines the interface for repositories
type RepositoryInterface interface {
	// Basic CRUD operations
	Find(id uint) (ModelInterface, error)
	FindBy(field string, value interface{}) (ModelInterface, error)
	Create(model ModelInterface) error
	Update(model ModelInterface) error
	Delete(model ModelInterface) error
	All() ([]ModelInterface, error)
	Where(query interface{}, args ...interface{}) RepositoryInterface
	First() (ModelInterface, error)
	Get() ([]ModelInterface, error)
}

// Repository provides base repository functionality
type Repository struct {
	modelType reflect.Type
	model     ModelInterface
}

// NewRepository creates a new repository
func NewRepository(model ModelInterface) *Repository {
	return &Repository{
		modelType: reflect.TypeOf(model),
		model:     model,
	}
}

// Find finds a model by ID
func (r *Repository) Find(id uint) (ModelInterface, error) {
	// This should be overridden by specific repository implementations
	return nil, fmt.Errorf("Find method not implemented")
}

// FindBy finds a model by field and value
func (r *Repository) FindBy(field string, value interface{}) (ModelInterface, error) {
	// This should be overridden by specific repository implementations
	return nil, fmt.Errorf("FindBy method not implemented")
}

// Create creates a new model
func (r *Repository) Create(model ModelInterface) error {
	// This should be overridden by specific repository implementations
	return fmt.Errorf("Create method not implemented")
}

// Update updates an existing model
func (r *Repository) Update(model ModelInterface) error {
	// This should be overridden by specific repository implementations
	return fmt.Errorf("Update method not implemented")
}

// Delete deletes a model
func (r *Repository) Delete(model ModelInterface) error {
	// This should be overridden by specific repository implementations
	return fmt.Errorf("Delete method not implemented")
}

// All retrieves all models
func (r *Repository) All() ([]ModelInterface, error) {
	// This should be overridden by specific repository implementations
	return nil, fmt.Errorf("All method not implemented")
}

// Where adds a where clause to the query
func (r *Repository) Where(query interface{}, args ...interface{}) RepositoryInterface {
	// This should be overridden by specific repository implementations
	return r
}

// First retrieves the first model from the query
func (r *Repository) First() (ModelInterface, error) {
	// This should be overridden by specific repository implementations
	return nil, fmt.Errorf("First method not implemented")
}

// Get retrieves all models from the query
func (r *Repository) Get() ([]ModelInterface, error) {
	// This should be overridden by specific repository implementations
	return nil, fmt.Errorf("Get method not implemented")
}

// GetModelType returns the model type
func (r *Repository) GetModelType() reflect.Type {
	return r.modelType
}

// GetModel returns the model instance
func (r *Repository) GetModel() ModelInterface {
	return r.model
}

// RepositoryContainer holds registered repositories
type RepositoryContainer struct {
	repositories map[string]RepositoryInterface
}

// NewRepositoryContainer creates a new repository container
func NewRepositoryContainer() *RepositoryContainer {
	return &RepositoryContainer{
		repositories: make(map[string]RepositoryInterface),
	}
}

// Register registers a repository
func (c *RepositoryContainer) Register(name string, repository RepositoryInterface) {
	c.repositories[name] = repository
}

// Get retrieves a repository by name
func (c *RepositoryContainer) Get(name string) (RepositoryInterface, error) {
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
func RegisterRepository(name string, repository RepositoryInterface) {
	RepositoryContainerInstance.Register(name, repository)
}

// GetRepository retrieves a repository globally
func GetRepository(name string) (RepositoryInterface, error) {
	return RepositoryContainerInstance.Get(name)
}

// HasRepository checks if a repository exists globally
func HasRepository(name string) bool {
	return RepositoryContainerInstance.Has(name)
}
