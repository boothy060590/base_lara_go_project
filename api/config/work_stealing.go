package config

import (
	"base_lara_go_project/app/core/laravel_core/env"
	"os"
	"runtime"
	"strconv"
	"time"
)

// WorkStealingConfig returns the work stealing configuration with fallback values
func WorkStealingConfig() map[string]interface{} {
	return map[string]interface{}{
		"enabled": env.GetBool("WORK_STEALING_ENABLED", true),
		"workers": map[string]interface{}{
			"num_workers":      env.GetInt("WORK_STEALING_NUM_WORKERS", runtime.NumCPU()),
			"queue_size":       env.GetInt("WORK_STEALING_QUEUE_SIZE", 1024),
			"steal_threshold":  env.GetInt("WORK_STEALING_THRESHOLD", 2),
			"steal_batch_size": env.GetInt("WORK_STEALING_BATCH_SIZE", 10),
			"idle_timeout":     env.GetInt("WORK_STEALING_IDLE_TIMEOUT", 100),
		},
		"profiles": map[string]interface{}{
			"web": map[string]interface{}{
				"num_workers":  runtime.NumCPU() * 2,
				"queue_size":   2048,
				"idle_timeout": 50,
			},
			"api": map[string]interface{}{
				"num_workers":  runtime.NumCPU() * 3,
				"queue_size":   4096,
				"idle_timeout": 25,
			},
			"background": map[string]interface{}{
				"num_workers":  runtime.NumCPU() * 4,
				"queue_size":   8192,
				"idle_timeout": 200,
			},
			"streaming": map[string]interface{}{
				"num_workers":  runtime.NumCPU() * 2,
				"queue_size":   16384,
				"idle_timeout": 10,
			},
			"batch": map[string]interface{}{
				"num_workers":  runtime.NumCPU() * 6,
				"queue_size":   32768,
				"idle_timeout": 500,
			},
		},
		"optimizations": map[string]interface{}{
			"enable_metrics":   env.GetBool("WORK_STEALING_METRICS", true),
			"enable_profiling": env.GetBool("WORK_STEALING_PROFILING", true),
		},
	}
}

// Helper functions for environment variable parsing
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
