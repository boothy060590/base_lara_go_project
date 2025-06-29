package providers

import (
	facades_core "base_lara_go_project/app/core/facades"
	"base_lara_go_project/app/services"
	"log"
	"sync"
)

// ServiceContainer holds all registered services
type ServiceContainer struct {
	services map[string]interface{}
	mutex    sync.RWMutex
}

// NewServiceContainer creates a new service container
func NewServiceContainer() *ServiceContainer {
	return &ServiceContainer{
		services: make(map[string]interface{}),
	}
}

// Register registers a service with a name
func (sc *ServiceContainer) Register(name string, service interface{}) {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()
	sc.services[name] = service
}

// Get retrieves a service by name
func (sc *ServiceContainer) Get(name string) (interface{}, bool) {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()
	service, exists := sc.services[name]
	return service, exists
}

// Global service container instance
var GlobalServiceContainer = NewServiceContainer()

// RegisterServices registers all services with facades
func RegisterServices() {
	// Create base user service
	userService, err := services.NewUserService()
	if err == nil {
		// Register the base service
		GlobalServiceContainer.Register("user", userService)

		// Set up the service facade
		facades_core.SetUserService(userService)

		log.Println("User service registered successfully")
	} else {
		log.Printf("Failed to register user service: %v", err)
	}

	// Add more services here as they are created
	// Example:
	// roleService, err := services.NewRoleService()
	// if err == nil {
	//     GlobalServiceContainer.Register("role", roleService)
	//     facades.SetRoleService(roleService)
	// }
}

// GetUserService is a global helper to get the user service
func GetUserService() (*services.UserService, bool) {
	if service, exists := GlobalServiceContainer.Get("user"); exists {
		if userService, ok := service.(*services.UserService); ok {
			return userService, true
		}
	}
	return nil, false
}
