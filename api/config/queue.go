package config

import (
	"base_lara_go_project/app/core/go_core"
	"base_lara_go_project/app/core/laravel_core/env"
)

// QueueConfig returns the queue configuration with environment variable fallbacks
// This config defines queue connections, drivers, and processing parameters
func QueueConfig() map[string]interface{} {
	return map[string]interface{}{
		"default": env.Get("QUEUE_CONNECTION", "sync"),
		"connections": map[string]interface{}{
			"sync": map[string]interface{}{
				"driver": "sync",
			},
			"database": map[string]interface{}{
				"driver":      "database",
				"table":       env.Get("QUEUE_TABLE", "jobs"),
				"queue":       env.Get("QUEUE_QUEUE", "default"),
				"retry_after": env.GetInt("QUEUE_RETRY_AFTER", 90),
			},
		},
	}
}

// init automatically registers this config with the global config loader
// This ensures the queue config is available via config.Get("queue") and dot notation
func init() {
	go_core.RegisterGlobalConfig("queue", QueueConfig)
}
