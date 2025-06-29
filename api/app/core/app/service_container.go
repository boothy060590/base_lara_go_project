package app_core

import (
	"fmt"
	"sync"
)

// ServiceContainer provides Laravel-style dependency injection and service management
type ServiceContainer struct {
	bindings   map[string]interface{}
	singletons map[string]interface{}
	resolvers  map[string]func() interface{}
	mutex      sync.RWMutex
}

// NewServiceContainer creates a new service container
func NewServiceContainer() *ServiceContainer {
	return &ServiceContainer{
		bindings:   make(map[string]interface{}),
		singletons: make(map[string]interface{}),
		resolvers:  make(map[string]func() interface{}),
	}
}

// Bind registers a binding in the container
func (c *ServiceContainer) Bind(abstract string, concrete interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.bindings[abstract] = concrete
}

// BindWithResolver registers a binding with a resolver function
func (c *ServiceContainer) BindWithResolver(abstract string, resolver func() interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.resolvers[abstract] = resolver
}

// Singleton registers a singleton binding
func (c *ServiceContainer) Singleton(abstract string, concrete interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.singletons[abstract] = concrete
}

// SingletonWithResolver registers a singleton with a resolver function
func (c *ServiceContainer) SingletonWithResolver(abstract string, resolver func() interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Store the resolver, but we'll only call it once
	c.resolvers[abstract] = func() interface{} {
		c.mutex.Lock()
		defer c.mutex.Unlock()

		// Check if already resolved
		if instance, exists := c.singletons[abstract]; exists {
			return instance
		}

		// Resolve and store
		instance := resolver()
		c.singletons[abstract] = instance
		return instance
	}
}

// Resolve resolves a binding from the container
func (c *ServiceContainer) Resolve(abstract string) (interface{}, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// Check singletons first
	if instance, exists := c.singletons[abstract]; exists {
		return instance, nil
	}

	// Check bindings
	if concrete, exists := c.bindings[abstract]; exists {
		return concrete, nil
	}

	// Check resolvers
	if resolver, exists := c.resolvers[abstract]; exists {
		return resolver(), nil
	}

	return nil, fmt.Errorf("no binding found for %s", abstract)
}

// ResolveOrFail resolves a binding or panics if not found
func (c *ServiceContainer) ResolveOrFail(abstract string) interface{} {
	instance, err := c.Resolve(abstract)
	if err != nil {
		panic(err)
	}
	return instance
}

// Has checks if a binding exists
func (c *ServiceContainer) Has(abstract string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	_, exists := c.singletons[abstract]
	if exists {
		return true
	}

	_, exists = c.bindings[abstract]
	if exists {
		return true
	}

	_, exists = c.resolvers[abstract]
	return exists
}

// Forget removes a binding from the container
func (c *ServiceContainer) Forget(abstract string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.singletons, abstract)
	delete(c.bindings, abstract)
	delete(c.resolvers, abstract)
}

// Flush clears all bindings
func (c *ServiceContainer) Flush() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.singletons = make(map[string]interface{})
	c.bindings = make(map[string]interface{})
	c.resolvers = make(map[string]func() interface{})
}

// Global service container instance
var App *ServiceContainer

// Initialize the global service container
func init() {
	App = NewServiceContainer()
}

// Helper functions for global access

// Bind globally binds a service
func Bind(abstract string, concrete interface{}) {
	App.Bind(abstract, concrete)
}

// BindWithResolver globally binds a service with resolver
func BindWithResolver(abstract string, resolver func() interface{}) {
	App.BindWithResolver(abstract, resolver)
}

// Singleton globally registers a singleton
func Singleton(abstract string, concrete interface{}) {
	App.Singleton(abstract, concrete)
}

// SingletonWithResolver globally registers a singleton with resolver
func SingletonWithResolver(abstract string, resolver func() interface{}) {
	App.SingletonWithResolver(abstract, resolver)
}

// Resolve globally resolves a service
func Resolve(abstract string) (interface{}, error) {
	return App.Resolve(abstract)
}

// ResolveOrFail globally resolves a service or panics
func ResolveOrFail(abstract string) interface{} {
	return App.ResolveOrFail(abstract)
}

// Has globally checks if a binding exists
func Has(abstract string) bool {
	return App.Has(abstract)
}

// Forget globally removes a binding
func Forget(abstract string) {
	App.Forget(abstract)
}

// Flush globally clears all bindings
func Flush() {
	App.Flush()
}

// Get gets a value from the container (alias for ResolveOrFail)
func (c *ServiceContainer) Get(key string) interface{} {
	return c.ResolveOrFail(key)
}

// Set sets a value in the container (alias for Bind)
func (c *ServiceContainer) Set(key string, value interface{}) {
	c.Bind(key, value)
}
