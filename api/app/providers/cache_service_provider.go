package providers

import (
	"context"
	"fmt"
	"log"
	"time"

	"base_lara_go_project/app/core"
	"base_lara_go_project/config"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

func RegisterCache() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Get cache configuration
	cacheConfig := config.GetCacheConfig()

	// Create cache driver based on configuration
	var cacheDriver core.CacheInterface

	switch cacheConfig.Store {
	case "redis":
		cacheDriver = createRedisDriver(cacheConfig)
	case "file":
		cacheDriver = createFileDriver(cacheConfig)
	case "array":
		fallthrough
	default:
		cacheDriver = createArrayDriver(cacheConfig)
	}

	// Set up the global cache instance
	core.CacheInstance = cacheDriver

	log.Printf("Cache configured with %s driver", cacheConfig.Store)
}

// createRedisDriver creates a Redis cache driver
func createRedisDriver(config config.CacheConfig) core.CacheInterface {
	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port),
		Password: config.Redis.Password,
		DB:       config.Redis.Database,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Printf("Warning: Redis connection failed: %v", err)
		log.Println("Falling back to array cache driver")
		return createArrayDriver(config)
	}

	log.Println("Redis cache connected successfully")
	return core.NewRedisCacheDriver(client, config.Prefix, config.TTL)
}

// createFileDriver creates a file cache driver
func createFileDriver(config config.CacheConfig) core.CacheInterface {
	return core.NewFileCacheDriver(config.File.Path, config.Prefix, config.TTL)
}

// createArrayDriver creates an array cache driver
func createArrayDriver(config config.CacheConfig) core.CacheInterface {
	return core.NewArrayCacheDriver(config.Prefix, config.TTL)
}
