package config

import "base_lara_go_project/app/core/laravel_core/env"

// AppConfig returns the app configuration with fallback values
func AppConfig() map[string]interface{} {
	return map[string]interface{}{
		"name":                env.Get("APP_NAME", "Base Laravel Go Project"),
		"debug":               env.GetBool("APP_DEBUG", false),
		"url":                 env.Get("APP_URL", "http://localhost"),
		"env":                 env.Get("APP_ENV", "development"),
		"port":                env.Get("APP_PORT", "8080"),
		"secret":              env.Get("APP_SECRET", "changeme"),
		"token_hour_lifespan": env.GetInt("TOKEN_HOUR_LIFESPAN", 1),
	}
}
