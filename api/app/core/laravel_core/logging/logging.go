package logging_core

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Define local interfaces and types instead of importing from go_core

// LoggerInterface defines the interface for logging
type LoggerInterface interface {
	Error(message string, context map[string]interface{}) error
	Info(message string, context map[string]interface{}) error
	Debug(message string, context map[string]interface{}) error
	Warning(message string, context map[string]interface{}) error
	Fatal(message string, context map[string]interface{}) error
	Log(level LogLevel, message string, context map[string]interface{}) error
	WithContext(ctx context.Context) LoggerInterface
	WithFields(fields map[string]interface{}) LoggerInterface
}

// LogDriver defines the interface for logging drivers
type LogDriver interface {
	Log(level LogLevel, message string, context map[string]interface{}) error
	Close() error
}

// LogLevel defines the logging level
type LogLevel int

const (
	LogLevelDebug   LogLevel = iota
	LogLevelInfo    LogLevel = iota
	LogLevelWarning LogLevel = iota
	LogLevelError   LogLevel = iota
	LogLevelFatal   LogLevel = iota
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarning:
		return "WARNING"
	case LogLevelError:
		return "ERROR"
	case LogLevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// LoggerConfig represents logger configuration
type LoggerConfig struct {
	Driver   string            `json:"driver"`
	Level    LogLevel          `json:"level"`
	Channels []string          `json:"channels"`
	Path     string            `json:"path"`
	MaxFiles int               `json:"max_files"`
	MaxSize  int64             `json:"max_size"`
	Options  map[string]string `json:"options"`
}

// Logger provides logging functionality with multiple drivers
type Logger struct {
	config  *LoggerConfig
	drivers map[string]LogDriver
	context context.Context
	fields  map[string]interface{}
}

// NewLogger creates a new logger instance
func NewLogger(config *LoggerConfig) *Logger {
	return &Logger{
		config:  config,
		drivers: make(map[string]LogDriver),
		context: context.Background(),
		fields:  make(map[string]interface{}),
	}
}

// Debug logs a debug message
func (l *Logger) Debug(message string, context map[string]interface{}) error {
	return l.Log(LogLevelDebug, message, context)
}

// Info logs an info message
func (l *Logger) Info(message string, context map[string]interface{}) error {
	return l.Log(LogLevelInfo, message, context)
}

// Warning logs a warning message
func (l *Logger) Warning(message string, context map[string]interface{}) error {
	return l.Log(LogLevelWarning, message, context)
}

// Error logs an error message
func (l *Logger) Error(message string, context map[string]interface{}) error {
	return l.Log(LogLevelError, message, context)
}

// Fatal logs a fatal message
func (l *Logger) Fatal(message string, context map[string]interface{}) error {
	return l.Log(LogLevelFatal, message, context)
}

// Log logs a message with the specified level
func (l *Logger) Log(level LogLevel, message string, context map[string]interface{}) error {
	// Merge fields with context
	mergedContext := make(map[string]interface{})
	for k, v := range l.fields {
		mergedContext[k] = v
	}
	for k, v := range context {
		mergedContext[k] = v
	}

	// Log to all configured drivers
	for _, driverName := range l.config.Channels {
		driver, exists := l.drivers[driverName]
		if !exists {
			// Initialize driver if not exists
			driver = l.initializeDriver(driverName)
			if driver != nil {
				l.drivers[driverName] = driver
			}
		}

		if driver != nil {
			if err := driver.Log(level, message, mergedContext); err != nil {
				// Fallback to standard log if driver fails
				log.Printf("Logger driver %s failed: %v", driverName, err)
			}
		}
	}

	return nil
}

// Emergency logs an emergency message (same as Fatal)
func (l *Logger) Emergency(message string, context map[string]interface{}) error {
	return l.Fatal(message, context)
}

// Alert logs an alert message (same as Error)
func (l *Logger) Alert(message string, context map[string]interface{}) error {
	return l.Error(message, context)
}

// Critical logs a critical message (same as Error)
func (l *Logger) Critical(message string, context map[string]interface{}) error {
	return l.Error(message, context)
}

// Notice logs a notice message (same as Info)
func (l *Logger) Notice(message string, context map[string]interface{}) error {
	return l.Info(message, context)
}

// WithContext returns a logger with the specified context
func (l *Logger) WithContext(ctx context.Context) LoggerInterface {
	newLogger := *l
	newLogger.context = ctx
	return &newLogger
}

// WithFields returns a logger with additional fields
func (l *Logger) WithFields(fields map[string]interface{}) LoggerInterface {
	newLogger := *l
	newLogger.fields = make(map[string]interface{})
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	for k, v := range fields {
		newLogger.fields[k] = v
	}
	return &newLogger
}

// initializeDriver initializes a log driver
func (l *Logger) initializeDriver(driverName string) LogDriver {
	switch driverName {
	case "single":
		return NewSingleLogDriver(l.config)
	case "daily":
		return NewDailyLogDriver(l.config)
	case "stack":
		return NewStackLogDriver(l.config)
	case "null":
		return NewNullLogDriver()
	default:
		return NewSingleLogDriver(l.config)
	}
}

// SingleLogDriver provides single file logging
type SingleLogDriver struct {
	config *LoggerConfig
	file   *os.File
}

// NewSingleLogDriver creates a new single log driver
func NewSingleLogDriver(config *LoggerConfig) *SingleLogDriver {
	// Ensure log directory exists
	logDir := filepath.Dir(config.Path)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("Failed to create log directory: %v", err)
		return nil
	}

	file, err := os.OpenFile(config.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		return nil
	}

	return &SingleLogDriver{
		config: config,
		file:   file,
	}
}

// Log implements LogDriver interface
func (d *SingleLogDriver) Log(level LogLevel, message string, context map[string]interface{}) error {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] %s: %s", timestamp, level, message)

	if len(context) > 0 {
		logEntry += fmt.Sprintf(" %+v", context)
	}

	logEntry += "\n"

	_, err := d.file.WriteString(logEntry)
	return err
}

// Close closes the log file
func (d *SingleLogDriver) Close() error {
	if d.file != nil {
		return d.file.Close()
	}
	return nil
}

// DailyLogDriver provides daily rotating file logging
type DailyLogDriver struct {
	config *LoggerConfig
	file   *os.File
}

// NewDailyLogDriver creates a new daily log driver
func NewDailyLogDriver(config *LoggerConfig) *DailyLogDriver {
	// Use daily file naming
	today := time.Now().Format("2006-01-02")
	logPath := fmt.Sprintf("%s-%s.log", config.Path, today)

	// Ensure log directory exists
	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("Failed to create log directory: %v", err)
		return nil
	}

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		return nil
	}

	return &DailyLogDriver{
		config: config,
		file:   file,
	}
}

// Log implements LogDriver interface
func (d *DailyLogDriver) Log(level LogLevel, message string, context map[string]interface{}) error {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] %s: %s", timestamp, level, message)

	if len(context) > 0 {
		logEntry += fmt.Sprintf(" %+v", context)
	}

	logEntry += "\n"

	_, err := d.file.WriteString(logEntry)
	return err
}

// Close closes the log file
func (d *DailyLogDriver) Close() error {
	if d.file != nil {
		return d.file.Close()
	}
	return nil
}

// StackLogDriver provides multiple driver logging
type StackLogDriver struct {
	drivers []LogDriver
}

// NewStackLogDriver creates a new stack log driver
func NewStackLogDriver(config *LoggerConfig) *StackLogDriver {
	drivers := make([]LogDriver, 0)

	// Add drivers based on configuration
	for _, driverName := range config.Channels {
		var driver LogDriver
		switch driverName {
		case "single":
			driver = NewSingleLogDriver(config)
		case "daily":
			driver = NewDailyLogDriver(config)
		}

		if driver != nil {
			drivers = append(drivers, driver)
		}
	}

	return &StackLogDriver{
		drivers: drivers,
	}
}

// Log implements LogDriver interface
func (d *StackLogDriver) Log(level LogLevel, message string, context map[string]interface{}) error {
	for _, driver := range d.drivers {
		if err := driver.Log(level, message, context); err != nil {
			// Continue with other drivers even if one fails
			log.Printf("Stack driver failed: %v", err)
		}
	}
	return nil
}

// Close implements LogDriver interface
func (d *StackLogDriver) Close() error {
	for _, driver := range d.drivers {
		if err := driver.Close(); err != nil {
			log.Printf("Failed to close stack driver: %v", err)
		}
	}
	return nil
}

// NullLogDriver provides no-op logging
type NullLogDriver struct{}

// NewNullLogDriver creates a new null log driver
func NewNullLogDriver() *NullLogDriver {
	return &NullLogDriver{}
}

// Log implements LogDriver interface
func (d *NullLogDriver) Log(level LogLevel, message string, context map[string]interface{}) error {
	// Do nothing
	return nil
}

// Close implements LogDriver interface
func (d *NullLogDriver) Close() error {
	return nil
}

// Global logger instance
var LoggerInstance LoggerInterface

// SetLogger sets the global logger instance
func SetLogger(logger LoggerInterface) {
	LoggerInstance = logger
}

// Helper functions for global logging (Laravel-style)

// Log logs a message with the global logger
func Log(level LogLevel, message string, context map[string]interface{}) error {
	if LoggerInstance == nil {
		return fmt.Errorf("logger not initialized")
	}
	return LoggerInstance.Log(level, message, context)
}

// Debug logs a debug message
func Debug(message string, context map[string]interface{}) error {
	return Log(LogLevelDebug, message, context)
}

// Info logs an info message
func Info(message string, context map[string]interface{}) error {
	return Log(LogLevelInfo, message, context)
}

// Warning logs a warning message
func Warning(message string, context map[string]interface{}) error {
	return Log(LogLevelWarning, message, context)
}

// Error logs an error message
func Error(message string, context map[string]interface{}) error {
	return Log(LogLevelError, message, context)
}

// Fatal logs a fatal message
func Fatal(message string, context map[string]interface{}) error {
	return Log(LogLevelFatal, message, context)
}

// Emergency logs an emergency message
func Emergency(message string, context map[string]interface{}) error {
	return Fatal(message, context)
}

// Alert logs an alert message
func Alert(message string, context map[string]interface{}) error {
	return Error(message, context)
}

// Critical logs a critical message
func Critical(message string, context map[string]interface{}) error {
	return Error(message, context)
}

// Notice logs a notice message
func Notice(message string, context map[string]interface{}) error {
	return Info(message, context)
}
