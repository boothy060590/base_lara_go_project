package config

import "base_lara_go_project/app/core/env"

// AppConfig returns the app configuration with fallback values
func AppConfig() map[string]interface{} {
	return map[string]interface{}{
		"name":                env.GetEnv("APP_NAME", "Base Laravel Go Project"),
		"debug":               env.GetEnv("APP_DEBUG", "false"),
		"url":                 env.GetEnv("APP_URL", "http://localhost"),
		"env":                 env.GetEnv("APP_ENV", "development"),
		"port":                env.GetEnv("APP_PORT", "8080"),
		"secret":              env.GetEnv("API_SECRET", "changeme"),
		"token_hour_lifespan": env.GetEnv("TOKEN_HOUR_LIFESPAN", "1"),
	}
}
