package config

import "base_lara_go_project/app/core/env"

// DatabaseConfig returns the database configuration with fallback values
func DatabaseConfig() map[string]interface{} {
	return map[string]interface{}{
		"default": env.GetEnv("DB_CONNECTION", "mysql"),
		"connections": map[string]interface{}{
			"mysql": map[string]interface{}{
				"driver":   "mysql",
				"host":     env.GetEnv("DB_HOST", "localhost"),
				"port":     env.GetEnv("DB_PORT", "3306"),
				"database": env.GetEnv("DB_NAME", "laravel"),
				"username": env.GetEnv("DB_USER", "root"),
				"password": env.GetEnv("DB_PASSWORD", ""),
				"charset":  env.GetEnv("DB_CHARSET", "utf8mb4"),
				"collation": env.GetEnv("DB_COLLATION", "utf8mb4_unicode_ci"),
				"prefix":   env.GetEnv("DB_PREFIX", ""),
				"strict":   env.GetEnv("DB_STRICT", "true"),
				"engine":   env.GetEnv("DB_ENGINE", "InnoDB"),
			},
			"postgres": map[string]interface{}{
				"driver":   "postgres",
				"host":     env.GetEnv("DB_HOST", "localhost"),
				"port":     env.GetEnv("DB_PORT", "5432"),
				"database": env.GetEnv("DB_NAME", "laravel"),
				"username": env.GetEnv("DB_USER", "postgres"),
				"password": env.GetEnv("DB_PASSWORD", ""),
				"charset":  env.GetEnv("DB_CHARSET", "utf8"),
				"prefix":   env.GetEnv("DB_PREFIX", ""),
				"sslmode":  env.GetEnv("DB_SSLMODE", "disable"),
			},
			"sqlite": map[string]interface{}{
				"driver":   "sqlite",
				"database": env.GetEnv("DB_DATABASE", "database/database.sqlite"),
				"prefix":   env.GetEnv("DB_PREFIX", ""),
			},
		},
	}
}
