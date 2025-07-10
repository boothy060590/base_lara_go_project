package go_core

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ============================================================================
// ADVANCED INTERFACE COMPOSITION
// ============================================================================

// BaseInterface defines the most basic interface that all components implement
type BaseInterface interface {
	GetType() string
	GetID() string
	IsEnabled() bool
}

// ContextualInterface defines interfaces that can work with context
type ContextualInterface interface {
	WithContext(ctx context.Context) interface{}
	GetContext() context.Context
}

// PerformanceInterface defines interfaces that can track performance
type PerformanceInterface interface {
	GetPerformanceStats() map[string]interface{}
	GetOptimizationStats() map[string]interface{}
	Track(name string, fn func() error) error
}

// ConfigurableInterface defines interfaces that can be configured
type ConfigurableInterface interface {
	GetConfig() map[string]interface{}
	SetConfig(key string, value interface{}) error
	ValidateConfig() error
}

// LifecycleInterface defines interfaces with lifecycle management
type LifecycleInterface interface {
	Initialize() error
	Start() error
	Stop() error
	IsRunning() bool
}

// ============================================================================
// COMPOSITE INTERFACES
// ============================================================================

// RepositoryInterface combines repository functionality with performance and context
type RepositoryInterface[T any] interface {
	BaseInterface
	PerformanceInterface
	ConfigurableInterface

	// Repository-specific methods
	Find(id uint) (*T, error)
	Create(model *T) error
	Update(model *T) error
	Delete(id uint) error
	WithContext(ctx context.Context) RepositoryInterface[T]
}

// CacheInterface combines cache functionality with performance and context
type CacheInterface[T any] interface {
	BaseInterface
	PerformanceInterface
	ConfigurableInterface

	// Cache-specific methods
	Get(key string) (*T, error)
	Set(key string, value *T, ttl time.Duration) error
	Delete(key string) error
	WithContext(ctx context.Context) CacheInterface[T]
}

// EventInterface combines event functionality with performance and context
type EventInterface[T any] interface {
	BaseInterface
	PerformanceInterface
	ConfigurableInterface
	LifecycleInterface

	// Event-specific methods
	Dispatch(event *Event[T]) error
	Listen(eventName string, listener EventListener[T]) error
	WithContext(ctx context.Context) EventInterface[T]
}

// QueueInterface combines queue functionality with performance and context
type QueueInterface[T any] interface {
	BaseInterface
	PerformanceInterface
	ConfigurableInterface
	LifecycleInterface

	// Queue-specific methods
	Push(job *Job[T]) error
	Pop() (*Job[T], error)
	WithContext(ctx context.Context) QueueInterface[T]
}

// ============================================================================
// INTERFACE COMPOSITION UTILITIES
// ============================================================================

// InterfaceComposer provides utilities for composing interfaces
type InterfaceComposer struct {
	mu sync.RWMutex
}

// NewInterfaceComposer creates a new interface composer
func NewInterfaceComposer() *InterfaceComposer {
	return &InterfaceComposer{}
}

// ComposeRepository creates a composed repository interface
func ComposeRepository[T any](base Repository[T], config map[string]interface{}) RepositoryInterface[T] {
	return &ComposedRepository[T]{
		base:   base,
		config: config,
		ctx:    context.Background(),
	}
}

// ComposeCache creates a composed cache interface
func ComposeCache[T any](base Cache[T], config map[string]interface{}) CacheInterface[T] {
	return &ComposedCache[T]{
		base:   base,
		config: config,
		ctx:    context.Background(),
	}
}

// ComposeEvent creates a composed event interface
func ComposeEvent[T any](base EventDispatcher[T], config map[string]interface{}) EventInterface[T] {
	return &ComposedEvent[T]{
		base:   base,
		config: config,
		ctx:    context.Background(),
	}
}

// ComposeQueue creates a composed queue interface
func ComposeQueue[T any](base Queue[T], config map[string]interface{}) QueueInterface[T] {
	return &ComposedQueue[T]{
		base:   base,
		config: config,
		ctx:    context.Background(),
	}
}

// ============================================================================
// COMPOSED IMPLEMENTATIONS
// ============================================================================

// ComposedRepository implements RepositoryInterface with composition
type ComposedRepository[T any] struct {
	base   Repository[T]
	config map[string]interface{}
	ctx    context.Context
	mu     sync.RWMutex
}

// GetType returns the type of this component
func (cr *ComposedRepository[T]) GetType() string {
	return "repository"
}

// GetID returns the unique identifier
func (cr *ComposedRepository[T]) GetID() string {
	return fmt.Sprintf("repository_%T", *new(T))
}

// IsEnabled returns whether this component is enabled
func (cr *ComposedRepository[T]) IsEnabled() bool {
	if enabled, ok := cr.config["enabled"].(bool); ok {
		return enabled
	}
	return true
}

// GetContext returns the current context
func (cr *ComposedRepository[T]) GetContext() context.Context {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	return cr.ctx
}

// GetPerformanceStats returns performance statistics
func (cr *ComposedRepository[T]) GetPerformanceStats() map[string]interface{} {
	if perf, ok := cr.base.(PerformanceInterface); ok {
		return perf.GetPerformanceStats()
	}
	return make(map[string]interface{})
}

// GetOptimizationStats returns optimization statistics
func (cr *ComposedRepository[T]) GetOptimizationStats() map[string]interface{} {
	if perf, ok := cr.base.(PerformanceInterface); ok {
		return perf.GetOptimizationStats()
	}
	return make(map[string]interface{})
}

// Track tracks performance of an operation
func (cr *ComposedRepository[T]) Track(name string, fn func() error) error {
	if perf, ok := cr.base.(PerformanceInterface); ok {
		return perf.Track(name, fn)
	}
	return fn()
}

// GetConfig returns the configuration
func (cr *ComposedRepository[T]) GetConfig() map[string]interface{} {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	config := make(map[string]interface{})
	for k, v := range cr.config {
		config[k] = v
	}
	return config
}

// SetConfig sets a configuration value
func (cr *ComposedRepository[T]) SetConfig(key string, value interface{}) error {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	cr.config[key] = value
	return nil
}

// ValidateConfig validates the configuration
func (cr *ComposedRepository[T]) ValidateConfig() error {
	// Basic validation - can be extended
	if cr.config == nil {
		return fmt.Errorf("configuration is nil")
	}
	return nil
}

// Repository methods
func (cr *ComposedRepository[T]) Find(id uint) (*T, error) {
	var result *T
	err := cr.Track("repository.find", func() error {
		var findErr error
		result, findErr = cr.base.Find(id)
		return findErr
	})
	return result, err
}

func (cr *ComposedRepository[T]) Create(model *T) error {
	return cr.Track("repository.create", func() error {
		return cr.base.Create(model)
	})
}

func (cr *ComposedRepository[T]) Update(model *T) error {
	return cr.Track("repository.update", func() error {
		return cr.base.Update(model)
	})
}

func (cr *ComposedRepository[T]) Delete(id uint) error {
	return cr.Track("repository.delete", func() error {
		return cr.base.Delete(id)
	})
}

func (cr *ComposedRepository[T]) WithContext(ctx context.Context) RepositoryInterface[T] {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	cr.ctx = ctx
	return cr
}

// ComposedCache implements CacheInterface with composition
type ComposedCache[T any] struct {
	base   Cache[T]
	config map[string]interface{}
	ctx    context.Context
	mu     sync.RWMutex
}

// GetType returns the type of this component
func (cc *ComposedCache[T]) GetType() string {
	return "cache"
}

// GetID returns the unique identifier
func (cc *ComposedCache[T]) GetID() string {
	return fmt.Sprintf("cache_%T", *new(T))
}

// IsEnabled returns whether this component is enabled
func (cc *ComposedCache[T]) IsEnabled() bool {
	if enabled, ok := cc.config["enabled"].(bool); ok {
		return enabled
	}
	return true
}

// GetContext returns the current context
func (cc *ComposedCache[T]) GetContext() context.Context {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.ctx
}

// GetPerformanceStats returns performance statistics
func (cc *ComposedCache[T]) GetPerformanceStats() map[string]interface{} {
	if perf, ok := cc.base.(PerformanceInterface); ok {
		return perf.GetPerformanceStats()
	}
	return make(map[string]interface{})
}

// GetOptimizationStats returns optimization statistics
func (cc *ComposedCache[T]) GetOptimizationStats() map[string]interface{} {
	if perf, ok := cc.base.(PerformanceInterface); ok {
		return perf.GetOptimizationStats()
	}
	return make(map[string]interface{})
}

// Track tracks performance of an operation
func (cc *ComposedCache[T]) Track(name string, fn func() error) error {
	if perf, ok := cc.base.(PerformanceInterface); ok {
		return perf.Track(name, fn)
	}
	return fn()
}

// GetConfig returns the configuration
func (cc *ComposedCache[T]) GetConfig() map[string]interface{} {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	config := make(map[string]interface{})
	for k, v := range cc.config {
		config[k] = v
	}
	return config
}

// SetConfig sets a configuration value
func (cc *ComposedCache[T]) SetConfig(key string, value interface{}) error {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.config[key] = value
	return nil
}

// ValidateConfig validates the configuration
func (cc *ComposedCache[T]) ValidateConfig() error {
	if cc.config == nil {
		return fmt.Errorf("configuration is nil")
	}
	return nil
}

// Cache methods
func (cc *ComposedCache[T]) Get(key string) (*T, error) {
	var result *T
	err := cc.Track("cache.get", func() error {
		var getErr error
		result, getErr = cc.base.Get(key)
		return getErr
	})
	return result, err
}

func (cc *ComposedCache[T]) Set(key string, value *T, ttl time.Duration) error {
	return cc.Track("cache.set", func() error {
		return cc.base.Set(key, value, ttl)
	})
}

func (cc *ComposedCache[T]) Delete(key string) error {
	return cc.Track("cache.delete", func() error {
		return cc.base.Delete(key)
	})
}

func (cc *ComposedCache[T]) WithContext(ctx context.Context) CacheInterface[T] {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.ctx = ctx
	return cc
}

// ComposedEvent implements EventInterface with composition
type ComposedEvent[T any] struct {
	base    EventDispatcher[T]
	config  map[string]interface{}
	ctx     context.Context
	mu      sync.RWMutex
	running bool
}

// GetType returns the type of this component
func (ce *ComposedEvent[T]) GetType() string {
	return "event"
}

// GetID returns the unique identifier
func (ce *ComposedEvent[T]) GetID() string {
	return fmt.Sprintf("event_%T", *new(T))
}

// IsEnabled returns whether this component is enabled
func (ce *ComposedEvent[T]) IsEnabled() bool {
	if enabled, ok := ce.config["enabled"].(bool); ok {
		return enabled
	}
	return true
}

// GetContext returns the current context
func (ce *ComposedEvent[T]) GetContext() context.Context {
	ce.mu.RLock()
	defer ce.mu.RUnlock()
	return ce.ctx
}

// GetPerformanceStats returns performance statistics
func (ce *ComposedEvent[T]) GetPerformanceStats() map[string]interface{} {
	if perf, ok := ce.base.(PerformanceInterface); ok {
		return perf.GetPerformanceStats()
	}
	return make(map[string]interface{})
}

// GetOptimizationStats returns optimization statistics
func (ce *ComposedEvent[T]) GetOptimizationStats() map[string]interface{} {
	if perf, ok := ce.base.(PerformanceInterface); ok {
		return perf.GetOptimizationStats()
	}
	return make(map[string]interface{})
}

// Track tracks performance of an operation
func (ce *ComposedEvent[T]) Track(name string, fn func() error) error {
	if perf, ok := ce.base.(PerformanceInterface); ok {
		return perf.Track(name, fn)
	}
	return fn()
}

// GetConfig returns the configuration
func (ce *ComposedEvent[T]) GetConfig() map[string]interface{} {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	config := make(map[string]interface{})
	for k, v := range ce.config {
		config[k] = v
	}
	return config
}

// SetConfig sets a configuration value
func (ce *ComposedEvent[T]) SetConfig(key string, value interface{}) error {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	ce.config[key] = value
	return nil
}

// ValidateConfig validates the configuration
func (ce *ComposedEvent[T]) ValidateConfig() error {
	if ce.config == nil {
		return fmt.Errorf("configuration is nil")
	}
	return nil
}

// Initialize initializes the event system
func (ce *ComposedEvent[T]) Initialize() error {
	return ce.Track("event.initialize", func() error {
		return nil
	})
}

// Start starts the event system
func (ce *ComposedEvent[T]) Start() error {
	return ce.Track("event.start", func() error {
		ce.mu.Lock()
		defer ce.mu.Unlock()
		ce.running = true
		return nil
	})
}

// Stop stops the event system
func (ce *ComposedEvent[T]) Stop() error {
	return ce.Track("event.stop", func() error {
		ce.mu.Lock()
		defer ce.mu.Unlock()
		ce.running = false
		return nil
	})
}

// IsRunning returns whether the event system is running
func (ce *ComposedEvent[T]) IsRunning() bool {
	ce.mu.RLock()
	defer ce.mu.RUnlock()
	return ce.running
}

// Event methods
func (ce *ComposedEvent[T]) Dispatch(event *Event[T]) error {
	return ce.Track("event.dispatch", func() error {
		return ce.base.Dispatch(event)
	})
}

func (ce *ComposedEvent[T]) Listen(eventName string, listener EventListener[T]) error {
	return ce.Track("event.listen", func() error {
		return ce.base.Listen(eventName, listener)
	})
}

func (ce *ComposedEvent[T]) WithContext(ctx context.Context) EventInterface[T] {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	ce.ctx = ctx
	return ce
}

// ComposedQueue implements QueueInterface with composition
type ComposedQueue[T any] struct {
	base    Queue[T]
	config  map[string]interface{}
	ctx     context.Context
	mu      sync.RWMutex
	running bool
}

// GetType returns the type of this component
func (cq *ComposedQueue[T]) GetType() string {
	return "queue"
}

// GetID returns the unique identifier
func (cq *ComposedQueue[T]) GetID() string {
	return fmt.Sprintf("queue_%T", *new(T))
}

// IsEnabled returns whether this component is enabled
func (cq *ComposedQueue[T]) IsEnabled() bool {
	if enabled, ok := cq.config["enabled"].(bool); ok {
		return enabled
	}
	return true
}

// GetContext returns the current context
func (cq *ComposedQueue[T]) GetContext() context.Context {
	cq.mu.RLock()
	defer cq.mu.RUnlock()
	return cq.ctx
}

// GetPerformanceStats returns performance statistics
func (cq *ComposedQueue[T]) GetPerformanceStats() map[string]interface{} {
	if perf, ok := cq.base.(PerformanceInterface); ok {
		return perf.GetPerformanceStats()
	}
	return make(map[string]interface{})
}

// GetOptimizationStats returns optimization statistics
func (cq *ComposedQueue[T]) GetOptimizationStats() map[string]interface{} {
	if perf, ok := cq.base.(PerformanceInterface); ok {
		return perf.GetOptimizationStats()
	}
	return make(map[string]interface{})
}

// Track tracks performance of an operation
func (cq *ComposedQueue[T]) Track(name string, fn func() error) error {
	if perf, ok := cq.base.(PerformanceInterface); ok {
		return perf.Track(name, fn)
	}
	return fn()
}

// GetConfig returns the configuration
func (cq *ComposedQueue[T]) GetConfig() map[string]interface{} {
	cq.mu.RLock()
	defer cq.mu.RUnlock()

	config := make(map[string]interface{})
	for k, v := range cq.config {
		config[k] = v
	}
	return config
}

// SetConfig sets a configuration value
func (cq *ComposedQueue[T]) SetConfig(key string, value interface{}) error {
	cq.mu.Lock()
	defer cq.mu.Unlock()

	cq.config[key] = value
	return nil
}

// ValidateConfig validates the configuration
func (cq *ComposedQueue[T]) ValidateConfig() error {
	if cq.config == nil {
		return fmt.Errorf("configuration is nil")
	}
	return nil
}

// Initialize initializes the queue system
func (cq *ComposedQueue[T]) Initialize() error {
	return cq.Track("queue.initialize", func() error {
		return nil
	})
}

// Start starts the queue system
func (cq *ComposedQueue[T]) Start() error {
	return cq.Track("queue.start", func() error {
		cq.mu.Lock()
		defer cq.mu.Unlock()
		cq.running = true
		return nil
	})
}

// Stop stops the queue system
func (cq *ComposedQueue[T]) Stop() error {
	return cq.Track("queue.stop", func() error {
		cq.mu.Lock()
		defer cq.mu.Unlock()
		cq.running = false
		return nil
	})
}

// IsRunning returns whether the queue system is running
func (cq *ComposedQueue[T]) IsRunning() bool {
	cq.mu.RLock()
	defer cq.mu.RUnlock()
	return cq.running
}

// Queue methods
func (cq *ComposedQueue[T]) Push(job *Job[T]) error {
	return cq.Track("queue.push", func() error {
		return cq.base.Push(job)
	})
}

func (cq *ComposedQueue[T]) Pop() (*Job[T], error) {
	var result *Job[T]
	err := cq.Track("queue.pop", func() error {
		var popErr error
		result, popErr = cq.base.Pop()
		return popErr
	})
	return result, err
}

func (cq *ComposedQueue[T]) WithContext(ctx context.Context) QueueInterface[T] {
	cq.mu.Lock()
	defer cq.mu.Unlock()

	cq.ctx = ctx
	return cq
}

// ============================================================================
// GLOBAL INTERFACE COMPOSER
// ============================================================================

// Global interface composer instance
var GlobalInterfaceComposer = NewInterfaceComposer()
