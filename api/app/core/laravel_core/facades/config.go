package facades_core

import (
	config_core "base_lara_go_project/app/core/laravel_core/config"
)

var configInstance *config_core.ConfigFacade

// Config returns the config facade instance
func Config() *config_core.ConfigFacade {
	if configInstance == nil {
		configInstance = &config_core.ConfigFacade{}
	}
	return configInstance
}

// GetConfig retrieves a configuration value using dot notation
func GetConfig(key string, defaultValue ...interface{}) interface{} {
	return Config().Get(key, defaultValue...)
}

// GetString retrieves a configuration value as string
func GetString(key string, defaultValue ...string) string {
	return Config().GetString(key, defaultValue...)
}

// GetInt retrieves a configuration value as int
func GetInt(key string, defaultValue ...int) int {
	return Config().GetInt(key, defaultValue...)
}

// GetBool retrieves a configuration value as bool
func GetBool(key string, defaultValue ...bool) bool {
	return Config().GetBool(key, defaultValue...)
}

// HasConfig checks if a configuration key exists
func HasConfig(key string) bool {
	return Config().Has(key)
}

// SetConfig sets a configuration value
func SetConfig(key string, value interface{}) {
	Config().Set(key, value)
}
