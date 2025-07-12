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
				"driver":             "mysql",
				"host":               env.Get("DB_HOST", "127.0.0.1"),
				"port":               env.Get("DB_PORT", "3306"),
				"database":           env.Get("DB_NAME", "laravel"),
				"username":           env.Get("DB_USER", "root"),
				"password":           env.Get("DB_PASSWORD", ""),
				"charset":            env.Get("DB_CHARSET", "utf8mb4"),
				"prefix":             env.Get("DB_PREFIX", ""),
				"max_idle_conns":     env.GetInt("DB_MAX_IDLE_CONNS", 10),       // Optimal for most workloads
				"max_open_conns":     env.GetInt("DB_MAX_OPEN_CONNS", 100),      // Optimal for most workloads
				"conn_max_lifetime":  env.GetInt("DB_CONN_MAX_LIFETIME", 3600),  // seconds - optimal
				"conn_max_idle_time": env.GetInt("DB_CONN_MAX_IDLE_TIME", 1800), // seconds - optimal
				"autocommit":         env.GetBool("DB_AUTOCOMMIT", true),
				"sql_mode":           env.Get("DB_SQL_MODE", "NO_ENGINE_SUBSTITUTION"),
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
