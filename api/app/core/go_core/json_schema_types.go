package go_core

import (
	"encoding/json"
	"fmt"
	"time"
)

// JSONSchemaConfig contains configuration for JSON optimization schemas
type JSONSchemaConfig struct {
	// Enable or disable JSON optimization
	Enabled bool `json:"enabled"`
	
	// Buffer pool configuration
	BufferPool BufferPoolConfig `json:"buffer_pool"`
	
	// Response pool configuration
	ResponsePool ResponsePoolConfig `json:"response_pool"`
	
	// Pre-compiled schema patterns
	Schemas map[string]JSONSchemaPattern `json:"schemas"`
	
	// Metrics configuration
	Metrics MetricsConfig `json:"metrics"`
}

// BufferPoolConfig configures the JSON buffer pool
type BufferPoolConfig struct {
	MaxBufferSize int `json:"max_buffer_size"`
	InitialSize   int `json:"initial_size"`
}

// ResponsePoolConfig configures the response object pool
type ResponsePoolConfig struct {
	PoolSize int `json:"pool_size"`
}

// JSONSchemaPattern defines a pre-compiled JSON pattern for optimization
type JSONSchemaPattern struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"` // "response", "request", "generic"
	Template    string                 `json:"template"`
	Fields      map[string]FieldConfig `json:"fields"`
	Enabled     bool                   `json:"enabled"`
}

// FieldConfig defines configuration for a JSON field
type FieldConfig struct {
	Type        string      `json:"type"`        // "string", "number", "boolean", "object", "array"
	Required    bool        `json:"required"`
	DefaultValue interface{} `json:"default_value,omitempty"`
	Format      string      `json:"format,omitempty"` // "timestamp", "uuid", etc.
}

// MetricsConfig configures JSON metrics collection
type MetricsConfig struct {
	Enabled         bool          `json:"enabled"`
	CollectionInterval time.Duration `json:"collection_interval"`
	ReportInterval  time.Duration `json:"report_interval"`
}

// GetDefaultJSONSchemaConfig returns the default JSON schema configuration
// This should only be used as a fallback when config system is not available
func GetDefaultJSONSchemaConfig() *JSONSchemaConfig {
	return &JSONSchemaConfig{
		Enabled: true,
		BufferPool: BufferPoolConfig{
			MaxBufferSize: 64 * 1024, // 64KB
			InitialSize:   4 * 1024,  // 4KB
		},
		ResponsePool: ResponsePoolConfig{
			PoolSize: 500,
		},
		Schemas: getMinimalDefaultSchemas(), // Only essential schemas
		Metrics: MetricsConfig{
			Enabled:            true,
			CollectionInterval: 5 * time.Minute,
			ReportInterval:     15 * time.Minute,
		},
	}
}

// getMinimalDefaultSchemas returns minimal default schemas for fallback
// These are only used when the config system is not available
func getMinimalDefaultSchemas() map[string]JSONSchemaPattern {
	return map[string]JSONSchemaPattern{
		"api_success": {
			Name:        "API Success Response",
			Description: "Standard successful API response (fallback)",
			Type:        "response",
			Template:    `{"success":true,"message":"{{.message}}","data":{{.data}}}`,
			Fields: map[string]FieldConfig{
				"success": {Type: "boolean", Required: true, DefaultValue: true},
				"message": {Type: "string", Required: true},
				"data":    {Type: "object", Required: false},
			},
			Enabled: true,
		},
		"api_error": {
			Name:        "API Error Response",
			Description: "Standard error API response (fallback)",
			Type:        "response",
			Template:    `{"success":false,"message":"{{.message}}","errors":{{.errors}}}`,
			Fields: map[string]FieldConfig{
				"success": {Type: "boolean", Required: true, DefaultValue: false},
				"message": {Type: "string", Required: true},
				"errors":  {Type: "object", Required: false},
			},
			Enabled: true,
		},
	}
}

// ValidateSchema validates a JSON schema pattern
func (jsp *JSONSchemaPattern) ValidateSchema() error {
	if jsp.Name == "" {
		return fmt.Errorf("schema name cannot be empty")
	}
	
	if jsp.Type == "" {
		return fmt.Errorf("schema type cannot be empty")
	}
	
	if jsp.Template == "" {
		return fmt.Errorf("schema template cannot be empty")
	}
	
	// Validate template is valid JSON (with placeholders)
	if !jsp.isValidTemplate() {
		return fmt.Errorf("schema template is not valid JSON")
	}
	
	return nil
}

// isValidTemplate checks if the template is valid JSON with placeholders
func (jsp *JSONSchemaPattern) isValidTemplate() bool {
	// Simple validation - check if it's valid JSON structure
	return len(jsp.Template) > 0 && jsp.Template[0] == '{' && jsp.Template[len(jsp.Template)-1] == '}'
}

// GetSchemaByName gets a schema by name
func (jsc *JSONSchemaConfig) GetSchemaByName(name string) (*JSONSchemaPattern, bool) {
	schema, exists := jsc.Schemas[name]
	if !exists || !schema.Enabled {
		return nil, false
	}
	return &schema, true
}

// GetSchemasByType gets all schemas of a specific type
func (jsc *JSONSchemaConfig) GetSchemasByType(schemaType string) []JSONSchemaPattern {
	var schemas []JSONSchemaPattern
	for _, schema := range jsc.Schemas {
		if schema.Type == schemaType && schema.Enabled {
			schemas = append(schemas, schema)
		}
	}
	return schemas
}

// AddSchema adds a new schema to the configuration
func (jsc *JSONSchemaConfig) AddSchema(name string, schema JSONSchemaPattern) error {
	if err := schema.ValidateSchema(); err != nil {
		return fmt.Errorf("invalid schema: %w", err)
	}
	
	if jsc.Schemas == nil {
		jsc.Schemas = make(map[string]JSONSchemaPattern)
	}
	
	jsc.Schemas[name] = schema
	return nil
}

// RemoveSchema removes a schema from the configuration
func (jsc *JSONSchemaConfig) RemoveSchema(name string) {
	if jsc.Schemas != nil {
		delete(jsc.Schemas, name)
	}
}

// EnableSchema enables a schema
func (jsc *JSONSchemaConfig) EnableSchema(name string) {
	if schema, exists := jsc.Schemas[name]; exists {
		schema.Enabled = true
		jsc.Schemas[name] = schema
	}
}

// DisableSchema disables a schema
func (jsc *JSONSchemaConfig) DisableSchema(name string) {
	if schema, exists := jsc.Schemas[name]; exists {
		schema.Enabled = false
		jsc.Schemas[name] = schema
	}
}

// LoadFromJSON loads configuration from JSON
func (jsc *JSONSchemaConfig) LoadFromJSON(data []byte) error {
	return json.Unmarshal(data, jsc)
}

// SaveToJSON saves configuration to JSON
func (jsc *JSONSchemaConfig) SaveToJSON() ([]byte, error) {
	return json.MarshalIndent(jsc, "", "  ")
}

// MergeWithDefaults merges the current configuration with defaults
func (jsc *JSONSchemaConfig) MergeWithDefaults() {
	defaults := GetDefaultJSONSchemaConfig()
	
	// Merge schemas
	if jsc.Schemas == nil {
		jsc.Schemas = make(map[string]JSONSchemaPattern)
	}
	
	for name, defaultSchema := range defaults.Schemas {
		if _, exists := jsc.Schemas[name]; !exists {
			jsc.Schemas[name] = defaultSchema
		}
	}
	
	// Set default values if not configured
	if jsc.BufferPool.MaxBufferSize == 0 {
		jsc.BufferPool = defaults.BufferPool
	}
	if jsc.ResponsePool.PoolSize == 0 {
		jsc.ResponsePool = defaults.ResponsePool
	}
	if jsc.Metrics.CollectionInterval == 0 {
		jsc.Metrics = defaults.Metrics
	}
}

// GetEnabledSchemas returns all enabled schemas
func (jsc *JSONSchemaConfig) GetEnabledSchemas() map[string]JSONSchemaPattern {
	enabled := make(map[string]JSONSchemaPattern)
	for name, schema := range jsc.Schemas {
		if schema.Enabled {
			enabled[name] = schema
		}
	}
	return enabled
}

// GetResponseSchemas returns all response schemas
func (jsc *JSONSchemaConfig) GetResponseSchemas() map[string]JSONSchemaPattern {
	responses := make(map[string]JSONSchemaPattern)
	for name, schema := range jsc.Schemas {
		if schema.Type == "response" && schema.Enabled {
			responses[name] = schema
		}
	}
	return responses
}

// GetRequestSchemas returns all request schemas
func (jsc *JSONSchemaConfig) GetRequestSchemas() map[string]JSONSchemaPattern {
	requests := make(map[string]JSONSchemaPattern)
	for name, schema := range jsc.Schemas {
		if schema.Type == "request" && schema.Enabled {
			requests[name] = schema
		}
	}
	return requests
}

// Clone creates a deep copy of the configuration
func (jsc *JSONSchemaConfig) Clone() *JSONSchemaConfig {
	clone := &JSONSchemaConfig{
		Enabled:      jsc.Enabled,
		BufferPool:   jsc.BufferPool,
		ResponsePool: jsc.ResponsePool,
		Metrics:      jsc.Metrics,
		Schemas:      make(map[string]JSONSchemaPattern),
	}
	
	for name, schema := range jsc.Schemas {
		clone.Schemas[name] = schema
	}
	
	return clone
}