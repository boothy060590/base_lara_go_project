package config

import (
	"base_lara_go_project/app/core/laravel_core/env"
)

// CustomAllocatorsConfig returns the custom allocators configuration with fallback values
func CustomAllocatorsConfig() map[string]interface{} {
	return map[string]interface{}{
		"enabled": env.GetBool("CUSTOM_ALLOCATORS_ENABLED", true),
		"pools": map[string]interface{}{
			"size":             env.GetInt("CUSTOM_ALLOCATORS_POOL_SIZE", 1000),
			"max_object_size":  env.GetInt("CUSTOM_ALLOCATORS_MAX_OBJECT_SIZE", 1024*1024), // 1MB
			"cleanup_interval": env.GetInt("CUSTOM_ALLOCATORS_CLEANUP_INTERVAL", 300),      // 5 minutes
		},
		"strategies": map[string]interface{}{
			"default": env.Get("CUSTOM_ALLOCATORS_STRATEGY", "pool"),
			"pool": map[string]interface{}{
				"enabled": true,
				"size":    env.GetInt("CUSTOM_ALLOCATORS_POOL_SIZE", 1000),
			},
			"slab": map[string]interface{}{
				"enabled":     true,
				"slab_size":   env.GetInt("CUSTOM_ALLOCATORS_SLAB_SIZE", 1024),
				"object_size": env.GetInt("CUSTOM_ALLOCATORS_OBJECT_SIZE", 1024),
			},
			"custom": map[string]interface{}{
				"enabled": false,
			},
		},
		"profiles": map[string]interface{}{
			"web": map[string]interface{}{
				"strategy":         "pool",
				"pool_size":        500,
				"max_object_size":  512 * 1024, // 512KB
				"cleanup_interval": 180,        // 3 minutes
			},
			"api": map[string]interface{}{
				"strategy":         "pool",
				"pool_size":        1000,
				"max_object_size":  1024 * 1024, // 1MB
				"cleanup_interval": 300,         // 5 minutes
			},
			"background": map[string]interface{}{
				"strategy":         "slab",
				"pool_size":        2000,
				"max_object_size":  2048 * 1024, // 2MB
				"cleanup_interval": 600,         // 10 minutes
			},
			"streaming": map[string]interface{}{
				"strategy":         "pool",
				"pool_size":        5000,
				"max_object_size":  4096 * 1024, // 4MB
				"cleanup_interval": 120,         // 2 minutes
			},
			"batch": map[string]interface{}{
				"strategy":         "slab",
				"pool_size":        10000,
				"max_object_size":  8192 * 1024, // 8MB
				"cleanup_interval": 1800,        // 30 minutes
			},
		},
		"optimizations": map[string]interface{}{
			"enable_metrics":   env.GetBool("CUSTOM_ALLOCATORS_METRICS", true),
			"enable_profiling": env.GetBool("CUSTOM_ALLOCATORS_PROFILING", true),
		},
		"memory_limits": map[string]interface{}{
			"max_total_memory": env.GetInt("CUSTOM_ALLOCATORS_MAX_TOTAL_MEMORY", 1024*1024*1024), // 1GB
			"max_pool_memory":  env.GetInt("CUSTOM_ALLOCATORS_MAX_POOL_MEMORY", 256*1024*1024),   // 256MB
			"max_slab_memory":  env.GetInt("CUSTOM_ALLOCATORS_MAX_SLAB_MEMORY", 512*1024*1024),   // 512MB
		},
		"cleanup": map[string]interface{}{
			"enabled":           env.GetBool("CUSTOM_ALLOCATORS_CLEANUP_ENABLED", true),
			"interval":          env.GetInt("CUSTOM_ALLOCATORS_CLEANUP_INTERVAL", 300),
			"usage_threshold":   env.GetFloat("CUSTOM_ALLOCATORS_CLEANUP_THRESHOLD", 0.1), // 10%
			"force_cleanup_age": env.GetInt("CUSTOM_ALLOCATORS_FORCE_CLEANUP_AGE", 3600),  // 1 hour
		},
	}
}
