package config

import (
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// CacheConfig holds the cache configuration
type CacheConfig struct {
	Store  string        `json:"store"`
	Prefix string        `json:"prefix"`
	TTL    time.Duration `json:"ttl"`
	Redis  RedisConfig   `json:"redis"`
	File   FileConfig    `json:"file"`
}

// RedisConfig holds Redis-specific configuration
type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	Database int    `json:"database"`
}

// FileConfig holds file cache configuration
type FileConfig struct {
	Path string `json:"path"`
}

// GetCacheConfig returns the cache configuration
func GetCacheConfig() CacheConfig {
	// Load environment variables
	godotenv.Load()

	// Parse TTL from config (default 1 hour)
	ttlSeconds := 3600
	if ttlStr := getEnv("CACHE_TTL", ""); ttlStr != "" {
		if ttl, err := strconv.Atoi(ttlStr); err == nil {
			ttlSeconds = ttl
		}
	}

	// Parse Redis port
	redisPort := 6379
	if portStr := getEnv("REDIS_PORT", ""); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			redisPort = port
		}
	}

	// Parse Redis database
	redisDB := 0
	if dbStr := getEnv("REDIS_DB", ""); dbStr != "" {
		if db, err := strconv.Atoi(dbStr); err == nil {
			redisDB = db
		}
	}

	// Handle Redis password - treat "null" as empty string
	redisPassword := getEnv("REDIS_PASSWORD", "")
	if redisPassword == "null" {
		redisPassword = ""
	}

	return CacheConfig{
		Store:  getEnv("CACHE_STORE", "array"),
		Prefix: getEnv("CACHE_PREFIX", "base_lara_go_cache_"),
		TTL:    time.Duration(ttlSeconds) * time.Second,
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "redis"),
			Port:     redisPort,
			Password: redisPassword,
			Database: redisDB,
		},
		File: FileConfig{
			Path: getEnv("CACHE_FILE_PATH", "storage/framework/cache/data"),
		},
	}
}
