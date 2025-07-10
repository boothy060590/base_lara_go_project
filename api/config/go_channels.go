package config

import "base_lara_go_project/app/core/laravel_core/env"

// GoChannelsConfig returns the Go channels configuration with fallback values
func GoChannelsConfig() map[string]interface{} {
	return map[string]interface{}{
		"enabled": env.GetBool("GO_CHANNELS_ENABLED", true),
		"defaults": map[string]interface{}{
			"buffer_size": env.GetInt("GO_CHANNELS_BUFFER_SIZE", 1000),
			"timeout":     env.GetInt("GO_CHANNELS_TIMEOUT", 30),
			"max_workers": env.GetInt("GO_CHANNELS_MAX_WORKERS", 10),
		},
		"patterns": map[string]interface{}{
			"fan_out": map[string]interface{}{
				"enabled":     env.GetBool("GO_CHANNELS_FAN_OUT_ENABLED", true),
				"max_workers": env.GetInt("GO_CHANNELS_FAN_OUT_MAX_WORKERS", 20),
				"buffer_size": env.GetInt("GO_CHANNELS_FAN_OUT_BUFFER_SIZE", 500),
				"timeout":     env.GetInt("GO_CHANNELS_FAN_OUT_TIMEOUT", 60),
			},
			"fan_in": map[string]interface{}{
				"enabled":     env.GetBool("GO_CHANNELS_FAN_IN_ENABLED", true),
				"max_workers": env.GetInt("GO_CHANNELS_FAN_IN_MAX_WORKERS", 10),
				"buffer_size": env.GetInt("GO_CHANNELS_FAN_IN_BUFFER_SIZE", 1000),
				"timeout":     env.GetInt("GO_CHANNELS_FAN_IN_TIMEOUT", 60),
			},
			"pipeline": map[string]interface{}{
				"enabled":     env.GetBool("GO_CHANNELS_PIPELINE_ENABLED", true),
				"max_stages":  env.GetInt("GO_CHANNELS_PIPELINE_MAX_STAGES", 10),
				"buffer_size": env.GetInt("GO_CHANNELS_PIPELINE_BUFFER_SIZE", 500),
				"timeout":     env.GetInt("GO_CHANNELS_PIPELINE_TIMEOUT", 120),
			},
			"batch": map[string]interface{}{
				"enabled":     env.GetBool("GO_CHANNELS_BATCH_ENABLED", true),
				"batch_size":  env.GetInt("GO_CHANNELS_BATCH_SIZE", 100),
				"max_workers": env.GetInt("GO_CHANNELS_BATCH_MAX_WORKERS", 5),
				"timeout":     env.GetInt("GO_CHANNELS_BATCH_TIMEOUT", 60),
			},
			"rate_limiting": map[string]interface{}{
				"enabled":             env.GetBool("GO_CHANNELS_RATE_LIMITING_ENABLED", true),
				"requests_per_second": env.GetInt("GO_CHANNELS_RATE_LIMIT_RPS", 100),
				"burst_size":          env.GetInt("GO_CHANNELS_RATE_LIMIT_BURST", 10),
			},
		},
		"operations": map[string]interface{}{
			"filter": map[string]interface{}{
				"enabled":     env.GetBool("GO_CHANNELS_FILTER_ENABLED", true),
				"buffer_size": env.GetInt("GO_CHANNELS_FILTER_BUFFER_SIZE", 100),
				"timeout":     env.GetInt("GO_CHANNELS_FILTER_TIMEOUT", 30),
			},
			"map": map[string]interface{}{
				"enabled":     env.GetBool("GO_CHANNELS_MAP_ENABLED", true),
				"max_workers": env.GetInt("GO_CHANNELS_MAP_MAX_WORKERS", 10),
				"buffer_size": env.GetInt("GO_CHANNELS_MAP_BUFFER_SIZE", 100),
				"timeout":     env.GetInt("GO_CHANNELS_MAP_TIMEOUT", 30),
			},
			"reduce": map[string]interface{}{
				"enabled":     env.GetBool("GO_CHANNELS_REDUCE_ENABLED", true),
				"buffer_size": env.GetInt("GO_CHANNELS_REDUCE_BUFFER_SIZE", 100),
				"timeout":     env.GetInt("GO_CHANNELS_REDUCE_TIMEOUT", 30),
			},
		},
		"profiles": map[string]interface{}{
			"high_throughput": map[string]interface{}{
				"buffer_size": env.GetInt("GO_CHANNELS_HT_BUFFER_SIZE", 10000),
				"max_workers": env.GetInt("GO_CHANNELS_HT_MAX_WORKERS", 50),
				"timeout":     env.GetInt("GO_CHANNELS_HT_TIMEOUT", 120),
			},
			"low_latency": map[string]interface{}{
				"buffer_size": env.GetInt("GO_CHANNELS_LL_BUFFER_SIZE", 100),
				"max_workers": env.GetInt("GO_CHANNELS_LL_MAX_WORKERS", 20),
				"timeout":     env.GetInt("GO_CHANNELS_LL_TIMEOUT", 10),
			},
			"memory_efficient": map[string]interface{}{
				"buffer_size": env.GetInt("GO_CHANNELS_ME_BUFFER_SIZE", 10),
				"max_workers": env.GetInt("GO_CHANNELS_ME_MAX_WORKERS", 5),
				"timeout":     env.GetInt("GO_CHANNELS_ME_TIMEOUT", 60),
			},
		},
		"optimizations": map[string]interface{}{
			"channel_pooling":      env.GetBool("GO_CHANNELS_POOLING", true),
			"worker_pooling":       env.GetBool("GO_CHANNELS_WORKER_POOLING", true),
			"performance_tracking": env.GetBool("GO_CHANNELS_PERFORMANCE_TRACKING", true),
			"memory_optimization":  env.GetBool("GO_CHANNELS_MEMORY_OPTIMIZATION", true),
		},
		"monitoring": map[string]interface{}{
			"enabled":          env.GetBool("GO_CHANNELS_MONITORING_ENABLED", true),
			"metrics_enabled":  env.GetBool("GO_CHANNELS_METRICS_ENABLED", true),
			"logging_enabled":  env.GetBool("GO_CHANNELS_LOGGING_ENABLED", true),
			"alerting_enabled": env.GetBool("GO_CHANNELS_ALERTING_ENABLED", false),
		},
	}
}
