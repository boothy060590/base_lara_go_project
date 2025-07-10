package go_core

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// LogLevel represents logging levels with atomic operations
type LogLevel int32

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarning
	LogLevelError
	LogLevelFatal
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

// LogEntry represents a log entry with generic context
type LogEntry[T any] struct {
	Level     LogLevel
	Message   string
	Context   T
	Timestamp time.Time
	TraceID   string
	SpanID    string
}

// LogHandler defines a generic log handler interface
type LogHandler[T any] interface {
	Handle(entry LogEntry[T]) error
	ShouldHandle(level LogLevel) bool
	GetLevel() LogLevel
	Close() error
}

// LoggerConfig represents optimized logger configuration
type LoggerConfig struct {
	DefaultLevel    LogLevel
	MaxConcurrency  int
	BufferSize      int
	FlushInterval   time.Duration
	ObjectPoolSize  int
	ContextTimeout  time.Duration
	PerformanceMode bool
}

// DefaultLoggerConfig returns optimized default configuration
func DefaultLoggerConfig() *LoggerConfig {
	return &LoggerConfig{
		DefaultLevel:    LogLevelInfo,
		MaxConcurrency:  100,
		BufferSize:      1000,
		FlushInterval:   time.Second,
		ObjectPoolSize:  100,
		ContextTimeout:  30 * time.Second,
		PerformanceMode: true,
	}
}

// Logger provides high-performance, generic logging with automatic optimizations
type Logger[T any] struct {
	config    *LoggerConfig
	handlers  map[string]LogHandler[T]
	entryPool *ObjectPool[LogEntry[T]]
	metrics   *LoggingMetrics
	mu        sync.RWMutex
	closed    int32
}

// LoggingMetrics tracks logging performance metrics
type LoggingMetrics struct {
	EntriesLogged  int64
	EntriesDropped int64
	HandlerErrors  int64
	AverageLatency int64
	LastFlushTime  time.Time
	mu             sync.RWMutex
}

// NewLogger creates a new optimized logger instance
func NewLogger[T any](config *LoggerConfig) *Logger[T] {
	if config == nil {
		config = DefaultLoggerConfig()
	}

	// Create object pool for log entries
	entryPool := NewObjectPool[LogEntry[T]](config.ObjectPoolSize, func() LogEntry[T] {
		return LogEntry[T]{
			Timestamp: time.Now(),
		}
	}, func(entry LogEntry[T]) LogEntry[T] {
		// Reset entry for reuse
		entry.Level = LogLevelDebug
		entry.Message = ""
		entry.Context = *new(T)
		entry.Timestamp = time.Now()
		entry.TraceID = ""
		entry.SpanID = ""
		return entry
	})

	logger := &Logger[T]{
		config:    config,
		handlers:  make(map[string]LogHandler[T]),
		entryPool: entryPool,
		metrics:   &LoggingMetrics{},
	}

	// Start background flush routine
	logger.startBackgroundFlush()

	return logger
}

// Log logs a message with generic context using automatic optimizations
func (l *Logger[T]) Log(level LogLevel, message string, context T) error {
	if atomic.LoadInt32(&l.closed) == 1 {
		return fmt.Errorf("logger is closed")
	}

	// Get entry from pool
	entry := l.entryPool.Get()
	entry.Level = level
	entry.Message = message
	entry.Context = context
	entry.Timestamp = time.Now()

	// Add trace and span IDs if available
	if traceID := l.getTraceID(); traceID != "" {
		entry.TraceID = traceID
	}
	if spanID := l.getSpanID(); spanID != "" {
		entry.SpanID = spanID
	}

	// Track performance
	start := time.Now()
	defer func() {
		atomic.AddInt64(&l.metrics.EntriesLogged, 1)
		atomic.StoreInt64(&l.metrics.AverageLatency, int64(time.Since(start)))
	}()

	// Process log entry
	return l.processLogEntry(entry)
}

// processLogEntry processes a log entry with automatic optimizations
func (l *Logger[T]) processLogEntry(entry LogEntry[T]) error {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// Route to appropriate handlers
	for name, handler := range l.handlers {
		if handler.ShouldHandle(entry.Level) {
			// Execute handler with context timeout
			ctx, cancel := context.WithTimeout(context.Background(), l.config.ContextTimeout)
			defer cancel()

			// Execute handler
			err := l.executeHandlerWithContext(ctx, handler, entry)
			if err != nil {
				atomic.AddInt64(&l.metrics.HandlerErrors, 1)
				// Don't fail the entire log operation for handler errors
				log.Printf("Log handler %s failed: %v", name, err)
			}
		}
	}

	// Return entry to pool
	l.entryPool.Put(entry)
	return nil
}

// executeHandlerWithContext executes a handler with context
func (l *Logger[T]) executeHandlerWithContext(ctx context.Context, handler LogHandler[T], entry LogEntry[T]) error {
	// Create result channel
	resultChan := make(chan error, 1)

	go func() {
		resultChan <- handler.Handle(entry)
	}()

	select {
	case err := <-resultChan:
		return err
	case <-ctx.Done():
		return fmt.Errorf("handler execution timed out: %w", ctx.Err())
	}
}

// AddHandler adds a log handler with automatic optimization
func (l *Logger[T]) AddHandler(name string, handler LogHandler[T]) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.handlers[name] = handler
}

// RemoveHandler removes a log handler
func (l *Logger[T]) RemoveHandler(name string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.handlers, name)
}

// GetHandler returns a handler by name
func (l *Logger[T]) GetHandler(name string) (LogHandler[T], bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	handler, exists := l.handlers[name]
	return handler, exists
}

// Close closes the logger and all handlers
func (l *Logger[T]) Close() error {
	if !atomic.CompareAndSwapInt32(&l.closed, 0, 1) {
		return nil // Already closed
	}

	// Close all handlers
	l.mu.RLock()
	for name, handler := range l.handlers {
		if err := handler.Close(); err != nil {
			log.Printf("Failed to close handler %s: %v", name, err)
		}
	}
	l.mu.RUnlock()

	return nil
}

// GetMetrics returns logging metrics
func (l *Logger[T]) GetMetrics() *LoggingMetrics {
	l.metrics.mu.RLock()
	defer l.metrics.mu.RUnlock()

	// Create a copy to avoid race conditions
	metrics := &LoggingMetrics{
		EntriesLogged:  atomic.LoadInt64(&l.metrics.EntriesLogged),
		EntriesDropped: atomic.LoadInt64(&l.metrics.EntriesDropped),
		HandlerErrors:  atomic.LoadInt64(&l.metrics.HandlerErrors),
		AverageLatency: atomic.LoadInt64(&l.metrics.AverageLatency),
		LastFlushTime:  l.metrics.LastFlushTime,
	}

	return metrics
}

// startBackgroundFlush starts the background flush routine
func (l *Logger[T]) startBackgroundFlush() {
	go func() {
		ticker := time.NewTicker(l.config.FlushInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				l.flushHandlers()
			case <-time.After(l.config.FlushInterval * 10): // Safety timeout
				return
			}
		}
	}()
}

// flushHandlers flushes all handlers
func (l *Logger[T]) flushHandlers() {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for name, handler := range l.handlers {
		if flushable, ok := handler.(interface{ Flush() error }); ok {
			if err := flushable.Flush(); err != nil {
				log.Printf("Failed to flush handler %s: %v", name, err)
			}
		}
	}

	l.metrics.mu.Lock()
	l.metrics.LastFlushTime = time.Now()
	l.metrics.mu.Unlock()
}

// getTraceID gets the current trace ID from context
func (l *Logger[T]) getTraceID() string {
	// Implementation would integrate with tracing system
	return ""
}

// getSpanID gets the current span ID from context
func (l *Logger[T]) getSpanID() string {
	// Implementation would integrate with tracing system
	return ""
}

// Convenience methods for different log levels
func (l *Logger[T]) Debug(message string, context T) error {
	return l.Log(LogLevelDebug, message, context)
}

func (l *Logger[T]) Info(message string, context T) error {
	return l.Log(LogLevelInfo, message, context)
}

func (l *Logger[T]) Warning(message string, context T) error {
	return l.Log(LogLevelWarning, message, context)
}

func (l *Logger[T]) Error(message string, context T) error {
	return l.Log(LogLevelError, message, context)
}

func (l *Logger[T]) Fatal(message string, context T) error {
	return l.Log(LogLevelFatal, message, context)
}

// FileLogHandler provides optimized file logging
type FileLogHandler[T any] struct {
	config *LoggerConfig
	file   *os.File
	mu     sync.Mutex
	level  LogLevel
	path   string
}

// NewFileLogHandler creates a new optimized file log handler
func NewFileLogHandler[T any](config *LoggerConfig, path string) (*FileLogHandler[T], error) {
	// Ensure log directory exists
	logDir := filepath.Dir(path)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &FileLogHandler[T]{
		config: config,
		file:   file,
		level:  config.DefaultLevel,
		path:   path,
	}, nil
}

// Handle implements LogHandler interface with performance tracking
func (h *FileLogHandler[T]) Handle(entry LogEntry[T]) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] %s: %s", timestamp, entry.Level, entry.Message)

	// Add context if not empty
	if !isEmpty(entry.Context) {
		logEntry += fmt.Sprintf(" %+v", entry.Context)
	}

	// Add trace and span IDs if available
	if entry.TraceID != "" {
		logEntry += fmt.Sprintf(" trace_id=%s", entry.TraceID)
	}
	if entry.SpanID != "" {
		logEntry += fmt.Sprintf(" span_id=%s", entry.SpanID)
	}

	logEntry += "\n"

	_, err := h.file.WriteString(logEntry)
	return err
}

// ShouldHandle determines if the handler should handle the given level
func (h *FileLogHandler[T]) ShouldHandle(level LogLevel) bool {
	return level >= h.level
}

// GetLevel returns the handler's level
func (h *FileLogHandler[T]) GetLevel() LogLevel {
	return h.level
}

// Close closes the file handler
func (h *FileLogHandler[T]) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.file != nil {
		return h.file.Close()
	}
	return nil
}

// Flush flushes the file buffer
func (h *FileLogHandler[T]) Flush() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.file != nil {
		return h.file.Sync()
	}
	return nil
}

// isEmpty checks if a generic value is empty
func isEmpty[T any](value T) bool {
	var zero T
	return fmt.Sprintf("%v", value) == fmt.Sprintf("%v", zero)
}

// NullLogHandler provides a null handler that discards all messages
type NullLogHandler[T any] struct {
	level LogLevel
}

// NewNullLogHandler creates a new null log handler
func NewNullLogHandler[T any]() *NullLogHandler[T] {
	return &NullLogHandler[T]{
		level: LogLevelDebug,
	}
}

// Handle implements LogHandler interface (discards all messages)
func (h *NullLogHandler[T]) Handle(entry LogEntry[T]) error {
	return nil
}

// ShouldHandle determines if the handler should handle the given level
func (h *NullLogHandler[T]) ShouldHandle(level LogLevel) bool {
	return level >= h.level
}

// GetLevel returns the handler's level
func (h *NullLogHandler[T]) GetLevel() LogLevel {
	return h.level
}

// Close closes the null handler
func (h *NullLogHandler[T]) Close() error {
	return nil
}

// Flush flushes the null handler
func (h *NullLogHandler[T]) Flush() error {
	return nil
}
