package logging_core

import (
	"context"
	"fmt"
	"time"

	go_core "base_lara_go_project/app/core/go_core"
	config_core "base_lara_go_project/app/core/laravel_core/config"
)

// GenericLoggingFacade provides Laravel-style logging with go_core optimizations
type GenericLoggingFacade[T any] struct {
	logger *go_core.Logger[T]
	config *config_core.ConfigFacade
}

// NewGenericLoggingFacade creates a new generic logging facade
func NewGenericLoggingFacade[T any](config *config_core.ConfigFacade) *GenericLoggingFacade[T] {
	// Convert Laravel config to go_core config
	goCoreConfig := convertToGoCoreConfig(config)

	return &GenericLoggingFacade[T]{
		logger: go_core.NewLogger[T](goCoreConfig),
		config: config,
	}
}

// convertToGoCoreConfig converts Laravel config to go_core config
func convertToGoCoreConfig(config *config_core.ConfigFacade) *go_core.LoggerConfig {
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

	return &go_core.LoggerConfig{
		DefaultLevel:    defaultLevel,
		MaxConcurrency:  maxConcurrency,
		BufferSize:      bufferSize,
		FlushInterval:   flushInterval,
		ObjectPoolSize:  objectPoolSize,
		ContextTimeout:  contextTimeout,
		PerformanceMode: performanceMode,
	}
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

// Log logs a message with generic context
func (f *GenericLoggingFacade[T]) Log(level go_core.LogLevel, message string, context T) error {
	return f.logger.Log(level, message, context)
}

// Debug logs a debug message
func (f *GenericLoggingFacade[T]) Debug(message string, context T) error {
	return f.logger.Debug(message, context)
}

// Info logs an info message
func (f *GenericLoggingFacade[T]) Info(message string, context T) error {
	return f.logger.Info(message, context)
}

// Warning logs a warning message
func (f *GenericLoggingFacade[T]) Warning(message string, context T) error {
	return f.logger.Warning(message, context)
}

// Error logs an error message
func (f *GenericLoggingFacade[T]) Error(message string, context T) error {
	return f.logger.Error(message, context)
}

// Fatal logs a fatal message
func (f *GenericLoggingFacade[T]) Fatal(message string, context T) error {
	return f.logger.Fatal(message, context)
}

// Emergency logs an emergency message (same as Fatal)
func (f *GenericLoggingFacade[T]) Emergency(message string, context T) error {
	return f.logger.Fatal(message, context)
}

// Alert logs an alert message (same as Error)
func (f *GenericLoggingFacade[T]) Alert(message string, context T) error {
	return f.logger.Error(message, context)
}

// Critical logs a critical message (same as Error)
func (f *GenericLoggingFacade[T]) Critical(message string, context T) error {
	return f.logger.Error(message, context)
}

// Notice logs a notice message (same as Info)
func (f *GenericLoggingFacade[T]) Notice(message string, context T) error {
	return f.logger.Info(message, context)
}

// AddHandler adds a log handler
func (f *GenericLoggingFacade[T]) AddHandler(name string, handler go_core.LogHandler[T]) {
	f.logger.AddHandler(name, handler)
}

// RemoveHandler removes a log handler
func (f *GenericLoggingFacade[T]) RemoveHandler(name string) {
	f.logger.RemoveHandler(name)
}

// GetHandler returns a handler by name
func (f *GenericLoggingFacade[T]) GetHandler(name string) (go_core.LogHandler[T], bool) {
	return f.logger.GetHandler(name)
}

// GetMetrics returns logging metrics
func (f *GenericLoggingFacade[T]) GetMetrics() *go_core.LoggingMetrics {
	return f.logger.GetMetrics()
}

// Close closes the logger
func (f *GenericLoggingFacade[T]) Close() error {
	return f.logger.Close()
}

// WithContext returns a logger with context (Laravel-style)
func (f *GenericLoggingFacade[T]) WithContext(ctx context.Context) *GenericLoggingFacade[T] {
	// Create a new facade with context-aware logging
	newFacade := &GenericLoggingFacade[T]{
		logger: f.logger,
		config: f.config,
	}

	// Add context to all future log entries
	// This would require extending the go_core logger to support context
	return newFacade
}

// WithFields returns a logger with fields (Laravel-style)
func (f *GenericLoggingFacade[T]) WithFields(fields map[string]interface{}) *GenericLoggingFacade[T] {
	// Create a new facade with field-aware logging
	newFacade := &GenericLoggingFacade[T]{
		logger: f.logger,
		config: f.config,
	}

	// Add fields to all future log entries
	// This would require extending the go_core logger to support fields
	return newFacade
}

// Channel returns a logger for a specific channel
func (f *GenericLoggingFacade[T]) Channel(channelName string) (*GenericLoggingFacade[T], error) {
	// Get channel configuration
	channels := f.config.Get("logging.channels").(map[string]interface{})
	_, exists := channels[channelName]
	if !exists {
		return nil, fmt.Errorf("channel '%s' not found", channelName)
	}

	// For now, return the same facade since config is read-only
	// In a real implementation, you'd create a new config instance
	return f, nil
}

// Stack returns a logger that writes to multiple channels
func (f *GenericLoggingFacade[T]) Stack(channels ...string) (*GenericLoggingFacade[T], error) {
	// For now, return the same facade since config is read-only
	// In a real implementation, you'd create a new config instance
	return f, nil
}

// Report logs an exception to specified channels (Laravel-style)
func (f *GenericLoggingFacade[T]) Report(exception error, channels ...string) error {
	// Create a simple context from the error
	var context T
	// This is a limitation - we need a proper way to convert error to generic context
	// For now, we'll just log the error message
	return f.Error("Exception occurred", context)
}

// ReportWithLevel logs an exception to specified channels with custom level
func (f *GenericLoggingFacade[T]) ReportWithLevel(exception error, level go_core.LogLevel, channels ...string) error {
	// Create a simple context from the error
	var context T
	// This is a limitation - we need a proper way to convert error to generic context
	// For now, we'll just log the error message
	return f.Log(level, "Exception occurred", context)
}
