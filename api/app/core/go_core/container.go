package go_core

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

// Container is a lightweight, thread-safe dependency injection container
type Container struct {
	services map[string]serviceEntry
	mu       sync.RWMutex
}

// serviceEntry represents a registered service
type serviceEntry struct {
	instance    any
	factory     func() (any, error)
	isSingleton bool
	resolved    bool
}

// NewContainer creates a new service container
func NewContainer() *Container {
	return &Container{
		services: make(map[string]serviceEntry),
	}
}

// Bind registers a factory function for a service
func (c *Container) Bind(key string, factory func() (any, error)) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.services[key] = serviceEntry{
		factory:     factory,
		isSingleton: false,
		resolved:    false,
	}
}

// Singleton registers a factory function for a singleton service
func (c *Container) Singleton(key string, factory func() (any, error)) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.services[key] = serviceEntry{
		factory:     factory,
		isSingleton: true,
		resolved:    false,
	}
}

// Instance registers an existing instance
func (c *Container) Instance(key string, instance any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.services[key] = serviceEntry{
		instance:    instance,
		isSingleton: true,
		resolved:    true,
	}
}

// Resolve retrieves a service from the container
func (c *Container) Resolve(key string) (any, error) {
	c.mu.RLock()
	entry, exists := c.services[key]
	c.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("service '%s' not found", key)
	}

	// If singleton and already resolved, return instance
	if entry.isSingleton && entry.resolved {
		return entry.instance, nil
	}

	// Need to create instance
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	entry, exists = c.services[key]
	if !exists {
		return nil, fmt.Errorf("service '%s' not found", key)
	}

	// If singleton and already resolved (another goroutine created it), return instance
	if entry.isSingleton && entry.resolved {
		return entry.instance, nil
	}

	// Create new instance
	var instance any
	var err error

	if entry.factory != nil {
		instance, err = entry.factory()
		if err != nil {
			return nil, fmt.Errorf("failed to create service '%s': %w", key, err)
		}
	} else {
		instance = entry.instance
	}

	// Update entry
	if entry.isSingleton {
		c.services[key] = serviceEntry{
			instance:    instance,
			isSingleton: true,
			resolved:    true,
		}
	}

	return instance, nil
}

// ResolveTyped retrieves a service and casts it to the specified type
func ResolveTyped[T any](c *Container, key string) (T, error) {
	instance, err := c.Resolve(key)
	if err != nil {
		var zero T
		return zero, err
	}

	typed, ok := instance.(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("service '%s' cannot be cast to %T", key, typed)
	}

	return typed, nil
}

// Has checks if a service is registered
func (c *Container) Has(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, exists := c.services[key]
	return exists
}

// Forget removes a service from the container
func (c *Container) Forget(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.services, key)
}

// Flush removes all services from the container
func (c *Container) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.services = make(map[string]serviceEntry)
}

// Keys returns all registered service keys
func (c *Container) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.services))
	for key := range c.services {
		keys = append(keys, key)
	}

	return keys
}

// ContextContainer is a container that supports context-aware resolution
type ContextContainer struct {
	*Container
	contextServices map[string]contextServiceEntry
	ctxMu           sync.RWMutex
}

type contextServiceEntry struct {
	factory func(context.Context) (any, error)
}

// NewContextContainer creates a new context-aware service container
func NewContextContainer() *ContextContainer {
	return &ContextContainer{
		Container:       NewContainer(),
		contextServices: make(map[string]contextServiceEntry),
	}
}

// BindContext registers a context-aware factory function
func (c *ContextContainer) BindContext(key string, factory func(context.Context) (any, error)) {
	c.ctxMu.Lock()
	defer c.ctxMu.Unlock()

	c.contextServices[key] = contextServiceEntry{
		factory: factory,
	}
}

// ResolveContext retrieves a service with context
func (c *ContextContainer) ResolveContext(ctx context.Context, key string) (any, error) {
	// Check context services first
	c.ctxMu.RLock()
	entry, exists := c.contextServices[key]
	c.ctxMu.RUnlock()

	if exists {
		return entry.factory(ctx)
	}

	// Fall back to regular container
	return c.Container.Resolve(key)
}

// ResolveContextTyped retrieves a context-aware service and casts it to the specified type
func ResolveContextTyped[T any](c *ContextContainer, ctx context.Context, key string) (T, error) {
	instance, err := c.ResolveContext(ctx, key)
	if err != nil {
		var zero T
		return zero, err
	}

	typed, ok := instance.(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("service '%s' cannot be cast to %T", key, typed)
	}

	return typed, nil
}

// AutoWire automatically resolves dependencies for a struct
func (c *Container) AutoWire(target any) error {
	val := reflect.ValueOf(target)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to a struct")
	}

	val = val.Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Check for inject tag
		tag := fieldType.Tag.Get("inject")
		if tag == "" {
			continue
		}

		// Resolve dependency
		instance, err := c.Resolve(tag)
		if err != nil {
			return fmt.Errorf("failed to inject field '%s': %w", fieldType.Name, err)
		}

		// Set field value
		if !field.CanSet() {
			return fmt.Errorf("field '%s' is not settable", fieldType.Name)
		}

		field.Set(reflect.ValueOf(instance))
	}

	return nil
}

// Call invokes a function with automatically resolved dependencies
func (c *Container) Call(fn any, args ...any) ([]any, error) {
	fnVal := reflect.ValueOf(fn)
	if fnVal.Kind() != reflect.Func {
		return nil, fmt.Errorf("fn must be a function")
	}

	fnType := fnVal.Type()
	numIn := fnType.NumIn()

	// Prepare arguments
	callArgs := make([]reflect.Value, numIn)

	// Use provided args first
	for i := 0; i < len(args) && i < numIn; i++ {
		callArgs[i] = reflect.ValueOf(args[i])
	}

	// Resolve remaining dependencies
	for i := len(args); i < numIn; i++ {
		paramType := fnType.In(i)

		// Try to resolve by type name
		instance, err := c.Resolve(paramType.String())
		if err != nil {
			return nil, fmt.Errorf("failed to resolve parameter %d: %w", i, err)
		}

		callArgs[i] = reflect.ValueOf(instance)
	}

	// Call function
	results := fnVal.Call(callArgs)

	// Convert results to []any
	resultValues := make([]any, len(results))
	for i, result := range results {
		resultValues[i] = result.Interface()
	}

	return resultValues, nil
}
