package go_core

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"text/template"
	"time"
)

// JSONSchemaManager manages JSON schema patterns and optimization
type JSONSchemaManager struct {
	config           *JSONSchemaConfig
	schemas          map[string]*JSONSchemaTemplate
	templateCache    map[string]*template.Template
	mu               sync.RWMutex
	metrics          *SchemaMetrics
}

// JSONSchemaTemplate represents a compiled JSON schema template
type JSONSchemaTemplate struct {
	Name         string
	Type         string
	Template     *template.Template
	Fields       map[string]FieldConfig
	Enabled      bool
	UsageCount   int64
	LastUsed     time.Time
	AvgRenderTime time.Duration
}

// SchemaMetrics tracks schema usage metrics
type SchemaMetrics struct {
	TotalRenders     int64
	TotalRenderTime  time.Duration
	SchemaUsage      map[string]int64
	RenderErrors     int64
	CacheHits        int64
	CacheMisses      int64
	mu               sync.RWMutex
}

// NewJSONSchemaManager creates a new JSON schema manager
func NewJSONSchemaManager(config *JSONSchemaConfig) *JSONSchemaManager {
	return &JSONSchemaManager{
		config:        config,
		schemas:       make(map[string]*JSONSchemaTemplate),
		templateCache: make(map[string]*template.Template),
		metrics: &SchemaMetrics{
			SchemaUsage: make(map[string]int64),
		},
	}
}

// RegisterSchema registers a new schema pattern
func (jsm *JSONSchemaManager) RegisterSchema(name string, schema JSONSchemaPattern) error {
	jsm.mu.Lock()
	defer jsm.mu.Unlock()
	
	// Validate schema
	if err := schema.ValidateSchema(); err != nil {
		return fmt.Errorf("invalid schema '%s': %w", name, err)
	}
	
	// Parse template
	tmpl, err := template.New(name).Parse(schema.Template)
	if err != nil {
		return fmt.Errorf("failed to parse template for schema '%s': %w", name, err)
	}
	
	// Create schema template
	schemaTemplate := &JSONSchemaTemplate{
		Name:     schema.Name,
		Type:     schema.Type,
		Template: tmpl,
		Fields:   schema.Fields,
		Enabled:  schema.Enabled,
	}
	
	jsm.schemas[name] = schemaTemplate
	jsm.templateCache[name] = tmpl
	
	return nil
}

// UnregisterSchema removes a schema
func (jsm *JSONSchemaManager) UnregisterSchema(name string) {
	jsm.mu.Lock()
	defer jsm.mu.Unlock()
	
	delete(jsm.schemas, name)
	delete(jsm.templateCache, name)
}

// GetSchema retrieves a schema by name
func (jsm *JSONSchemaManager) GetSchema(name string) (*JSONSchemaTemplate, bool) {
	jsm.mu.RLock()
	defer jsm.mu.RUnlock()
	
	schema, exists := jsm.schemas[name]
	if !exists || !schema.Enabled {
		return nil, false
	}
	
	return schema, true
}

// RenderSchema renders a schema with the provided data
func (jsm *JSONSchemaManager) RenderSchema(name string, data interface{}) ([]byte, error) {
	start := time.Now()
	
	// Get schema
	schema, exists := jsm.GetSchema(name)
	if !exists {
		jsm.updateMetrics(name, start, true)
		return nil, fmt.Errorf("schema '%s' not found or not enabled", name)
	}
	
	// Render template
	var buf bytes.Buffer
	if err := schema.Template.Execute(&buf, data); err != nil {
		jsm.updateMetrics(name, start, true)
		return nil, fmt.Errorf("failed to render schema '%s': %w", name, err)
	}
	
	// Update metrics
	jsm.updateMetrics(name, start, false)
	jsm.updateSchemaUsage(name)
	
	return buf.Bytes(), nil
}

// RenderSchemaToWriter renders a schema directly to a writer
func (jsm *JSONSchemaManager) RenderSchemaToWriter(name string, data interface{}, writer io.Writer) error {
	start := time.Now()
	
	// Get schema
	schema, exists := jsm.GetSchema(name)
	if !exists {
		jsm.updateMetrics(name, start, true)
		return fmt.Errorf("schema '%s' not found or not enabled", name)
	}
	
	// Render template directly to writer
	if err := schema.Template.Execute(writer, data); err != nil {
		jsm.updateMetrics(name, start, true)
		return fmt.Errorf("failed to render schema '%s': %w", name, err)
	}
	
	// Update metrics
	jsm.updateMetrics(name, start, false)
	jsm.updateSchemaUsage(name)
	
	return nil
}

// RenderSchemaOptimized renders a schema with optimization based on data type
func (jsm *JSONSchemaManager) RenderSchemaOptimized(name string, data interface{}) ([]byte, error) {
	// Try to use pre-compiled patterns for common data types
	switch name {
	case "api_success":
		return jsm.renderAPISuccess(data)
	case "api_error":
		return jsm.renderAPIError(data)
	case "health_check":
		return jsm.renderHealthCheck(data)
	case "login_success":
		return jsm.renderLoginSuccess(data)
	default:
		return jsm.RenderSchema(name, data)
	}
}

// renderAPISuccess renders API success response with optimization
func (jsm *JSONSchemaManager) renderAPISuccess(data interface{}) ([]byte, error) {
	start := time.Now()
	
	// Try to cast to known structure
	if apiData, ok := data.(map[string]interface{}); ok {
		message := getString(apiData, "message", "Success")
		dataField := apiData["data"]
		
		// Build JSON manually for better performance
		var buf bytes.Buffer
		buf.WriteString(`{"success":true,"message":"`)
		buf.WriteString(message)
		buf.WriteString(`","data":`)
		
		if dataField != nil {
			// Use JSON processor for data field
			if jsonBytes, err := EncodeJSON(dataField); err == nil {
				buf.Write(jsonBytes)
			} else {
				buf.WriteString("null")
			}
		} else {
			buf.WriteString("null")
		}
		
		buf.WriteString("}")
		
		jsm.updateMetrics("api_success", start, false)
		jsm.updateSchemaUsage("api_success")
		
		return buf.Bytes(), nil
	}
	
	// Fall back to template rendering
	return jsm.RenderSchema("api_success", data)
}

// renderAPIError renders API error response with optimization
func (jsm *JSONSchemaManager) renderAPIError(data interface{}) ([]byte, error) {
	start := time.Now()
	
	if apiData, ok := data.(map[string]interface{}); ok {
		message := getString(apiData, "message", "Error")
		errors := apiData["errors"]
		
		var buf bytes.Buffer
		buf.WriteString(`{"success":false,"message":"`)
		buf.WriteString(message)
		buf.WriteString(`","errors":`)
		
		if errors != nil {
			if jsonBytes, err := EncodeJSON(errors); err == nil {
				buf.Write(jsonBytes)
			} else {
				buf.WriteString("null")
			}
		} else {
			buf.WriteString("null")
		}
		
		buf.WriteString("}")
		
		jsm.updateMetrics("api_error", start, false)
		jsm.updateSchemaUsage("api_error")
		
		return buf.Bytes(), nil
	}
	
	return jsm.RenderSchema("api_error", data)
}

// renderHealthCheck renders health check response with optimization
func (jsm *JSONSchemaManager) renderHealthCheck(data interface{}) ([]byte, error) {
	start := time.Now()
	
	if healthData, ok := data.(map[string]interface{}); ok {
		status := getString(healthData, "status", "ok")
		service := getString(healthData, "service", "unknown")
		server := getString(healthData, "server", "unknown")
		timestamp := getInt64(healthData, "timestamp", time.Now().Unix())
		
		var buf bytes.Buffer
		buf.WriteString(`{"status":"`)
		buf.WriteString(status)
		buf.WriteString(`","timestamp":`)
		buf.WriteString(fmt.Sprintf("%d", timestamp))
		buf.WriteString(`,"service":"`)
		buf.WriteString(service)
		buf.WriteString(`","server":"`)
		buf.WriteString(server)
		buf.WriteString(`"}`)
		
		jsm.updateMetrics("health_check", start, false)
		jsm.updateSchemaUsage("health_check")
		
		return buf.Bytes(), nil
	}
	
	return jsm.RenderSchema("health_check", data)
}

// renderLoginSuccess renders login success response with optimization
func (jsm *JSONSchemaManager) renderLoginSuccess(data interface{}) ([]byte, error) {
	start := time.Now()
	
	if loginData, ok := data.(map[string]interface{}); ok {
		message := getString(loginData, "message", "Login successful")
		token := getString(loginData, "token", "")
		user := loginData["user"]
		
		var buf bytes.Buffer
		buf.WriteString(`{"message":"`)
		buf.WriteString(message)
		buf.WriteString(`","user":`)
		
		if user != nil {
			if jsonBytes, err := EncodeJSON(user); err == nil {
				buf.Write(jsonBytes)
			} else {
				buf.WriteString("null")
			}
		} else {
			buf.WriteString("null")
		}
		
		buf.WriteString(`,"token":"`)
		buf.WriteString(token)
		buf.WriteString(`"}`)
		
		jsm.updateMetrics("login_success", start, false)
		jsm.updateSchemaUsage("login_success")
		
		return buf.Bytes(), nil
	}
	
	return jsm.RenderSchema("login_success", data)
}

// Helper functions
func getString(data map[string]interface{}, key, defaultValue string) string {
	if value, exists := data[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

func getInt64(data map[string]interface{}, key string, defaultValue int64) int64 {
	if value, exists := data[key]; exists {
		switch v := value.(type) {
		case int64:
			return v
		case int:
			return int64(v)
		case float64:
			return int64(v)
		}
	}
	return defaultValue
}

// updateMetrics updates schema rendering metrics
func (jsm *JSONSchemaManager) updateMetrics(schemaName string, start time.Time, isError bool) {
	jsm.metrics.mu.Lock()
	defer jsm.metrics.mu.Unlock()
	
	renderTime := time.Since(start)
	jsm.metrics.TotalRenders++
	jsm.metrics.TotalRenderTime += renderTime
	
	if isError {
		jsm.metrics.RenderErrors++
	}
	
	// Update schema-specific metrics
	if schema, exists := jsm.schemas[schemaName]; exists {
		schema.UsageCount++
		schema.LastUsed = time.Now()
		schema.AvgRenderTime = (schema.AvgRenderTime + renderTime) / 2
	}
}

// updateSchemaUsage updates schema usage count
func (jsm *JSONSchemaManager) updateSchemaUsage(schemaName string) {
	jsm.metrics.mu.Lock()
	defer jsm.metrics.mu.Unlock()
	
	jsm.metrics.SchemaUsage[schemaName]++
}

// GetMetrics returns schema manager metrics
func (jsm *JSONSchemaManager) GetMetrics() SchemaMetrics {
	jsm.metrics.mu.RLock()
	defer jsm.metrics.mu.RUnlock()
	
	// Return a copy to avoid lock issues
	return SchemaMetrics{
		TotalRenders:    jsm.metrics.TotalRenders,
		TotalRenderTime: jsm.metrics.TotalRenderTime,
		RenderErrors:    jsm.metrics.RenderErrors,
		CacheHits:       jsm.metrics.CacheHits,
		CacheMisses:     jsm.metrics.CacheMisses,
		SchemaUsage:     copyUsageMap(jsm.metrics.SchemaUsage),
	}
}

// copyUsageMap creates a copy of the usage map
func copyUsageMap(original map[string]int64) map[string]int64 {
	copy := make(map[string]int64)
	for k, v := range original {
		copy[k] = v
	}
	return copy
}

// GetSchemaList returns a list of all registered schemas
func (jsm *JSONSchemaManager) GetSchemaList() []string {
	jsm.mu.RLock()
	defer jsm.mu.RUnlock()
	
	var schemas []string
	for name, schema := range jsm.schemas {
		if schema.Enabled {
			schemas = append(schemas, name)
		}
	}
	return schemas
}

// GetSchemasByType returns schemas of a specific type
func (jsm *JSONSchemaManager) GetSchemasByType(schemaType string) []string {
	jsm.mu.RLock()
	defer jsm.mu.RUnlock()
	
	var schemas []string
	for name, schema := range jsm.schemas {
		if schema.Type == schemaType && schema.Enabled {
			schemas = append(schemas, name)
		}
	}
	return schemas
}

// EnableSchema enables a schema
func (jsm *JSONSchemaManager) EnableSchema(name string) error {
	jsm.mu.Lock()
	defer jsm.mu.Unlock()
	
	if schema, exists := jsm.schemas[name]; exists {
		schema.Enabled = true
		return nil
	}
	
	return fmt.Errorf("schema '%s' not found", name)
}

// DisableSchema disables a schema
func (jsm *JSONSchemaManager) DisableSchema(name string) error {
	jsm.mu.Lock()
	defer jsm.mu.Unlock()
	
	if schema, exists := jsm.schemas[name]; exists {
		schema.Enabled = false
		return nil
	}
	
	return fmt.Errorf("schema '%s' not found", name)
}

// ValidateSchemaData validates data against a schema
func (jsm *JSONSchemaManager) ValidateSchemaData(schemaName string, data interface{}) error {
	schema, exists := jsm.GetSchema(schemaName)
	if !exists {
		return fmt.Errorf("schema '%s' not found", schemaName)
	}
	
	// Basic validation against field definitions
	if dataMap, ok := data.(map[string]interface{}); ok {
		for fieldName, fieldConfig := range schema.Fields {
			value, exists := dataMap[fieldName]
			
			// Check required fields
			if fieldConfig.Required && !exists {
				return fmt.Errorf("required field '%s' is missing", fieldName)
			}
			
			// Type validation (basic)
			if exists && value != nil {
				if err := jsm.validateFieldType(fieldName, value, fieldConfig); err != nil {
					return err
				}
			}
		}
	}
	
	return nil
}

// validateFieldType validates a field type
func (jsm *JSONSchemaManager) validateFieldType(fieldName string, value interface{}, config FieldConfig) error {
	switch config.Type {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("field '%s' must be a string", fieldName)
		}
	case "number":
		switch value.(type) {
		case int, int64, float64:
			// Valid number types
		default:
			return fmt.Errorf("field '%s' must be a number", fieldName)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("field '%s' must be a boolean", fieldName)
		}
	case "array":
		if _, ok := value.([]interface{}); !ok {
			return fmt.Errorf("field '%s' must be an array", fieldName)
		}
	case "object":
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Errorf("field '%s' must be an object", fieldName)
		}
	}
	
	return nil
}

// GetSchemaInfo returns detailed information about a schema
func (jsm *JSONSchemaManager) GetSchemaInfo(name string) (*JSONSchemaTemplate, error) {
	schema, exists := jsm.GetSchema(name)
	if !exists {
		return nil, fmt.Errorf("schema '%s' not found", name)
	}
	
	return schema, nil
}

// ResetMetrics resets all metrics
func (jsm *JSONSchemaManager) ResetMetrics() {
	jsm.metrics.mu.Lock()
	defer jsm.metrics.mu.Unlock()
	
	jsm.metrics.TotalRenders = 0
	jsm.metrics.TotalRenderTime = 0
	jsm.metrics.RenderErrors = 0
	jsm.metrics.CacheHits = 0
	jsm.metrics.CacheMisses = 0
	jsm.metrics.SchemaUsage = make(map[string]int64)
}

// Global schema manager
var GlobalSchemaManager *JSONSchemaManager

// SetGlobalSchemaManager sets the global schema manager
func SetGlobalSchemaManager(manager *JSONSchemaManager) {
	GlobalSchemaManager = manager
}

// GetGlobalSchemaManager returns the global schema manager
func GetGlobalSchemaManager() *JSONSchemaManager {
	return GlobalSchemaManager
}

// Helper functions for global schema manager

// RenderGlobalSchema renders a schema using the global manager
func RenderGlobalSchema(name string, data interface{}) ([]byte, error) {
	if GlobalSchemaManager == nil {
		return nil, fmt.Errorf("global schema manager not initialized")
	}
	return GlobalSchemaManager.RenderSchemaOptimized(name, data)
}

// RenderGlobalSchemaToWriter renders a schema to a writer using the global manager
func RenderGlobalSchemaToWriter(name string, data interface{}, writer io.Writer) error {
	if GlobalSchemaManager == nil {
		return fmt.Errorf("global schema manager not initialized")
	}
	return GlobalSchemaManager.RenderSchemaToWriter(name, data, writer)
}

// ValidateGlobalSchemaData validates data against a global schema
func ValidateGlobalSchemaData(schemaName string, data interface{}) error {
	if GlobalSchemaManager == nil {
		return fmt.Errorf("global schema manager not initialized")
	}
	return GlobalSchemaManager.ValidateSchemaData(schemaName, data)
}

// GetGlobalSchemaList returns a list of all global schemas
func GetGlobalSchemaList() []string {
	if GlobalSchemaManager == nil {
		return nil
	}
	return GlobalSchemaManager.GetSchemaList()
}

// GetGlobalSchemaMetrics returns global schema metrics
func GetGlobalSchemaMetrics() SchemaMetrics {
	if GlobalSchemaManager == nil {
		return SchemaMetrics{}
	}
	return GlobalSchemaManager.GetMetrics()
}