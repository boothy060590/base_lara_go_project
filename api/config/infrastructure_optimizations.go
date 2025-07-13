package config

import (
	"base_lara_go_project/app/core/go_core"
	"base_lara_go_project/app/core/laravel_core/env"
)

// InfrastructureOptimizationsConfig returns the infrastructure optimization configuration
// This config enables batch operations, connection pooling, async operations, and pipeline optimizations
func InfrastructureOptimizationsConfig() map[string]interface{} {
	return map[string]interface{}{
		"enabled": env.GetBool("INFRASTRUCTURE_OPTIMIZATIONS_ENABLED", true),
		"batch_operations": map[string]interface{}{
			"enabled": env.GetBool("BATCH_OPERATIONS_ENABLED", true),
			"database": map[string]interface{}{
				"batch_size":       env.GetInt("DB_BATCH_SIZE", 100),
				"batch_timeout":    env.GetInt("DB_BATCH_TIMEOUT", 5),
				"auto_flush":       env.GetBool("DB_AUTO_FLUSH", true),
				"flush_interval":   env.GetInt("DB_FLUSH_INTERVAL", 100), // milliseconds
				"max_batch_size":   env.GetInt("DB_MAX_BATCH_SIZE", 1000),
				"parallel_batches": env.GetInt("DB_PARALLEL_BATCHES", 4),
			},
			"redis": map[string]interface{}{
				"batch_size":       env.GetInt("REDIS_BATCH_SIZE", 50),
				"pipeline_enabled": env.GetBool("REDIS_PIPELINE_ENABLED", true),
				"pipeline_size":    env.GetInt("REDIS_PIPELINE_SIZE", 100),
				"auto_flush":       env.GetBool("REDIS_AUTO_FLUSH", true),
				"flush_interval":   env.GetInt("REDIS_FLUSH_INTERVAL", 50), // milliseconds
			},
			"sqs": map[string]interface{}{
				"batch_size":     env.GetInt("SQS_BATCH_SIZE", 10),
				"batch_timeout":  env.GetInt("SQS_BATCH_TIMEOUT", 5),
				"auto_flush":     env.GetBool("SQS_AUTO_FLUSH", true),
				"flush_interval": env.GetInt("SQS_FLUSH_INTERVAL", 200), // milliseconds
				"max_batch_size": env.GetInt("SQS_MAX_BATCH_SIZE", 10),
			},
		},
		"connection_pooling": map[string]interface{}{
			"enabled": env.GetBool("CONNECTION_POOLING_ENABLED", true),
			"database": map[string]interface{}{
				"pool_size":             env.GetInt("DB_POOL_SIZE", 50),
				"max_idle_conns":        env.GetInt("DB_MAX_IDLE_CONNS", 25),
				"max_open_conns":        env.GetInt("DB_MAX_OPEN_CONNS", 100),
				"conn_max_lifetime":     env.GetInt("DB_CONN_MAX_LIFETIME", 3600),
				"conn_max_idle_time":    env.GetInt("DB_CONN_MAX_IDLE_TIME", 1800),
				"health_check_interval": env.GetInt("DB_HEALTH_CHECK_INTERVAL", 30),
			},
			"redis": map[string]interface{}{
				"pool_size":      env.GetInt("REDIS_POOL_SIZE", 100),
				"min_idle_conns": env.GetInt("REDIS_MIN_IDLE_CONNS", 10),
				"max_retries":    env.GetInt("REDIS_MAX_RETRIES", 3),
				"dial_timeout":   env.GetInt("REDIS_DIAL_TIMEOUT", 5),
				"read_timeout":   env.GetInt("REDIS_READ_TIMEOUT", 3),
				"write_timeout":  env.GetInt("REDIS_WRITE_TIMEOUT", 3),
				"pool_timeout":   env.GetInt("REDIS_POOL_TIMEOUT", 4),
				"idle_timeout":   env.GetInt("REDIS_IDLE_TIMEOUT", 300),
			},
		},
		"async_operations": map[string]interface{}{
			"enabled": env.GetBool("ASYNC_OPERATIONS_ENABLED", true),
			"database": map[string]interface{}{
				"async_writes":    env.GetBool("DB_ASYNC_WRITES", true),
				"async_reads":     env.GetBool("DB_ASYNC_READS", false),
				"fire_and_forget": env.GetBool("DB_FIRE_AND_FORGET", false),
				"max_async_ops":   env.GetInt("DB_MAX_ASYNC_OPS", 1000),
				"async_timeout":   env.GetInt("DB_ASYNC_TIMEOUT", 30),
			},
			"redis": map[string]interface{}{
				"async_operations": env.GetBool("REDIS_ASYNC_OPERATIONS", true),
				"fire_and_forget":  env.GetBool("REDIS_FIRE_AND_FORGET", false),
				"max_async_ops":    env.GetInt("REDIS_MAX_ASYNC_OPS", 5000),
				"async_timeout":    env.GetInt("REDIS_ASYNC_TIMEOUT", 10),
			},
			"sqs": map[string]interface{}{
				"async_send":      env.GetBool("SQS_ASYNC_SEND", true),
				"fire_and_forget": env.GetBool("SQS_FIRE_AND_FORGET", false),
				"max_async_ops":   env.GetInt("SQS_MAX_ASYNC_OPS", 1000),
				"async_timeout":   env.GetInt("SQS_ASYNC_TIMEOUT", 30),
			},
		},
		"pipeline_operations": map[string]interface{}{
			"enabled": env.GetBool("PIPELINE_OPERATIONS_ENABLED", true),
			"redis": map[string]interface{}{
				"pipeline_enabled":  env.GetBool("REDIS_PIPELINE_ENABLED", true),
				"pipeline_size":     env.GetInt("REDIS_PIPELINE_SIZE", 100),
				"pipeline_timeout":  env.GetInt("REDIS_PIPELINE_TIMEOUT", 5),
				"auto_pipeline":     env.GetBool("REDIS_AUTO_PIPELINE", true),
				"pipeline_commands": []string{"SET", "GET", "DEL", "EXPIRE"},
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
			"enabled":          env.GetBool("INFRASTRUCTURE_MONITORING_ENABLED", true),
			"metrics_enabled":  env.GetBool("INFRASTRUCTURE_METRICS_ENABLED", true),
			"logging_enabled":  env.GetBool("INFRASTRUCTURE_LOGGING_ENABLED", true),
			"alerting_enabled": env.GetBool("INFRASTRUCTURE_ALERTING_ENABLED", false),
		},
		"performance": map[string]interface{}{
			"target_events_per_second": env.GetInt("TARGET_EVENTS_PER_SECOND", 10000),
			"target_latency_ms":        env.GetInt("TARGET_LATENCY_MS", 10),
			"max_memory_usage_mb":      env.GetInt("MAX_MEMORY_USAGE_MB", 1024),
			"auto_tune":                env.GetBool("INFRASTRUCTURE_AUTO_TUNE", true),
		},
	}
}

// init automatically registers this config with the global config loader
func init() {
	go_core.RegisterGlobalConfig("infrastructure_optimizations", InfrastructureOptimizationsConfig)
}
