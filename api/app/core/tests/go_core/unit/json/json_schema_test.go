package go_core

import (
	"base_lara_go_project/app/core/go_core"
	"base_lara_go_project/app/core/laravel_core/providers"
	"base_lara_go_project/config"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

// Test schema configuration
func TestJSONSchemaConfig(t *testing.T) {
	// Test default configuration
	config := go_core.GetDefaultJSONSchemaConfig()
	if !config.Enabled {
		t.Error("Default configuration should be enabled")
	}
	
	if len(config.Schemas) == 0 {
		t.Error("Default configuration should have schemas")
	}
	
	// The basic configuration from the service provider should have at least these core schemas
	coreSchemas := []string{"api_success", "api_error"}
	for _, schemaName := range coreSchemas {
		if _, exists := config.GetSchemaByName(schemaName); !exists {
			t.Errorf("Default configuration should have %s schema", schemaName)
		}
	}
	
	// Log available schemas for debugging
	var schemaNames []string
	for name := range config.Schemas {
		schemaNames = append(schemaNames, name)
	}
	t.Logf("Available schemas: %v", schemaNames)
}

// Test schema manager
func TestJSONSchemaManager(t *testing.T) {
	config := go_core.GetDefaultJSONSchemaConfig()
	manager := go_core.NewJSONSchemaManager(config)
	
	// Test schema registration
	for name, schema := range config.Schemas {
		if err := manager.RegisterSchema(name, schema); err != nil {
			t.Errorf("Failed to register schema '%s': %v", name, err)
		}
	}
	
	// Test schema retrieval
	schema, exists := manager.GetSchema("api_success")
	if !exists {
		t.Error("api_success schema should exist")
	}
	if schema.Name != "API Success Response" {
		t.Errorf("Expected 'API Success Response', got '%s'", schema.Name)
	}
	
	// Test schema rendering
	data := map[string]interface{}{
		"message": "Test message",
		"data":    map[string]interface{}{"key": "value"},
	}
	
	result, err := manager.RenderSchemaOptimized("api_success", data)
	if err != nil {
		t.Errorf("Failed to render schema: %v", err)
	}
	
	// Parse and validate result
	var parsed map[string]interface{}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Errorf("Failed to parse rendered JSON: %v", err)
	}
	
	if parsed["success"] != true {
		t.Error("Expected success to be true")
	}
	
	if parsed["message"] != "Test message" {
		t.Errorf("Expected message 'Test message', got '%v'", parsed["message"])
	}
}

// Test optimized schema rendering
func TestOptimizedSchemaRendering(t *testing.T) {
	config := go_core.GetDefaultJSONSchemaConfig()
	manager := go_core.NewJSONSchemaManager(config)
	
	// Register schemas
	for name, schema := range config.Schemas {
		if err := manager.RegisterSchema(name, schema); err != nil {
			t.Errorf("Failed to register schema '%s': %v", name, err)
		}
	}
	
	// Test optimized rendering for common schemas
	testCases := []struct {
		schema string
		data   map[string]interface{}
	}{
		{
			schema: "api_success",
			data:   map[string]interface{}{"message": "Success", "data": map[string]interface{}{"id": 1}},
		},
		{
			schema: "api_error",
			data:   map[string]interface{}{"message": "Error", "errors": map[string]interface{}{"field": "error"}},
		},
		{
			schema: "health_check",
			data:   map[string]interface{}{"status": "ok", "timestamp": time.Now().Unix(), "service": "test", "server": "test"},
		},
		{
			schema: "login_success",
			data:   map[string]interface{}{"message": "Login successful", "user": map[string]interface{}{"id": 1}, "token": "abc123"},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.schema, func(t *testing.T) {
			result, err := manager.RenderSchemaOptimized(tc.schema, tc.data)
			if err != nil {
				t.Errorf("Failed to render schema '%s': %v", tc.schema, err)
			}
			
			// Validate JSON is valid
			var parsed map[string]interface{}
			if err := json.Unmarshal(result, &parsed); err != nil {
				t.Errorf("Failed to parse rendered JSON for schema '%s': %v", tc.schema, err)
			}
			
			t.Logf("Schema '%s' rendered: %s", tc.schema, string(result))
		})
	}
}

// Test schema validation
func TestSchemaValidation(t *testing.T) {
	config := go_core.GetDefaultJSONSchemaConfig()
	manager := go_core.NewJSONSchemaManager(config)
	
	// Register schemas
	for name, schema := range config.Schemas {
		if err := manager.RegisterSchema(name, schema); err != nil {
			t.Errorf("Failed to register schema '%s': %v", name, err)
		}
	}
	
	// Test valid data - api_success schema requires success, message, and optional data
	validData := map[string]interface{}{
		"success": true,
		"message": "Test message",
		"data":    map[string]interface{}{"key": "value"},
	}
	
	if err := manager.ValidateSchemaData("api_success", validData); err != nil {
		t.Errorf("Valid data should pass validation: %v", err)
	}
	
	// Test invalid data (missing required field)
	invalidData := map[string]interface{}{
		"data": map[string]interface{}{"key": "value"},
		// Missing required "success" and "message" fields
	}
	
	if err := manager.ValidateSchemaData("api_success", invalidData); err == nil {
		t.Error("Invalid data should fail validation")
	}
}

// Test config loader
func TestConfigLoader(t *testing.T) {
	// Test loading from main config
	configData := config.JSONOptimizationsConfig()
	if configData == nil {
		t.Error("Config data should not be nil")
	}
	
	// Test enabled flag
	if enabled, ok := configData["enabled"].(bool); !ok || !enabled {
		t.Error("Config should be enabled")
	}
	
	// Test schemas section
	if schemas, ok := configData["schemas"].(map[string]interface{}); !ok || len(schemas) == 0 {
		t.Error("Config should have schemas")
	}
}

// Test JSON service provider
func TestJSONServiceProvider(t *testing.T) {
	// Initialize service provider
	provider := providers.NewJSONServiceProvider()
	
	// Test registration
	if err := provider.Register(); err != nil {
		t.Errorf("Failed to register JSON service provider: %v", err)
	}
	
	// Test boot
	if err := provider.Boot(); err != nil {
		t.Errorf("Failed to boot JSON service provider: %v", err)
	}
	
	// Test configuration retrieval
	config := provider.GetConfiguration()
	if config == nil {
		t.Error("Configuration should not be nil")
	}
	
	if !config.Enabled {
		t.Error("Configuration should be enabled")
	}
	
	// Test schema operations
	schemas := provider.GetSchemaList()
	if len(schemas) == 0 {
		t.Error("Should have schemas")
	}
	
	// Test metrics
	if provider.IsEnabled() {
		metrics := provider.GetMetrics()
		if metrics == nil {
			t.Error("Metrics should not be nil when enabled")
		}
	}
}

// Test global schema functions
func TestGlobalSchemaFunctions(t *testing.T) {
	// Initialize global components
	go_core.InitializeGlobalJSONProcessor(nil)
	go_core.InitializeGlobalResponsePoolManager()
	
	// Create and set schema manager
	config := go_core.GetDefaultJSONSchemaConfig()
	manager := go_core.NewJSONSchemaManager(config)
	
	// Register schemas
	for name, schema := range config.Schemas {
		if err := manager.RegisterSchema(name, schema); err != nil {
			t.Errorf("Failed to register schema '%s': %v", name, err)
		}
	}
	
	go_core.SetGlobalSchemaManager(manager)
	
	// Test global schema rendering
	data := map[string]interface{}{
		"message": "Global test",
		"data":    map[string]interface{}{"test": true},
	}
	
	result, err := go_core.RenderGlobalSchema("api_success", data)
	if err != nil {
		t.Errorf("Failed to render global schema: %v", err)
	}
	
	// Validate result
	var parsed map[string]interface{}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Errorf("Failed to parse global schema result: %v", err)
	}
	
	if parsed["success"] != true {
		t.Error("Expected success to be true")
	}
	
	if parsed["message"] != "Global test" {
		t.Errorf("Expected message 'Global test', got '%v'", parsed["message"])
	}
	
	// Test global schema validation with complete data
	validationData := map[string]interface{}{
		"success": true,
		"message": "Global test",
		"data":    map[string]interface{}{"test": true},
	}
	if err := go_core.ValidateGlobalSchemaData("api_success", validationData); err != nil {
		t.Errorf("Global schema validation failed: %v", err)
	}
	
	// Test global schema list
	schemaList := go_core.GetGlobalSchemaList()
	if len(schemaList) == 0 {
		t.Error("Global schema list should not be empty")
	}
}

// Test performance with schema optimization
func TestSchemaPerformance(t *testing.T) {
	config := go_core.GetDefaultJSONSchemaConfig()
	manager := go_core.NewJSONSchemaManager(config)
	
	// Register schemas
	for name, schema := range config.Schemas {
		if err := manager.RegisterSchema(name, schema); err != nil {
			t.Errorf("Failed to register schema '%s': %v", name, err)
		}
	}
	
	// Test data
	data := map[string]interface{}{
		"message": "Performance test",
		"data":    map[string]interface{}{"benchmark": true},
	}
	
	// Benchmark schema rendering
	iterations := 1000
	start := time.Now()
	
	for i := 0; i < iterations; i++ {
		if _, err := manager.RenderSchemaOptimized("api_success", data); err != nil {
			t.Errorf("Schema rendering failed on iteration %d: %v", i, err)
		}
	}
	
	duration := time.Since(start)
	avgTime := duration / time.Duration(iterations)
	
	t.Logf("Schema rendering performance: %d iterations in %v (avg: %v per operation)", 
		iterations, duration, avgTime)
	
	// Get metrics
	metrics := manager.GetMetrics()
	if metrics.TotalRenders != int64(iterations) {
		t.Errorf("Expected %d renders, got %d", iterations, metrics.TotalRenders)
	}
	
	if metrics.RenderErrors > 0 {
		t.Errorf("Expected 0 render errors, got %d", metrics.RenderErrors)
	}
}

// Test concurrent schema operations
func TestConcurrentSchemaOperations(t *testing.T) {
	config := go_core.GetDefaultJSONSchemaConfig()
	manager := go_core.NewJSONSchemaManager(config)
	
	// Register schemas
	for name, schema := range config.Schemas {
		if err := manager.RegisterSchema(name, schema); err != nil {
			t.Errorf("Failed to register schema '%s': %v", name, err)
		}
	}
	
	// Test concurrent rendering
	done := make(chan bool)
	errors := make(chan error, 100)
	
	for i := 0; i < 100; i++ {
		go func(id int) {
			defer func() { done <- true }()
			
			data := map[string]interface{}{
				"message": fmt.Sprintf("Concurrent test %d", id),
				"data":    map[string]interface{}{"id": id},
			}
			
			if _, err := manager.RenderSchemaOptimized("api_success", data); err != nil {
				errors <- err
			}
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 100; i++ {
		<-done
	}
	
	close(errors)
	
	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent rendering error: %v", err)
	}
	
	// Check metrics
	metrics := manager.GetMetrics()
	if metrics.TotalRenders != 100 {
		t.Errorf("Expected 100 renders, got %d", metrics.TotalRenders)
	}
}

// Helper function to run all schema tests
func TestJSONSchemaSystem(t *testing.T) {
	t.Run("Config", TestJSONSchemaConfig)
	t.Run("Manager", TestJSONSchemaManager)
	t.Run("OptimizedRendering", TestOptimizedSchemaRendering)
	t.Run("Validation", TestSchemaValidation)
	t.Run("ConfigLoader", TestConfigLoader)
	t.Run("ServiceProvider", TestJSONServiceProvider)
	t.Run("GlobalFunctions", TestGlobalSchemaFunctions)
	t.Run("Performance", TestSchemaPerformance)
	t.Run("Concurrency", TestConcurrentSchemaOperations)
}