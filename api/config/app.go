package config

import (
	"base_lara_go_project/app/core/go_core"
	"base_lara_go_project/app/core/laravel_core/env"
)

// AppConfig returns the application configuration with environment variable fallbacks
// This config contains core application settings like name, debug mode, port, etc.
func AppConfig() map[string]interface{} {
	return map[string]interface{}{
		"name":                env.Get("APP_NAME", "Base Laravel Go Project"),
		"debug":               env.GetBool("APP_DEBUG", false),
		"url":                 env.Get("APP_URL", "http://localhost"),
		"env":                 env.Get("APP_ENV", "development"),
		"port":                env.Get("APP_PORT", "8080"),
		"secret":              env.Get("API_SECRET", "changeme"),
		"token_hour_lifespan": env.GetInt("TOKEN_HOUR_LIFESPAN", 1),
	}
}

// init automatically registers this config with the global config loader
// This ensures the app config is available via config.Get("app") and dot notation
func init() {
	go_core.RegisterGlobalConfig("app", AppConfig)
}
