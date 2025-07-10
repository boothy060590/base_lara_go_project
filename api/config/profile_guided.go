package config

import (
	"base_lara_go_project/app/core/laravel_core/env"
)

// ProfileGuidedConfig returns the profile-guided optimization configuration with fallback values
func ProfileGuidedConfig() map[string]interface{} {
	return map[string]interface{}{
		"enabled": env.GetBool("PROFILE_GUIDED_ENABLED", true),
		"sampling": map[string]interface{}{
			"interval":    env.GetInt("PROFILE_GUIDED_SAMPLING_INTERVAL", 1),
			"min_samples": env.GetInt("PROFILE_GUIDED_MIN_SAMPLES", 100),
			"max_samples": env.GetInt("PROFILE_GUIDED_MAX_SAMPLES", 1000),
		},
		"optimization": map[string]interface{}{
			"interval":          env.GetInt("PROFILE_GUIDED_OPTIMIZATION_INTERVAL", 30),
			"max_optimizations": env.GetInt("PROFILE_GUIDED_MAX_OPTIMIZATIONS", 10),
			"auto_tuning":       env.GetBool("PROFILE_GUIDED_AUTO_TUNING", true),
		},
		"profiles": map[string]interface{}{
			"web": map[string]interface{}{
				"sampling_interval":     1,
				"optimization_interval": 30,
				"min_samples":           50,
				"max_optimizations":     5,
			},
			"api": map[string]interface{}{
				"sampling_interval":     2,
				"optimization_interval": 60,
				"min_samples":           100,
				"max_optimizations":     8,
			},
			"background": map[string]interface{}{
				"sampling_interval":     5,
				"optimization_interval": 120,
				"min_samples":           200,
				"max_optimizations":     15,
			},
			"streaming": map[string]interface{}{
				"sampling_interval":     1,
				"optimization_interval": 15,
				"min_samples":           25,
				"max_optimizations":     3,
			},
			"batch": map[string]interface{}{
				"sampling_interval":     10,
				"optimization_interval": 300,
				"min_samples":           500,
				"max_optimizations":     20,
			},
		},
		"optimizations": map[string]interface{}{
			"enable_metrics":   env.GetBool("PROFILE_GUIDED_METRICS", true),
			"enable_profiling": env.GetBool("PROFILE_GUIDED_PROFILING", true),
		},
		"thresholds": map[string]interface{}{
			"cpu_usage": map[string]interface{}{
				"min": env.GetFloat("PROFILE_GUIDED_CPU_MIN", 20.0),
				"max": env.GetFloat("PROFILE_GUIDED_CPU_MAX", 90.0),
			},
			"memory_usage": map[string]interface{}{
				"min": env.GetInt("PROFILE_GUIDED_MEMORY_MIN", 50*1024*1024),   // 50MB
				"max": env.GetInt("PROFILE_GUIDED_MEMORY_MAX", 1024*1024*1024), // 1GB
			},
			"goroutines": map[string]interface{}{
				"min": env.GetInt("PROFILE_GUIDED_GOROUTINES_MIN", 10),
				"max": env.GetInt("PROFILE_GUIDED_GOROUTINES_MAX", 1000),
			},
			"error_rate": map[string]interface{}{
				"min": env.GetFloat("PROFILE_GUIDED_ERROR_RATE_MIN", 0.01), // 1%
				"max": env.GetFloat("PROFILE_GUIDED_ERROR_RATE_MAX", 0.1),  // 10%
			},
		},
	}
}
