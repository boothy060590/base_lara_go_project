package config_core

import (
	"base_lara_go_project/config"
	"strings"
)

// ConfigFacade provides Laravel-style config access
type ConfigFacade struct{}

// Get retrieves a configuration value using dot notation
func (c *ConfigFacade) Get(key string, defaultValue ...interface{}) interface{} {
	parts := strings.Split(key, ".")
	if len(parts) < 2 {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return nil
	}

	configName := parts[0]
	configKey := strings.Join(parts[1:], ".")

	var configMap map[string]interface{}

	switch configName {
	case "app":
		configMap = config.AppConfig()
	case "database":
		configMap = config.DatabaseConfig()
	case "queue":
		configMap = config.QueueConfig()
	case "cache":
		configMap = config.CacheConfig()
	case "mail":
		configMap = config.MailConfig()
	case "logging":
		configMap = config.LoggingConfig()
	default:
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
func (c *ConfigFacade) GetString(key string, defaultValue ...string) string {
	val := c.Get(key)
	if str, ok := val.(string); ok {
		return str
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ""
}

// GetInt retrieves a configuration value as int
func (c *ConfigFacade) GetInt(key string, defaultValue ...int) int {
	val := c.Get(key)
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
func (c *ConfigFacade) GetBool(key string, defaultValue ...bool) bool {
	val := c.Get(key)
	if b, ok := val.(bool); ok {
		return b
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return false
}

// Has checks if a configuration key exists
func (c *ConfigFacade) Has(key string) bool {
	return c.Get(key) != nil
}

// Set sets a configuration value (not implemented for read-only configs)
func (c *ConfigFacade) Set(key string, value interface{}) {
	// This is a read-only config system, so Set is a no-op
	// In a real implementation, you might want to store runtime configs separately
}
