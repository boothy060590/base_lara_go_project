package config

import "base_lara_go_project/app/core/env"

// LoggingConfig returns the logging configuration with fallback values
func LoggingConfig() map[string]interface{} {
	return map[string]interface{}{
		"default": env.GetEnv("LOG_CHANNEL", "stack"),
		"deprecations": map[string]interface{}{
			"channel": env.GetEnv("LOG_DEPRECATIONS_CHANNEL", "null"),
			"trace":   env.GetEnvBool("LOG_DEPRECATIONS_TRACE", false),
		},
		"channels": map[string]interface{}{
			"stack": map[string]interface{}{
				"driver":   "stack",
				"channels": []string{"single", "daily"},
			},
			"single": map[string]interface{}{
				"driver": "single",
				"path":   env.GetEnv("LOG_PATH", "storage/logs/laravel.log"),
				"level":  env.GetEnv("LOG_LEVEL", "debug"),
			},
			"daily": map[string]interface{}{
				"driver":   "daily",
				"path":     env.GetEnv("LOG_PATH", "storage/logs/laravel.log"),
				"level":    env.GetEnv("LOG_LEVEL", "debug"),
				"days":     env.GetEnvInt("LOG_DAILY_DAYS", 14),
				"max_size": env.GetEnvInt("LOG_DAILY_MAX_SIZE", 10485760), // 10MB
			},
			"slack": map[string]interface{}{
				"driver":   "slack",
				"url":      env.GetEnv("LOG_SLACK_WEBHOOK_URL", ""),
				"username": env.GetEnv("LOG_SLACK_USERNAME", "Laravel Log"),
				"emoji":    env.GetEnv("LOG_SLACK_EMOJI", ":boom:"),
				"level":    env.GetEnv("LOG_LEVEL", "critical"),
			},
			"papertrail": map[string]interface{}{
				"driver": "papertrail",
				"level":   env.GetEnv("LOG_LEVEL", "debug"),
				"host":    env.GetEnv("PAPERTRAIL_URL", ""),
				"port":    env.GetEnv("PAPERTRAIL_PORT", ""),
			},
			"stderr": map[string]interface{}{
				"driver": "stderr",
				"level":  env.GetEnv("LOG_LEVEL", "debug"),
			},
			"sentry": map[string]interface{}{
				"driver": "sentry",
				"level":  env.GetEnv("LOG_LEVEL", "debug"),
				"dsn":    env.GetEnv("SENTRY_DSN", ""),
			},
			"null": map[string]interface{}{
				"driver": "null",
			},
			"emergency": map[string]interface{}{
				"driver": "emergency",
				"path":   env.GetEnv("LOG_EMERGENCY_PATH", "storage/logs/emergency.log"),
			},
		},
	}
}
