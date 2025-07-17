package providers

import (
	"base_lara_go_project/app/core/go_core"
	app_core "base_lara_go_project/app/core/go_core"
	"base_lara_go_project/config"
)

// ObjectPoolsServiceProvider provides object pool services
type ObjectPoolsServiceProvider struct {
	BaseServiceProvider
	container *app_core.Container
}

// NewObjectPoolsServiceProvider creates a new object pools service provider
func NewObjectPoolsServiceProvider(container *app_core.Container) *ObjectPoolsServiceProvider {
	return &ObjectPoolsServiceProvider{
		container: container,
	}
}

// Register registers object pool services with the container
func (sp *ObjectPoolsServiceProvider) Register(container *app_core.Container) error {
	// Get object pools configuration
	objectPoolsConfig := config.Get("object_pools")

	// Check if object pools are enabled
	if configMap, ok := objectPoolsConfig.(map[string]interface{}); ok {
		if enabled, ok := configMap["enabled"].(bool); !ok || !enabled {
			// Object pools are disabled, register a no-op manager
			container.Singleton("object_pool_manager", func() (any, error) {
				return go_core.NewObjectPoolManager(), nil
			})
			return nil
		}
	}

	// Merge default pools with custom pools
	mergedPools := sp.mergePools(objectPoolsConfig)

	// Initialize object pools from merged configuration
	if err := go_core.InitializeObjectPools(mergedPools); err != nil {
		// Log error but don't fail - pools will be disabled
		// Note: We'll need to implement proper logging here
		return err
	}

	// Register global object pool manager
	container.Singleton("object_pool_manager", func() (any, error) {
		return go_core.GetGlobalObjectPoolManager(), nil
	})

	// Register individual pools for easy access
	sp.registerIndividualPools(mergedPools, container)

	return nil
}

// Boot performs any necessary bootstrapping
func (sp *ObjectPoolsServiceProvider) Boot(container *app_core.Container) error {
	// Get the object pool manager
	managerInstance, err := container.Resolve("object_pool_manager")
	if err != nil {
		return err
	}

	manager := managerInstance.(*go_core.ObjectPoolManager)

	// Log pool statistics
	pools := manager.GetAllPools()
	// Note: We'll need to implement proper logging here
	_ = pools // Use pools to avoid unused variable error

	return nil
}

// mergePools merges default pools with custom pools from config
func (sp *ObjectPoolsServiceProvider) mergePools(config interface{}) map[string]interface{} {
	merged := make(map[string]interface{})

	configMap, ok := config.(map[string]interface{})
	if !ok {
		return merged
	}

	// Start with default pools (core framework pools)
	if defaultPools, ok := configMap["default_pools"].(map[string]interface{}); ok {
		for name, pool := range defaultPools {
			merged[name] = pool
		}
	}

	// Merge with custom pools (developer-defined pools)
	if customPools, ok := configMap["custom_pools"].(map[string]interface{}); ok {
		for name, pool := range customPools {
			// Custom pools can override default pools
			merged[name] = pool
		}
	}

	return merged
}

// registerIndividualPools registers individual pools for easy access
func (sp *ObjectPoolsServiceProvider) registerIndividualPools(pools map[string]interface{}, container *app_core.Container) {
	manager := go_core.GetGlobalObjectPoolManager()

	// Register only enabled pools
	for poolName, poolConfig := range pools {
		if poolMap, ok := poolConfig.(map[string]interface{}); ok {
			if enabled, ok := poolMap["enabled"].(bool); ok && enabled {
				sp.registerPool(poolName, poolMap, manager, container)
			}
		}
	}
}

// registerPool registers a specific pool type
func (sp *ObjectPoolsServiceProvider) registerPool(name string, poolConfig map[string]interface{}, manager *go_core.ObjectPoolManager, container *app_core.Container) {
	container.Singleton(name+"_pool", func() (any, error) {
		// Get pool type from config
		poolType, _ := poolConfig["type"].(string)

		// Try to get the pool based on type
		switch poolType {
		case "response":
			if pool, exists := manager.GetPool(name); exists {
				return pool, nil
			}
		case "event":
			if pool, exists := manager.GetPool(name); exists {
				return pool, nil
			}
		case "job":
			if pool, exists := manager.GetPool(name); exists {
				return pool, nil
			}
		case "cache":
			if pool, exists := manager.GetPool(name); exists {
				return pool, nil
			}
		case "db_row":
			if pool, exists := manager.GetPool(name); exists {
				return pool, nil
			}
		case "request":
			if pool, exists := manager.GetPool(name); exists {
				return pool, nil
			}
		case "validation":
			if pool, exists := manager.GetPool(name); exists {
				return pool, nil
			}
		case "log_entry":
			if pool, exists := manager.GetPool(name); exists {
				return pool, nil
			}
		case "mail_message":
			if pool, exists := manager.GetPool(name); exists {
				return pool, nil
			}
		case "user":
			// User pool is a custom type - developers need to implement their own pool
			// This is just a placeholder for the service registration
			return nil, nil
		default:
			// Custom pool type - developers need to implement their own pool
			return nil, nil
		}
		return nil, nil
	})
}

// Provides returns the services provided by this provider
func (sp *ObjectPoolsServiceProvider) Provides() []string {
	return []string{
		"object_pool_manager",
		// Individual pools will be registered dynamically based on config
	}
}

// When returns the conditions when this provider should be loaded
func (sp *ObjectPoolsServiceProvider) When() []string {
	return []string{}
}
