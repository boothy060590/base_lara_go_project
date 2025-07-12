package facades_core

import (
	"context"

	go_core "base_lara_go_project/app/core/go_core"
	logging_core "base_lara_go_project/app/core/laravel_core/logging"
)

// convertToGoCoreLevel converts logging_core.LogLevel to go_core.LogLevel
func convertToGoCoreLevel(level logging_core.LogLevel) go_core.LogLevel {
	switch level {
	case logging_core.LogLevelDebug:
		return go_core.LogLevelDebug
	case logging_core.LogLevelInfo:
		return go_core.LogLevelInfo
	case logging_core.LogLevelWarning:
		return go_core.LogLevelWarning
	case logging_core.LogLevelError:
		return go_core.LogLevelError
	case logging_core.LogLevelFatal:
		return go_core.LogLevelFatal
	default:
		return go_core.LogLevelInfo
	}
}

// LoggingWrapper wraps GenericLogFactory to implement LoggerInterface
type LoggingWrapper struct {
	factory *logging_core.GenericLogFactory[map[string]interface{}]
	logger  *go_core.Logger[map[string]interface{}]
}

// NewLoggingWrapper creates a new logging wrapper
func NewLoggingWrapper() *LoggingWrapper {
	factory := logging_core.NewGenericLogFactory[map[string]interface{}]()
	logger, err := factory.CreateLogger("default")
	if err != nil {
		// Fallback to a simple logger if creation fails
		logger = go_core.NewLogger[map[string]interface{}](&go_core.LoggerConfig{})
	}

	return &LoggingWrapper{
		factory: factory,
		logger:  logger,
	}
}

// Log implements LoggerInterface.Log
func (w *LoggingWrapper) Log(level logging_core.LogLevel, message string, context map[string]interface{}) error {
	goLevel := convertToGoCoreLevel(level)
	return w.logger.Log(goLevel, message, context)
}

// Debug implements LoggerInterface.Debug
func (w *LoggingWrapper) Debug(message string, context map[string]interface{}) error {
	return w.logger.Debug(message, context)
}

// Info implements LoggerInterface.Info
func (w *LoggingWrapper) Info(message string, context map[string]interface{}) error {
	return w.logger.Info(message, context)
}

// Warning implements LoggerInterface.Warning
func (w *LoggingWrapper) Warning(message string, context map[string]interface{}) error {
	return w.logger.Warning(message, context)
}

// Error implements LoggerInterface.Error
func (w *LoggingWrapper) Error(message string, context map[string]interface{}) error {
	return w.logger.Error(message, context)
}

// Fatal implements LoggerInterface.Fatal
func (w *LoggingWrapper) Fatal(message string, context map[string]interface{}) error {
	return w.logger.Fatal(message, context)
}

// WithContext implements LoggerInterface.WithContext
func (w *LoggingWrapper) WithContext(ctx context.Context) logging_core.LoggerInterface {
	// Create a new wrapper with the same factory
	return &LoggingWrapper{
		factory: w.factory,
		logger:  w.logger,
	}
}

// WithFields implements LoggerInterface.WithFields
func (w *LoggingWrapper) WithFields(fields map[string]interface{}) logging_core.LoggerInterface {
	// Create a new wrapper with the same factory
	return &LoggingWrapper{
		factory: w.factory,
		logger:  w.logger,
	}
}

// Logging provides a facade for logging operations with go_core optimizations
type Logging struct{}

// Global logging facade instance
var LoggingInstance = &Logging{}

// Debug logs a debug message
func (l *Logging) Debug(message string, context map[string]interface{}) error {
	wrapper := NewLoggingWrapper()
	return wrapper.Debug(message, context)
}

// Info logs an info message
func (l *Logging) Info(message string, context map[string]interface{}) error {
	wrapper := NewLoggingWrapper()
	return wrapper.Info(message, context)
}

// Warning logs a warning message
func (l *Logging) Warning(message string, context map[string]interface{}) error {
	wrapper := NewLoggingWrapper()
	return wrapper.Warning(message, context)
}

// Error logs an error message
func (l *Logging) Error(message string, context map[string]interface{}) error {
	wrapper := NewLoggingWrapper()
	return wrapper.Error(message, context)
}

// Fatal logs a fatal message
func (l *Logging) Fatal(message string, context map[string]interface{}) error {
	wrapper := NewLoggingWrapper()
	return wrapper.Fatal(message, context)
}

// Emergency logs an emergency message
func (l *Logging) Emergency(message string, context map[string]interface{}) error {
	wrapper := NewLoggingWrapper()
	return wrapper.Fatal(message, context) // Emergency is same as Fatal
}

// Alert logs an alert message
func (l *Logging) Alert(message string, context map[string]interface{}) error {
	wrapper := NewLoggingWrapper()
	return wrapper.Error(message, context) // Alert is same as Error
}

// Critical logs a critical message
func (l *Logging) Critical(message string, context map[string]interface{}) error {
	wrapper := NewLoggingWrapper()
	return wrapper.Error(message, context) // Critical is same as Error
}

// Notice logs a notice message
func (l *Logging) Notice(message string, context map[string]interface{}) error {
	wrapper := NewLoggingWrapper()
	return wrapper.Info(message, context) // Notice is same as Info
}

// Log logs a message with a specific level
func (l *Logging) Log(level logging_core.LogLevel, message string, context map[string]interface{}) error {
	wrapper := NewLoggingWrapper()
	return wrapper.Log(level, message, context)
}

// WithContext returns a logger with context
func (l *Logging) WithContext(ctx context.Context) logging_core.LoggerInterface {
	wrapper := NewLoggingWrapper()
	return wrapper.WithContext(ctx)
}

// WithFields returns a logger with fields
func (l *Logging) WithFields(fields map[string]interface{}) logging_core.LoggerInterface {
	wrapper := NewLoggingWrapper()
	return wrapper.WithFields(fields)
}

// Channel returns a logger for a specific channel
func (l *Logging) Channel(channelName string) (logging_core.LoggerInterface, error) {
	wrapper := NewLoggingWrapper()
	// For now, return the wrapper since Channel functionality isn't implemented
	return wrapper, nil
}

// Stack returns a logger that writes to multiple channels
func (l *Logging) Stack(channels ...string) (logging_core.LoggerInterface, error) {
	wrapper := NewLoggingWrapper()
	// For now, return the wrapper since Stack functionality isn't implemented
	return wrapper, nil
}

// Report logs an exception to specified channels (Laravel-style)
func (l *Logging) Report(exception error, channels ...string) error {
	wrapper := NewLoggingWrapper()
	// Convert exception to error message
	return wrapper.Error(exception.Error(), map[string]interface{}{
		"exception": exception,
		"channels":  channels,
	})
}

// ReportWithLevel logs an exception to specified channels with custom level
func (l *Logging) ReportWithLevel(exception error, level logging_core.LogLevel, channels ...string) error {
	wrapper := NewLoggingWrapper()
	// Convert exception to error message with custom level
	return wrapper.Log(level, exception.Error(), map[string]interface{}{
		"exception": exception,
		"channels":  channels,
	})
}

// Global logging functions for convenience

// Debug logs a debug message
func Debug(message string, context map[string]interface{}) error {
	return LoggingInstance.Debug(message, context)
}

// Info logs an info message
func Info(message string, context map[string]interface{}) error {
	return LoggingInstance.Info(message, context)
}

// Warning logs a warning message
func Warning(message string, context map[string]interface{}) error {
	return LoggingInstance.Warning(message, context)
}

// Error logs an error message
func Error(message string, context map[string]interface{}) error {
	return LoggingInstance.Error(message, context)
}

// Fatal logs a fatal message
func Fatal(message string, context map[string]interface{}) error {
	return LoggingInstance.Fatal(message, context)
}

// Emergency logs an emergency message
func Emergency(message string, context map[string]interface{}) error {
	return LoggingInstance.Emergency(message, context)
}

// Alert logs an alert message
func Alert(message string, context map[string]interface{}) error {
	return LoggingInstance.Alert(message, context)
}

// Critical logs a critical message
func Critical(message string, context map[string]interface{}) error {
	return LoggingInstance.Critical(message, context)
}

// Notice logs a notice message
func Notice(message string, context map[string]interface{}) error {
	return LoggingInstance.Notice(message, context)
}

// Log logs a message with a specific level
func Log(level logging_core.LogLevel, message string, context map[string]interface{}) error {
	return LoggingInstance.Log(level, message, context)
}

// WithContext returns a logger with context
func WithContext(ctx context.Context) logging_core.LoggerInterface {
	return LoggingInstance.WithContext(ctx)
}

// WithFields returns a logger with fields
func WithFields(fields map[string]interface{}) logging_core.LoggerInterface {
	return LoggingInstance.WithFields(fields)
}

// Channel returns a logger for a specific channel
func Channel(channelName string) (logging_core.LoggerInterface, error) {
	return LoggingInstance.Channel(channelName)
}

// Stack returns a logger that writes to multiple channels
func Stack(channels ...string) (logging_core.LoggerInterface, error) {
	return LoggingInstance.Stack(channels...)
}

// Report logs an exception to specified channels
func Report(exception error, channels ...string) error {
	return LoggingInstance.Report(exception, channels...)
}

// ReportWithLevel logs an exception to specified channels with custom level
func ReportWithLevel(exception error, level logging_core.LogLevel, channels ...string) error {
	return LoggingInstance.ReportWithLevel(exception, level, channels...)
}
