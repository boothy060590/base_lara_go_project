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
	// Execute with context awareness (respect context cancellation)
	resultChan := make(chan struct {
		result *T
		err    error
	}, 1)

	execErr := cac.manager.ExecuteWithContext(ctx, func(ctx context.Context) error {
		result, err := cac.cache.Get(key)
		resultChan <- struct {
			result *T
			err    error
		}{result, err}
		return err
	})

	if execErr != nil {
		return nil, execErr
	}

	result := <-resultChan
	return result.result, result.err
}

// Set stores a value with context awareness
func (cac *ContextAwareCache[T]) Set(ctx context.Context, key string, value *T, ttl time.Duration) error {
	// Execute with context awareness (respect context cancellation)
	return cac.manager.ExecuteWithContext(ctx, func(ctx context.Context) error {
		return cac.cache.Set(key, value, ttl)
	})
}

// Delete removes a value with context awareness
func (cac *ContextAwareCache[T]) Delete(ctx context.Context, key string) error {
	// Execute with context awareness (respect context cancellation)
	return cac.manager.ExecuteWithContext(ctx, func(ctx context.Context) error {
		return cac.cache.Delete(key)
	})
}

// Flush clears all values with context awareness
func (cac *ContextAwareCache[T]) Flush(ctx context.Context) error {
	// Execute with context awareness (respect context cancellation)
	return cac.manager.ExecuteWithContext(ctx, func(ctx context.Context) error {
		return cac.cache.Flush()
	})
}

// Has checks if a key exists with context awareness
func (cac *ContextAwareCache[T]) Has(ctx context.Context, key string) (bool, error) {
	// Execute with context awareness (respect context cancellation)
	resultChan := make(chan struct {
		result bool
		err    error
	}, 1)

	execErr := cac.manager.ExecuteWithContext(ctx, func(ctx context.Context) error {
		result, err := cac.cache.Has(key)
		resultChan <- struct {
			result bool
			err    error
		}{result, err}
		return err
	})

	if execErr != nil {
		return false, execErr
	}

	result := <-resultChan
	return result.result, result.err
}

// GetOrSet retrieves a value or stores the result of a callback with context awareness
func (cac *ContextAwareCache[T]) GetOrSet(ctx context.Context, key string, factory func() (*T, error), ttl time.Duration) (*T, error) {
	// Execute with context awareness (respect context cancellation)
	resultChan := make(chan struct {
		result *T
		err    error
	}, 1)

	execErr := cac.manager.ExecuteWithContext(ctx, func(ctx context.Context) error {
		result, err := cac.cache.GetOrSet(key, factory, ttl)
		resultChan <- struct {
			result *T
			err    error
		}{result, err}
		return err
	})

	if execErr != nil {
		return nil, execErr
	}

	result := <-resultChan
	return result.result, result.err
}

// GetPerformanceStats returns cache performance statistics with context awareness
func (cac *ContextAwareCache[T]) GetPerformanceStats(ctx context.Context) map[string]interface{} {
	// Execute with context awareness (respect context cancellation)
	resultChan := make(chan map[string]interface{}, 1)

	execErr := cac.manager.ExecuteWithContext(ctx, func(ctx context.Context) error {
		stats := cac.cache.GetPerformanceStats()
		resultChan <- stats
		return nil
	})

	if execErr != nil {
		return map[string]interface{}{"error": execErr.Error()}
	}

	return <-resultChan
}

// GetManyWithContext retrieves multiple values with context awareness
func (cac *ContextAwareCache[T]) GetManyWithContext(ctx context.Context, keys []string) (map[string]*T, error) {
	// Execute with context awareness (respect context cancellation)
	resultChan := make(chan struct {
		result map[string]*T
		err    error
	}, 1)

	execErr := cac.manager.ExecuteWithContext(ctx, func(ctx context.Context) error {
		result, err := cac.cache.GetManyWithContext(ctx, keys)
		resultChan <- struct {
			result map[string]*T
			err    error
		}{result, err}
		return err
	})

	if execErr != nil {
		return nil, execErr
	}

	result := <-resultChan
	return result.result, result.err
}

// SetManyWithContext stores multiple values with context awareness
func (cac *ContextAwareCache[T]) SetManyWithContext(ctx context.Context, values map[string]*T, ttl time.Duration) error {
	// Execute with context awareness (respect context cancellation)
	return cac.manager.ExecuteWithContext(ctx, func(ctx context.Context) error {
		return cac.cache.SetManyWithContext(ctx, values, ttl)
	})
}

// DeleteManyWithContext removes multiple values with context awareness
func (cac *ContextAwareCache[T]) DeleteManyWithContext(ctx context.Context, keys []string) error {
	// Execute with context awareness (respect context cancellation)
	return cac.manager.ExecuteWithContext(ctx, func(ctx context.Context) error {
		return cac.cache.DeleteManyWithContext(ctx, keys)
	})
}

// DeletePatternWithContext removes all keys matching a pattern with context awareness
func (cac *ContextAwareCache[T]) DeletePatternWithContext(ctx context.Context, pattern string) error {
	// Execute with context awareness (respect context cancellation)
	return cac.manager.ExecuteWithContext(ctx, func(ctx context.Context) error {
		return cac.cache.DeletePatternWithContext(ctx, pattern)
	})
}

// IncrementWithContext increments a numeric value with context awareness
func (cac *ContextAwareCache[T]) IncrementWithContext(ctx context.Context, key string, value int64) (int64, error) {
	// Execute with context awareness (respect context cancellation)
	resultChan := make(chan struct {
		result int64
		err    error
	}, 1)

	execErr := cac.manager.ExecuteWithContext(ctx, func(ctx context.Context) error {
		result, err := cac.cache.IncrementWithContext(ctx, key, value)
		resultChan <- struct {
			result int64
			err    error
		}{result, err}
		return err
	})

	if execErr != nil {
		return 0, execErr
	}

	result := <-resultChan
	return result.result, result.err
}

// DecrementWithContext decrements a numeric value with context awareness
func (cac *ContextAwareCache[T]) DecrementWithContext(ctx context.Context, key string, value int64) (int64, error) {
	// Execute with context awareness (respect context cancellation)
	resultChan := make(chan struct {
		result int64
		err    error
	}, 1)

	execErr := cac.manager.ExecuteWithContext(ctx, func(ctx context.Context) error {
		result, err := cac.cache.DecrementWithContext(ctx, key, value)
		resultChan <- struct {
			result int64
			err    error
		}{result, err}
		return err
	})

	if execErr != nil {
		return 0, execErr
	}

	result := <-resultChan
	return result.result, result.err
}
