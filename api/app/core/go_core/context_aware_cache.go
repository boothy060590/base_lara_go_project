package go_core

import (
	"context"
	"time"
)

// ContextAwareCache provides context-aware cache operations
type ContextAwareCache[T any] struct {
	cache   Cache[T]
	manager *ContextManager
}

// NewContextAwareCache creates a new context-aware cache
func NewContextAwareCache[T any](cache Cache[T], manager *ContextManager) *ContextAwareCache[T] {
	if manager == nil {
		manager = NewContextManager(DefaultContextConfig())
	}

	return &ContextAwareCache[T]{
		cache:   cache,
		manager: manager,
	}
}

// Get retrieves a value with context awareness
func (cac *ContextAwareCache[T]) Get(ctx context.Context, key string) (*T, error) {
	// Add context values for tracking
	ctx = context.WithValue(ctx, "cache_get_start", time.Now())
	ctx = context.WithValue(ctx, "cache_key", key)

	var result *T
	var err error

	// Execute with automatic timeout (config-driven)
	timeout := GetOperationTimeout(nil, "cache") // TODO: Pass config map
	err = cac.manager.ExecuteWithTimeout(ctx, timeout, func(ctx context.Context) error {
		result, err = cac.cache.Get(key)
		return err
	})

	// Add context values for completion
	_ = context.WithValue(ctx, "cache_get_end", time.Now())

	return result, err
}

// Set stores a value with context awareness
func (cac *ContextAwareCache[T]) Set(ctx context.Context, key string, value *T, ttl time.Duration) error {
	// Add context values for tracking
	ctx = context.WithValue(ctx, "cache_set_start", time.Now())
	ctx = context.WithValue(ctx, "cache_key", key)

	// Execute with automatic timeout (config-driven)
	timeout := GetOperationTimeout(nil, "cache") // TODO: Pass config map
	err := cac.manager.ExecuteWithTimeout(ctx, timeout, func(ctx context.Context) error {
		return cac.cache.Set(key, value, ttl)
	})

	// Add context values for completion
	_ = context.WithValue(ctx, "cache_set_end", time.Now())

	return err
}

// Delete removes a value with context awareness
func (cac *ContextAwareCache[T]) Delete(ctx context.Context, key string) error {
	// Add context values for tracking
	ctx = context.WithValue(ctx, "cache_delete_start", time.Now())
	ctx = context.WithValue(ctx, "cache_key", key)

	// Execute with automatic timeout (config-driven)
	timeout := GetOperationTimeout(nil, "cache") // TODO: Pass config map
	err := cac.manager.ExecuteWithTimeout(ctx, timeout, func(ctx context.Context) error {
		return cac.cache.Delete(key)
	})

	// Add context values for completion
	_ = context.WithValue(ctx, "cache_delete_end", time.Now())

	return err
}

// Flush clears all values with context awareness
func (cac *ContextAwareCache[T]) Flush(ctx context.Context) error {
	// Add context values for tracking
	ctx = context.WithValue(ctx, "cache_flush_start", time.Now())

	// Execute with automatic timeout (config-driven)
	timeout := GetOperationTimeout(nil, "cache") * 2 // Flush takes longer
	err := cac.manager.ExecuteWithTimeout(ctx, timeout, func(ctx context.Context) error {
		return cac.cache.Flush()
	})

	// Add context values for completion
	_ = context.WithValue(ctx, "cache_flush_end", time.Now())

	return err
}

// Has checks if a key exists with context awareness
func (cac *ContextAwareCache[T]) Has(ctx context.Context, key string) (bool, error) {
	// Add context values for tracking
	ctx = context.WithValue(ctx, "cache_has_start", time.Now())
	ctx = context.WithValue(ctx, "cache_key", key)

	var result bool
	var err error

	// Execute with automatic timeout
	err = cac.manager.ExecuteWithTimeout(ctx, 5*time.Second, func(ctx context.Context) error {
		result, err = cac.cache.Has(key)
		return err
	})

	// Add context values for completion
	_ = context.WithValue(ctx, "cache_has_end", time.Now())

	return result, err
}

// GetOrSet retrieves a value or stores the result of a callback with context awareness
func (cac *ContextAwareCache[T]) GetOrSet(ctx context.Context, key string, factory func() (*T, error), ttl time.Duration) (*T, error) {
	// Add context values for tracking
	ctx = context.WithValue(ctx, "cache_get_or_set_start", time.Now())
	ctx = context.WithValue(ctx, "cache_key", key)

	var result *T
	var err error

	// Execute with automatic timeout
	err = cac.manager.ExecuteWithTimeout(ctx, 30*time.Second, func(ctx context.Context) error {
		result, err = cac.cache.GetOrSet(key, factory, ttl)
		return err
	})

	// Add context values for completion
	_ = context.WithValue(ctx, "cache_get_or_set_end", time.Now())

	return result, err
}

// GetPerformanceStats returns cache performance statistics with context awareness
func (cac *ContextAwareCache[T]) GetPerformanceStats(ctx context.Context) map[string]interface{} {
	// Add context values for tracking
	ctx = context.WithValue(ctx, "cache_performance_stats_start", time.Now())

	var stats map[string]interface{}

	// Execute with automatic timeout
	err := cac.manager.ExecuteWithTimeout(ctx, 5*time.Second, func(ctx context.Context) error {
		stats = cac.cache.GetPerformanceStats()
		return nil
	})

	// Add context values for completion
	_ = context.WithValue(ctx, "cache_performance_stats_end", time.Now())

	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	return stats
}
