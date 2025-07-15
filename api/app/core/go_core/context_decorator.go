package go_core

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ContextDecorator provides cross-cutting context management with performance analytics
type ContextDecorator struct {
	// Performance tracking
	performanceTracker *PerformanceTracker
	contextPool        *ContextPool

	// Configuration
	config map[string]any

	// Thread safety
	mu sync.RWMutex
}

// NewContextDecorator creates a new context decorator with performance tracking
func NewContextDecorator(config map[string]any) *ContextDecorator {
	return &ContextDecorator{
		performanceTracker: NewPerformanceTracker(),
		contextPool:        NewContextPool(100), // Configurable pool size
		config:             config,
	}
}

// WithContext creates a context-aware wrapper for any operation
func (cd *ContextDecorator) WithContext(ctx context.Context, operation string, fn func(context.Context) error) error {
	// Start performance tracking
	start := time.Now()
	traceID := cd.generateTraceID()

	// Enhance context with tracing information
	enhancedCtx := context.WithValue(ctx, "trace_id", traceID)
	enhancedCtx = context.WithValue(enhancedCtx, "operation", operation)
	enhancedCtx = context.WithValue(enhancedCtx, "start_time", start)

	// Apply context optimizations based on config
	enhancedCtx = cd.applyContextOptimizations(enhancedCtx)

	// Execute the operation
	err := fn(enhancedCtx)

	// Track performance
	duration := time.Since(start)
	cd.performanceTracker.Track(operation, duration, err)

	// Log performance metrics if enabled
	if cd.isLoggingEnabled() {
		cd.logPerformanceMetrics(operation, duration, err, traceID)
	}

	return err
}

// WithTimeout creates a context with automatic timeout management
func (cd *ContextDecorator) WithTimeout(ctx context.Context, operation string, timeout time.Duration, fn func(context.Context) error) error {
	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute with context decorator
	return cd.WithContext(timeoutCtx, operation, fn)
}

// WithDeadline creates a context with automatic deadline management
func (cd *ContextDecorator) WithDeadline(ctx context.Context, operation string, deadline time.Time, fn func(context.Context) error) error {
	// Create deadline context
	deadlineCtx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	// Execute with context decorator
	return cd.WithContext(deadlineCtx, operation, fn)
}

// WithCancellation creates a context with cancellation support
func (cd *ContextDecorator) WithCancellation(ctx context.Context, operation string, fn func(context.Context) error) error {
	// Create cancellable context
	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Execute with context decorator
	return cd.WithContext(cancelCtx, operation, fn)
}

// WithValues creates a context with additional values for tracing
func (cd *ContextDecorator) WithValues(ctx context.Context, operation string, values map[string]any, fn func(context.Context) error) error {
	// Add values to context
	enhancedCtx := ctx
	for key, value := range values {
		enhancedCtx = context.WithValue(enhancedCtx, key, value)
	}

	// Execute with context decorator
	return cd.WithContext(enhancedCtx, operation, fn)
}

// GetPerformanceStats returns performance statistics
func (cd *ContextDecorator) GetPerformanceStats() map[string]any {
	cd.mu.RLock()
	defer cd.mu.RUnlock()

	stats := cd.performanceTracker.GetStats()
	stats["context_pool"] = cd.contextPool.GetStats()
	stats["config"] = cd.config

	return stats
}

// GetOptimizationStats returns optimization statistics
func (cd *ContextDecorator) GetOptimizationStats() map[string]any {
	cd.mu.RLock()
	defer cd.mu.RUnlock()

	return map[string]any{
		"context_decorator_enabled": true,
		"performance_tracking":      true,
		"context_pooling":           true,
		"automatic_timeout":         true,
		"tracing_enabled":           true,
		"metrics_enabled":           cd.isMetricsEnabled(),
		"logging_enabled":           cd.isLoggingEnabled(),
	}
}

// applyContextOptimizations applies context optimizations based on configuration
func (cd *ContextDecorator) applyContextOptimizations(ctx context.Context) context.Context {
	cd.mu.RLock()
	defer cd.mu.RUnlock()

	// Apply automatic timeout if enabled and not already set
	if cd.isAutoTimeoutEnabled() && ctx.Err() == nil {
		if _, ok := ctx.Deadline(); !ok {
			timeout := cd.getDefaultTimeout()
			timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
			// Store cancel function in context for later cleanup
			timeoutCtx = context.WithValue(timeoutCtx, "cancel_func", cancel)
			return timeoutCtx
		}
	}

	return ctx
}

// generateTraceID generates a unique trace ID for performance tracking
func (cd *ContextDecorator) generateTraceID() string {
	return fmt.Sprintf("trace_%d_%d", time.Now().UnixNano(), cd.performanceTracker.GetOperationCount())
}

// logPerformanceMetrics logs performance metrics if logging is enabled
func (cd *ContextDecorator) logPerformanceMetrics(operation string, duration time.Duration, err error, traceID string) {
	if !cd.isLoggingEnabled() {
		return
	}

	// This would integrate with your logging system
	// For now, we'll just track it in the performance tracker
	cd.performanceTracker.LogMetric(operation, duration, err, traceID)
}

// Configuration helper methods
func (cd *ContextDecorator) isAutoTimeoutEnabled() bool {
	if enabled, ok := cd.config["auto_timeout"].(bool); ok {
		return enabled
	}
	return true // Default to enabled
}

func (cd *ContextDecorator) getDefaultTimeout() time.Duration {
	if timeout, ok := cd.config["default_timeout"].(int); ok {
		return time.Duration(timeout) * time.Second
	}
	return 30 * time.Second // Default timeout
}

func (cd *ContextDecorator) isMetricsEnabled() bool {
	if enabled, ok := cd.config["metrics_enabled"].(bool); ok {
		return enabled
	}
	return true // Default to enabled
}

func (cd *ContextDecorator) isLoggingEnabled() bool {
	if enabled, ok := cd.config["logging_enabled"].(bool); ok {
		return enabled
	}
	return true // Default to enabled
}

// PerformanceTracker tracks performance metrics
type PerformanceTracker struct {
	operations map[string]*OperationStats
	mu         sync.RWMutex
	opCount    int64
}

// NewPerformanceTracker creates a new performance tracker
func NewPerformanceTracker() *PerformanceTracker {
	return &PerformanceTracker{
		operations: make(map[string]*OperationStats),
	}
}

// Track tracks performance of an operation
func (pt *PerformanceTracker) Track(operation string, duration time.Duration, err error) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	if pt.operations[operation] == nil {
		pt.operations[operation] = &OperationStats{
			Count:      0,
			TotalTime:  0,
			MinTime:    duration,
			MaxTime:    duration,
			ErrorCount: 0,
			LastError:  nil,
		}
	}

	stats := pt.operations[operation]
	stats.Count++
	stats.TotalTime += duration

	if duration < stats.MinTime {
		stats.MinTime = duration
	}
	if duration > stats.MaxTime {
		stats.MaxTime = duration
	}

	if err != nil {
		stats.ErrorCount++
		stats.LastError = err
	}

	pt.opCount++
}

// GetStats returns performance statistics
func (pt *PerformanceTracker) GetStats() map[string]any {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	stats := make(map[string]any)
	for operation, opStats := range pt.operations {
		stats[operation] = map[string]any{
			"count":       opStats.Count,
			"total_time":  opStats.TotalTime.String(),
			"avg_time":    (opStats.TotalTime / time.Duration(opStats.Count)).String(),
			"min_time":    opStats.MinTime.String(),
			"max_time":    opStats.MaxTime.String(),
			"error_count": opStats.ErrorCount,
			"error_rate":  float64(opStats.ErrorCount) / float64(opStats.Count),
			"last_error":  opStats.LastError,
		}
	}

	stats["total_operations"] = pt.opCount
	return stats
}

// GetOperationCount returns the total number of operations tracked
func (pt *PerformanceTracker) GetOperationCount() int64 {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	return pt.opCount
}

// LogMetric logs a performance metric
func (pt *PerformanceTracker) LogMetric(operation string, duration time.Duration, err error, traceID string) {
	// This would integrate with your logging system
	// For now, we'll just track it in the operations map
	pt.Track(operation, duration, err)
}

// OperationStats holds statistics for a single operation
type OperationStats struct {
	Count      int64
	TotalTime  time.Duration
	MinTime    time.Duration
	MaxTime    time.Duration
	ErrorCount int64
	LastError  error
}

// ContextPool provides context pooling for performance optimization
type ContextPool struct {
	pool    chan context.Context
	size    int
	created int64
	reused  int64
	mu      sync.RWMutex
}

// NewContextPool creates a new context pool
func NewContextPool(size int) *ContextPool {
	return &ContextPool{
		pool: make(chan context.Context, size),
		size: size,
	}
}

// Get gets a context from the pool or creates a new one
func (cp *ContextPool) Get() context.Context {
	select {
	case ctx := <-cp.pool:
		cp.mu.Lock()
		cp.reused++
		cp.mu.Unlock()
		return ctx
	default:
		cp.mu.Lock()
		cp.created++
		cp.mu.Unlock()
		return context.Background()
	}
}

// Put puts a context back into the pool
func (cp *ContextPool) Put(ctx context.Context) {
	select {
	case cp.pool <- ctx:
		// Successfully returned to pool
	default:
		// Pool is full, discard
	}
}

// GetStats returns pool statistics
func (cp *ContextPool) GetStats() map[string]any {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	return map[string]any{
		"pool_size":  cp.size,
		"pool_usage": len(cp.pool),
		"created":    cp.created,
		"reused":     cp.reused,
		"reuse_rate": float64(cp.reused) / float64(cp.created+cp.reused),
	}
}
