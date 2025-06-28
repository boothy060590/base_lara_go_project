package core

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

// BaseModel provides common functionality for all models
type BaseModel struct {
	gorm.Model
	BaseModelData
}

// BaseModelData holds the base model data structure
type BaseModelData struct {
	data map[string]interface{}
}

// NewBaseModel creates a new base model
func NewBaseModel() *BaseModelData {
	return &BaseModelData{
		data: make(map[string]interface{}),
	}
}

// Initialize initializes the base model data
func (b *BaseModelData) Initialize() {
	if b.data == nil {
		b.data = make(map[string]interface{})
	}
}

// Set sets a value in the base model
func (b *BaseModelData) Set(key string, value interface{}) {
	b.data[key] = value
}

// Get gets a value from the base model
func (b *BaseModelData) Get(key string) interface{} {
	return b.data[key]
}

// Has checks if a key exists in the base model
func (b *BaseModelData) Has(key string) bool {
	_, exists := b.data[key]
	return exists
}

// GetData returns all data from the base model
func (b *BaseModelData) GetData() map[string]interface{} {
	return b.data
}

// ModelInterface defines the interface for all models
type ModelInterface interface {
	GetID() uint
	GetTableName() string
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetDeletedAt() *time.Time
}

// BaseModelInterface extends ModelInterface with base functionality
type BaseModelInterface interface {
	ModelInterface
	Set(key string, value interface{})
	Get(key string) interface{}
	Has(key string) bool
	GetData() map[string]interface{}
}

// GetString retrieves a string value by key
func (b *BaseModelData) GetString(key string) string {
	if val, ok := b.data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// GetUint retrieves a uint value by key
func (b *BaseModelData) GetUint(key string) uint {
	if val, ok := b.data[key]; ok {
		if u, ok := val.(uint); ok {
			return u
		}
		// Try to convert from other numeric types
		if f, ok := val.(float64); ok {
			return uint(f)
		}
		if i, ok := val.(int); ok {
			return uint(i)
		}
	}
	return 0
}

// GetBool retrieves a bool value by key
func (b *BaseModelData) GetBool(key string) bool {
	if val, ok := b.data[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

// Fill fills the model with data from a map
func (b *BaseModelData) Fill(data map[string]interface{}) {
	for key, value := range data {
		b.Set(key, value)
	}
}

// ToMap converts the model to a map
func (b *BaseModelData) ToMap() map[string]interface{} {
	return b.data
}

// Magic getter/setter using reflection for struct fields
func (b *BaseModelData) GetField(fieldName string) interface{} {
	// Convert field name to snake_case for database columns
	dbField := toSnakeCase(fieldName)
	return b.Get(dbField)
}

func (b *BaseModelData) SetField(fieldName string, value interface{}) {
	// Convert field name to snake_case for database columns
	dbField := toSnakeCase(fieldName)
	b.Set(dbField, value)
}

// Helper function to convert camelCase to snake_case
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// Magic getter for common fields
func (b *BaseModelData) GetID() uint {
	return b.GetUint("id")
}

func (b *BaseModelData) GetEmail() string {
	return b.GetString("email")
}

func (b *BaseModelData) GetFirstName() string {
	return b.GetString("first_name")
}

func (b *BaseModelData) GetLastName() string {
	return b.GetString("last_name")
}

// Dynamic field access using reflection
func (b *BaseModelData) GetAttribute(name string) interface{} {
	return b.Get(name)
}

func (b *BaseModelData) SetAttribute(name string, value interface{}) {
	b.Set(name, value)
}

// Laravel-style accessors
func (b *BaseModelData) GetFullName() string {
	firstName := b.GetFirstName()
	lastName := b.GetLastName()
	if firstName != "" && lastName != "" {
		return firstName + " " + lastName
	}
	return firstName + lastName
}
