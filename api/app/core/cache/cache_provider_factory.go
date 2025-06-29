package cache_core

import (
	"fmt"

	app_core "base_lara_go_project/app/core/app"
	clients_core "base_lara_go_project/app/core/clients"
)

// CacheProviderFactory creates cache providers based on configuration
type CacheProviderFactory struct {
	container *app_core.ServiceContainer
}

// NewCacheProviderFactory creates a new cache provider factory
func NewCacheProviderFactory(container *app_core.ServiceContainer) *CacheProviderFactory {
	return &CacheProviderFactory{container: container}
}

var cacheProviderMap = map[string]func(cfg *clients_core.ClientConfig) app_core.CacheProviderServiceInterface{
	"local": func(cfg *clients_core.ClientConfig) app_core.CacheProviderServiceInterface {
		return NewLocalCacheProvider(NewLocalCacheClient(cfg))
	},
	"redis": func(cfg *clients_core.ClientConfig) app_core.CacheProviderServiceInterface {
		return NewRedisCacheProvider(NewRedisCacheClient(cfg))
	},
}

// Create creates a cache provider based on the driver
func (f *CacheProviderFactory) Create(driver string, cfg *clients_core.ClientConfig) (app_core.CacheProviderServiceInterface, error) {
	constructor, ok := cacheProviderMap[driver]
	if !ok {
		return nil, fmt.Errorf("unsupported cache driver: %s", driver)
	}
	return constructor(cfg), nil
}

// RegisterFromConfig registers a cache provider from configuration
func (f *CacheProviderFactory) RegisterFromConfig(config map[string]interface{}) error {
	// Get store from config
	store, ok := config["store"].(string)
	if !ok {
		return fmt.Errorf("cache store not set in config")
	}

	// Build client config
	clientConfig := f.buildClientConfig(store, config)

	provider, err := f.Create(store, clientConfig)
	if err != nil {
		return err
	}

	f.container.Singleton("cache.provider", provider)
	return nil
}

// buildClientConfig converts cache config to client config
func (f *CacheProviderFactory) buildClientConfig(store string, config map[string]interface{}) *clients_core.ClientConfig {
	clientConfig := &clients_core.ClientConfig{
		Driver:  store,
		Options: config,
	}

	// Set common fields
	if prefix, ok := config["prefix"].(string); ok {
		clientConfig.Options["prefix"] = prefix
	}
	if ttl, ok := config["ttl"].(int); ok {
		clientConfig.Options["ttl"] = ttl
	}

	// Set store-specific fields
	switch store {
	case "redis":
		if redisConfig, ok := config["redis"].(map[string]interface{}); ok {
			if host, ok := redisConfig["host"].(string); ok {
				clientConfig.Options["host"] = host
			}
			if port, ok := redisConfig["port"].(int); ok {
				clientConfig.Options["port"] = port
			}
			if password, ok := redisConfig["password"].(string); ok {
				clientConfig.Options["password"] = password
			}
			if database, ok := redisConfig["database"].(int); ok {
				clientConfig.Options["database"] = database
			}
		}
	case "local":
		// Local cache doesn't need additional configuration
	}

	return clientConfig
}
