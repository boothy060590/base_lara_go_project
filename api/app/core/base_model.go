package core

import (
	"strings"
)

// BaseModel provides Laravel Eloquent-style magic getters and setters
type BaseModel struct {
	attributes map[string]interface{}
}

// NewBaseModel creates a new base model
func NewBaseModel() *BaseModel {
	return &BaseModel{
		attributes: make(map[string]interface{}),
	}
}

// Get retrieves a value by key (magic getter)
func (m *BaseModel) Get(key string) interface{} {
	return m.attributes[key]
}

// Set sets a value by key (magic setter)
func (m *BaseModel) Set(key string, value interface{}) {
	m.attributes[key] = value
}

// GetString retrieves a string value by key
func (m *BaseModel) GetString(key string) string {
	if val, ok := m.attributes[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// GetUint retrieves a uint value by key
func (m *BaseModel) GetUint(key string) uint {
	if val, ok := m.attributes[key]; ok {
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
func (m *BaseModel) GetBool(key string) bool {
	if val, ok := m.attributes[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

// Fill fills the model with data from a map
func (m *BaseModel) Fill(data map[string]interface{}) {
	for key, value := range data {
		m.Set(key, value)
	}
}

// ToMap converts the model to a map
func (m *BaseModel) ToMap() map[string]interface{} {
	return m.attributes
}

// Magic getter/setter using reflection for struct fields
func (m *BaseModel) GetField(fieldName string) interface{} {
	// Convert field name to snake_case for database columns
	dbField := toSnakeCase(fieldName)
	return m.Get(dbField)
}

func (m *BaseModel) SetField(fieldName string, value interface{}) {
	// Convert field name to snake_case for database columns
	dbField := toSnakeCase(fieldName)
	m.Set(dbField, value)
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
func (m *BaseModel) GetID() uint {
	return m.GetUint("id")
}

func (m *BaseModel) GetEmail() string {
	return m.GetString("email")
}

func (m *BaseModel) GetFirstName() string {
	return m.GetString("first_name")
}

func (m *BaseModel) GetLastName() string {
	return m.GetString("last_name")
}

// Dynamic field access using reflection
func (m *BaseModel) GetAttribute(name string) interface{} {
	return m.Get(name)
}

func (m *BaseModel) SetAttribute(name string, value interface{}) {
	m.Set(name, value)
}

// Laravel-style accessors
func (m *BaseModel) GetFullName() string {
	firstName := m.GetFirstName()
	lastName := m.GetLastName()
	if firstName != "" && lastName != "" {
		return firstName + " " + lastName
	}
	return firstName + lastName
}
