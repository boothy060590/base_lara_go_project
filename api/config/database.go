package config

import (
	"base_lara_go_project/app/core/go_core"
	"base_lara_go_project/app/core/laravel_core/env"
)

// DatabaseConfig returns the database configuration with environment variable fallbacks
// This config defines database connections, credentials, and connection parameters
func DatabaseConfig() map[string]interface{} {
	return map[string]interface{}{
		"default": env.Get("DB_CONNECTION", "mysql"),
		"connections": map[string]interface{}{
			"mysql": map[string]interface{}{
				"driver":   "mysql",
				"host":     env.Get("DB_HOST", "127.0.0.1"),
				"port":     env.Get("DB_PORT", "3306"),
				"database": env.Get("DB_DATABASE", "laravel"),
				"username": env.Get("DB_USERNAME", "root"),
				"password": env.Get("DB_PASSWORD", ""),
				"charset":  env.Get("DB_CHARSET", "utf8mb4"),
				"prefix":   env.Get("DB_PREFIX", ""),
			},
			"postgres": map[string]interface{}{
				"driver":   "postgres",
				"host":     env.Get("DB_HOST", "127.0.0.1"),
				"port":     env.Get("DB_PORT", "5432"),
				"database": env.Get("DB_DATABASE", "laravel"),
				"username": env.Get("DB_USERNAME", "postgres"),
				"password": env.Get("DB_PASSWORD", ""),
				"charset":  env.Get("DB_CHARSET", "utf8"),
				"prefix":   env.Get("DB_PREFIX", ""),
			},
		},
	}
}

// init automatically registers this config with the global config loader
// This ensures the database config is available via config.Get("database") and dot notation
func init() {
	go_core.RegisterGlobalConfig("database", DatabaseConfig)
}
