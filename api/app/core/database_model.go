package core

import (
	"time"

	"gorm.io/gorm"
)

// DatabaseModel provides database functionality for models
type DatabaseModel struct {
	gorm.Model
	BaseModelData
}

// NewDatabaseModel creates a new database model
func NewDatabaseModel() *DatabaseModel {
	return &DatabaseModel{
		BaseModelData: *NewBaseModel(),
	}
}

// DatabaseModelInterface defines the interface for database models
type DatabaseModelInterface interface {
	BaseModelInterface
	GetTableName() string
	GetPrimaryKey() string
	GetConnection() string
}

// GetTableName returns the table name (should be overridden)
func (d *DatabaseModel) GetTableName() string {
	return ""
}

// GetPrimaryKey returns the primary key (defaults to "id")
func (d *DatabaseModel) GetPrimaryKey() string {
	return "id"
}

// GetConnection returns the database connection (defaults to "default")
func (d *DatabaseModel) GetConnection() string {
	return "default"
}

// GetCreatedAt returns the created at time
func (d *DatabaseModel) GetCreatedAt() time.Time {
	return d.Model.CreatedAt
}

// GetUpdatedAt returns the updated at time
func (d *DatabaseModel) GetUpdatedAt() time.Time {
	return d.Model.UpdatedAt
}

// GetDeletedAt returns the deleted at time
func (d *DatabaseModel) GetDeletedAt() *time.Time {
	if d.Model.DeletedAt.Valid {
		return &d.Model.DeletedAt.Time
	}
	return nil
}

// GetID returns the model ID
func (d *DatabaseModel) GetID() uint {
	return d.Model.ID
}

// SetID sets the model ID
func (d *DatabaseModel) SetID(id uint) {
	d.Model.ID = id
}

// IsNew checks if the model is new (has no ID)
func (d *DatabaseModel) IsNew() bool {
	return d.Model.ID == 0
}

// Exists checks if the model exists in the database
func (d *DatabaseModel) Exists() bool {
	return !d.IsNew()
}

// BeforeCreate is a GORM hook that runs before creation
func (d *DatabaseModel) BeforeCreate(tx *gorm.DB) error {
	// Initialize the data map if it's nil
	if d.data == nil {
		d.data = make(map[string]interface{})
	}

	// Populate base model data from GORM model
	d.Set("id", d.Model.ID)
	d.Set("created_at", d.Model.CreatedAt)
	d.Set("updated_at", d.Model.UpdatedAt)
	if d.Model.DeletedAt.Valid {
		d.Set("deleted_at", d.Model.DeletedAt.Time)
	}
	return nil
}

// AfterFind is a GORM hook that runs after finding
func (d *DatabaseModel) AfterFind(tx *gorm.DB) error {
	// Initialize the data map if it's nil
	if d.data == nil {
		d.data = make(map[string]interface{})
	}

	// Populate base model data from GORM model
	d.Set("id", d.Model.ID)
	d.Set("created_at", d.Model.CreatedAt)
	d.Set("updated_at", d.Model.UpdatedAt)
	if d.Model.DeletedAt.Valid {
		d.Set("deleted_at", d.Model.DeletedAt.Time)
	}
	return nil
}

// AfterCreate is a GORM hook that runs after creation
func (d *DatabaseModel) AfterCreate(tx *gorm.DB) error {
	return d.AfterFind(tx)
}

// AfterUpdate is a GORM hook that runs after update
func (d *DatabaseModel) AfterUpdate(tx *gorm.DB) error {
	return d.AfterFind(tx)
}

// AfterDelete is a GORM hook that runs after deletion
func (d *DatabaseModel) AfterDelete(tx *gorm.DB) error {
	return d.AfterFind(tx)
}
