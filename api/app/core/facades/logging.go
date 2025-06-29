package facades_core

import (
	logging_core "base_lara_go_project/app/core/logging"
	"context"
)

// Logging provides a facade for logging operations
type Logging struct{}

// Global logging facade instance
var LoggingInstance = &Logging{}

// Debug logs a debug message
func (l *Logging) Debug(message string, context map[string]interface{}) error {
	return logging_core.Debug(message, context)
}

// Info logs an info message
func (l *Logging) Info(message string, context map[string]interface{}) error {
	return logging_core.Info(message, context)
}

// Warning logs a warning message
func (l *Logging) Warning(message string, context map[string]interface{}) error {
	return logging_core.Warning(message, context)
}

// Error logs an error message
func (l *Logging) Error(message string, context map[string]interface{}) error {
	return logging_core.Error(message, context)
}

// Fatal logs a fatal message
func (l *Logging) Fatal(message string, context map[string]interface{}) error {
	return logging_core.Fatal(message, context)
}

// Emergency logs an emergency message
func (l *Logging) Emergency(message string, context map[string]interface{}) error {
	return logging_core.Emergency(message, context)
}

// Alert logs an alert message
func (l *Logging) Alert(message string, context map[string]interface{}) error {
	return logging_core.Alert(message, context)
}

// Critical logs a critical message
func (l *Logging) Critical(message string, context map[string]interface{}) error {
	return logging_core.Critical(message, context)
}

// Notice logs a notice message
func (l *Logging) Notice(message string, context map[string]interface{}) error {
	return logging_core.Notice(message, context)
}

// Log logs a message with a specific level
func (l *Logging) Log(level logging_core.LogLevel, message string, context map[string]interface{}) error {
	return logging_core.Log(level, message, context)
}

// WithContext returns a logger with context
func (l *Logging) WithContext(ctx context.Context) logging_core.LoggerInterface {
	if logging_core.LoggerInstance == nil {
		return nil
	}
	return logging_core.LoggerInstance.WithContext(ctx)
}

// WithFields returns a logger with fields
func (l *Logging) WithFields(fields map[string]interface{}) logging_core.LoggerInterface {
	if logging_core.LoggerInstance == nil {
		return nil
	}
	return logging_core.LoggerInstance.WithFields(fields)
}

// Channel returns a logger for a specific channel
func (l *Logging) Channel(channelName string) (logging_core.LoggerInterface, error) {
	return logging_core.GetLoggerForChannel(channelName)
}

// Stack returns a logger that writes to multiple channels
func (l *Logging) Stack(channels ...string) (logging_core.LoggerInterface, error) {
	// For now, just return the default logger
	return logging_core.GetLogger(), nil
}

// Report logs an exception to specified channels (Laravel-style)
func (l *Logging) Report(exception error, channels ...string) error {
	context := map[string]interface{}{
		"exception": exception.Error(),
		"type":      "exception",
	}

	if len(channels) == 0 {
		// Use default channels for exceptions
		return l.Error("Exception occurred", context)
	}

	// Log to specified channels
	for _, channel := range channels {
		if logger, err := l.Channel(channel); err == nil {
			logger.Error("Exception occurred", context)
		}
	}

	return nil
}

// ReportWithLevel logs an exception to specified channels with custom level
func (l *Logging) ReportWithLevel(exception error, level logging_core.LogLevel, channels ...string) error {
	context := map[string]interface{}{
		"exception": exception.Error(),
		"type":      "exception",
	}

	if len(channels) == 0 {
		// Use default channels for exceptions
		return l.Log(level, "Exception occurred", context)
	}

	// Log to specified channels
	for _, channel := range channels {
		if logger, err := l.Channel(channel); err == nil {
			logger.Log(level, "Exception occurred", context)
		}
	}

	return nil
}

// Global logging functions for convenience

// Debug logs a debug message
func Debug(message string, context map[string]interface{}) error {
	return logging_core.Debug(message, context)
}

// Info logs an info message
func Info(message string, context map[string]interface{}) error {
	return logging_core.Info(message, context)
}

// Warning logs a warning message
func Warning(message string, context map[string]interface{}) error {
	return logging_core.Warning(message, context)
}

// Error logs an error message
func Error(message string, context map[string]interface{}) error {
	return logging_core.Error(message, context)
}

// Fatal logs a fatal message
func Fatal(message string, context map[string]interface{}) error {
	return logging_core.Fatal(message, context)
}

// Emergency logs an emergency message
func Emergency(message string, context map[string]interface{}) error {
	return logging_core.Emergency(message, context)
}

// Alert logs an alert message
func Alert(message string, context map[string]interface{}) error {
	return logging_core.Alert(message, context)
}

// Critical logs a critical message
func Critical(message string, context map[string]interface{}) error {
	return logging_core.Critical(message, context)
}

// Notice logs a notice message
func Notice(message string, context map[string]interface{}) error {
	return logging_core.Notice(message, context)
}

// Log logs a message with a specific level
func Log(level logging_core.LogLevel, message string, context map[string]interface{}) error {
	return logging_core.Log(level, message, context)
}

// WithContext returns a logger with context
func WithContext(ctx context.Context) logging_core.LoggerInterface {
	return logging_core.LoggerInstance.WithContext(ctx)
}

// WithFields returns a logger with fields
func WithFields(fields map[string]interface{}) logging_core.LoggerInterface {
	return logging_core.LoggerInstance.WithFields(fields)
}

// Channel returns a logger for a specific channel
func Channel(channelName string) (logging_core.LoggerInterface, error) {
	return logging_core.GetLoggerForChannel(channelName)
}

// Stack returns a logger that writes to multiple channels
func Stack(channels ...string) (logging_core.LoggerInterface, error) {
	return logging_core.GetLogger(), nil
}

// Report logs an exception to specified channels (Laravel-style)
func Report(exception error, channels ...string) error {
	return LoggingInstance.Report(exception, channels...)
}

// ReportWithLevel logs an exception to specified channels with custom level
func ReportWithLevel(exception error, level logging_core.LogLevel, channels ...string) error {
	return LoggingInstance.ReportWithLevel(exception, level, channels...)
}
