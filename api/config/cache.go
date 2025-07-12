package config

import (
	"base_lara_go_project/app/core/go_core"
	"base_lara_go_project/app/core/laravel_core/env"
)

// CacheConfig returns the cache configuration with environment variable fallbacks
// This config defines cache stores, TTL settings, and connection parameters
func CacheConfig() map[string]interface{} {
	return map[string]interface{}{
		"default": env.Get("CACHE_STORE", "local"),
		"stores": map[string]interface{}{
			"local": map[string]interface{}{
				"driver": "local",
				"ttl":    env.GetInt("CACHE_TTL", 3600),
			},
			"redis": map[string]interface{}{
				"driver":     "redis",
				"host":       env.Get("REDIS_HOST", "127.0.0.1"),
				"port":       env.Get("REDIS_PORT", "6379"),
				"password":   env.Get("REDIS_PASSWORD", ""),
				"database":   env.GetInt("REDIS_DB", 0),
				"prefix":     env.Get("CACHE_PREFIX", "laravel_cache"),
				"connection": env.Get("REDIS_CONNECTION", "default"),
			},
		},
	}
}

// init automatically registers this config with the global config loader
// This ensures the cache config is available via config.Get("cache") and dot notation
func init() {
	go_core.RegisterGlobalConfig("cache", CacheConfig)
}
