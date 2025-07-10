package logging_core

import (
	"fmt"
	"os"
	"sync"
	"time"

	go_core "base_lara_go_project/app/core/go_core"
	config_core "base_lara_go_project/app/core/laravel_core/config"
)

// OptimizedFileHandler provides high-performance file logging with go_core optimizations
type OptimizedFileHandler[T any] struct {
	*go_core.FileLogHandler[T]
	config *config_core.ConfigFacade
	path   string
	mu     sync.Mutex
}

// NewOptimizedFileHandler creates a new optimized file handler
func NewOptimizedFileHandler[T any](config *config_core.ConfigFacade, path string) (*OptimizedFileHandler[T], error) {
	// Convert config to go_core config
	goCoreConfig := convertToGoCoreConfig(config)

	// Create go_core file handler
	fileHandler, err := go_core.NewFileLogHandler[T](goCoreConfig, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create file handler: %w", err)
	}

	return &OptimizedFileHandler[T]{
		FileLogHandler: fileHandler,
		config:         config,
		path:           path,
	}, nil
}

// Handle implements LogHandler interface with additional optimizations
func (h *OptimizedFileHandler[T]) Handle(entry go_core.LogEntry[T]) error {
	// Add Laravel-specific formatting
	formattedEntry := h.formatForLaravel(entry)

	// Use the underlying go_core handler
	return h.FileLogHandler.Handle(formattedEntry)
}

// formatForLaravel formats the log entry for Laravel compatibility
func (h *OptimizedFileHandler[T]) formatForLaravel(entry go_core.LogEntry[T]) go_core.LogEntry[T] {
	// Add Laravel-specific context if needed
	// This could include request ID, user ID, etc.
	return entry
}

// OptimizedDailyHandler provides high-performance daily rotating file logging
type OptimizedDailyHandler[T any] struct {
	*go_core.FileLogHandler[T]
	config     *config_core.ConfigFacade
	basePath   string
	currentDay string
	mu         sync.Mutex
}

// NewOptimizedDailyHandler creates a new optimized daily handler
func NewOptimizedDailyHandler[T any](config *config_core.ConfigFacade, basePath string) (*OptimizedDailyHandler[T], error) {
	// Convert config to go_core config
	goCoreConfig := convertToGoCoreConfig(config)

	// Get today's file path
	today := time.Now().Format("2006-01-02")
	path := fmt.Sprintf("%s-%s.log", basePath, today)

	// Create go_core file handler
	fileHandler, err := go_core.NewFileLogHandler[T](goCoreConfig, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create daily handler: %w", err)
	}

	return &OptimizedDailyHandler[T]{
		FileLogHandler: fileHandler,
		config:         config,
		basePath:       basePath,
		currentDay:     today,
	}, nil
}

// Handle implements LogHandler interface with daily rotation
func (h *OptimizedDailyHandler[T]) Handle(entry go_core.LogEntry[T]) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Check if we need to rotate to a new day
	today := time.Now().Format("2006-01-02")
	if today != h.currentDay {
		// Create new file handler for the new day
		path := fmt.Sprintf("%s-%s.log", h.basePath, today)
		newHandler, err := go_core.NewFileLogHandler[T](convertToGoCoreConfig(h.config), path)
		if err == nil {
			// Close old handler
			h.FileLogHandler.Close()
			// Update to new handler
			h.FileLogHandler = newHandler
			h.currentDay = today
		}
	}

	// Use the underlying go_core handler
	return h.FileLogHandler.Handle(entry)
}

// OptimizedStackHandler provides high-performance stack logging
type OptimizedStackHandler[T any] struct {
	handlers []go_core.LogHandler[T]
	config   *config_core.ConfigFacade
	mu       sync.RWMutex
}

// NewOptimizedStackHandler creates a new optimized stack handler
func NewOptimizedStackHandler[T any](config *config_core.ConfigFacade, handlers []go_core.LogHandler[T]) *OptimizedStackHandler[T] {
	return &OptimizedStackHandler[T]{
		handlers: handlers,
		config:   config,
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
	config *config_core.ConfigFacade
	level  go_core.LogLevel
	mu     sync.Mutex
}

// NewOptimizedSentryHandler creates a new optimized Sentry handler
func NewOptimizedSentryHandler[T any](config *config_core.ConfigFacade) *OptimizedSentryHandler[T] {
	return &OptimizedSentryHandler[T]{
		config: config,
		level:  go_core.LogLevelError, // Only handle errors and above
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
	config *config_core.ConfigFacade
	level  go_core.LogLevel
	mu     sync.Mutex
}

// NewOptimizedSlackHandler creates a new optimized Slack handler
func NewOptimizedSlackHandler[T any](config *config_core.ConfigFacade) *OptimizedSlackHandler[T] {
	return &OptimizedSlackHandler[T]{
		config: config,
		level:  go_core.LogLevelError, // Only handle errors and above
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
