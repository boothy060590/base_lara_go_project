package config

import (
	"base_lara_go_project/app/core/go_core"
	"base_lara_go_project/app/core/laravel_core/env"
)

// LoggingConfig returns the logging configuration with environment variable fallbacks
// This config defines logging channels, handlers, and log levels for different environments
func LoggingConfig() map[string]interface{} {
	return map[string]interface{}{
		"default": env.Get("LOG_CHANNEL", "stack"),
		"channels": map[string]interface{}{
			"stack": map[string]interface{}{
				"driver":   "stack",
				"channels": []string{"single", "daily"},
			},
			"single": map[string]interface{}{
				"driver": "single",
				"path":   env.Get("LOG_PATH", "storage/logs/laravel.log"),
				"level":  env.Get("LOG_LEVEL", "debug"),
			},
			"daily": map[string]interface{}{
				"driver": "daily",
				"path":   env.Get("LOG_PATH", "storage/logs/laravel.log"),
				"level":  env.Get("LOG_LEVEL", "debug"),
				"days":   env.GetInt("LOG_DAYS", 14),
			},
			"sentry": map[string]interface{}{
				"driver": "sentry",
				"dsn":    env.Get("SENTRY_DSN", ""),
				"level":  env.Get("LOG_LEVEL", "debug"),
			},
		},
	}
}

// init automatically registers this config with the global config loader
// This ensures the logging config is available via config.Get("logging") and dot notation
func init() {
	go_core.RegisterGlobalConfig("logging", LoggingConfig)
}
