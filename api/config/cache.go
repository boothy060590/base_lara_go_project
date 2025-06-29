package config

import "base_lara_go_project/app/core/env"

// CacheConfig returns the cache configuration with fallback values
func CacheConfig() map[string]interface{} {
	return map[string]interface{}{
		"default": env.GetEnv("CACHE_STORE", "local"),
		"stores": map[string]interface{}{
			"local": map[string]interface{}{
				"driver": "local",
				"path":   env.GetEnv("CACHE_PATH", "storage/cache"),
				"prefix": env.GetEnv("CACHE_PREFIX", "base_lara_go_cache_"),
				"ttl":    env.GetEnvInt("CACHE_TTL", 3600),
			},
			"redis": map[string]interface{}{
				"driver":     "redis",
				"host":       env.GetEnv("REDIS_HOST", "localhost"),
				"port":       env.GetEnv("REDIS_PORT", "6379"),
				"password":   env.GetEnv("REDIS_PASSWORD", ""),
				"database":   env.GetEnvInt("REDIS_DB", 0),
				"prefix":     env.GetEnv("CACHE_PREFIX", "base_lara_go_cache_"),
				"ttl":        env.GetEnvInt("CACHE_TTL", 3600),
			},
		},
	}
}
