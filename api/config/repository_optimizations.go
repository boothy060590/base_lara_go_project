package config

import (
	"base_lara_go_project/app/core/go_core"
	"base_lara_go_project/app/core/laravel_core/env"
)

// RepositoryOptimizationsConfig returns the repository optimization configuration
// This config enables batch operations, connection pooling, async operations, and pipeline optimizations for repositories
func RepositoryOptimizationsConfig() map[string]interface{} {
	return map[string]interface{}{
		"enabled": env.GetBool("REPOSITORY_OPTIMIZATIONS_ENABLED", true),
		"batch_operations": map[string]interface{}{
			"enabled": env.GetBool("REPOSITORY_BATCH_OPERATIONS_ENABLED", true),
			"create": map[string]interface{}{
				"batch_size":       env.GetInt("REPOSITORY_CREATE_BATCH_SIZE", 100),
				"batch_timeout":    env.GetInt("REPOSITORY_CREATE_BATCH_TIMEOUT", 5),
				"auto_flush":       env.GetBool("REPOSITORY_CREATE_AUTO_FLUSH", true),
				"flush_interval":   env.GetInt("REPOSITORY_CREATE_FLUSH_INTERVAL", 100), // milliseconds
				"max_batch_size":   env.GetInt("REPOSITORY_CREATE_MAX_BATCH_SIZE", 1000),
				"parallel_batches": env.GetInt("REPOSITORY_CREATE_PARALLEL_BATCHES", 4),
			},
			"update": map[string]interface{}{
				"batch_size":       env.GetInt("REPOSITORY_UPDATE_BATCH_SIZE", 100),
				"batch_timeout":    env.GetInt("REPOSITORY_UPDATE_BATCH_TIMEOUT", 5),
				"auto_flush":       env.GetBool("REPOSITORY_UPDATE_AUTO_FLUSH", true),
				"flush_interval":   env.GetInt("REPOSITORY_UPDATE_FLUSH_INTERVAL", 100), // milliseconds
				"max_batch_size":   env.GetInt("REPOSITORY_UPDATE_MAX_BATCH_SIZE", 1000),
				"parallel_batches": env.GetInt("REPOSITORY_UPDATE_PARALLEL_BATCHES", 4),
			},
			"delete": map[string]interface{}{
				"batch_size":       env.GetInt("REPOSITORY_DELETE_BATCH_SIZE", 100),
				"batch_timeout":    env.GetInt("REPOSITORY_DELETE_BATCH_TIMEOUT", 5),
				"auto_flush":       env.GetBool("REPOSITORY_DELETE_AUTO_FLUSH", true),
				"flush_interval":   env.GetInt("REPOSITORY_DELETE_FLUSH_INTERVAL", 100), // milliseconds
				"max_batch_size":   env.GetInt("REPOSITORY_DELETE_MAX_BATCH_SIZE", 1000),
				"parallel_batches": env.GetInt("REPOSITORY_DELETE_PARALLEL_BATCHES", 4),
			},
		},
		"connection_pooling": map[string]interface{}{
			"enabled":               env.GetBool("REPOSITORY_CONNECTION_POOLING_ENABLED", true),
			"pool_size":             env.GetInt("REPOSITORY_POOL_SIZE", 50),
			"max_idle_conns":        env.GetInt("REPOSITORY_MAX_IDLE_CONNS", 25),
			"max_open_conns":        env.GetInt("REPOSITORY_MAX_OPEN_CONNS", 100),
			"conn_max_lifetime":     env.GetInt("REPOSITORY_CONN_MAX_LIFETIME", 3600),
			"conn_max_idle_time":    env.GetInt("REPOSITORY_CONN_MAX_IDLE_TIME", 1800),
			"health_check_interval": env.GetInt("REPOSITORY_HEALTH_CHECK_INTERVAL", 30),
		},
		"async_operations": map[string]interface{}{
			"enabled": env.GetBool("REPOSITORY_ASYNC_OPERATIONS_ENABLED", true),
			"create": map[string]interface{}{
				"async_enabled":   env.GetBool("REPOSITORY_CREATE_ASYNC_ENABLED", true),
				"fire_and_forget": env.GetBool("REPOSITORY_CREATE_FIRE_AND_FORGET", false),
				"max_async_ops":   env.GetInt("REPOSITORY_CREATE_MAX_ASYNC_OPS", 1000),
				"async_timeout":   env.GetInt("REPOSITORY_CREATE_ASYNC_TIMEOUT", 30),
			},
			"update": map[string]interface{}{
				"async_enabled":   env.GetBool("REPOSITORY_UPDATE_ASYNC_ENABLED", true),
				"fire_and_forget": env.GetBool("REPOSITORY_UPDATE_FIRE_AND_FORGET", false),
				"max_async_ops":   env.GetInt("REPOSITORY_UPDATE_MAX_ASYNC_OPS", 1000),
				"async_timeout":   env.GetInt("REPOSITORY_UPDATE_ASYNC_TIMEOUT", 30),
			},
			"read": map[string]interface{}{
				"async_enabled":   env.GetBool("REPOSITORY_READ_ASYNC_ENABLED", false), // Default false for consistency
				"fire_and_forget": env.GetBool("REPOSITORY_READ_FIRE_AND_FORGET", false),
				"max_async_ops":   env.GetInt("REPOSITORY_READ_MAX_ASYNC_OPS", 500),
				"async_timeout":   env.GetInt("REPOSITORY_READ_ASYNC_TIMEOUT", 10),
			},
		},
		"pipeline_operations": map[string]interface{}{
			"enabled": env.GetBool("REPOSITORY_PIPELINE_OPERATIONS_ENABLED", true),
			"query": map[string]interface{}{
				"pipeline_enabled": env.GetBool("REPOSITORY_QUERY_PIPELINE_ENABLED", true),
				"pipeline_size":    env.GetInt("REPOSITORY_QUERY_PIPELINE_SIZE", 50),
				"pipeline_timeout": env.GetInt("REPOSITORY_QUERY_PIPELINE_TIMEOUT", 5),
				"auto_pipeline":    env.GetBool("REPOSITORY_QUERY_AUTO_PIPELINE", true),
			},
		},
		"profiles": map[string]interface{}{
			"web": map[string]interface{}{
				"batch_size":       50,
				"pool_size":        25,
				"async_operations": false,
				"pipeline_enabled": true,
			},
			"api": map[string]interface{}{
				"batch_size":       100,
				"pool_size":        50,
				"async_operations": true,
				"pipeline_enabled": true,
			},
			"background": map[string]interface{}{
				"batch_size":       200,
				"pool_size":        100,
				"async_operations": true,
				"pipeline_enabled": true,
			},
			"streaming": map[string]interface{}{
				"batch_size":       25,
				"pool_size":        75,
				"async_operations": true,
				"pipeline_enabled": true,
			},
			"batch": map[string]interface{}{
				"batch_size":       500,
				"pool_size":        200,
				"async_operations": true,
				"pipeline_enabled": true,
			},
		},
		"monitoring": map[string]interface{}{
			"enabled":          env.GetBool("REPOSITORY_MONITORING_ENABLED", true),
			"metrics_enabled":  env.GetBool("REPOSITORY_METRICS_ENABLED", true),
			"logging_enabled":  env.GetBool("REPOSITORY_LOGGING_ENABLED", true),
			"alerting_enabled": env.GetBool("REPOSITORY_ALERTING_ENABLED", false),
		},
		"performance": map[string]interface{}{
			"target_operations_per_second": env.GetInt("REPOSITORY_TARGET_OPS_PER_SECOND", 10000),
			"target_latency_ms":            env.GetInt("REPOSITORY_TARGET_LATENCY_MS", 10),
			"max_memory_usage_mb":          env.GetInt("REPOSITORY_MAX_MEMORY_USAGE_MB", 1024),
			"auto_tune":                    env.GetBool("REPOSITORY_AUTO_TUNE", true),
		},
	}
}

// init automatically registers this config with the global config loader
func init() {
	go_core.RegisterGlobalConfig("repository_optimizations", RepositoryOptimizationsConfig)
}
