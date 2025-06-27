package config

func DatabaseConfig() map[string]interface{} {
	return map[string]interface{}{
		"default": getEnv("DB_CONNECTION", "mysql"),
		"connections": map[string]interface{}{
			"mysql": map[string]interface{}{
				"driver":   "mysql",
				"host":     getEnv("DB_HOST", "localhost"),
				"port":     getEnv("DB_PORT", "3306"),
				"database": getEnv("DB_NAME", "app_db"),
				"username": getEnv("DB_USER", "root"),
				"password": getEnv("DB_PASSWORD", ""),
			},
			"sqlite": map[string]interface{}{
				"driver":   "sqlite",
				"database": getEnv("SQLITE_DB", "database.sqlite"),
			},
		},
	}
}
