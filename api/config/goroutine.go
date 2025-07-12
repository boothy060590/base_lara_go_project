package config

import (
	"base_lara_go_project/app/core/go_core"
	"base_lara_go_project/app/core/laravel_core/env"
)

// GoroutineConfig returns the goroutine optimization configuration with environment variable fallbacks
// This config defines goroutine pools, work-stealing settings, and performance optimization parameters
func GoroutineConfig() map[string]interface{} {
	return map[string]interface{}{
		"enabled": env.GetBool("GOROUTINE_ENABLED", true),
		"pools": map[string]interface{}{
			"default": map[string]interface{}{
				"min_workers":      env.GetInt("GOROUTINE_MIN_WORKERS", 2),
				"max_workers":      env.GetInt("GOROUTINE_MAX_WORKERS", 10),
				"queue_size":       env.GetInt("GOROUTINE_QUEUE_SIZE", 1000),
				"idle_timeout":     env.GetInt("GOROUTINE_IDLE_TIMEOUT", 60),
				"shutdown_timeout": env.GetInt("GOROUTINE_SHUTDOWN_TIMEOUT", 30),
			},
			"high_performance": map[string]interface{}{
				"min_workers":      env.GetInt("GOROUTINE_HP_MIN_WORKERS", 5),
				"max_workers":      env.GetInt("GOROUTINE_HP_MAX_WORKERS", 50),
				"queue_size":       env.GetInt("GOROUTINE_HP_QUEUE_SIZE", 5000),
				"idle_timeout":     env.GetInt("GOROUTINE_HP_IDLE_TIMEOUT", 120),
				"shutdown_timeout": env.GetInt("GOROUTINE_HP_SHUTDOWN_TIMEOUT", 60),
			},
			"low_latency": map[string]interface{}{
				"min_workers":      env.GetInt("GOROUTINE_LL_MIN_WORKERS", 10),
				"max_workers":      env.GetInt("GOROUTINE_LL_MAX_WORKERS", 100),
				"queue_size":       env.GetInt("GOROUTINE_LL_QUEUE_SIZE", 10000),
				"idle_timeout":     env.GetInt("GOROUTINE_LL_IDLE_TIMEOUT", 30),
				"shutdown_timeout": env.GetInt("GOROUTINE_LL_SHUTDOWN_TIMEOUT", 15),
			},
		},
		"optimizations": map[string]interface{}{
			"auto_scale":           env.GetBool("GOROUTINE_AUTO_SCALE", true),
			"work_stealing":        env.GetBool("GOROUTINE_WORK_STEALING", true),
			"performance_tracking": env.GetBool("GOROUTINE_PERFORMANCE_TRACKING", true),
			"metrics_enabled":      env.GetBool("GOROUTINE_METRICS_ENABLED", true),
		},
		"repository": map[string]interface{}{
			"async_enabled":  env.GetBool("GOROUTINE_REPO_ASYNC", true),
			"parallel_limit": env.GetInt("GOROUTINE_REPO_PARALLEL_LIMIT", 10),
			"batch_size":     env.GetInt("GOROUTINE_REPO_BATCH_SIZE", 100),
			"timeout":        env.GetInt("GOROUTINE_REPO_TIMEOUT", 30),
		},
		"events": map[string]interface{}{
			"async_enabled":  env.GetBool("GOROUTINE_EVENTS_ASYNC", true),
			"parallel_limit": env.GetInt("GOROUTINE_EVENTS_PARALLEL_LIMIT", 5),
			"timeout":        env.GetInt("GOROUTINE_EVENTS_TIMEOUT", 30),
			"retry_attempts": env.GetInt("GOROUTINE_EVENTS_RETRY", 3),
		},
		"jobs": map[string]interface{}{
			"async_enabled":      env.GetBool("GOROUTINE_JOBS_ASYNC", true),
			"parallel_limit":     env.GetInt("GOROUTINE_JOBS_PARALLEL_LIMIT", 20),
			"timeout":            env.GetInt("GOROUTINE_JOBS_TIMEOUT", 60),
			"retry_attempts":     env.GetInt("GOROUTINE_JOBS_RETRY", 3),
			"backoff_multiplier": env.GetFloat("GOROUTINE_JOBS_BACKOFF", 2.0),
		},
		"cache": map[string]interface{}{
			"async_enabled":  env.GetBool("GOROUTINE_CACHE_ASYNC", false),
			"parallel_limit": env.GetInt("GOROUTINE_CACHE_PARALLEL_LIMIT", 5),
			"timeout":        env.GetInt("GOROUTINE_CACHE_TIMEOUT", 5),
		},
	}
}

// init automatically registers this config with the global config loader
// This ensures the goroutine config is available via config.Get("goroutine") and dot notation
func init() {
	go_core.RegisterGlobalConfig("goroutine", GoroutineConfig)
}
