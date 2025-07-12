package go_core

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache defines a generic cache interface for any type
type Cache[T any] interface {
	// Basic operations
	Get(key string) (*T, error)
	Set(key string, value *T, ttl time.Duration) error
	Delete(key string) error
	Has(key string) (bool, error)

	// Context-aware basic operations
	GetWithContext(ctx context.Context, key string) (*T, error)
	SetWithContext(ctx context.Context, key string, value *T, ttl time.Duration) error
	DeleteWithContext(ctx context.Context, key string) error
	HasWithContext(ctx context.Context, key string) (bool, error)

	// Advanced operations
	GetOrSet(key string, factory func() (*T, error), ttl time.Duration) (*T, error)
	Increment(key string, value int64) (int64, error)
	Decrement(key string, value int64) (int64, error)

	// Context-aware advanced operations
	GetOrSetWithContext(ctx context.Context, key string, factory func() (*T, error), ttl time.Duration) (*T, error)
	IncrementWithContext(ctx context.Context, key string, value int64) (int64, error)
	DecrementWithContext(ctx context.Context, key string, value int64) (int64, error)

	// Batch operations
	GetMany(keys []string) (map[string]*T, error)
	SetMany(values map[string]*T, ttl time.Duration) error
	DeleteMany(keys []string) error
	DeletePattern(pattern string) error

	// Context-aware batch operations
	GetManyWithContext(ctx context.Context, keys []string) (map[string]*T, error)
	SetManyWithContext(ctx context.Context, values map[string]*T, ttl time.Duration) error
	DeleteManyWithContext(ctx context.Context, keys []string) error
	DeletePatternWithContext(ctx context.Context, pattern string) error

	// Utility operations
	Flush() error
	FlushWithContext(ctx context.Context) error
	WithContext(ctx context.Context) Cache[T]

	// Performance operations
	GetPerformanceStats() map[string]interface{}
	GetOptimizationStats() map[string]interface{}
}

// redisCache implements Cache[T] with Redis and performance optimizations
type redisCache[T any] struct {
	client *redis.Client
	ctx    context.Context
	// Performance optimizations (safe for cache operations)
	atomicCounter     *AtomicCounter
	jsonEncoderPool   *ObjectPool[json.Encoder]
	jsonDecoderPool   *ObjectPool[json.Decoder]
	performanceFacade *PerformanceFacade
}

// NewRedisCache creates a new Redis cache instance with performance optimizations
func NewRedisCache[T any](client *redis.Client) Cache[T] {
	// Create performance optimizations
	atomicCounter := NewAtomicCounter()
	performanceFacade := NewPerformanceFacade()

	// Create object pools for JSON operations (safe - no database state)
	jsonEncoderPool := NewObjectPool[json.Encoder](50,
		func() json.Encoder { return *json.NewEncoder(nil) },
		func(encoder json.Encoder) json.Encoder { return *json.NewEncoder(nil) },
	)

	jsonDecoderPool := NewObjectPool[json.Decoder](50,
		func() json.Decoder { return *json.NewDecoder(nil) },
		func(decoder json.Decoder) json.Decoder { return *json.NewDecoder(nil) },
	)

	return &redisCache[T]{
		client:            client,
		ctx:               context.Background(),
		atomicCounter:     atomicCounter,
		jsonEncoderPool:   jsonEncoderPool,
		jsonDecoderPool:   jsonDecoderPool,
		performanceFacade: performanceFacade,
	}
}

// Get retrieves a value from cache with performance tracking and atomic counter
func (c *redisCache[T]) Get(key string) (*T, error) {
	return c.GetWithContext(c.ctx, key)
}

// GetWithContext retrieves a value from cache with context support
func (c *redisCache[T]) GetWithContext(ctx context.Context, key string) (*T, error) {
	// Track operation count atomically
	c.atomicCounter.Increment()

	var result *T
	err := c.performanceFacade.Track("cache.get", func() error {
		data, err := c.client.Get(ctx, key).Bytes()
		if err != nil {
			if err == redis.Nil {
				return nil // Key not found
			}
			return err
		}

		var value T
		err = json.Unmarshal(data, &value)
		if err != nil {
			return fmt.Errorf("failed to unmarshal cache value: %w", err)
		}

		result = &value
		return nil
	})

	return result, err
}

// Set stores a value in cache with performance tracking and atomic counter
func (c *redisCache[T]) Set(key string, value *T, ttl time.Duration) error {
	return c.SetWithContext(c.ctx, key, value, ttl)
}

// SetWithContext stores a value in cache with context support
func (c *redisCache[T]) SetWithContext(ctx context.Context, key string, value *T, ttl time.Duration) error {
	// Track operation count atomically
	c.atomicCounter.Increment()

	return c.performanceFacade.Track("cache.set", func() error {
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal cache value: %w", err)
		}

		return c.client.Set(ctx, key, data, ttl).Err()
	})
}

// Delete removes a value from cache
func (c *redisCache[T]) Delete(key string) error {
	return c.DeleteWithContext(c.ctx, key)
}

// DeleteWithContext removes a value from cache with context support
func (c *redisCache[T]) DeleteWithContext(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// Has checks if a key exists in cache
func (c *redisCache[T]) Has(key string) (bool, error) {
	return c.HasWithContext(c.ctx, key)
}

// HasWithContext checks if a key exists in cache with context support
func (c *redisCache[T]) HasWithContext(ctx context.Context, key string) (bool, error) {
	exists, err := c.client.Exists(ctx, key).Result()
	return exists > 0, err
}

// GetOrSet retrieves a value or sets it using a factory function
func (c *redisCache[T]) GetOrSet(key string, factory func() (*T, error), ttl time.Duration) (*T, error) {
	return c.GetOrSetWithContext(c.ctx, key, factory, ttl)
}

// GetOrSetWithContext retrieves a value or sets it using a factory function with context support
func (c *redisCache[T]) GetOrSetWithContext(ctx context.Context, key string, factory func() (*T, error), ttl time.Duration) (*T, error) {
	// Try to get from cache first
	if value, err := c.GetWithContext(ctx, key); err != nil {
		return nil, err
	} else if value != nil {
		return value, nil
	}

	// Value not in cache, create it
	value, err := factory()
	if err != nil {
		return nil, err
	}

	// Store in cache
	err = c.SetWithContext(ctx, key, value, ttl)
	if err != nil {
		// Log error but return value anyway
		// TODO: Add proper logging
	}

	return value, nil
}

// Increment increments a numeric value
func (c *redisCache[T]) Increment(key string, value int64) (int64, error) {
	return c.IncrementWithContext(c.ctx, key, value)
}

// IncrementWithContext increments a numeric value with context support
func (c *redisCache[T]) IncrementWithContext(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.IncrBy(ctx, key, value).Result()
}

// Decrement decrements a numeric value
func (c *redisCache[T]) Decrement(key string, value int64) (int64, error) {
	return c.DecrementWithContext(c.ctx, key, value)
}

// DecrementWithContext decrements a numeric value with context support
func (c *redisCache[T]) DecrementWithContext(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.DecrBy(ctx, key, value).Result()
}

// GetMany retrieves multiple values from cache
func (c *redisCache[T]) GetMany(keys []string) (map[string]*T, error) {
	return c.GetManyWithContext(c.ctx, keys)
}

// GetManyWithContext retrieves multiple values from cache with context support
func (c *redisCache[T]) GetManyWithContext(ctx context.Context, keys []string) (map[string]*T, error) {
	if len(keys) == 0 {
		return make(map[string]*T), nil
	}

	// Get all keys at once
	results, err := c.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	values := make(map[string]*T)
	for i, result := range results {
		if result == nil {
			continue // Key not found
		}

		// Convert interface{} to []byte
		data, ok := result.(string)
		if !ok {
			continue // Invalid data
		}

		var value T
		err := json.Unmarshal([]byte(data), &value)
		if err != nil {
			continue // Invalid JSON
		}

		values[keys[i]] = &value
	}

	return values, nil
}

// SetMany stores multiple values in cache
func (c *redisCache[T]) SetMany(values map[string]*T, ttl time.Duration) error {
	return c.SetManyWithContext(c.ctx, values, ttl)
}

// SetManyWithContext stores multiple values in cache with context support
func (c *redisCache[T]) SetManyWithContext(ctx context.Context, values map[string]*T, ttl time.Duration) error {
	if len(values) == 0 {
		return nil
	}

	// Prepare pipeline
	pipe := c.client.Pipeline()

	for key, value := range values {
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
		}

		pipe.Set(ctx, key, data, ttl)
	}

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	return err
}

// DeleteMany removes multiple values from cache
func (c *redisCache[T]) DeleteMany(keys []string) error {
	return c.DeleteManyWithContext(c.ctx, keys)
}

// DeleteManyWithContext removes multiple values from cache with context support
func (c *redisCache[T]) DeleteManyWithContext(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	return c.client.Del(ctx, keys...).Err()
}

// DeletePattern removes all keys matching a pattern
func (c *redisCache[T]) DeletePattern(pattern string) error {
	return c.DeletePatternWithContext(c.ctx, pattern)
}

// DeletePatternWithContext removes all keys matching a pattern with context support
func (c *redisCache[T]) DeletePatternWithContext(ctx context.Context, pattern string) error {
	// Scan for keys matching pattern
	var keys []string
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}

	// Delete all matching keys
	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}

	return nil
}

// Flush clears all values from cache
func (c *redisCache[T]) Flush() error {
	return c.FlushWithContext(c.ctx)
}

// FlushWithContext clears all values from cache with context support
func (c *redisCache[T]) FlushWithContext(ctx context.Context) error {
	return c.client.FlushAll(ctx).Err()
}

// GetPerformanceStats returns cache performance statistics
func (c *redisCache[T]) GetPerformanceStats() map[string]interface{} {
	stats := c.performanceFacade.GetStats()

	// Add cache-specific stats
	stats["cache"] = map[string]interface{}{
		"operations_count":       c.atomicCounter.Get(),
		"json_encoder_pool_size": len(c.jsonEncoderPool.pool),
		"json_decoder_pool_size": len(c.jsonDecoderPool.pool),
	}

	return stats
}

// GetOptimizationStats returns cache optimization statistics
func (c *redisCache[T]) GetOptimizationStats() map[string]interface{} {
	return map[string]interface{}{
		"atomic_operations":       c.atomicCounter.Get(),
		"json_encoder_pool_usage": len(c.jsonEncoderPool.pool),
		"json_decoder_pool_usage": len(c.jsonDecoderPool.pool),
	}
}

// WithContext returns a cache with context
func (c *redisCache[T]) WithContext(ctx context.Context) Cache[T] {
	return &redisCache[T]{
		client:            c.client,
		ctx:               ctx,
		atomicCounter:     c.atomicCounter,
		jsonEncoderPool:   c.jsonEncoderPool,
		jsonDecoderPool:   c.jsonDecoderPool,
		performanceFacade: c.performanceFacade,
	}
}

// localCache implements Cache[T] with in-memory storage and performance optimizations
type localCache[T any] struct {
	data map[string]cacheItem[T]
	ctx  context.Context
	mu   sync.RWMutex
	// Performance optimizations (safe for cache operations)
	atomicCounter     *AtomicCounter
	performanceFacade *PerformanceFacade
}

type cacheItem[T any] struct {
	value      *T
	expiration time.Time
}

// NewLocalCache creates a new local cache instance
func NewLocalCache[T any]() Cache[T] {
	// Create performance optimizations
	atomicCounter := NewAtomicCounter()
	performanceFacade := NewPerformanceFacade()

	return &localCache[T]{
		data:              make(map[string]cacheItem[T]),
		ctx:               context.Background(),
		atomicCounter:     atomicCounter,
		performanceFacade: performanceFacade,
	}
}

// Get retrieves a value from local cache
func (c *localCache[T]) Get(key string) (*T, error) {
	return c.GetWithContext(c.ctx, key)
}

// GetWithContext retrieves a value from local cache with context support
func (c *localCache[T]) GetWithContext(ctx context.Context, key string) (*T, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.data[key]
	if !exists {
		return nil, nil
	}

	// Check expiration
	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		// Need write lock to delete
		c.mu.RUnlock()
		c.mu.Lock()
		delete(c.data, key)
		c.mu.Unlock()
		c.mu.RLock()
		return nil, nil
	}

	return item.value, nil
}

// Set stores a value in local cache
func (c *localCache[T]) Set(key string, value *T, ttl time.Duration) error {
	return c.SetWithContext(c.ctx, key, value, ttl)
}

// SetWithContext stores a value in local cache with context support
func (c *localCache[T]) SetWithContext(ctx context.Context, key string, value *T, ttl time.Duration) error {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	var expiration time.Time
	if ttl > 0 {
		expiration = time.Now().Add(ttl)
	}

	c.data[key] = cacheItem[T]{
		value:      value,
		expiration: expiration,
	}

	return nil
}

// Delete removes a value from local cache
func (c *localCache[T]) Delete(key string) error {
	return c.DeleteWithContext(c.ctx, key)
}

// DeleteWithContext removes a value from local cache with context support
func (c *localCache[T]) DeleteWithContext(ctx context.Context, key string) error {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
	return nil
}

// Has checks if a key exists in local cache
func (c *localCache[T]) Has(key string) (bool, error) {
	return c.HasWithContext(c.ctx, key)
}

// HasWithContext checks if a key exists in local cache with context support
func (c *localCache[T]) HasWithContext(ctx context.Context, key string) (bool, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.data[key]
	if !exists {
		return false, nil
	}

	// Check expiration
	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		// Need write lock to delete
		c.mu.RUnlock()
		c.mu.Lock()
		delete(c.data, key)
		c.mu.Unlock()
		c.mu.RLock()
		return false, nil
	}

	return true, nil
}

// GetOrSet retrieves a value or sets it using a factory function
func (c *localCache[T]) GetOrSet(key string, factory func() (*T, error), ttl time.Duration) (*T, error) {
	return c.GetOrSetWithContext(c.ctx, key, factory, ttl)
}

// GetOrSetWithContext retrieves a value or sets it using a factory function with context support
func (c *localCache[T]) GetOrSetWithContext(ctx context.Context, key string, factory func() (*T, error), ttl time.Duration) (*T, error) {
	// Try to get from cache first
	if value, err := c.GetWithContext(ctx, key); err != nil {
		return nil, err
	} else if value != nil {
		return value, nil
	}

	// Value not in cache, create it
	value, err := factory()
	if err != nil {
		return nil, err
	}

	// Store in cache
	err = c.SetWithContext(ctx, key, value, ttl)
	if err != nil {
		// Log error but return value anyway
		// TODO: Add proper logging
	}

	return value, nil
}

// Increment increments a numeric value
func (c *localCache[T]) Increment(key string, value int64) (int64, error) {
	return c.IncrementWithContext(c.ctx, key, value)
}

// IncrementWithContext increments a numeric value with context support
func (c *localCache[T]) IncrementWithContext(ctx context.Context, key string, value int64) (int64, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Get current value
	item, exists := c.data[key]
	var currentValue int64 = 0

	if exists && item.value != nil {
		// Try to convert current value to int64
		switch v := any(*item.value).(type) {
		case int64:
			currentValue = v
		case int:
			currentValue = int64(v)
		case int32:
			currentValue = int64(v)
		case int16:
			currentValue = int64(v)
		case int8:
			currentValue = int64(v)
		case uint64:
			currentValue = int64(v)
		case uint:
			currentValue = int64(v)
		case uint32:
			currentValue = int64(v)
		case uint16:
			currentValue = int64(v)
		case uint8:
			currentValue = int64(v)
		default:
			// If not a numeric type, start from 0
			currentValue = 0
		}
	}

	// Calculate new value
	newValue := currentValue + value

	// Store new value using reflection for type conversion
	newValueInterface := any(newValue)
	newValueTyped := newValueInterface.(T)
	c.data[key] = cacheItem[T]{
		value:      &newValueTyped,
		expiration: item.expiration, // Keep existing expiration
	}

	return newValue, nil
}

// Decrement decrements a numeric value
func (c *localCache[T]) Decrement(key string, value int64) (int64, error) {
	return c.DecrementWithContext(c.ctx, key, value)
}

// DecrementWithContext decrements a numeric value with context support
func (c *localCache[T]) DecrementWithContext(ctx context.Context, key string, value int64) (int64, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Get current value
	item, exists := c.data[key]
	var currentValue int64 = 0

	if exists && item.value != nil {
		// Try to convert current value to int64
		switch v := any(*item.value).(type) {
		case int64:
			currentValue = v
		case int:
			currentValue = int64(v)
		case int32:
			currentValue = int64(v)
		case int16:
			currentValue = int64(v)
		case int8:
			currentValue = int64(v)
		case uint64:
			currentValue = int64(v)
		case uint:
			currentValue = int64(v)
		case uint32:
			currentValue = int64(v)
		case uint16:
			currentValue = int64(v)
		case uint8:
			currentValue = int64(v)
		default:
			// If not a numeric type, start from 0
			currentValue = 0
		}
	}

	// Calculate new value
	newValue := currentValue - value

	// Store new value using reflection for type conversion
	newValueInterface := any(newValue)
	newValueTyped := newValueInterface.(T)
	c.data[key] = cacheItem[T]{
		value:      &newValueTyped,
		expiration: item.expiration, // Keep existing expiration
	}

	return newValue, nil
}

// GetMany retrieves multiple values from local cache
func (c *localCache[T]) GetMany(keys []string) (map[string]*T, error) {
	return c.GetManyWithContext(c.ctx, keys)
}

// GetManyWithContext retrieves multiple values from local cache with context support
func (c *localCache[T]) GetManyWithContext(ctx context.Context, keys []string) (map[string]*T, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if len(keys) == 0 {
		return make(map[string]*T), nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	values := make(map[string]*T)
	for _, key := range keys {
		item, exists := c.data[key]
		if !exists {
			continue
		}

		// Check expiration
		if !item.expiration.IsZero() && time.Now().After(item.expiration) {
			continue
		}

		values[key] = item.value
	}

	return values, nil
}

// SetMany stores multiple values in local cache
func (c *localCache[T]) SetMany(values map[string]*T, ttl time.Duration) error {
	return c.SetManyWithContext(c.ctx, values, ttl)
}

// SetManyWithContext stores multiple values in local cache with context support
func (c *localCache[T]) SetManyWithContext(ctx context.Context, values map[string]*T, ttl time.Duration) error {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if len(values) == 0 {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	var expiration time.Time
	if ttl > 0 {
		expiration = time.Now().Add(ttl)
	}

	for key, value := range values {
		c.data[key] = cacheItem[T]{
			value:      value,
			expiration: expiration,
		}
	}

	return nil
}

// DeleteMany removes multiple values from local cache
func (c *localCache[T]) DeleteMany(keys []string) error {
	return c.DeleteManyWithContext(c.ctx, keys)
}

// DeleteManyWithContext removes multiple values from local cache with context support
func (c *localCache[T]) DeleteManyWithContext(ctx context.Context, keys []string) error {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if len(keys) == 0 {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, key := range keys {
		delete(c.data, key)
	}

	return nil
}

// DeletePattern removes all keys matching a pattern
func (c *localCache[T]) DeletePattern(pattern string) error {
	return c.DeletePatternWithContext(c.ctx, pattern)
}

// DeletePatternWithContext removes all keys matching a pattern with context support
func (c *localCache[T]) DeletePatternWithContext(ctx context.Context, pattern string) error {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	var keysToDelete []string
	for key := range c.data {
		if c.matchesPattern(key, pattern) {
			keysToDelete = append(keysToDelete, key)
		}
	}

	for _, key := range keysToDelete {
		delete(c.data, key)
	}

	return nil
}

// Flush clears all values from local cache
func (c *localCache[T]) Flush() error {
	return c.FlushWithContext(c.ctx)
}

// FlushWithContext clears all values from local cache with context support
func (c *localCache[T]) FlushWithContext(ctx context.Context) error {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]cacheItem[T])
	return nil
}

// matchesPattern checks if a key matches a pattern
func (c *localCache[T]) matchesPattern(key, pattern string) bool {
	// Handle exact match
	if pattern == key {
		return true
	}

	// Handle wildcard patterns
	if pattern == "*" {
		return true
	}

	// Handle patterns like "user:*:profile"
	patternParts := strings.Split(pattern, "*")
	if len(patternParts) == 2 {
		prefix := patternParts[0]
		suffix := patternParts[1]

		if strings.HasPrefix(key, prefix) && strings.HasSuffix(key, suffix) {
			return true
		}
	}

	return false
}

// GetPerformanceStats returns local cache performance statistics
func (c *localCache[T]) GetPerformanceStats() map[string]interface{} {
	stats := c.performanceFacade.GetStats()

	// Add cache-specific stats
	stats["cache"] = map[string]interface{}{
		"operations_count": c.atomicCounter.Get(),
		"cache_size":       len(c.data),
	}

	return stats
}

// GetOptimizationStats returns local cache optimization statistics
func (c *localCache[T]) GetOptimizationStats() map[string]interface{} {
	return map[string]interface{}{
		"atomic_operations": c.atomicCounter.Get(),
		"cache_size":        len(c.data),
	}
}

// WithContext returns a cache with context
func (c *localCache[T]) WithContext(ctx context.Context) Cache[T] {
	return &localCache[T]{
		data:              c.data,
		ctx:               ctx,
		atomicCounter:     c.atomicCounter,
		performanceFacade: c.performanceFacade,
	}
}
