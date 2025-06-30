package config

import "base_lara_go_project/app/core/laravel_core/env"

// DatabaseConfig returns the database configuration with fallback values
func DatabaseConfig() map[string]interface{} {
	return map[string]interface{}{
		"default": env.Get("DB_CONNECTION", "mysql"),
		"connections": map[string]interface{}{
			"mysql": map[string]interface{}{
				"driver":    "mysql",
				"host":      env.Get("DB_HOST", "localhost"),
				"port":      env.Get("DB_PORT", "3306"),
				"database":  env.Get("DB_NAME", "laravel"),
				"username":  env.Get("DB_USER", "root"),
				"password":  env.Get("DB_PASSWORD", ""),
				"charset":   env.Get("DB_CHARSET", "utf8mb4"),
				"collation": env.Get("DB_COLLATION", "utf8mb4_unicode_ci"),
				"prefix":    env.Get("DB_PREFIX", ""),
				"strict":    env.Get("DB_STRICT", "true"),
				"engine":    env.Get("DB_ENGINE", "InnoDB"),
			},
			"postgres": map[string]interface{}{
				"driver":   "postgres",
				"host":     env.Get("DB_HOST", "localhost"),
				"port":     env.Get("DB_PORT", "5432"),
				"database": env.Get("DB_NAME", "laravel"),
				"username": env.Get("DB_USER", "postgres"),
				"password": env.Get("DB_PASSWORD", ""),
				"charset":  env.Get("DB_CHARSET", "utf8"),
				"prefix":   env.Get("DB_PREFIX", ""),
				"sslmode":  env.Get("DB_SSLMODE", "disable"),
			},
			"sqlite": map[string]interface{}{
				"driver":   "sqlite",
				"database": env.Get("DB_DATABASE", "database/database.sqlite"),
				"prefix":   env.Get("DB_PREFIX", ""),
			},
		},
	}
}
