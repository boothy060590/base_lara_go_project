package config

import "base_lara_go_project/app/core/laravel_core/env"

// LoggingConfig returns the logging configuration with fallback values
func LoggingConfig() map[string]interface{} {
	return map[string]interface{}{
		"default": env.Get("LOG_CHANNEL", "stack"),
		"deprecations": map[string]interface{}{
			"channel": env.Get("LOG_DEPRECATIONS_CHANNEL", "null"),
			"trace":   env.GetBool("LOG_DEPRECATIONS_TRACE", false),
		},
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
				"driver":   "daily",
				"path":     env.Get("LOG_PATH", "storage/logs/laravel.log"),
				"level":    env.Get("LOG_LEVEL", "debug"),
				"days":     env.GetInt("LOG_DAILY_DAYS", 14),
				"max_size": env.GetInt("LOG_DAILY_MAX_SIZE", 10485760), // 10MB
			},
			"slack": map[string]interface{}{
				"driver":   "slack",
				"url":      env.Get("LOG_SLACK_WEBHOOK_URL", ""),
				"username": env.Get("LOG_SLACK_USERNAME", "Laravel Log"),
				"emoji":    env.Get("LOG_SLACK_EMOJI", ":boom:"),
				"level":    env.Get("LOG_LEVEL", "critical"),
			},
			"papertrail": map[string]interface{}{
				"driver": "papertrail",
				"level":  env.Get("LOG_LEVEL", "debug"),
				"host":   env.Get("PAPERTRAIL_URL", ""),
				"port":   env.Get("PAPERTRAIL_PORT", ""),
			},
			"stderr": map[string]interface{}{
				"driver": "stderr",
				"level":  env.Get("LOG_LEVEL", "debug"),
			},
			"sentry": map[string]interface{}{
				"driver": "sentry",
				"level":  env.Get("LOG_LEVEL", "debug"),
				"dsn":    env.Get("SENTRY_DSN", ""),
			},
			"null": map[string]interface{}{
				"driver": "null",
			},
			"emergency": map[string]interface{}{
				"driver": "emergency",
				"path":   env.Get("LOG_EMERGENCY_PATH", "storage/logs/emergency.log"),
			},
		},
	}
}
