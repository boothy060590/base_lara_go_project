package core

import (
	"os"
	"strings"
)

var configRegistry = map[string]interface{}{}

// LoadConfig loads all config maps into the registry
func LoadConfig(configs map[string]map[string]interface{}) {
	for k, v := range configs {
		configRegistry[k] = v
	}
}

// Get retrieves a config value using dot notation (e.g. "database.username")
func Get(key string, defaultValue ...interface{}) interface{} {
	value := os.Getenv(key)
	if value == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return nil
	}
	parts := strings.Split(key, ".")
	var current interface{} = configRegistry
	for _, part := range parts {
		m, ok := current.(map[string]interface{})
		if !ok {
			if len(defaultValue) > 0 {
				return defaultValue[0]
			}
			return nil
		}
		current, ok = m[part]
		if !ok {
			if len(defaultValue) > 0 {
				return defaultValue[0]
			}
			return nil
		}
	}
	return current
}

// Set sets a config value using dot notation (e.g. "app.debug")
func Set(key string, value interface{}) {
	parts := strings.Split(key, ".")
	last := len(parts) - 1
	var current interface{} = configRegistry
	for i, part := range parts {
		m, ok := current.(map[string]interface{})
		if !ok {
			return
		}
		if i == last {
			m[part] = value
			return
		}
		if _, exists := m[part]; !exists {
			m[part] = map[string]interface{}{}
		}
		current = m[part]
	}
}
