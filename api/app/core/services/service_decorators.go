package services_core

import (
	"fmt"
	"log"
	"time"

	app_core "base_lara_go_project/app/core/app"
)

// ServiceDecorator wraps a service with additional functionality
type ServiceDecorator[T any] struct {
	service app_core.BaseServiceInterface[T]
}

// NewServiceDecorator creates a new service decorator
func NewServiceDecorator[T any](service app_core.BaseServiceInterface[T]) *ServiceDecorator[T] {
	return &ServiceDecorator[T]{service: service}
}

// LoggingDecorator adds logging to service operations
type LoggingDecorator[T any] struct {
	*ServiceDecorator[T]
	logger *log.Logger
}

// NewLoggingDecorator creates a new logging decorator
func NewLoggingDecorator[T any](service app_core.BaseServiceInterface[T], logger *log.Logger) *LoggingDecorator[T] {
	return &LoggingDecorator[T]{
		ServiceDecorator: NewServiceDecorator(service),
		logger:           logger,
	}
}

// Create logs the create operation
func (l *LoggingDecorator[T]) Create(data map[string]interface{}) (T, error) {
	start := time.Now()
	l.logger.Printf("Creating %T with data: %v", *new(T), data)

	result, err := l.service.Create(data)

	duration := time.Since(start)
	if err != nil {
		l.logger.Printf("Failed to create %T after %v: %v", *new(T), duration, err)
	} else {
		l.logger.Printf("Successfully created %T after %v", *new(T), duration)
	}

	return result, err
}

// FindByID logs the find by ID operation
func (l *LoggingDecorator[T]) FindByID(id uint) (T, error) {
	start := time.Now()
	l.logger.Printf("Finding %T by ID: %d", *new(T), id)

	result, err := l.service.FindByID(id)

	duration := time.Since(start)
	if err != nil {
		l.logger.Printf("Failed to find %T by ID %d after %v: %v", *new(T), id, duration, err)
	} else {
		l.logger.Printf("Successfully found %T by ID %d after %v", *new(T), id, duration)
	}

	return result, err
}

// CachingDecorator adds caching to service operations
type CachingDecorator[T any] struct {
	*ServiceDecorator[T]
	cache app_core.CacheInterface
	ttl   time.Duration
}

// NewCachingDecorator creates a new caching decorator
func NewCachingDecorator[T any](service app_core.BaseServiceInterface[T], cache app_core.CacheInterface, ttl time.Duration) *CachingDecorator[T] {
	return &CachingDecorator[T]{
		ServiceDecorator: NewServiceDecorator(service),
		cache:            cache,
		ttl:              ttl,
	}
}

// FindByIDCached finds by ID with caching
func (c *CachingDecorator[T]) FindByIDCached(id uint) (T, error) {
	cacheKey := fmt.Sprintf("%T:%d", *new(T), id)

	// Try to get from cache first
	if cached, exists := c.cache.Get(cacheKey); exists {
		if result, ok := cached.(T); ok {
			return result, nil
		}
	}

	// If not in cache, get from service
	result, err := c.service.FindByID(id)
	if err != nil {
		return result, err
	}

	// Store in cache
	c.cache.Set(cacheKey, result, c.ttl)

	return result, nil
}

// FindByFieldCached finds by field with caching
func (c *CachingDecorator[T]) FindByFieldCached(field string, value interface{}) (T, error) {
	cacheKey := fmt.Sprintf("%T:%s:%v", *new(T), field, value)

	// Try to get from cache first
	if cached, exists := c.cache.Get(cacheKey); exists {
		if result, ok := cached.(T); ok {
			return result, nil
		}
	}

	// If not in cache, get from service
	result, err := c.service.FindByField(field, value)
	if err != nil {
		return result, err
	}

	// Store in cache
	c.cache.Set(cacheKey, result, c.ttl)

	return result, nil
}

// InvalidateCache invalidates cache for a specific ID
func (c *CachingDecorator[T]) InvalidateCache(id uint) error {
	cacheKey := fmt.Sprintf("%T:%d", *new(T), id)
	return c.cache.Delete(cacheKey)
}

// InvalidateAllCache invalidates all cache for this type
func (c *CachingDecorator[T]) InvalidateAllCache() error {
	// This is a simplified implementation
	// In a real scenario, you might want to use cache tags or patterns
	return nil
}

// AuditingDecorator adds auditing to service operations
type AuditingDecorator[T any] struct {
	*ServiceDecorator[T]
	auditLogger AuditLogger
}

// AuditLogger defines the interface for audit logging
type AuditLogger interface {
	Log(action string, table string, recordID uint, oldValues, newValues interface{}) error
	GetAuditLog(table string, recordID uint) ([]AuditLog, error)
}

// NewAuditingDecorator creates a new auditing decorator
func NewAuditingDecorator[T any](service app_core.BaseServiceInterface[T], auditLogger AuditLogger) *AuditingDecorator[T] {
	return &AuditingDecorator[T]{
		ServiceDecorator: NewServiceDecorator(service),
		auditLogger:      auditLogger,
	}
}

// Create audits the create operation
func (a *AuditingDecorator[T]) Create(data map[string]interface{}) (T, error) {
	result, err := a.service.Create(data)
	if err != nil {
		return result, err
	}

	// Extract ID from result (assuming it has a GetID method)
	if withID, ok := any(result).(interface{ GetID() uint }); ok {
		a.auditLogger.Log("CREATE", fmt.Sprintf("%T", *new(T)), withID.GetID(), nil, data)
	}

	return result, nil
}

// Update audits the update operation
func (a *AuditingDecorator[T]) Update(id uint, data map[string]interface{}) (T, error) {
	// Get old values before update
	oldResult, _ := a.service.FindByID(id)

	result, err := a.service.Update(id, data)
	if err != nil {
		return result, err
	}

	// Log the audit
	a.auditLogger.Log("UPDATE", fmt.Sprintf("%T", *new(T)), id, oldResult, data)

	return result, nil
}

// Delete audits the delete operation
func (a *AuditingDecorator[T]) Delete(id uint) error {
	// Get old values before delete
	oldResult, _ := a.service.FindByID(id)

	err := a.service.Delete(id)
	if err != nil {
		return err
	}

	// Log the audit
	a.auditLogger.Log("DELETE", fmt.Sprintf("%T", *new(T)), id, oldResult, nil)

	return nil
}

// CompositeDecorator combines multiple decorators
type CompositeDecorator[T any] struct {
	*ServiceDecorator[T]
	decorators []app_core.BaseServiceInterface[T]
}

// NewCompositeDecorator creates a new composite decorator
func NewCompositeDecorator[T any](service app_core.BaseServiceInterface[T], decorators ...app_core.BaseServiceInterface[T]) *CompositeDecorator[T] {
	return &CompositeDecorator[T]{
		ServiceDecorator: NewServiceDecorator(service),
		decorators:       decorators,
	}
}

// Create applies all decorators to create operation
func (c *CompositeDecorator[T]) Create(data map[string]interface{}) (T, error) {
	result, err := c.service.Create(data)
	if err != nil {
		return result, err
	}

	// Apply all decorators
	for _, decorator := range c.decorators {
		if createDecorator, ok := decorator.(interface {
			Create(data map[string]interface{}) (T, error)
		}); ok {
			_, _ = createDecorator.Create(data)
		}
	}

	return result, nil
}

// FindByID applies all decorators to find by ID operation
func (c *CompositeDecorator[T]) FindByID(id uint) (T, error) {
	result, err := c.service.FindByID(id)
	if err != nil {
		return result, err
	}

	// Apply all decorators
	for _, decorator := range c.decorators {
		if findDecorator, ok := decorator.(interface {
			FindByID(id uint) (T, error)
		}); ok {
			_, _ = findDecorator.FindByID(id)
		}
	}

	return result, nil
}
