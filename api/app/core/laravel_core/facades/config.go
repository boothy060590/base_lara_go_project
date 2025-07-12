package facades_core

import (
	"base_lara_go_project/config"
)

// ConfigFacade provides Laravel-style config access
type ConfigFacade struct{}

// Config returns the config facade instance
func Config() *ConfigFacade {
	return &ConfigFacade{}
}

// Get retrieves a configuration value using dot notation
func (c *ConfigFacade) Get(key string, defaultValue ...interface{}) interface{} {
	return config.Get(key, defaultValue...)
}

// GetString retrieves a configuration value as string
func (c *ConfigFacade) GetString(key string, defaultValue ...string) string {
	return config.GetString(key, defaultValue...)
}

// GetInt retrieves a configuration value as int
func (c *ConfigFacade) GetInt(key string, defaultValue ...int) int {
	return config.GetInt(key, defaultValue...)
}

// GetBool retrieves a configuration value as bool
func (c *ConfigFacade) GetBool(key string, defaultValue ...bool) bool {
	return config.GetBool(key, defaultValue...)
}

// Has checks if a configuration key exists
func (c *ConfigFacade) Has(key string) bool {
	return config.Has(key)
}

// Set sets a configuration value
func (c *ConfigFacade) Set(key string, value interface{}) {
	config.Set(key, value)
}

// GetConfig retrieves a configuration value using dot notation
func GetConfig(key string, defaultValue ...interface{}) interface{} {
	return config.Get(key, defaultValue...)
}

// GetString retrieves a configuration value as string
func GetString(key string, defaultValue ...string) string {
	return config.GetString(key, defaultValue...)
}

// GetInt retrieves a configuration value as int
func GetInt(key string, defaultValue ...int) int {
	return config.GetInt(key, defaultValue...)
}

// GetBool retrieves a configuration value as bool
func GetBool(key string, defaultValue ...bool) bool {
	return config.GetBool(key, defaultValue...)
}

// HasConfig checks if a configuration key exists
func HasConfig(key string) bool {
	return config.Has(key)
}

// SetConfig sets a configuration value
func SetConfig(key string, value interface{}) {
	config.Set(key, value)
}
