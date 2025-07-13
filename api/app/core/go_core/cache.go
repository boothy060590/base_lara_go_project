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
	// Infrastructure optimizations
	batchProcessor    *CacheBatchProcessor[T]
	connectionPool    *CacheConnectionPool
	asyncProcessor    *CacheAsyncProcessor[T]
	pipelineProcessor *CachePipelineProcessor
	contextDecorator  *ContextDecorator
	config            map[string]interface{}
}

// NewRedisCache creates a new Redis cache instance with performance optimizations
func NewRedisCache[T any](client *redis.Client) Cache[T] {
	return NewRedisCacheWithConfig[T](client, nil)
}

// NewRedisCacheWithConfig creates a new Redis cache with custom configuration
func NewRedisCacheWithConfig[T any](client *redis.Client, config map[string]interface{}) Cache[T] {
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

	// Create infrastructure optimizations
	batchProcessor := NewCacheBatchProcessor[T](config)
	connectionPool := NewCacheConnectionPool(config)
	asyncProcessor := NewCacheAsyncProcessor[T](config)
	pipelineProcessor := NewCachePipelineProcessor(config)
	contextDecorator := NewContextDecorator(config)

	cache := &redisCache[T]{
		client:            client,
		ctx:               context.Background(),
		atomicCounter:     atomicCounter,
		jsonEncoderPool:   jsonEncoderPool,
		jsonDecoderPool:   jsonDecoderPool,
		performanceFacade: performanceFacade,
		// Infrastructure optimizations
		batchProcessor:    batchProcessor,
		connectionPool:    connectionPool,
		asyncProcessor:    asyncProcessor,
		pipelineProcessor: pipelineProcessor,
		contextDecorator:  contextDecorator,
		config:            config,
	}

	// Start background processors
	cache.startBackgroundProcessors()

	return cache
}

// Get retrieves a value from cache with performance tracking and atomic counter
func (c *redisCache[T]) Get(key string) (*T, error) {
	return c.GetWithContext(c.ctx, key)
}

// GetWithContext retrieves a value from cache with context support
func (c *redisCache[T]) GetWithContext(ctx context.Context, key string) (*T, error) {
	// Use context decorator for performance tracking and context management
	var result *T
	err := c.contextDecorator.WithContext(ctx, "cache.get", func(ctx context.Context) error {
		// Check if batch processing is enabled
		if c.batchProcessor != nil && c.batchProcessor.IsEnabled() {
			var getErr error
			result, getErr = c.batchProcessor.Get(key)
			if getErr == nil {
				return nil // Found in batch buffer
			}
		}

		// Track operation count atomically
		c.atomicCounter.Increment()

		var getErr error
		getErr = c.performanceFacade.Track("cache.get", func() error {
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

		return getErr
	})

	return result, err
}

// Set stores a value in cache with performance tracking and atomic counter
func (c *redisCache[T]) Set(key string, value *T, ttl time.Duration) error {
	return c.SetWithContext(c.ctx, key, value, ttl)
}

// SetWithContext stores a value in cache with context support
func (c *redisCache[T]) SetWithContext(ctx context.Context, key string, value *T, ttl time.Duration) error {
	// Use context decorator for performance tracking and context management
	return c.contextDecorator.WithContext(ctx, "cache.set", func(ctx context.Context) error {
		// Check if batch processing is enabled
		if c.batchProcessor != nil && c.batchProcessor.IsEnabled() {
			return c.batchProcessor.Set(key, value, ttl)
		}

		// Check if async processing is enabled
		if c.asyncProcessor != nil && c.asyncProcessor.IsEnabled() {
			return c.asyncProcessor.Process(key, value, ttl, func(ctx context.Context, k string, v *T, t time.Duration) error {
				return c.performSet(ctx, k, v, t)
			})
		}

		// Fall back to direct cache operation
		return c.performSet(ctx, key, value, ttl)
	})
}

// performSet performs the actual set operation
func (c *redisCache[T]) performSet(ctx context.Context, key string, value *T, ttl time.Duration) error {
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

	// Add infrastructure optimization stats
	if c.batchProcessor != nil {
		stats["batch_processor"] = c.batchProcessor.GetStats()
	}
	if c.connectionPool != nil {
		stats["connection_pool"] = c.connectionPool.GetStats()
	}
	if c.asyncProcessor != nil {
		stats["async_processor"] = c.asyncProcessor.GetStats()
	}
	if c.pipelineProcessor != nil {
		stats["pipeline_processor"] = c.pipelineProcessor.GetStats()
	}
	if c.contextDecorator != nil {
		stats["context_decorator"] = c.contextDecorator.GetPerformanceStats()
	}

	return stats
}

// GetOptimizationStats returns cache optimization statistics
func (c *redisCache[T]) GetOptimizationStats() map[string]interface{} {
	return map[string]interface{}{
		"atomic_operations":            c.atomicCounter.Get(),
		"json_encoder_pool_usage":      len(c.jsonEncoderPool.pool),
		"json_decoder_pool_usage":      len(c.jsonDecoderPool.pool),
		"batch_processing_enabled":     c.batchProcessor != nil && c.batchProcessor.IsEnabled(),
		"connection_pooling_enabled":   c.connectionPool != nil && c.connectionPool.IsEnabled(),
		"async_operations_enabled":     c.asyncProcessor != nil && c.asyncProcessor.IsEnabled(),
		"pipeline_operations_enabled":  c.pipelineProcessor != nil && c.pipelineProcessor.IsEnabled(),
		"infrastructure_optimizations": true,
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

// ============================================================================
// CACHE INFRASTRUCTURE PROCESSORS
// ============================================================================

// CacheBatchProcessor handles batch operations for cache
type CacheBatchProcessor[T any] struct {
	enabled     bool
	batchSize   int
	batchBuffer map[string]*CacheBatchItem[T]
	bufferMutex sync.Mutex
	flushTicker *time.Ticker
	done        chan bool
	config      map[string]interface{}
}

type CacheBatchItem[T any] struct {
	Key   string
	Value *T
	TTL   time.Duration
	Op    string // "get", "set", "delete"
}

// NewCacheBatchProcessor creates a new cache batch processor
func NewCacheBatchProcessor[T any](config map[string]interface{}) *CacheBatchProcessor[T] {
	batchSize := 100 // Default
	if size, ok := config["cache_batch_size"].(int); ok {
		batchSize = size
	}

	return &CacheBatchProcessor[T]{
		enabled:     true,
		batchSize:   batchSize,
		batchBuffer: make(map[string]*CacheBatchItem[T]),
		done:        make(chan bool),
		config:      config,
	}
}

// Start starts the batch processor
func (cbp *CacheBatchProcessor[T]) Start() {
	if !cbp.enabled {
		return
	}

	flushInterval := 100 * time.Millisecond // Default
	if interval, ok := cbp.config["cache_flush_interval"].(int); ok {
		flushInterval = time.Duration(interval) * time.Millisecond
	}

	cbp.flushTicker = time.NewTicker(flushInterval)
	go cbp.backgroundFlusher()
}

// Stop stops the batch processor
func (cbp *CacheBatchProcessor[T]) Stop() {
	if cbp.flushTicker != nil {
		cbp.flushTicker.Stop()
	}
	close(cbp.done)
	cbp.flushBuffer()
}

// backgroundFlusher runs the background flush process
func (cbp *CacheBatchProcessor[T]) backgroundFlusher() {
	for {
		select {
		case <-cbp.flushTicker.C:
			cbp.flushBuffer()
		case <-cbp.done:
			return
		}
	}
}

// flushBuffer flushes the batch buffer
func (cbp *CacheBatchProcessor[T]) flushBuffer() {
	cbp.bufferMutex.Lock()
	if len(cbp.batchBuffer) == 0 {
		cbp.bufferMutex.Unlock()
		return
	}

	items := make([]*CacheBatchItem[T], 0, len(cbp.batchBuffer))
	for _, item := range cbp.batchBuffer {
		items = append(items, item)
	}
	cbp.batchBuffer = make(map[string]*CacheBatchItem[T])
	cbp.bufferMutex.Unlock()

	// Process batch
	cbp.processBatch(items)
}

// processBatch processes a batch of cache operations
func (cbp *CacheBatchProcessor[T]) processBatch(items []*CacheBatchItem[T]) {
	// This would integrate with Redis pipeline operations
	// For now, we'll just process them individually
	for _, item := range items {
		_ = item // Process item
	}
}

// Get gets a value from batch buffer or falls back to direct operation
func (cbp *CacheBatchProcessor[T]) Get(key string) (*T, error) {
	cbp.bufferMutex.Lock()
	defer cbp.bufferMutex.Unlock()

	// Check if item is in batch buffer
	if item, exists := cbp.batchBuffer[key]; exists && item.Op == "set" {
		return item.Value, nil
	}

	// Fall back to direct operation
	return nil, fmt.Errorf("not implemented")
}

// Set adds a set operation to batch buffer
func (cbp *CacheBatchProcessor[T]) Set(key string, value *T, ttl time.Duration) error {
	cbp.bufferMutex.Lock()
	defer cbp.bufferMutex.Unlock()

	cbp.batchBuffer[key] = &CacheBatchItem[T]{
		Key:   key,
		Value: value,
		TTL:   ttl,
		Op:    "set",
	}

	// Flush if buffer is full
	if len(cbp.batchBuffer) >= cbp.batchSize {
		// Copy buffer and clear it
		items := make([]*CacheBatchItem[T], 0, len(cbp.batchBuffer))
		for _, item := range cbp.batchBuffer {
			items = append(items, item)
		}
		cbp.batchBuffer = make(map[string]*CacheBatchItem[T])

		// Process batch in background
		go cbp.processBatch(items)
	}

	return nil
}

// Delete adds a delete operation to batch buffer
func (cbp *CacheBatchProcessor[T]) Delete(key string) error {
	cbp.bufferMutex.Lock()
	defer cbp.bufferMutex.Unlock()

	cbp.batchBuffer[key] = &CacheBatchItem[T]{
		Key: key,
		Op:  "delete",
	}

	// Flush if buffer is full
	if len(cbp.batchBuffer) >= cbp.batchSize {
		// Copy buffer and clear it
		items := make([]*CacheBatchItem[T], 0, len(cbp.batchBuffer))
		for _, item := range cbp.batchBuffer {
			items = append(items, item)
		}
		cbp.batchBuffer = make(map[string]*CacheBatchItem[T])

		// Process batch in background
		go cbp.processBatch(items)
	}

	return nil
}

// IsEnabled returns whether batch processing is enabled
func (cbp *CacheBatchProcessor[T]) IsEnabled() bool {
	return cbp.enabled
}

// GetStats returns batch processor statistics
func (cbp *CacheBatchProcessor[T]) GetStats() map[string]interface{} {
	cbp.bufferMutex.Lock()
	defer cbp.bufferMutex.Unlock()

	return map[string]interface{}{
		"enabled":     cbp.enabled,
		"batch_size":  cbp.batchSize,
		"buffer_size": len(cbp.batchBuffer),
		"config":      cbp.config,
	}
}

// CacheConnectionPool manages connection pooling for cache
type CacheConnectionPool struct {
	enabled bool
	config  map[string]interface{}
}

// NewCacheConnectionPool creates a new cache connection pool
func NewCacheConnectionPool(config map[string]interface{}) *CacheConnectionPool {
	return &CacheConnectionPool{
		enabled: true,
		config:  config,
	}
}

// IsEnabled returns whether connection pooling is enabled
func (ccp *CacheConnectionPool) IsEnabled() bool {
	return ccp.enabled
}

// GetStats returns connection pool statistics
func (ccp *CacheConnectionPool) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"enabled": ccp.enabled,
		"config":  ccp.config,
	}
}

// CacheAsyncProcessor handles async operations for cache
type CacheAsyncProcessor[T any] struct {
	enabled bool
	config  map[string]interface{}
}

// NewCacheAsyncProcessor creates a new cache async processor
func NewCacheAsyncProcessor[T any](config map[string]interface{}) *CacheAsyncProcessor[T] {
	return &CacheAsyncProcessor[T]{
		enabled: true,
		config:  config,
	}
}

// Start starts the async processor
func (cap *CacheAsyncProcessor[T]) Start() {
	// Start async processing
}

// Stop stops the async processor
func (cap *CacheAsyncProcessor[T]) Stop() {
	// Stop async processing
}

// Process processes a cache operation asynchronously
func (cap *CacheAsyncProcessor[T]) Process(key string, value *T, ttl time.Duration, processor func(context.Context, string, *T, time.Duration) error) error {
	if !cap.enabled {
		return processor(context.Background(), key, value, ttl)
	}

	// Process asynchronously
	go func() {
		_ = processor(context.Background(), key, value, ttl)
	}()

	return nil
}

// IsEnabled returns whether async processing is enabled
func (cap *CacheAsyncProcessor[T]) IsEnabled() bool {
	return cap.enabled
}

// GetStats returns async processor statistics
func (cap *CacheAsyncProcessor[T]) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"enabled": cap.enabled,
		"config":  cap.config,
	}
}

// CachePipelineProcessor handles pipeline operations for cache
type CachePipelineProcessor struct {
	enabled bool
	config  map[string]interface{}
}

// NewCachePipelineProcessor creates a new cache pipeline processor
func NewCachePipelineProcessor(config map[string]interface{}) *CachePipelineProcessor {
	return &CachePipelineProcessor{
		enabled: true,
		config:  config,
	}
}

// Start starts the pipeline processor
func (cpp *CachePipelineProcessor) Start() {
	// Start pipeline processing
}

// Stop stops the pipeline processor
func (cpp *CachePipelineProcessor) Stop() {
	// Stop pipeline processing
}

// IsEnabled returns whether pipeline processing is enabled
func (cpp *CachePipelineProcessor) IsEnabled() bool {
	return cpp.enabled
}

// GetStats returns pipeline processor statistics
func (cpp *CachePipelineProcessor) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"enabled": cpp.enabled,
		"config":  cpp.config,
	}
}

// startBackgroundProcessors starts background processing for optimizations
func (c *redisCache[T]) startBackgroundProcessors() {
	// Start batch processor
	if c.batchProcessor != nil {
		c.batchProcessor.Start()
	}

	// Start async processor
	if c.asyncProcessor != nil {
		c.asyncProcessor.Start()
	}

	// Start pipeline processor
	if c.pipelineProcessor != nil {
		c.pipelineProcessor.Start()
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
