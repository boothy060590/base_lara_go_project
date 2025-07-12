package config

import (
	go_core "base_lara_go_project/app/core/go_core"
	"sync"
)

// Global config loader instance with thread-safe initialization
var (
	globalLoader *go_core.ConfigLoader
	loaderOnce   sync.Once
)

// GetConfigLoader returns the global config loader instance
// This ensures thread-safe singleton access to the config loader
func GetConfigLoader() *go_core.ConfigLoader {
	loaderOnce.Do(func() {
		globalLoader = go_core.GetGlobalConfigLoader()
	})
	return globalLoader
}

// Get retrieves a configuration value using dot notation
// Examples: config.Get("app.name"), config.Get("database.connections.mysql.host")
func Get(key string, defaultValue ...interface{}) interface{} {
	loader := GetConfigLoader()
	return loader.Get(key, defaultValue...)
}

// GetString retrieves a configuration value as string
// Examples: config.GetString("app.name"), config.GetString("app.name", "default")
func GetString(key string, defaultValue ...string) string {
	loader := GetConfigLoader()
	return loader.GetString(key, defaultValue...)
}

// GetInt retrieves a configuration value as int
// Examples: config.GetInt("app.port"), config.GetInt("app.port", 8080)
func GetInt(key string, defaultValue ...int) int {
	loader := GetConfigLoader()
	return loader.GetInt(key, defaultValue...)
}

// GetBool retrieves a configuration value as bool
// Examples: config.GetBool("app.debug"), config.GetBool("app.debug", false)
func GetBool(key string, defaultValue ...bool) bool {
	loader := GetConfigLoader()
	return loader.GetBool(key, defaultValue...)
}

// Has checks if a configuration key exists
// Returns true if the key exists and has a non-nil value
func Has(key string) bool {
	loader := GetConfigLoader()
	return loader.Has(key)
}

// Set sets a configuration value using dot notation
// Examples: config.Set("app.name", "MyApp"), config.Set("database.host", "localhost")
func Set(key string, value interface{}) {
	loader := GetConfigLoader()
	loader.Set(key, value)
}

// Load loads a specific configuration file by name
// Returns the entire config map for the specified config name
func Load(configName string) (map[string]interface{}, error) {
	loader := GetConfigLoader()
	return loader.Load(configName)
}

// ClearCache clears the configuration cache
// Forces reload of all configs on next access
func ClearCache() {
	loader := GetConfigLoader()
	loader.ClearCache()
}

// Reload reloads a specific configuration
// Clears the cache for the specified config and reloads it
func Reload(configName string) error {
	loader := GetConfigLoader()
	return loader.Reload(configName)
}

// ListAvailableConfigs returns a list of available configuration files
// Returns all registered config names that can be accessed
func ListAvailableConfigs() []string {
	loader := GetConfigLoader()
	return loader.ListAvailableConfigs()
}
