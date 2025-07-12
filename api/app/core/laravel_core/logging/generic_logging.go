package logging_core

import (
	"sync"
	"time"

	go_core "base_lara_go_project/app/core/go_core"
	"base_lara_go_project/config"
)

// GenericLogFactory creates generic loggers with config-driven settings
type GenericLogFactory[T any] struct {
	handlers map[string]go_core.LogHandler[T]
	metrics  *go_core.LoggingMetrics
	mu       sync.RWMutex
}

// NewGenericLogFactory creates a new generic log factory
func NewGenericLogFactory[T any]() *GenericLogFactory[T] {
	return &GenericLogFactory[T]{
		handlers: make(map[string]go_core.LogHandler[T]),
		metrics:  &go_core.LoggingMetrics{},
	}
}

// CreateLogger creates a new logger with the specified configuration
func (f *GenericLogFactory[T]) CreateLogger(name string) (*go_core.Logger[T], error) {
	// Get logging configuration
	loggingConfig := config.Get("logging").(map[string]interface{})

	// Get default level
	defaultLevel := go_core.LogLevelInfo
	if levelStr, ok := loggingConfig["level"].(string); ok {
		defaultLevel = parseLogLevel(levelStr)
	}

	// Get performance settings
	maxConcurrency := 100
	if max, ok := loggingConfig["max_concurrency"].(int); ok {
		maxConcurrency = max
	}

	bufferSize := 1000
	if buffer, ok := loggingConfig["buffer_size"].(int); ok {
		bufferSize = buffer
	}

	flushInterval := time.Second
	if interval, ok := loggingConfig["flush_interval"].(int); ok {
		flushInterval = time.Duration(interval) * time.Second
	}

	objectPoolSize := 100
	if poolSize, ok := loggingConfig["object_pool_size"].(int); ok {
		objectPoolSize = poolSize
	}

	contextTimeout := 30 * time.Second
	if timeout, ok := loggingConfig["context_timeout"].(int); ok {
		contextTimeout = time.Duration(timeout) * time.Second
	}

	performanceMode := true
	if mode, ok := loggingConfig["performance_mode"].(bool); ok {
		performanceMode = mode
	}

	return go_core.NewLogger[T](&go_core.LoggerConfig{
		DefaultLevel:    defaultLevel,
		MaxConcurrency:  maxConcurrency,
		BufferSize:      bufferSize,
		FlushInterval:   flushInterval,
		ObjectPoolSize:  objectPoolSize,
		ContextTimeout:  contextTimeout,
		PerformanceMode: performanceMode,
	}), nil
}

// parseLogLevel converts string level to go_core LogLevel
func parseLogLevel(level string) go_core.LogLevel {
	switch level {
	case "debug":
		return go_core.LogLevelDebug
	case "info":
		return go_core.LogLevelInfo
	case "warning":
		return go_core.LogLevelWarning
	case "error":
		return go_core.LogLevelError
	case "fatal":
		return go_core.LogLevelFatal
	default:
		return go_core.LogLevelInfo
	}
}

// AddHandler adds a log handler to the factory
func (f *GenericLogFactory[T]) AddHandler(name string, handler go_core.LogHandler[T]) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.handlers[name] = handler
}

// RemoveHandler removes a log handler from the factory
func (f *GenericLogFactory[T]) RemoveHandler(name string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.handlers, name)
}

// GetHandler returns a handler by name
func (f *GenericLogFactory[T]) GetHandler(name string) (go_core.LogHandler[T], bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	handler, exists := f.handlers[name]
	return handler, exists
}

// GetMetrics returns logging metrics
func (f *GenericLogFactory[T]) GetMetrics() *go_core.LoggingMetrics {
	return f.metrics
}
