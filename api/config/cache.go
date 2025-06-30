package config

import "base_lara_go_project/app/core/laravel_core/env"

// CacheConfig returns the cache configuration with fallback values
func CacheConfig() map[string]interface{} {
	return map[string]interface{}{
		"default": env.Get("CACHE_STORE", "local"),
		"stores": map[string]interface{}{
			"local": map[string]interface{}{
				"driver": "local",
				"path":   env.Get("CACHE_PATH", "storage/cache"),
				"prefix": env.Get("CACHE_PREFIX", "base_lara_go_cache_"),
				"ttl":    env.GetInt("CACHE_TTL", 3600),
			},
			"redis": map[string]interface{}{
				"driver":   "redis",
				"host":     env.Get("REDIS_HOST", "localhost"),
				"port":     env.Get("REDIS_PORT", "6379"),
				"password": env.Get("REDIS_PASSWORD", ""),
				"database": env.GetInt("REDIS_DB", 0),
				"prefix":   env.Get("CACHE_PREFIX", "base_lara_go_cache_"),
				"ttl":      env.GetInt("CACHE_TTL", 3600),
			},
		},
	}
}
