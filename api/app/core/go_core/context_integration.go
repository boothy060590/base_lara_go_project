package go_core

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ============================================================================
// ADVANCED CONTEXT INTEGRATION
// ============================================================================

// ContextConfig defines configuration for context management
type ContextConfig struct {
	DefaultTimeout     time.Duration `json:"default_timeout"`
	MaxTimeout         time.Duration `json:"max_timeout"`
	EnableDeadline     bool          `json:"enable_deadline"`
	EnableCancellation bool          `json:"enable_cancellation"`
	PropagateValues    bool          `json:"propagate_values"`
}

// DefaultContextConfig returns sensible defaults for context management
func DefaultContextConfig() *ContextConfig {
	return &ContextConfig{
		DefaultTimeout:     30 * time.Second,
		MaxTimeout:         5 * time.Minute,
		EnableDeadline:     true,
		EnableCancellation: true,
		PropagateValues:    true,
	}
}

// NewContextConfigFromConfig creates a context config from Laravel-style config
func NewContextConfigFromConfig(configMap map[string]interface{}) *ContextConfig {
	if configMap == nil {
		return DefaultContextConfig()
	}

	config := &ContextConfig{
		DefaultTimeout:     30 * time.Second,
		MaxTimeout:         5 * time.Minute,
		EnableDeadline:     true,
		EnableCancellation: true,
		PropagateValues:    true,
	}

	// Load from config if available
	if defaults, ok := configMap["defaults"].(map[string]interface{}); ok {
		if timeout, ok := defaults["timeout"].(int); ok {
			config.DefaultTimeout = time.Duration(timeout) * time.Second
		}
		if maxTimeout, ok := defaults["max_timeout"].(int); ok {
			config.MaxTimeout = time.Duration(maxTimeout) * time.Second
		}
		if enableDeadline, ok := defaults["enable_deadline"].(bool); ok {
			config.EnableDeadline = enableDeadline
		}
		if enableCancellation, ok := defaults["enable_cancellation"].(bool); ok {
			config.EnableCancellation = enableCancellation
		}
		if propagateValues, ok := defaults["propagate_values"].(bool); ok {
			config.PropagateValues = propagateValues
		}
	}

	return config
}

// GetOperationTimeout gets the timeout for a specific operation from config
func GetOperationTimeout(configMap map[string]interface{}, operation string) time.Duration {
	if configMap == nil {
		return 30 * time.Second
	}

	if operations, ok := configMap["operations"].(map[string]interface{}); ok {
		if opConfig, ok := operations[operation].(map[string]interface{}); ok {
			if timeout, ok := opConfig["timeout"].(int); ok {
				return time.Duration(timeout) * time.Second
			}
		}
	}

	return 30 * time.Second
}

// GetProfileTimeout gets the timeout for a specific profile from config
func GetProfileTimeout(configMap map[string]interface{}, profile string) time.Duration {
	if configMap == nil {
		return 30 * time.Second
	}

	if profiles, ok := configMap["profiles"].(map[string]interface{}); ok {
		if profileConfig, ok := profiles[profile].(map[string]interface{}); ok {
			if timeout, ok := profileConfig["timeout"].(int); ok {
				return time.Duration(timeout) * time.Second
			}
		}
	}

	return 30 * time.Second
}

// ContextManager manages context lifecycle and provides automatic timeout/cancellation
type ContextManager struct {
	config *ContextConfig
	mu     sync.RWMutex
}

// NewContextManager creates a new context manager
func NewContextManager(config *ContextConfig) *ContextManager {
	if config == nil {
		config = DefaultContextConfig()
	}

	return &ContextManager{
		config: config,
	}
}

// WithTimeout creates a context with automatic timeout
func (cm *ContextManager) WithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	// Apply timeout limits
	if timeout > cm.config.MaxTimeout {
		timeout = cm.config.MaxTimeout
	}
	if timeout <= 0 {
		timeout = cm.config.DefaultTimeout
	}

	return context.WithTimeout(ctx, timeout)
}

// WithDeadline creates a context with automatic deadline
func (cm *ContextManager) WithDeadline(ctx context.Context, deadline time.Time) (context.Context, context.CancelFunc) {
	if !cm.config.EnableDeadline {
		return context.WithCancel(ctx)
	}

	return context.WithDeadline(ctx, deadline)
}

// WithValues creates a context with propagated values
func (cm *ContextManager) WithValues(ctx context.Context, values map[string]interface{}) context.Context {
	if !cm.config.PropagateValues {
		return ctx
	}

	for key, value := range values {
		ctx = context.WithValue(ctx, key, value)
	}

	return ctx
}

// ExecuteWithTimeout executes a function with automatic timeout
func (cm *ContextManager) ExecuteWithTimeout(ctx context.Context, timeout time.Duration, fn func(context.Context) error) error {
	ctx, cancel := cm.WithTimeout(ctx, timeout)
	defer cancel()

	// Create result channel
	resultChan := make(chan error, 1)

	go func() {
		resultChan <- fn(ctx)
	}()

	select {
	case err := <-resultChan:
		return err
	case <-ctx.Done():
		return fmt.Errorf("operation timed out after %v: %w", timeout, ctx.Err())
	}
}

// ExecuteWithDeadline executes a function with automatic deadline
func (cm *ContextManager) ExecuteWithDeadline(ctx context.Context, deadline time.Time, fn func(context.Context) error) error {
	ctx, cancel := cm.WithDeadline(ctx, deadline)
	defer cancel()

	// Create result channel
	resultChan := make(chan error, 1)

	go func() {
		resultChan <- fn(ctx)
	}()

	select {
	case err := <-resultChan:
		return err
	case <-ctx.Done():
		return fmt.Errorf("operation deadline exceeded: %w", ctx.Err())
	}
}

// ExecuteWithContext executes a function with context awareness (respects context cancellation)
func (cm *ContextManager) ExecuteWithContext(ctx context.Context, fn func(context.Context) error) error {
	// Create result channel
	resultChan := make(chan error, 1)

	go func() {
		resultChan <- fn(ctx)
	}()

	select {
	case err := <-resultChan:
		return err
	case <-ctx.Done():
		return fmt.Errorf("operation cancelled: %w", ctx.Err())
	}
}

// ============================================================================
// CONTEXT-AWARE INTERFACES
// ============================================================================

// ContextAware defines types that can work with context
type ContextAware interface {
	WithContext(ctx context.Context) interface{}
	GetContext() context.Context
}

// TimeoutAware defines types that can handle timeouts
type TimeoutAware interface {
	WithTimeout(timeout time.Duration) interface{}
	GetTimeout() time.Duration
}

// Cancellable defines types that can be cancelled
type Cancellable interface {
	Cancel() error
	IsCancelled() bool
}

// ContextAwareOperation represents an operation that can be executed with context
type ContextAwareOperation[T any] struct {
	operation func(context.Context) (T, error)
	timeout   time.Duration
	ctx       context.Context
	manager   *ContextManager
}

// NewContextAwareOperation creates a new context-aware operation
func NewContextAwareOperation[T any](operation func(context.Context) (T, error), manager *ContextManager) *ContextAwareOperation[T] {
	return &ContextAwareOperation[T]{
		operation: operation,
		manager:   manager,
		ctx:       context.Background(),
	}
}

// WithContext sets the context for the operation
func (cao *ContextAwareOperation[T]) WithContext(ctx context.Context) *ContextAwareOperation[T] {
	cao.ctx = ctx
	return cao
}

// WithTimeout sets the timeout for the operation
func (cao *ContextAwareOperation[T]) WithTimeout(timeout time.Duration) *ContextAwareOperation[T] {
	cao.timeout = timeout
	return cao
}

// Execute executes the operation with context and timeout
func (cao *ContextAwareOperation[T]) Execute() (T, error) {
	if cao.timeout > 0 {
		// Execute with timeout
		var result T
		err := cao.manager.ExecuteWithTimeout(cao.ctx, cao.timeout, func(ctx context.Context) error {
			var execErr error
			result, execErr = cao.operation(ctx)
			return execErr
		})
		return result, err
	}

	// Execute without timeout
	return cao.operation(cao.ctx)
}

// ============================================================================
// CONTEXT INTEGRATION DECORATORS
// ============================================================================

// WithContextDecorator decorates an operation with context awareness
func WithContextDecorator[T any](operation func(context.Context) (T, error)) func(context.Context) (T, error) {
	return func(ctx context.Context) (T, error) {
		// Add context values for tracking
		ctx = context.WithValue(ctx, "operation_start", time.Now())

		result, err := operation(ctx)

		// Add context values for completion (for potential future use)
		_ = context.WithValue(ctx, "operation_end", time.Now())

		return result, err
	}
}

// WithTimeoutDecorator decorates an operation with timeout
func WithTimeoutDecorator[T any](timeout time.Duration) func(func(context.Context) (T, error)) func(context.Context) (T, error) {
	return func(operation func(context.Context) (T, error)) func(context.Context) (T, error) {
		return func(ctx context.Context) (T, error) {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			resultChan := make(chan struct {
				result T
				err    error
			}, 1)

			go func() {
				result, err := operation(ctx)
				resultChan <- struct {
					result T
					err    error
				}{result, err}
			}()

			select {
			case result := <-resultChan:
				return result.result, result.err
			case <-ctx.Done():
				var zero T
				return zero, fmt.Errorf("operation timed out after %v", timeout)
			}
		}
	}
}

// WithRetryDecorator decorates an operation with retry logic
func WithRetryDecorator[T any](maxAttempts int, delay time.Duration) func(func(context.Context) (T, error)) func(context.Context) (T, error) {
	return func(operation func(context.Context) (T, error)) func(context.Context) (T, error) {
		return func(ctx context.Context) (T, error) {
			var lastErr error

			for attempt := 1; attempt <= maxAttempts; attempt++ {
				result, err := operation(ctx)
				if err == nil {
					return result, nil
				}

				lastErr = err
				if attempt < maxAttempts {
					select {
					case <-time.After(delay):
						// Continue to next attempt
					case <-ctx.Done():
						var zero T
						return zero, ctx.Err()
					}
				}
			}

			var zero T
			return zero, fmt.Errorf("operation failed after %d attempts: %w", maxAttempts, lastErr)
		}
	}
}

// ============================================================================
// CONTEXT UTILITIES
// ============================================================================

// ContextUtils provides utility functions for context management
type ContextUtils struct {
	manager *ContextManager
}

// NewContextUtils creates new context utilities
func NewContextUtils(manager *ContextManager) *ContextUtils {
	return &ContextUtils{
		manager: manager,
	}
}

// MergeContexts merges multiple contexts into one
func (cu *ContextUtils) MergeContexts(ctxs ...context.Context) context.Context {
	if len(ctxs) == 0 {
		return context.Background()
	}

	// Use the first context as base
	merged := ctxs[0]

	// Merge values from other contexts
	for _, ctx := range ctxs[1:] {
		merged = cu.mergeContextValues(merged, ctx)
	}

	return merged
}

// mergeContextValues merges values from two contexts
func (cu *ContextUtils) mergeContextValues(ctx1, ctx2 context.Context) context.Context {
	// This is a simplified implementation
	// In practice, you'd want to handle value conflicts more carefully
	return ctx1
}

// IsContextExpired checks if a context has expired
func (cu *ContextUtils) IsContextExpired(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

// GetContextTimeout returns the timeout from a context
func (cu *ContextUtils) GetContextTimeout(ctx context.Context) (time.Duration, bool) {
	if deadline, ok := ctx.Deadline(); ok {
		return time.Until(deadline), true
	}
	return 0, false
}

// ============================================================================
// GLOBAL CONTEXT MANAGER
// ============================================================================

// Global context manager instance
var GlobalContextManager = NewContextManager(DefaultContextConfig())

// WithGlobalTimeout creates a context with global timeout settings
func WithGlobalTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return GlobalContextManager.WithTimeout(ctx, timeout)
}

// ExecuteWithGlobalTimeout executes a function with global timeout settings
func ExecuteWithGlobalTimeout(ctx context.Context, timeout time.Duration, fn func(context.Context) error) error {
	return GlobalContextManager.ExecuteWithTimeout(ctx, timeout, fn)
}

// ============================================================================
// CONTEXT-AWARE SERVICE INTEGRATIONS
// ============================================================================

// ContextAwareEventDispatcher provides context-aware event dispatching
type ContextAwareEventDispatcher[T any] struct {
	dispatcher EventManagerInterface[T]
	manager    *ContextManager
}

// NewContextAwareEventDispatcher creates a new context-aware event dispatcher
func NewContextAwareEventDispatcher[T any](dispatcher EventManagerInterface[T]) *ContextAwareEventDispatcher[T] {
	return &ContextAwareEventDispatcher[T]{
		dispatcher: dispatcher,
		manager:    NewContextManager(DefaultContextConfig()),
	}
}

// Dispatch dispatches an event with context awareness
func (caed *ContextAwareEventDispatcher[T]) Dispatch(ctx context.Context, event *Event[T]) error {
	// Add context values for tracking
	ctx = context.WithValue(ctx, "event_dispatch_start", time.Now())
	ctx = context.WithValue(ctx, "event_name", event.Name)

	// Execute with automatic timeout (config-driven)
	timeout := GetOperationTimeout(nil, "events") // TODO: Pass config map
	return caed.manager.ExecuteWithTimeout(ctx, timeout, func(ctx context.Context) error {
		return caed.dispatcher.Dispatch(event)
	})
}

// Listen registers a listener with context awareness
func (caed *ContextAwareEventDispatcher[T]) Listen(eventName string, listener EventListener[T]) {
	caed.dispatcher.Listen(eventName, func(ctx context.Context, event *Event[T]) error {
		// Add context values for tracking
		ctx = context.WithValue(ctx, "listener_start", time.Now())
		ctx = context.WithValue(ctx, "event_name", eventName)

		// Execute the listener with context optimization
		err := listener(ctx, event)

		// Add context values for completion
		_ = context.WithValue(ctx, "listener_end", time.Now())

		return err
	})
}

// ContextAwareJobDispatcher provides context-aware job dispatching
type ContextAwareJobDispatcher[T any] struct {
	dispatcher JobDispatcher[T]
	manager    *ContextManager
}

// NewContextAwareJobDispatcher creates a new context-aware job dispatcher
func NewContextAwareJobDispatcher[T any](dispatcher JobDispatcher[T]) *ContextAwareJobDispatcher[T] {
	return &ContextAwareJobDispatcher[T]{
		dispatcher: dispatcher,
		manager:    NewContextManager(DefaultContextConfig()),
	}
}

// Dispatch dispatches a job with context awareness
func (cajd *ContextAwareJobDispatcher[T]) Dispatch(ctx context.Context, job T) error {
	// Add context values for tracking
	ctx = context.WithValue(ctx, "job_dispatch_start", time.Now())

	// Execute with automatic timeout (config-driven)
	timeout := GetOperationTimeout(nil, "jobs") // TODO: Pass config map
	return cajd.manager.ExecuteWithTimeout(ctx, timeout, func(ctx context.Context) error {
		return cajd.dispatcher.Dispatch(job)
	})
}

// DispatchSync dispatches a job synchronously with context awareness
func (cajd *ContextAwareJobDispatcher[T]) DispatchSync(ctx context.Context, job T) error {
	// Add context values for tracking
	ctx = context.WithValue(ctx, "job_dispatch_sync_start", time.Now())

	// Execute with automatic timeout (config-driven)
	timeout := GetOperationTimeout(nil, "jobs") // TODO: Pass config map
	return cajd.manager.ExecuteWithTimeout(ctx, timeout, func(ctx context.Context) error {
		return cajd.dispatcher.DispatchSync(job)
	})
}

// ContextAwareRepository provides context-aware repository operations
type ContextAwareRepository[T any] struct {
	repository Repository[T]
	manager    *ContextManager
}

// NewContextAwareRepository creates a new context-aware repository
func NewContextAwareRepository[T any](repository Repository[T], manager *ContextManager) *ContextAwareRepository[T] {
	if manager == nil {
		manager = NewContextManager(DefaultContextConfig())
	}

	return &ContextAwareRepository[T]{
		repository: repository,
		manager:    manager,
	}
}

// Find finds a model by ID with context awareness
func (car *ContextAwareRepository[T]) Find(ctx context.Context, id uint) (*T, error) {
	// Add context values for tracking
	ctx = context.WithValue(ctx, "repository_find_start", time.Now())
	ctx = context.WithValue(ctx, "repository_id", id)

	// Execute with automatic timeout (config-driven)
	timeout := GetOperationTimeout(nil, "repository") // TODO: Pass config map
	var result *T
	err := car.manager.ExecuteWithTimeout(ctx, timeout, func(ctx context.Context) error {
		var findErr error
		result, findErr = car.repository.Find(id)
		return findErr
	})

	// Add context values for completion
	_ = context.WithValue(ctx, "repository_find_end", time.Now())

	return result, err
}

// FindAll finds all models with context awareness
func (car *ContextAwareRepository[T]) FindAll(ctx context.Context) ([]T, error) {
	// Add context values for tracking
	ctx = context.WithValue(ctx, "repository_find_all_start", time.Now())

	// Execute with automatic timeout (config-driven)
	timeout := GetOperationTimeout(nil, "repository") // TODO: Pass config map
	var result []T
	err := car.manager.ExecuteWithTimeout(ctx, timeout, func(ctx context.Context) error {
		var findErr error
		result, findErr = car.repository.FindAll()
		return findErr
	})

	// Add context values for completion
	_ = context.WithValue(ctx, "repository_find_all_end", time.Now())

	return result, err
}

// Create creates a model with context awareness
func (car *ContextAwareRepository[T]) Create(ctx context.Context, model *T) error {
	// Add context values for tracking
	ctx = context.WithValue(ctx, "repository_create_start", time.Now())

	// Execute with automatic timeout (config-driven)
	timeout := GetOperationTimeout(nil, "repository") // TODO: Pass config map
	err := car.manager.ExecuteWithTimeout(ctx, timeout, func(ctx context.Context) error {
		return car.repository.Create(model)
	})

	// Add context values for completion
	_ = context.WithValue(ctx, "repository_create_end", time.Now())

	return err
}

// Update updates a model with context awareness
func (car *ContextAwareRepository[T]) Update(ctx context.Context, model *T) error {
	// Add context values for tracking
	ctx = context.WithValue(ctx, "repository_update_start", time.Now())

	// Execute with automatic timeout (config-driven)
	timeout := GetOperationTimeout(nil, "repository") // TODO: Pass config map
	err := car.manager.ExecuteWithTimeout(ctx, timeout, func(ctx context.Context) error {
		return car.repository.Update(model)
	})

	// Add context values for completion
	_ = context.WithValue(ctx, "repository_update_end", time.Now())

	return err
}

// Delete deletes a model with context awareness
func (car *ContextAwareRepository[T]) Delete(ctx context.Context, id uint) error {
	// Add context values for tracking
	ctx = context.WithValue(ctx, "repository_delete_start", time.Now())
	ctx = context.WithValue(ctx, "repository_delete_id", id)

	// Execute with automatic timeout (config-driven)
	timeout := GetOperationTimeout(nil, "repository") // TODO: Pass config map
	err := car.manager.ExecuteWithTimeout(ctx, timeout, func(ctx context.Context) error {
		return car.repository.Delete(id)
	})

	// Add context values for completion
	_ = context.WithValue(ctx, "repository_delete_end", time.Now())

	return err
}

// FindAsync finds a model by ID asynchronously with context awareness
func (car *ContextAwareRepository[T]) FindAsync(ctx context.Context, id uint) <-chan RepositoryResult[T] {
	resultChan := make(chan RepositoryResult[T], 1)

	go func() {
		defer close(resultChan)

		// Add context values for tracking
		ctx = context.WithValue(ctx, "repository_find_async_start", time.Now())
		ctx = context.WithValue(ctx, "repository_find_async_id", id)

		result, err := car.Find(ctx, id)

		// Add context values for completion
		_ = context.WithValue(ctx, "repository_find_async_end", time.Now())

		if result != nil {
			resultChan <- RepositoryResult[T]{
				Data:  *result,
				Error: err,
			}
		} else {
			var zero T
			resultChan <- RepositoryResult[T]{
				Data:  zero,
				Error: err,
			}
		}
	}()

	return resultChan
}

// FindManyAsync finds multiple models asynchronously with context awareness
func (car *ContextAwareRepository[T]) FindManyAsync(ctx context.Context, ids []uint) <-chan RepositoryResult[[]T] {
	resultChan := make(chan RepositoryResult[[]T], 1)

	go func() {
		defer close(resultChan)

		// Add context values for tracking
		ctx = context.WithValue(ctx, "repository_find_many_async_start", time.Now())
		ctx = context.WithValue(ctx, "repository_find_many_async_ids", ids)

		var results []*T
		var err error

		// Execute with automatic timeout
		err = car.manager.ExecuteWithTimeout(ctx, 30*time.Second, func(ctx context.Context) error {
			for _, id := range ids {
				result, findErr := car.repository.Find(id)
				if findErr != nil {
					return findErr
				}
				results = append(results, result)
			}
			return nil
		})

		// Add context values for completion
		_ = context.WithValue(ctx, "repository_find_many_async_end", time.Now())

		// Convert []*T to []T
		var resultSlice []T
		for _, result := range results {
			if result != nil {
				resultSlice = append(resultSlice, *result)
			}
		}

		resultChan <- RepositoryResult[[]T]{
			Data:  resultSlice,
			Error: err,
		}
	}()

	return resultChan
}
