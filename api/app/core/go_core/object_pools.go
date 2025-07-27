package go_core

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

// PoolableObject interface for objects that can be pooled
type PoolableObject interface {
	Reset() // Reset method to clear object state
}

// ObjectPoolConfig represents configuration for a single object pool
type ObjectPoolConfig struct {
	Name            string
	Description     string
	Type            string
	PoolSize        int
	MaxIdleTime     time.Duration
	CleanupStrategy string
	ResetMethod     string
	Enabled         bool
	Metrics         PoolMetricsConfig
}

// PoolMetricsConfig represents metrics configuration for a pool
type PoolMetricsConfig struct {
	TrackHits        bool
	TrackMisses      bool
	TrackAllocations bool
	TrackReuses      bool
}

// ConfigurableObjectPool represents a configurable object pool
type ConfigurableObjectPool[T PoolableObject] struct {
	config        ObjectPoolConfig
	pool          *sync.Pool
	metrics       *PoolMetrics
	cleanupTicker *time.Ticker
	ctx           context.Context
	cancel        context.CancelFunc
	mu            sync.RWMutex
}

// PoolMetrics tracks pool usage metrics
type PoolMetrics struct {
	Hits        int64
	Misses      int64
	Allocations int64
	Reuses      int64
	CurrentSize int64
	MaxSize     int64
	LastAccess  time.Time
	mu          sync.RWMutex
}

// ObjectPoolManager manages multiple object pools
type ObjectPoolManager struct {
	pools        map[string]interface{}
	configs      map[string]ObjectPoolConfig
	globalConfig map[string]interface{}
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewConfigurableObjectPool creates a new configurable object pool
func NewConfigurableObjectPool[T PoolableObject](config ObjectPoolConfig, newFunc func() T) *ConfigurableObjectPool[T] {
	ctx, cancel := context.WithCancel(context.Background())

	pool := &ConfigurableObjectPool[T]{
		config: config,
		metrics: &PoolMetrics{
			MaxSize: int64(config.PoolSize),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// Create the sync.Pool with the provided new function
	pool.pool = &sync.Pool{
		New: func() interface{} {
			atomic.AddInt64(&pool.metrics.Allocations, 1)
			atomic.AddInt64(&pool.metrics.CurrentSize, 1)
			return newFunc()
		},
	}

	// Start cleanup routine if enabled
	if config.Enabled && config.MaxIdleTime > 0 {
		pool.startCleanupRoutine()
	}

	return pool
}

// Get retrieves an object from the pool
func (op *ConfigurableObjectPool[T]) Get() T {
	op.metrics.mu.Lock()
	op.metrics.LastAccess = time.Now()
	op.metrics.mu.Unlock()

	obj := op.pool.Get().(T)

	// Check if this is a reuse (not a new allocation)
	if op.metrics.Allocations > op.metrics.Reuses {
		atomic.AddInt64(&op.metrics.Reuses, 1)
		if op.config.Metrics.TrackReuses {
			atomic.AddInt64(&op.metrics.Hits, 1)
		}
	} else {
		if op.config.Metrics.TrackMisses {
			atomic.AddInt64(&op.metrics.Misses, 1)
		}
	}

	// Reset the object before returning
	obj.Reset()

	return obj
}

// Put returns an object to the pool
func (op *ConfigurableObjectPool[T]) Put(obj T) {
	// Check if obj is zero value (equivalent to nil for our purposes)
	var zero T
	if reflect.DeepEqual(obj, zero) {
		return
	}

	op.pool.Put(obj)
	atomic.AddInt64(&op.metrics.CurrentSize, -1)
}

// GetMetrics returns current pool metrics
func (op *ConfigurableObjectPool[T]) GetMetrics() PoolMetrics {
	op.metrics.mu.RLock()
	defer op.metrics.mu.RUnlock()

	return PoolMetrics{
		Hits:        op.metrics.Hits,
		Misses:      op.metrics.Misses,
		Allocations: op.metrics.Allocations,
		Reuses:      op.metrics.Reuses,
		CurrentSize: op.metrics.CurrentSize,
		MaxSize:     op.metrics.MaxSize,
		LastAccess:  op.metrics.LastAccess,
	}
}

// ResetMetrics resets all metrics to zero
func (op *ConfigurableObjectPool[T]) ResetMetrics() {
	op.metrics.mu.Lock()
	defer op.metrics.mu.Unlock()

	op.metrics.Hits = 0
	op.metrics.Misses = 0
	op.metrics.Allocations = 0
	op.metrics.Reuses = 0
	op.metrics.CurrentSize = 0
}

// Close shuts down the pool and cleanup routines
func (op *ConfigurableObjectPool[T]) Close() {
	op.cancel()
	if op.cleanupTicker != nil {
		op.cleanupTicker.Stop()
	}
}

// startCleanupRoutine starts the cleanup routine based on the strategy
func (op *ConfigurableObjectPool[T]) startCleanupRoutine() {
	switch op.config.CleanupStrategy {
	case "lazy":
		// Lazy cleanup - only when pool is full
		// This is handled by sync.Pool automatically
	case "eager":
		// Eager cleanup - periodic cleanup
		op.cleanupTicker = time.NewTicker(op.config.MaxIdleTime / 2)
		go op.eagerCleanupRoutine()
	case "adaptive":
		// Adaptive cleanup - based on usage patterns
		op.cleanupTicker = time.NewTicker(op.config.MaxIdleTime / 4)
		go op.adaptiveCleanupRoutine()
	}
}

// eagerCleanupRoutine performs eager cleanup
func (op *ConfigurableObjectPool[T]) eagerCleanupRoutine() {
	for {
		select {
		case <-op.cleanupTicker.C:
			op.performCleanup()
		case <-op.ctx.Done():
			return
		}
	}
}

// adaptiveCleanupRoutine performs adaptive cleanup based on usage
func (op *ConfigurableObjectPool[T]) adaptiveCleanupRoutine() {
	for {
		select {
		case <-op.cleanupTicker.C:
			metrics := op.GetMetrics()
			usageRatio := float64(metrics.CurrentSize) / float64(metrics.MaxSize)

			// If usage is low, perform cleanup
			if usageRatio < 0.3 {
				op.performCleanup()
			}
		case <-op.ctx.Done():
			return
		}
	}
}

// performCleanup performs the actual cleanup
func (op *ConfigurableObjectPool[T]) performCleanup() {
	// For sync.Pool, cleanup is mostly automatic
	// We can reset metrics periodically
	op.ResetMetrics()
}

// NewObjectPoolManager creates a new object pool manager
func NewObjectPoolManager() *ObjectPoolManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &ObjectPoolManager{
		pools:   make(map[string]interface{}),
		configs: make(map[string]ObjectPoolConfig),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// RegisterPool registers a new object pool with the manager
func (om *ObjectPoolManager) RegisterPool(
	name string,
	config ObjectPoolConfig,
	newFunc interface{},
) interface{} {
	om.mu.Lock()
	defer om.mu.Unlock()

	// Create pool based on type using reflection
	var pool interface{}
	switch config.Type {
	case "user":
		if fn, ok := newFunc.(func() *PooledUser); ok {
			pool = NewConfigurableObjectPool[*PooledUser](config, fn)
		}
	case "event":
		if fn, ok := newFunc.(func() *PooledEvent); ok {
			pool = NewConfigurableObjectPool[*PooledEvent](config, fn)
		}
	case "response":
		if fn, ok := newFunc.(func() *GenericResponse); ok {
			pool = NewConfigurableObjectPool[*GenericResponse](config, fn)
		}
	case "job":
		if fn, ok := newFunc.(func() *PooledJob); ok {
			pool = NewConfigurableObjectPool[*PooledJob](config, fn)
		}
	case "cache":
		if fn, ok := newFunc.(func() *PooledCacheEntry); ok {
			pool = NewConfigurableObjectPool[*PooledCacheEntry](config, fn)
		}
	case "db_row":
		if fn, ok := newFunc.(func() *PooledDBRow); ok {
			pool = NewConfigurableObjectPool[*PooledDBRow](config, fn)
		}
	case "request":
		if fn, ok := newFunc.(func() *PooledRequest); ok {
			pool = NewConfigurableObjectPool[*PooledRequest](config, fn)
		}
	case "validation":
		if fn, ok := newFunc.(func() *PooledValidationResult); ok {
			pool = NewConfigurableObjectPool[*PooledValidationResult](config, fn)
		}
	case "log_entry":
		if fn, ok := newFunc.(func() *PooledLogEntry); ok {
			pool = NewConfigurableObjectPool[*PooledLogEntry](config, fn)
		}
	case "mail_message":
		if fn, ok := newFunc.(func() *PooledMailMessage); ok {
			pool = NewConfigurableObjectPool[*PooledMailMessage](config, fn)
		}
	}

	if pool != nil {
		om.pools[name] = pool
		om.configs[name] = config
	}

	return pool
}

// GetPool retrieves a pool by name
func (om *ObjectPoolManager) GetPool(name string) (interface{}, bool) {
	om.mu.RLock()
	defer om.mu.RUnlock()

	pool, exists := om.pools[name]
	return pool, exists
}

// GetAllPools returns all registered pools
func (om *ObjectPoolManager) GetAllPools() map[string]interface{} {
	om.mu.RLock()
	defer om.mu.RUnlock()

	result := make(map[string]interface{})
	for name, pool := range om.pools {
		result[name] = pool
	}
	return result
}

// GetPoolMetrics returns metrics for all pools
func (om *ObjectPoolManager) GetPoolMetrics() map[string]PoolMetrics {
	om.mu.RLock()
	defer om.mu.RUnlock()

	metrics := make(map[string]PoolMetrics)
	for name, pool := range om.pools {
		// Use reflection to get metrics from any pool type
		if poolValue := reflect.ValueOf(pool); poolValue.IsValid() {
			if getMetricsMethod := poolValue.MethodByName("GetMetrics"); getMetricsMethod.IsValid() {
				if results := getMetricsMethod.Call(nil); len(results) > 0 {
					if poolMetrics, ok := results[0].Interface().(PoolMetrics); ok {
						metrics[name] = poolMetrics
					}
				}
			}
		}
	}
	return metrics
}

// Close closes all pools and cleanup routines
func (om *ObjectPoolManager) Close() {
	om.mu.Lock()
	defer om.mu.Unlock()

	for _, pool := range om.pools {
		// Use reflection to call Close on any pool type
		if poolValue := reflect.ValueOf(pool); poolValue.IsValid() {
			if closeMethod := poolValue.MethodByName("Close"); closeMethod.IsValid() {
				closeMethod.Call(nil)
			}
		}
	}

	om.cancel()
}

// CreatePoolFromConfig creates a pool from configuration
func (om *ObjectPoolManager) CreatePoolFromConfig(
	config ObjectPoolConfig,
	newFunc interface{},
) interface{} {
	return om.RegisterPool(config.Type, config, newFunc)
}

// Global object pool manager instance
var globalObjectPoolManager *ObjectPoolManager
var globalObjectPoolManagerOnce sync.Once

// GetGlobalObjectPoolManager returns the global object pool manager
func GetGlobalObjectPoolManager() *ObjectPoolManager {
	globalObjectPoolManagerOnce.Do(func() {
		globalObjectPoolManager = NewObjectPoolManager()
	})
	return globalObjectPoolManager
}

// InitializeObjectPools initializes object pools from configuration
// The config should contain merged default_pools and custom_pools
func InitializeObjectPools(config map[string]interface{}) error {
	manager := GetGlobalObjectPoolManager()

	// Initialize each pool from the merged configuration
	for poolName, poolConfig := range config {
		if err := initializePool(manager, poolName, poolConfig); err != nil {
			return fmt.Errorf("failed to initialize pool %s: %w", poolName, err)
		}
	}

	return nil
}

// initializePool initializes a single pool from configuration
func initializePool(manager *ObjectPoolManager, poolName string, poolConfig interface{}) error {
	configMap, ok := poolConfig.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid pool configuration for %s", poolName)
	}

	// Parse pool configuration
	config := ObjectPoolConfig{
		Name:            getPoolConfigString(configMap, "name", poolName),
		Description:     getPoolConfigString(configMap, "description", ""),
		Type:            getPoolConfigString(configMap, "type", poolName),
		PoolSize:        getPoolConfigInt(configMap, "pool_size", 100),
		MaxIdleTime:     time.Duration(getPoolConfigInt(configMap, "max_idle_time", 1800)) * time.Second,
		CleanupStrategy: getPoolConfigString(configMap, "cleanup_strategy", "lazy"),
		ResetMethod:     getPoolConfigString(configMap, "reset_method", "Reset"),
		Enabled:         getPoolConfigBool(configMap, "enabled", true),
	}

	// Parse metrics configuration
	if metricsConfig, ok := configMap["metrics"].(map[string]interface{}); ok {
		config.Metrics = PoolMetricsConfig{
			TrackHits:        getPoolConfigBool(metricsConfig, "track_hits", true),
			TrackMisses:      getPoolConfigBool(metricsConfig, "track_misses", true),
			TrackAllocations: getPoolConfigBool(metricsConfig, "track_allocations", true),
			TrackReuses:      getPoolConfigBool(metricsConfig, "track_reuses", true),
		}
	}

	// Create pool factory function based on type
	newFunc, err := createPoolFactory(config.Type)
	if err != nil {
		return fmt.Errorf("failed to create pool factory for %s: %w", poolName, err)
	}

	// Register the pool using the non-generic RegisterPool method
	manager.RegisterPool(poolName, config, newFunc)

	return nil
}

// Registry for custom pool factory functions
var customPoolFactories = make(map[string]func() PoolableObject)
var customPoolFactoriesMutex sync.RWMutex

// RegisterCustomPoolFactory allows developers to register custom pool types
// This enables config-driven extensibility without touching core code
func RegisterCustomPoolFactory(poolType string, factory func() PoolableObject) {
	customPoolFactoriesMutex.Lock()
	defer customPoolFactoriesMutex.Unlock()
	customPoolFactories[poolType] = factory
}

// GetRegisteredPoolTypes returns all registered pool types (core + custom)
func GetRegisteredPoolTypes() []string {
	customPoolFactoriesMutex.RLock()
	defer customPoolFactoriesMutex.RUnlock()
	
	coreTypes := []string{"user", "event", "response", "job", "cache", "db_row", "request", "validation", "log_entry", "mail_message"}
	customTypes := make([]string, 0, len(customPoolFactories))
	for poolType := range customPoolFactories {
		customTypes = append(customTypes, poolType)
	}
	
	return append(coreTypes, customTypes...)
}

// createPoolFactory creates a factory function for the given pool type
// Now supports both core types and custom registered types
func createPoolFactory(poolType string) (func() PoolableObject, error) {
	// Check core types first
	switch poolType {
	case "user":
		return func() PoolableObject { return &PooledUser{} }, nil
	case "event":
		return func() PoolableObject { return &PooledEvent{} }, nil
	case "response":
		return func() PoolableObject { return &GenericResponse{} }, nil
	case "job":
		return func() PoolableObject { return &PooledJob{} }, nil
	case "cache":
		return func() PoolableObject { return &PooledCacheEntry{} }, nil
	case "db_row":
		return func() PoolableObject { return &PooledDBRow{} }, nil
	case "request":
		return func() PoolableObject { return &PooledRequest{} }, nil
	case "validation":
		return func() PoolableObject { return &PooledValidationResult{} }, nil
	case "log_entry":
		return func() PoolableObject { return &PooledLogEntry{} }, nil
	case "mail_message":
		return func() PoolableObject { return &PooledMailMessage{} }, nil
	}
	
	// Check custom registered types
	customPoolFactoriesMutex.RLock()
	factory, exists := customPoolFactories[poolType]
	customPoolFactoriesMutex.RUnlock()
	
	if exists {
		return factory, nil
	}
	
	return nil, fmt.Errorf("unknown pool type: %s. Available types: %v", poolType, GetRegisteredPoolTypes())
}

// Helper functions for configuration parsing
func getPoolConfigString(config map[string]interface{}, key, defaultValue string) string {
	if value, ok := config[key].(string); ok {
		return value
	}
	return defaultValue
}

func getPoolConfigInt(config map[string]interface{}, key string, defaultValue int) int {
	if value, ok := config[key].(int); ok {
		return value
	}
	return defaultValue
}

func getPoolConfigBool(config map[string]interface{}, key string, defaultValue bool) bool {
	if value, ok := config[key].(bool); ok {
		return value
	}
	return defaultValue
}

// Poolable object types (these would be defined in your application)
type PooledUser struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"-"`
}

func (u *PooledUser) Reset() {
	u.ID = 0
	u.Name = ""
	u.Email = ""
	u.Password = ""
}

type PooledEvent struct {
	ID      string                 `json:"id"`
	Type    string                 `json:"type"`
	Data    map[string]interface{} `json:"data"`
	Created time.Time              `json:"created"`
}

func (e *PooledEvent) Reset() {
	e.ID = ""
	e.Type = ""
	e.Data = nil
	e.Created = time.Time{}
}

type PooledJob struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Payload  map[string]interface{} `json:"payload"`
	Attempts int                    `json:"attempts"`
}

func (j *PooledJob) Reset() {
	j.ID = ""
	j.Type = ""
	j.Payload = nil
	j.Attempts = 0
}

type PooledCacheEntry struct {
	Key      string      `json:"key"`
	Value    interface{} `json:"value"`
	Expires  time.Time   `json:"expires"`
	Accessed time.Time   `json:"accessed"`
}

func (c *PooledCacheEntry) Reset() {
	c.Key = ""
	c.Value = nil
	c.Expires = time.Time{}
	c.Accessed = time.Time{}
}

type PooledDBRow struct {
	Columns map[string]interface{} `json:"columns"`
	RowNum  int                    `json:"row_num"`
}

func (d *PooledDBRow) Reset() {
	d.Columns = nil
	d.RowNum = 0
}

type PooledRequest struct {
	Method  string                 `json:"method"`
	URL     string                 `json:"url"`
	Headers map[string]string      `json:"headers"`
	Body    map[string]interface{} `json:"body"`
}

func (r *PooledRequest) Reset() {
	r.Method = ""
	r.URL = ""
	r.Headers = nil
	r.Body = nil
}

type PooledValidationResult struct {
	Valid    bool              `json:"valid"`
	Errors   map[string]string `json:"errors"`
	Warnings map[string]string `json:"warnings"`
}

func (v *PooledValidationResult) Reset() {
	v.Valid = false
	v.Errors = nil
	v.Warnings = nil
}

type PooledLogEntry struct {
	Level   string                 `json:"level"`
	Message string                 `json:"message"`
	Context map[string]interface{} `json:"context"`
	Time    time.Time              `json:"time"`
}

func (l *PooledLogEntry) Reset() {
	l.Level = ""
	l.Message = ""
	l.Context = nil
	l.Time = time.Time{}
}

type PooledMailMessage struct {
	To      []string          `json:"to"`
	From    string            `json:"from"`
	Subject string            `json:"subject"`
	Body    string            `json:"body"`
	Headers map[string]string `json:"headers"`
}

func (m *PooledMailMessage) Reset() {
	m.To = nil
	m.From = ""
	m.Subject = ""
	m.Body = ""
	m.Headers = nil
}
