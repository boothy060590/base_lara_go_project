package config

import (
	"base_lara_go_project/app/core/go_core"
	"base_lara_go_project/app/core/laravel_core/env"
)

// ContextConfig returns the context optimization configuration with environment variable fallbacks
// This config defines context timeouts, retry policies, and optimization settings for different operations
func ContextConfig() map[string]interface{} {
	return map[string]interface{}{
		"enabled": env.GetBool("CONTEXT_ENABLED", true),
		"defaults": map[string]interface{}{
			"timeout":             env.GetInt("CONTEXT_DEFAULT_TIMEOUT", 30),
			"max_timeout":         env.GetInt("CONTEXT_MAX_TIMEOUT", 300),
			"enable_deadline":     env.GetBool("CONTEXT_ENABLE_DEADLINE", true),
			"enable_cancellation": env.GetBool("CONTEXT_ENABLE_CANCELLATION", true),
			"propagate_values":    env.GetBool("CONTEXT_PROPAGATE_VALUES", true),
		},
		"operations": map[string]interface{}{
			"repository": map[string]interface{}{
				"timeout":        env.GetInt("CONTEXT_REPO_TIMEOUT", 30),
				"retry_attempts": env.GetInt("CONTEXT_REPO_RETRY", 3),
				"retry_delay":    env.GetInt("CONTEXT_REPO_RETRY_DELAY", 1),
			},
			"events": map[string]interface{}{
				"timeout":        env.GetInt("CONTEXT_EVENTS_TIMEOUT", 30),
				"retry_attempts": env.GetInt("CONTEXT_EVENTS_RETRY", 3),
				"retry_delay":    env.GetInt("CONTEXT_EVENTS_RETRY_DELAY", 1),
			},
			"jobs": map[string]interface{}{
				"timeout":        env.GetInt("CONTEXT_JOBS_TIMEOUT", 60),
				"retry_attempts": env.GetInt("CONTEXT_JOBS_RETRY", 3),
				"retry_delay":    env.GetInt("CONTEXT_JOBS_RETRY_DELAY", 5),
			},
			"cache": map[string]interface{}{
				"timeout":        env.GetInt("CONTEXT_CACHE_TIMEOUT", 5),
				"retry_attempts": env.GetInt("CONTEXT_CACHE_RETRY", 2),
				"retry_delay":    env.GetInt("CONTEXT_CACHE_RETRY_DELAY", 1),
			},
			"mail": map[string]interface{}{
				"timeout":        env.GetInt("CONTEXT_MAIL_TIMEOUT", 30),
				"retry_attempts": env.GetInt("CONTEXT_MAIL_RETRY", 3),
				"retry_delay":    env.GetInt("CONTEXT_MAIL_RETRY_DELAY", 5),
			},
			"http": map[string]interface{}{
				"timeout":        env.GetInt("CONTEXT_HTTP_TIMEOUT", 30),
				"retry_attempts": env.GetInt("CONTEXT_HTTP_RETRY", 3),
				"retry_delay":    env.GetInt("CONTEXT_HTTP_RETRY_DELAY", 1),
			},
		},
		"profiles": map[string]interface{}{
			"web": map[string]interface{}{
				"timeout":     env.GetInt("CONTEXT_PROFILE_WEB_TIMEOUT", 30),
				"max_timeout": env.GetInt("CONTEXT_PROFILE_WEB_MAX_TIMEOUT", 60),
			},
			"api": map[string]interface{}{
				"timeout":     env.GetInt("CONTEXT_PROFILE_API_TIMEOUT", 60),
				"max_timeout": env.GetInt("CONTEXT_PROFILE_API_MAX_TIMEOUT", 300),
			},
			"background": map[string]interface{}{
				"timeout":     env.GetInt("CONTEXT_PROFILE_BACKGROUND_TIMEOUT", 300),
				"max_timeout": env.GetInt("CONTEXT_PROFILE_BACKGROUND_MAX_TIMEOUT", 1800),
			},
			"streaming": map[string]interface{}{
				"timeout":     env.GetInt("CONTEXT_PROFILE_STREAMING_TIMEOUT", 1800),
				"max_timeout": env.GetInt("CONTEXT_PROFILE_STREAMING_MAX_TIMEOUT", 3600),
			},
			"batch": map[string]interface{}{
				"timeout":     env.GetInt("CONTEXT_PROFILE_BATCH_TIMEOUT", 1800),
				"max_timeout": env.GetInt("CONTEXT_PROFILE_BATCH_MAX_TIMEOUT", 7200),
			},
		},
		"tracking": map[string]interface{}{
			"enabled":             env.GetBool("CONTEXT_TRACKING_ENABLED", true),
			"performance_metrics": env.GetBool("CONTEXT_PERFORMANCE_METRICS", true),
			"request_tracing":     env.GetBool("CONTEXT_REQUEST_TRACING", true),
			"error_tracking":      env.GetBool("CONTEXT_ERROR_TRACKING", true),
		},
		"optimizations": map[string]interface{}{
			"auto_timeout":             env.GetBool("CONTEXT_AUTO_TIMEOUT", true),
			"context_pooling":          env.GetBool("CONTEXT_POOLING", true),
			"deadline_propagation":     env.GetBool("CONTEXT_DEADLINE_PROPAGATION", true),
			"cancellation_propagation": env.GetBool("CONTEXT_CANCELLATION_PROPAGATION", true),
		},
	}
}

// init automatically registers this config with the global config loader
// This ensures the context config is available via config.Get("context") and dot notation
func init() {
	go_core.RegisterGlobalConfig("context", ContextConfig)
}
