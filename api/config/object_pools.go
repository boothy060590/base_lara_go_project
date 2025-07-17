package config

import (
	"base_lara_go_project/app/core/go_core"
	"base_lara_go_project/app/core/laravel_core/env"
)

// ObjectPoolsConfig returns the object pool configuration with environment variable fallbacks
// This config defines object pools for different types with configurable sizing and cleanup strategies
// Core provides default pools, developers can add custom pools in the custom_pools section
func ObjectPoolsConfig() map[string]interface{} {
	return map[string]interface{}{
		// Global enable/disable for object pools
		"enabled": env.GetBool("OBJECT_POOLS_ENABLED", true),

		// Global pool settings
		"global": map[string]interface{}{
			"max_pool_size":    env.GetInt("OBJECT_POOLS_MAX_SIZE", 1000),
			"min_pool_size":    env.GetInt("OBJECT_POOLS_MIN_SIZE", 10),
			"cleanup_interval": env.GetInt("OBJECT_POOLS_CLEANUP_INTERVAL", 300), // 5 minutes
			"max_idle_time":    env.GetInt("OBJECT_POOLS_MAX_IDLE_TIME", 1800),   // 30 minutes
			"enable_metrics":   env.GetBool("OBJECT_POOLS_METRICS_ENABLED", true),
		},

		// Core default pools - these are provided by the framework
		"default_pools": map[string]interface{}{
			// Response object pool (core framework pool)
			"response": map[string]interface{}{
				"name":             "Response Object Pool",
				"description":      "Pool for HTTP response objects",
				"type":             "response",
				"pool_size":        env.GetInt("OBJECT_POOL_RESPONSE_SIZE", 500),
				"max_idle_time":    env.GetInt("OBJECT_POOL_RESPONSE_MAX_IDLE_TIME", 600), // 10 minutes
				"cleanup_strategy": "eager",
				"reset_method":     "Reset",
				"enabled":          env.GetBool("OBJECT_POOL_RESPONSE_ENABLED", true),
				"metrics": map[string]interface{}{
					"track_hits":        true,
					"track_misses":      true,
					"track_allocations": true,
					"track_reuses":      true,
				},
			},

			// Event object pool (core framework pool)
			"event": map[string]interface{}{
				"name":             "Event Object Pool",
				"description":      "Pool for Event objects used in event processing",
				"type":             "event",
				"pool_size":        env.GetInt("OBJECT_POOL_EVENT_SIZE", 200),
				"max_idle_time":    env.GetInt("OBJECT_POOL_EVENT_MAX_IDLE_TIME", 900), // 15 minutes
				"cleanup_strategy": "adaptive",
				"reset_method":     "Reset",
				"enabled":          env.GetBool("OBJECT_POOL_EVENT_ENABLED", true),
				"metrics": map[string]interface{}{
					"track_hits":        true,
					"track_misses":      true,
					"track_allocations": true,
					"track_reuses":      true,
				},
			},

			// Job object pool (core framework pool)
			"job": map[string]interface{}{
				"name":             "Job Object Pool",
				"description":      "Pool for background job objects",
				"type":             "job",
				"pool_size":        env.GetInt("OBJECT_POOL_JOB_SIZE", 50),
				"max_idle_time":    env.GetInt("OBJECT_POOL_JOB_MAX_IDLE_TIME", 3600), // 1 hour
				"cleanup_strategy": "lazy",
				"reset_method":     "Reset",
				"enabled":          env.GetBool("OBJECT_POOL_JOB_ENABLED", true),
				"metrics": map[string]interface{}{
					"track_hits":        true,
					"track_misses":      true,
					"track_allocations": true,
					"track_reuses":      true,
				},
			},

			// Cache object pool (core framework pool)
			"cache": map[string]interface{}{
				"name":             "Cache Object Pool",
				"description":      "Pool for cache entry objects",
				"type":             "cache",
				"pool_size":        env.GetInt("OBJECT_POOL_CACHE_SIZE", 300),
				"max_idle_time":    env.GetInt("OBJECT_POOL_CACHE_MAX_IDLE_TIME", 1200), // 20 minutes
				"cleanup_strategy": "adaptive",
				"reset_method":     "Reset",
				"enabled":          env.GetBool("OBJECT_POOL_CACHE_ENABLED", true),
				"metrics": map[string]interface{}{
					"track_hits":        true,
					"track_misses":      true,
					"track_allocations": true,
					"track_reuses":      true,
				},
			},

			// Database row pool (core framework pool)
			"db_row": map[string]interface{}{
				"name":             "Database Row Pool",
				"description":      "Pool for database row objects",
				"type":             "db_row",
				"pool_size":        env.GetInt("OBJECT_POOL_DB_ROW_SIZE", 1000),
				"max_idle_time":    env.GetInt("OBJECT_POOL_DB_ROW_MAX_IDLE_TIME", 300), // 5 minutes
				"cleanup_strategy": "eager",
				"reset_method":     "Reset",
				"enabled":          env.GetBool("OBJECT_POOL_DB_ROW_ENABLED", true),
				"metrics": map[string]interface{}{
					"track_hits":        true,
					"track_misses":      true,
					"track_allocations": true,
					"track_reuses":      true,
				},
			},

			// Request object pool (core framework pool)
			"request": map[string]interface{}{
				"name":             "Request Object Pool",
				"description":      "Pool for HTTP request objects",
				"type":             "request",
				"pool_size":        env.GetInt("OBJECT_POOL_REQUEST_SIZE", 200),
				"max_idle_time":    env.GetInt("OBJECT_POOL_REQUEST_MAX_IDLE_TIME", 600), // 10 minutes
				"cleanup_strategy": "adaptive",
				"reset_method":     "Reset",
				"enabled":          env.GetBool("OBJECT_POOL_REQUEST_ENABLED", true),
				"metrics": map[string]interface{}{
					"track_hits":        true,
					"track_misses":      true,
					"track_allocations": true,
					"track_reuses":      true,
				},
			},

			// Validation object pool (core framework pool)
			"validation": map[string]interface{}{
				"name":             "Validation Object Pool",
				"description":      "Pool for validation result objects",
				"type":             "validation",
				"pool_size":        env.GetInt("OBJECT_POOL_VALIDATION_SIZE", 150),
				"max_idle_time":    env.GetInt("OBJECT_POOL_VALIDATION_MAX_IDLE_TIME", 900), // 15 minutes
				"cleanup_strategy": "lazy",
				"reset_method":     "Reset",
				"enabled":          env.GetBool("OBJECT_POOL_VALIDATION_ENABLED", true),
				"metrics": map[string]interface{}{
					"track_hits":        true,
					"track_misses":      true,
					"track_allocations": true,
					"track_reuses":      true,
				},
			},

			// Log entry pool (core framework pool)
			"log_entry": map[string]interface{}{
				"name":             "Log Entry Pool",
				"description":      "Pool for log entry objects",
				"type":             "log_entry",
				"pool_size":        env.GetInt("OBJECT_POOL_LOG_ENTRY_SIZE", 100),
				"max_idle_time":    env.GetInt("OBJECT_POOL_LOG_ENTRY_MAX_IDLE_TIME", 1800), // 30 minutes
				"cleanup_strategy": "lazy",
				"reset_method":     "Reset",
				"enabled":          env.GetBool("OBJECT_POOL_LOG_ENTRY_ENABLED", true),
				"metrics": map[string]interface{}{
					"track_hits":        true,
					"track_misses":      true,
					"track_allocations": true,
					"track_reuses":      true,
				},
			},

			// Mail message pool (core framework pool)
			"mail_message": map[string]interface{}{
				"name":             "Mail Message Pool",
				"description":      "Pool for email message objects",
				"type":             "mail_message",
				"pool_size":        env.GetInt("OBJECT_POOL_MAIL_MESSAGE_SIZE", 50),
				"max_idle_time":    env.GetInt("OBJECT_POOL_MAIL_MESSAGE_MAX_IDLE_TIME", 3600), // 1 hour
				"cleanup_strategy": "lazy",
				"reset_method":     "Reset",
				"enabled":          env.GetBool("OBJECT_POOL_MAIL_MESSAGE_ENABLED", true),
				"metrics": map[string]interface{}{
					"track_hits":        true,
					"track_misses":      true,
					"track_allocations": true,
					"track_reuses":      true,
				},
			},
		},

		// Custom pools - developers can add their own pools here
		// These will be merged with the default pools by the service provider
		"custom_pools": map[string]interface{}{
			// Example custom pool - developers can add their own
			"example_custom": map[string]interface{}{
				"name":             "Example Custom Pool",
				"description":      "Example of how to define a custom pool",
				"type":             "custom_type",
				"pool_size":        env.GetInt("OBJECT_POOL_CUSTOM_SIZE", 100),
				"max_idle_time":    env.GetInt("OBJECT_POOL_CUSTOM_MAX_IDLE_TIME", 1800),
				"cleanup_strategy": "adaptive",
				"reset_method":     "Reset",
				"enabled":          env.GetBool("OBJECT_POOL_CUSTOM_ENABLED", false), // disabled by default
				"metrics": map[string]interface{}{
					"track_hits":        true,
					"track_misses":      true,
					"track_allocations": true,
					"track_reuses":      true,
				},
			},

			// Example User pool - developers can add this if they need it
			"user": map[string]interface{}{
				"name":             "User Object Pool",
				"description":      "Pool for User objects used in authentication and user management",
				"type":             "user",
				"pool_size":        env.GetInt("OBJECT_POOL_USER_SIZE", 100),
				"max_idle_time":    env.GetInt("OBJECT_POOL_USER_MAX_IDLE_TIME", 1800), // 30 minutes
				"cleanup_strategy": "lazy",                                             // lazy, eager, or adaptive
				"reset_method":     "Reset",                                            // method name to call on object reset
				"enabled":          env.GetBool("OBJECT_POOL_USER_ENABLED", false),     // disabled by default
				"metrics": map[string]interface{}{
					"track_hits":        true,
					"track_misses":      true,
					"track_allocations": true,
					"track_reuses":      true,
				},
			},
		},

		// Metrics configuration
		"metrics": map[string]interface{}{
			"enabled":             env.GetBool("OBJECT_POOLS_METRICS_ENABLED", true),
			"collection_interval": env.GetInt("OBJECT_POOLS_METRICS_COLLECTION_INTERVAL", 300), // 5 minutes
			"report_interval":     env.GetInt("OBJECT_POOLS_METRICS_REPORT_INTERVAL", 900),     // 15 minutes
			"export_prometheus":   env.GetBool("OBJECT_POOLS_METRICS_PROMETHEUS", false),
			"export_statsd":       env.GetBool("OBJECT_POOLS_METRICS_STATSD", false),
		},

		// Cleanup strategies
		"cleanup_strategies": map[string]interface{}{
			"lazy": map[string]interface{}{
				"description": "Clean up objects only when pool is full",
				"trigger":     "pool_full",
				"batch_size":  env.GetInt("OBJECT_POOLS_CLEANUP_LAZY_BATCH_SIZE", 10),
			},
			"eager": map[string]interface{}{
				"description": "Clean up objects immediately when idle time expires",
				"trigger":     "idle_timeout",
				"batch_size":  env.GetInt("OBJECT_POOLS_CLEANUP_EAGER_BATCH_SIZE", 5),
			},
			"adaptive": map[string]interface{}{
				"description": "Adaptive cleanup based on pool usage patterns",
				"trigger":     "usage_pattern",
				"batch_size":  env.GetInt("OBJECT_POOLS_CLEANUP_ADAPTIVE_BATCH_SIZE", 15),
				"threshold":   env.GetFloat("OBJECT_POOLS_CLEANUP_ADAPTIVE_THRESHOLD", 0.7), // 70% usage
			},
		},
	}
}

// init automatically registers this config with the global config loader
// This ensures the config is available via config.Get("object_pools") and dot notation
func init() {
	go_core.RegisterGlobalConfig("object_pools", ObjectPoolsConfig)
}
