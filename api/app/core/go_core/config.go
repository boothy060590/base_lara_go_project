package go_core

import (
	"strings"
	"sync"
)

// Global config registration system
var (
	globalConfigLoader *ConfigLoader
	globalLoaderOnce   sync.Once
	globalFuncMap      = make(map[string]func() map[string]interface{})
	globalFuncMapMu    sync.RWMutex
)

// RegisterGlobalConfig registers a config function globally
func RegisterGlobalConfig(configName string, configFunc func() map[string]interface{}) {
	globalFuncMapMu.Lock()
	defer globalFuncMapMu.Unlock()
	globalFuncMap[configName] = configFunc
}

// GetGlobalConfigLoader returns the global config loader instance
func GetGlobalConfigLoader() *ConfigLoader {
	globalLoaderOnce.Do(func() {
		globalConfigLoader = NewConfigLoader("api/config")
		// Register all global configs
		globalFuncMapMu.RLock()
		for name, fn := range globalFuncMap {
			globalConfigLoader.RegisterConfig(name, fn)
		}
		globalFuncMapMu.RUnlock()
	})
	return globalConfigLoader
}

// ConfigLoader provides a flexible way to load configuration files dynamically
type ConfigLoader struct {
	configDir string
	cache     map[string]map[string]interface{}
	funcMap   map[string]func() map[string]interface{}
	mu        sync.RWMutex
}

// NewConfigLoader creates a new config loader instance
func NewConfigLoader(configDir string) *ConfigLoader {
	if configDir == "" {
		configDir = "api/config"
	}

	loader := &ConfigLoader{
		configDir: configDir,
		cache:     make(map[string]map[string]interface{}),
		funcMap:   make(map[string]func() map[string]interface{}),
	}

	return loader
}

// RegisterConfig registers a config function for a given name
func (cl *ConfigLoader) RegisterConfig(configName string, configFunc func() map[string]interface{}) {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	cl.funcMap[configName] = configFunc
}

// Load loads a configuration file by name dynamically
func (cl *ConfigLoader) Load(configName string) (map[string]interface{}, error) {
	cl.mu.RLock()
	if cached, exists := cl.cache[configName]; exists {
		cl.mu.RUnlock()
		return cached, nil
	}
	cl.mu.RUnlock()

	cl.mu.Lock()
	defer cl.mu.Unlock()

	// Double-check after acquiring write lock
	if cached, exists := cl.cache[configName]; exists {
		return cached, nil
	}

	// Try to load the config dynamically
	config, err := cl.loadDynamicConfig(configName)
	if err != nil {
		return nil, err
	}

	cl.cache[configName] = config
	return config, nil
}

// loadDynamicConfig loads a config file dynamically
func (cl *ConfigLoader) loadDynamicConfig(configName string) (map[string]interface{}, error) {
	// Check if we have a registered config function
	if configFunc, exists := cl.funcMap[configName]; exists {
		return configFunc(), nil
	}
	return nil, nil
}

// Get retrieves a configuration value using dot notation
func (cl *ConfigLoader) Get(key string, defaultValue ...interface{}) interface{} {
	parts := strings.Split(key, ".")
	if len(parts) < 1 {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return nil
	}

	configName := parts[0]

	// If there's only one part, return the entire config
	if len(parts) == 1 {
		configMap, err := cl.Load(configName)
		if err != nil {
			if len(defaultValue) > 0 {
				return defaultValue[0]
			}
			return nil
		}
		return configMap
	}

	configKey := strings.Join(parts[1:], ".")

	configMap, err := cl.Load(configName)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return nil
	}

	// Navigate through the config map
	current := configMap
	keyParts := strings.Split(configKey, ".")

	for i, part := range keyParts {
		if val, ok := current[part]; ok {
			if i == len(keyParts)-1 {
				return val
			}
			if mapVal, isMap := val.(map[string]interface{}); isMap {
				current = mapVal
			} else {
				if len(defaultValue) > 0 {
					return defaultValue[0]
				}
				return nil
			}
		} else {
			if len(defaultValue) > 0 {
				return defaultValue[0]
			}
			return nil
		}
	}

	return current
}

// GetString retrieves a configuration value as string
func (cl *ConfigLoader) GetString(key string, defaultValue ...string) string {
	val := cl.Get(key)
	if str, ok := val.(string); ok {
		return str
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ""
}

// GetInt retrieves a configuration value as int
func (cl *ConfigLoader) GetInt(key string, defaultValue ...int) int {
	val := cl.Get(key)
	switch v := val.(type) {
	case int:
		return v
	case float64:
		return int(v)
	default:
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}
}

// GetBool retrieves a configuration value as bool
func (cl *ConfigLoader) GetBool(key string, defaultValue ...bool) bool {
	val := cl.Get(key)
	if b, ok := val.(bool); ok {
		return b
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return false
}

// Has checks if a configuration key exists
func (cl *ConfigLoader) Has(key string) bool {
	return cl.Get(key) != nil
}

// ClearCache clears the configuration cache
func (cl *ConfigLoader) ClearCache() {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	cl.cache = make(map[string]map[string]interface{})
}

// Reload reloads a specific configuration
func (cl *ConfigLoader) Reload(configName string) error {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	delete(cl.cache, configName)
	_, err := cl.Load(configName)
	return err
}

// ListAvailableConfigs returns a list of available configuration files
func (cl *ConfigLoader) ListAvailableConfigs() []string {
	var configs []string

	// Return all discovered config names
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	for configName := range cl.funcMap {
		configs = append(configs, configName)
	}

	return configs
}
