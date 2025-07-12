package logging_core

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	go_core "base_lara_go_project/app/core/go_core"
)

// OptimizedFileHandler provides high-performance file logging
type OptimizedFileHandler[T any] struct {
	path    string
	file    *os.File
	mu      sync.Mutex
	metrics *go_core.LoggingMetrics
}

// NewOptimizedFileHandler creates a new optimized file handler
func NewOptimizedFileHandler[T any](path string) (*OptimizedFileHandler[T], error) {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open file
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &OptimizedFileHandler[T]{
		path:    path,
		file:    file,
		metrics: &go_core.LoggingMetrics{},
	}, nil
}

// Handle implements LogHandler interface
func (h *OptimizedFileHandler[T]) Handle(entry go_core.LogEntry[T]) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Format log entry
	logLine := fmt.Sprintf("[%s] %s: %s\n",
		entry.Timestamp.Format("2006-01-02 15:04:05"),
		entry.Level.String(),
		entry.Message)

	// Write to file
	_, err := h.file.WriteString(logLine)
	if err != nil {
		return fmt.Errorf("failed to write to log file: %w", err)
	}

	// Update metrics
	h.metrics.EntriesLogged++

	return nil
}

// ShouldHandle determines if the handler should handle the given level
func (h *OptimizedFileHandler[T]) ShouldHandle(level go_core.LogLevel) bool {
	return level >= go_core.LogLevelInfo
}

// GetLevel returns the handler's level
func (h *OptimizedFileHandler[T]) GetLevel() go_core.LogLevel {
	return go_core.LogLevelInfo
}

// Close closes the file handler
func (h *OptimizedFileHandler[T]) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.file.Close()
}

// Flush flushes the file handler
func (h *OptimizedFileHandler[T]) Flush() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.file.Sync()
}

// GetMetrics returns handler metrics
func (h *OptimizedFileHandler[T]) GetMetrics() *go_core.LoggingMetrics {
	return h.metrics
}

// OptimizedDailyHandler provides high-performance daily file logging
type OptimizedDailyHandler[T any] struct {
	basePath   string
	currentDay string
	handler    *OptimizedFileHandler[T]
	mu         sync.Mutex
}

// NewOptimizedDailyHandler creates a new optimized daily handler
func NewOptimizedDailyHandler[T any](basePath string) (*OptimizedDailyHandler[T], error) {
	return &OptimizedDailyHandler[T]{
		basePath: basePath,
	}, nil
}

// Handle implements LogHandler interface
func (h *OptimizedDailyHandler[T]) Handle(entry go_core.LogEntry[T]) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Check if we need to rotate to a new day
	today := entry.Timestamp.Format("2006-01-02")
	if h.currentDay != today {
		// Close old handler if exists
		if h.handler != nil {
			h.handler.Close()
		}

		// Create new handler for today
		logPath := fmt.Sprintf("%s-%s.log", h.basePath, today)
		handler, err := NewOptimizedFileHandler[T](logPath)
		if err != nil {
			return fmt.Errorf("failed to create daily handler: %w", err)
		}

		h.handler = handler
		h.currentDay = today
	}

	// Handle the entry
	return h.handler.Handle(entry)
}

// ShouldHandle determines if the handler should handle the given level
func (h *OptimizedDailyHandler[T]) ShouldHandle(level go_core.LogLevel) bool {
	return level >= go_core.LogLevelInfo
}

// GetLevel returns the handler's level
func (h *OptimizedDailyHandler[T]) GetLevel() go_core.LogLevel {
	return go_core.LogLevelInfo
}

// Close closes the daily handler
func (h *OptimizedDailyHandler[T]) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.handler != nil {
		return h.handler.Close()
	}
	return nil
}

// Flush flushes the daily handler
func (h *OptimizedDailyHandler[T]) Flush() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.handler != nil {
		return h.handler.Flush()
	}
	return nil
}

// GetMetrics returns handler metrics
func (h *OptimizedDailyHandler[T]) GetMetrics() *go_core.LoggingMetrics {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.handler != nil {
		return h.handler.GetMetrics()
	}
	return &go_core.LoggingMetrics{}
}

// OptimizedStackHandler provides high-performance stack logging
type OptimizedStackHandler[T any] struct {
	handlers []go_core.LogHandler[T]
	mu       sync.RWMutex
}

// NewOptimizedStackHandler creates a new optimized stack handler
func NewOptimizedStackHandler[T any](handlers []go_core.LogHandler[T]) *OptimizedStackHandler[T] {
	return &OptimizedStackHandler[T]{
		handlers: handlers,
	}
}

// Handle implements LogHandler interface with concurrent processing
func (h *OptimizedStackHandler[T]) Handle(entry go_core.LogEntry[T]) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Process all handlers concurrently
	errors := make(chan error, len(h.handlers))

	for _, handler := range h.handlers {
		go func(hdl go_core.LogHandler[T]) {
			if hdl.ShouldHandle(entry.Level) {
				errors <- hdl.Handle(entry)
			} else {
				errors <- nil
			}
		}(handler)
	}

	// Collect errors
	var lastError error
	for i := 0; i < len(h.handlers); i++ {
		if err := <-errors; err != nil {
			lastError = err
		}
	}

	return lastError
}

// ShouldHandle determines if the handler should handle the given level
func (h *OptimizedStackHandler[T]) ShouldHandle(level go_core.LogLevel) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Check if any handler should handle this level
	for _, handler := range h.handlers {
		if handler.ShouldHandle(level) {
			return true
		}
	}
	return false
}

// GetLevel returns the handler's level (minimum level of all handlers)
func (h *OptimizedStackHandler[T]) GetLevel() go_core.LogLevel {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.handlers) == 0 {
		return go_core.LogLevelDebug
	}

	minLevel := h.handlers[0].GetLevel()
	for _, handler := range h.handlers {
		if handler.GetLevel() < minLevel {
			minLevel = handler.GetLevel()
		}
	}
	return minLevel
}

// Close closes all handlers
func (h *OptimizedStackHandler[T]) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	var lastError error
	for _, handler := range h.handlers {
		if err := handler.Close(); err != nil {
			lastError = err
		}
	}
	return lastError
}

// Flush flushes all handlers
func (h *OptimizedStackHandler[T]) Flush() error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var lastError error
	for _, handler := range h.handlers {
		if flushable, ok := handler.(interface{ Flush() error }); ok {
			if err := flushable.Flush(); err != nil {
				lastError = err
			}
		}
	}
	return lastError
}

// OptimizedNullHandler provides high-performance null logging
type OptimizedNullHandler[T any] struct {
	*go_core.NullLogHandler[T]
}

// NewOptimizedNullHandler creates a new optimized null handler
func NewOptimizedNullHandler[T any]() *OptimizedNullHandler[T] {
	return &OptimizedNullHandler[T]{
		NullLogHandler: go_core.NewNullLogHandler[T](),
	}
}

// Handle implements LogHandler interface (discards all messages)
func (h *OptimizedNullHandler[T]) Handle(entry go_core.LogEntry[T]) error {
	// Use the underlying go_core null handler
	return h.NullLogHandler.Handle(entry)
}

// OptimizedSentryHandler provides high-performance Sentry integration
type OptimizedSentryHandler[T any] struct {
	level go_core.LogLevel
	mu    sync.Mutex
}

// NewOptimizedSentryHandler creates a new optimized Sentry handler
func NewOptimizedSentryHandler[T any]() *OptimizedSentryHandler[T] {
	return &OptimizedSentryHandler[T]{
		level: go_core.LogLevelError, // Only handle errors and above
	}
}

// Handle implements LogHandler interface with Sentry integration
func (h *OptimizedSentryHandler[T]) Handle(entry go_core.LogEntry[T]) error {
	// Only handle error levels and above
	if entry.Level < go_core.LogLevelError {
		return nil
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// TODO: Implement actual Sentry integration
	// For now, just log to stderr
	fmt.Fprintf(os.Stderr, "[SENTRY] %s: %s\n", entry.Level, entry.Message)

	return nil
}

// ShouldHandle determines if the handler should handle the given level
func (h *OptimizedSentryHandler[T]) ShouldHandle(level go_core.LogLevel) bool {
	return level >= h.level
}

// GetLevel returns the handler's level
func (h *OptimizedSentryHandler[T]) GetLevel() go_core.LogLevel {
	return h.level
}

// Close closes the Sentry handler
func (h *OptimizedSentryHandler[T]) Close() error {
	// TODO: Implement Sentry flush
	return nil
}

// Flush flushes the Sentry handler
func (h *OptimizedSentryHandler[T]) Flush() error {
	// TODO: Implement Sentry flush
	return nil
}

// OptimizedSlackHandler provides high-performance Slack integration
type OptimizedSlackHandler[T any] struct {
	level go_core.LogLevel
	mu    sync.Mutex
}

// NewOptimizedSlackHandler creates a new optimized Slack handler
func NewOptimizedSlackHandler[T any]() *OptimizedSlackHandler[T] {
	return &OptimizedSlackHandler[T]{
		level: go_core.LogLevelError, // Only handle errors and above
	}
}

// Handle implements LogHandler interface with Slack integration
func (h *OptimizedSlackHandler[T]) Handle(entry go_core.LogEntry[T]) error {
	// Only handle error levels and above
	if entry.Level < go_core.LogLevelError {
		return nil
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// TODO: Implement actual Slack integration
	// For now, just log to stderr
	fmt.Fprintf(os.Stderr, "[SLACK] %s: %s\n", entry.Level, entry.Message)

	return nil
}

// ShouldHandle determines if the handler should handle the given level
func (h *OptimizedSlackHandler[T]) ShouldHandle(level go_core.LogLevel) bool {
	return level >= h.level
}

// GetLevel returns the handler's level
func (h *OptimizedSlackHandler[T]) GetLevel() go_core.LogLevel {
	return h.level
}

// Close closes the Slack handler
func (h *OptimizedSlackHandler[T]) Close() error {
	// TODO: Implement Slack flush
	return nil
}

// Flush flushes the Slack handler
func (h *OptimizedSlackHandler[T]) Flush() error {
	// TODO: Implement Slack flush
	return nil
}
