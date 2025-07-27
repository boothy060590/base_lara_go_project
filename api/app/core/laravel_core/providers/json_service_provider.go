package providers

import (
	"base_lara_go_project/app/core/go_core"
	app_core "base_lara_go_project/app/core/go_core"
	"base_lara_go_project/config"
	"fmt"
	"log"
	"time"
)

// JSONServiceProvider registers JSON optimization services
type JSONServiceProvider struct {
	config *go_core.JSONSchemaConfig
}

// NewJSONServiceProvider creates a new JSON service provider
func NewJSONServiceProvider() *JSONServiceProvider {
	return &JSONServiceProvider{}
}

// Register registers the JSON optimization services
func (jsp *JSONServiceProvider) Register(container *app_core.Container) error {
	// Load configuration
	if err := jsp.loadConfiguration(); err != nil {
		log.Printf("Failed to load JSON configuration: %v", err)
		// Fall back to defaults
		jsp.config = go_core.GetDefaultJSONSchemaConfig()
	}

	// Initialize JSON processor with config
	if err := jsp.initializeJSONProcessor(); err != nil {
		return fmt.Errorf("failed to initialize JSON processor: %w", err)
	}

	// Initialize response pools
	if err := jsp.initializeResponsePools(); err != nil {
		return fmt.Errorf("failed to initialize response pools: %w", err)
	}

	// Initialize JSON schema manager
	if err := jsp.initializeSchemaManager(); err != nil {
		return fmt.Errorf("failed to initialize schema manager: %w", err)
	}

	// Initialize metrics collection
	if err := jsp.initializeMetrics(); err != nil {
		return fmt.Errorf("failed to initialize metrics: %w", err)
	}

	log.Println("JSON optimization services registered successfully")
	return nil
}

// Boot boots the JSON optimization services
func (jsp *JSONServiceProvider) Boot(container *app_core.Container) error {
	// Start metrics monitoring if enabled
	if jsp.config != nil && jsp.config.Metrics.Enabled {
		go_core.StartGlobalJSONMetricsMonitor()
		log.Println("JSON metrics monitoring started")
	}

	// Register schema templates
	if err := jsp.registerSchemaTemplates(); err != nil {
		return fmt.Errorf("failed to register schema templates: %w", err)
	}

	log.Println("JSON optimization services booted successfully")
	return nil
}

func (jsp *JSONServiceProvider) Provides() []string {
	return []string{"json_optimizations"}
}

func (jsp *JSONServiceProvider) When() []string {
	return []string{}
}

// loadConfiguration loads the JSON schema configuration
func (jsp *JSONServiceProvider) loadConfiguration() error {
	// Load from Laravel-style config system
	configData := config.Get("json_optimizations")

	// Convert config data to JSONSchemaConfig
	if configMap, ok := configData.(map[string]interface{}); ok {
		jsp.config = jsp.parseConfigMap(configMap)
	} else {
		// Fall back to default if config is not available
		jsp.config = go_core.GetDefaultJSONSchemaConfig()
	}

	return nil
}

// parseConfigMap converts config map to JSONSchemaConfig
func (jsp *JSONServiceProvider) parseConfigMap(configMap map[string]interface{}) *go_core.JSONSchemaConfig {
	config := &go_core.JSONSchemaConfig{
		Enabled: getBoolFromConfig(configMap, "enabled", true),
		Schemas: make(map[string]go_core.JSONSchemaPattern),
	}

	// Parse buffer pool config
	if bufferPoolData, ok := configMap["buffer_pool"].(map[string]interface{}); ok {
		config.BufferPool = go_core.BufferPoolConfig{
			MaxBufferSize: getIntFromConfig(bufferPoolData, "max_buffer_size", 64*1024),
			InitialSize:   getIntFromConfig(bufferPoolData, "initial_size", 4*1024),
		}
	}

	// Parse response pool config
	if responsePoolData, ok := configMap["response_pool"].(map[string]interface{}); ok {
		config.ResponsePool = go_core.ResponsePoolConfig{
			PoolSize: getIntFromConfig(responsePoolData, "pool_size", 500),
		}
	}

	// Parse metrics config
	if metricsData, ok := configMap["metrics"].(map[string]interface{}); ok {
		config.Metrics = go_core.MetricsConfig{
			Enabled:            getBoolFromConfig(metricsData, "enabled", true),
			CollectionInterval: time.Duration(getIntFromConfig(metricsData, "collection_interval", 300)) * time.Second,
			ReportInterval:     time.Duration(getIntFromConfig(metricsData, "report_interval", 900)) * time.Second,
		}
	}

	// Parse schemas
	if schemasData, ok := configMap["schemas"].(map[string]interface{}); ok {
		for schemaName, schemaData := range schemasData {
			if schemaMap, ok := schemaData.(map[string]interface{}); ok {
				schema := jsp.parseSchemaMap(schemaMap)
				config.Schemas[schemaName] = schema
			}
		}
	}

	return config
}

// parseSchemaMap converts schema map to JSONSchemaPattern
func (jsp *JSONServiceProvider) parseSchemaMap(schemaMap map[string]interface{}) go_core.JSONSchemaPattern {
	schema := go_core.JSONSchemaPattern{
		Name:        getStringFromConfig(schemaMap, "name", ""),
		Description: getStringFromConfig(schemaMap, "description", ""),
		Type:        getStringFromConfig(schemaMap, "type", "response"),
		Template:    getStringFromConfig(schemaMap, "template", "{}"),
		Enabled:     getBoolFromConfig(schemaMap, "enabled", true),
		Fields:      make(map[string]go_core.FieldConfig),
	}

	// Parse fields
	if fieldsData, ok := schemaMap["fields"].(map[string]interface{}); ok {
		for fieldName, fieldData := range fieldsData {
			if fieldMap, ok := fieldData.(map[string]interface{}); ok {
				field := go_core.FieldConfig{
					Type:         getStringFromConfig(fieldMap, "type", "string"),
					Required:     getBoolFromConfig(fieldMap, "required", false),
					DefaultValue: fieldMap["default_value"],
					Format:       getStringFromConfig(fieldMap, "format", ""),
				}
				schema.Fields[fieldName] = field
			}
		}
	}

	return schema
}

// Helper functions for config parsing
func getStringFromConfig(configMap map[string]interface{}, key, defaultValue string) string {
	if value, exists := configMap[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

func getIntFromConfig(configMap map[string]interface{}, key string, defaultValue int) int {
	if value, exists := configMap[key]; exists {
		if num, ok := value.(int); ok {
			return num
		}
		if num, ok := value.(float64); ok {
			return int(num)
		}
	}
	return defaultValue
}

func getBoolFromConfig(configMap map[string]interface{}, key string, defaultValue bool) bool {
	if value, exists := configMap[key]; exists {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return defaultValue
}

// initializeJSONProcessor initializes the global JSON processor
func (jsp *JSONServiceProvider) initializeJSONProcessor() error {
	if !jsp.config.Enabled {
		log.Println("JSON optimization is disabled")
		return nil
	}

	// Create processor config from schema config
	processorConfig := &go_core.JSONProcessorConfig{
		MaxBufferSize:    jsp.config.BufferPool.MaxBufferSize,
		InitialSize:      jsp.config.BufferPool.InitialSize,
		ResponsePoolSize: jsp.config.ResponsePool.PoolSize,
	}

	// Initialize global processor
	go_core.InitializeGlobalJSONProcessor(processorConfig)

	log.Printf("JSON processor initialized with max buffer: %d bytes", processorConfig.MaxBufferSize)
	return nil
}

// initializeResponsePools initializes the response object pools
func (jsp *JSONServiceProvider) initializeResponsePools() error {
	if !jsp.config.Enabled {
		return nil
	}

	// Initialize global response pool manager
	go_core.InitializeGlobalResponsePoolManager()

	log.Printf("Response pools initialized with pool size: %d", jsp.config.ResponsePool.PoolSize)
	return nil
}

// initializeSchemaManager initializes the schema manager
func (jsp *JSONServiceProvider) initializeSchemaManager() error {
	if !jsp.config.Enabled {
		return nil
	}

	// Create schema manager with loaded configuration
	manager := go_core.NewJSONSchemaManager(jsp.config)
	go_core.SetGlobalSchemaManager(manager)

	log.Printf("Schema manager initialized with %d schemas", len(jsp.config.Schemas))
	return nil
}

// initializeMetrics initializes the metrics collection
func (jsp *JSONServiceProvider) initializeMetrics() error {
	if !jsp.config.Enabled || !jsp.config.Metrics.Enabled {
		return nil
	}

	// Initialize global metrics
	go_core.InitializeGlobalJSONMetrics()

	log.Println("JSON metrics collection initialized")
	return nil
}

// registerSchemaTemplates registers all schema templates
func (jsp *JSONServiceProvider) registerSchemaTemplates() error {
	if jsp.config == nil {
		log.Println("JSON config is nil, using default config")
		jsp.config = go_core.GetDefaultJSONSchemaConfig()
	}
	
	if !jsp.config.Enabled {
		return nil
	}

	schemaManager := go_core.GetGlobalSchemaManager()
	if schemaManager == nil {
		return fmt.Errorf("schema manager not initialized")
	}

	// Register all configured schemas
	for name, schema := range jsp.config.GetEnabledSchemas() {
		if err := schemaManager.RegisterSchema(name, schema); err != nil {
			log.Printf("Failed to register schema '%s': %v", name, err)
			continue
		}
		log.Printf("Registered schema: %s", name)
	}

	return nil
}

// GetConfiguration returns the current configuration
func (jsp *JSONServiceProvider) GetConfiguration() *go_core.JSONSchemaConfig {
	return jsp.config
}

// UpdateConfiguration updates the configuration
func (jsp *JSONServiceProvider) UpdateConfiguration(newConfig *go_core.JSONSchemaConfig) error {
	jsp.config = newConfig

	// Reinitialize services with new configuration
	if err := jsp.initializeJSONProcessor(); err != nil {
		return err
	}

	if err := jsp.initializeResponsePools(); err != nil {
		return err
	}

	if err := jsp.initializeSchemaManager(); err != nil {
		return err
	}

	if err := jsp.registerSchemaTemplates(); err != nil {
		return err
	}

	log.Println("JSON configuration updated successfully")
	return nil
}

// SaveConfiguration saves the current configuration
// Note: In Laravel-style config system, configuration is managed through config files
// and environment variables. Runtime configuration changes are not typically persisted.
func (jsp *JSONServiceProvider) SaveConfiguration() error {
	log.Println("Configuration is managed through config files and environment variables")
	log.Println("To modify configuration, update config/json_optimizations.go or set environment variables")
	return nil
}

// AddCustomSchema adds a custom schema to the configuration
func (jsp *JSONServiceProvider) AddCustomSchema(name string, schema go_core.JSONSchemaPattern) error {
	if err := jsp.config.AddSchema(name, schema); err != nil {
		return fmt.Errorf("failed to add schema: %w", err)
	}

	// Register the schema with the schema manager
	schemaManager := go_core.GetGlobalSchemaManager()
	if schemaManager != nil {
		if err := schemaManager.RegisterSchema(name, schema); err != nil {
			return fmt.Errorf("failed to register schema: %w", err)
		}
	}

	log.Printf("Custom schema '%s' added successfully", name)
	return nil
}

// RemoveCustomSchema removes a custom schema from the configuration
func (jsp *JSONServiceProvider) RemoveCustomSchema(name string) error {
	jsp.config.RemoveSchema(name)

	// Unregister the schema from the schema manager
	schemaManager := go_core.GetGlobalSchemaManager()
	if schemaManager != nil {
		schemaManager.UnregisterSchema(name)
	}

	log.Printf("Custom schema '%s' removed successfully", name)
	return nil
}

// EnableSchema enables a schema
func (jsp *JSONServiceProvider) EnableSchema(name string) error {
	jsp.config.EnableSchema(name)

	// Re-register the schema
	if schema, exists := jsp.config.GetSchemaByName(name); exists {
		schemaManager := go_core.GetGlobalSchemaManager()
		if schemaManager != nil {
			if err := schemaManager.RegisterSchema(name, *schema); err != nil {
				return fmt.Errorf("failed to register schema: %w", err)
			}
		}
	}

	log.Printf("Schema '%s' enabled successfully", name)
	return nil
}

// DisableSchema disables a schema
func (jsp *JSONServiceProvider) DisableSchema(name string) error {
	jsp.config.DisableSchema(name)

	// Unregister the schema
	schemaManager := go_core.GetGlobalSchemaManager()
	if schemaManager != nil {
		schemaManager.UnregisterSchema(name)
	}

	log.Printf("Schema '%s' disabled successfully", name)
	return nil
}

// GetMetrics returns current JSON processing metrics
func (jsp *JSONServiceProvider) GetMetrics() map[string]interface{} {
	if !jsp.config.Enabled || !jsp.config.Metrics.Enabled {
		return nil
	}

	return go_core.GetGlobalJSONMetricsSummary()
}

// GetDetailedMetrics returns detailed JSON processing metrics
func (jsp *JSONServiceProvider) GetDetailedMetrics() go_core.JSONPerformanceReport {
	if !jsp.config.Enabled || !jsp.config.Metrics.Enabled {
		return go_core.JSONPerformanceReport{}
	}

	return go_core.GetGlobalJSONMetrics()
}

// GetSchemaList returns a list of all available schemas
func (jsp *JSONServiceProvider) GetSchemaList() []string {
	var schemas []string
	for name, schema := range jsp.config.Schemas {
		if schema.Enabled {
			schemas = append(schemas, name)
		}
	}
	return schemas
}

// GetSchemaInfo returns information about a specific schema
func (jsp *JSONServiceProvider) GetSchemaInfo(name string) (*go_core.JSONSchemaPattern, error) {
	schema, exists := jsp.config.GetSchemaByName(name)
	if !exists {
		return nil, fmt.Errorf("schema '%s' not found or not enabled", name)
	}
	return schema, nil
}

// IsEnabled returns whether JSON optimization is enabled
func (jsp *JSONServiceProvider) IsEnabled() bool {
	return jsp.config.Enabled
}

// Global JSON service provider instance
var GlobalJSONServiceProvider *JSONServiceProvider

// InitializeGlobalJSONServiceProvider initializes the global JSON service provider
func InitializeGlobalJSONServiceProvider() error {
	GlobalJSONServiceProvider = NewJSONServiceProvider()
	return GlobalJSONServiceProvider.Register(nil) // Pass nil for now, as app_core.Container is not available here
}

// BootGlobalJSONServiceProvider boots the global JSON service provider
func BootGlobalJSONServiceProvider() error {
	if GlobalJSONServiceProvider == nil {
		return fmt.Errorf("JSON service provider not initialized")
	}
	return GlobalJSONServiceProvider.Boot(nil) // Pass nil for now, as app_core.Container is not available here
}

// GetGlobalJSONServiceProvider returns the global JSON service provider
func GetGlobalJSONServiceProvider() *JSONServiceProvider {
	return GlobalJSONServiceProvider
}
