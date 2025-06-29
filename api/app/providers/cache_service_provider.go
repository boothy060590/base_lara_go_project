package providers

import (
	"log"

	app_core "base_lara_go_project/app/core/app"
	cache_core "base_lara_go_project/app/core/cache"
	"base_lara_go_project/config"
)

func RegisterCache(container *app_core.ServiceContainer) {
	// Get cache configuration from the new config system
	cacheConfig := config.CacheConfig()

	// Create cache provider factory
	factory := cache_core.NewCacheProviderFactory(container)

	// Register cache provider from config
	err := factory.RegisterFromConfig(cacheConfig)
	if err != nil {
		log.Printf("Error registering cache provider: %v", err)
		return
	}

	// Get the store name for logging
	store, _ := cacheConfig["default"].(string)
	log.Printf("Cache configured with %s driver", store)
}
