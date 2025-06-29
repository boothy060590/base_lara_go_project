package logging_core

import (
	"fmt"
	"time"

	app_core "base_lara_go_project/app/core/app"
	client_core "base_lara_go_project/app/core/clients"
	exceptions_core "base_lara_go_project/app/core/exceptions"
)

// LoggingClient provides a generic logging client implementation
type LoggingClient struct {
	*client_core.BaseClient
	level    string
	handlers map[string]app_core.LogHandler
}

// LogHandler defines the interface for log handlers
type LogHandler interface {
	Handle(level string, message string, context map[string]interface{}) error
	ShouldHandle(level string) bool
	GetLevel() string
}

// NewLoggingClient creates a new logging client
func NewLoggingClient(config *client_core.ClientConfig) *LoggingClient {
	return &LoggingClient{
		BaseClient: client_core.NewBaseClient(config, "logging"),
		level:      "info",
		handlers:   make(map[string]app_core.LogHandler),
	}
}

// convertToException converts a generic error to our framework exception
func convertToException(err error) app_core.Exception {
	return exceptions_core.ConvertToException(err)
}

// Log implements LoggingClientInterface
func (c *LoggingClient) Log(level string, message string, context map[string]interface{}) error {
	if !c.IsConnected() {
		return fmt.Errorf("logging client not connected")
	}

	// Add timestamp and client info to context
	if context == nil {
		context = make(map[string]interface{})
	}
	context["timestamp"] = time.Now().Format(time.RFC3339)
	context["client"] = c.GetName()
	context["level"] = level

	// Route to appropriate handlers
	for handlerName, handler := range c.handlers {
		if handler.ShouldHandle(level) {
			if err := handler.Handle(level, message, context); err != nil {
				// Log handler error but don't fail the main log
				fmt.Printf("Log handler %s failed: %v\n", handlerName, err)
			}
		}
	}

	return nil
}

// LogException implements LoggingClientInterface
func (c *LoggingClient) LogException(exception error) error {
	// Convert to our exception type if needed
	ex := convertToException(exception)

	// Add exception context
	context := map[string]interface{}{
		"exception_type":  "framework_exception",
		"exception_code":  ex.GetCode(),
		"exception_file":  ex.GetFile(),
		"exception_line":  ex.GetLine(),
		"exception_trace": ex.GetTrace(),
	}

	// Merge with exception context
	for k, v := range ex.GetContext() {
		context[k] = v
	}

	// Determine log level based on exception code
	level := "error"
	switch ex.GetCode() {
	case 404:
		level = "warning"
	case 401, 403:
		level = "info"
	case 422:
		level = "warning"
	case 500:
		level = "error"
	default:
		level = "error"
	}

	return c.Log(level, ex.GetMessage(), context)
}

// Flush implements LoggingClientInterface
func (c *LoggingClient) Flush() error {
	// Flush all handlers
	for handlerName, handler := range c.handlers {
		if flushable, ok := handler.(interface{ Flush() error }); ok {
			if err := flushable.Flush(); err != nil {
				fmt.Printf("Failed to flush handler %s: %v\n", handlerName, err)
			}
		}
	}
	return nil
}

// SetLevel implements LoggingClientInterface
func (c *LoggingClient) SetLevel(level string) error {
	validLevels := map[string]bool{
		"debug":   true,
		"info":    true,
		"warning": true,
		"error":   true,
		"fatal":   true,
	}

	if !validLevels[level] {
		return fmt.Errorf("invalid log level: %s", level)
	}

	c.level = level
	return nil
}

// GetLevel implements LoggingClientInterface
func (c *LoggingClient) GetLevel() string {
	return c.level
}

// AddHandler adds a log handler
func (c *LoggingClient) AddHandler(name string, handler app_core.LogHandler) {
	c.handlers[name] = handler
}

// RemoveHandler removes a log handler
func (c *LoggingClient) RemoveHandler(name string) {
	delete(c.handlers, name)
}

// GetHandler returns a log handler by name
func (c *LoggingClient) GetHandler(name string) (app_core.LogHandler, bool) {
	handler, exists := c.handlers[name]
	return handler, exists
}

// GetHandlers returns all log handlers
func (c *LoggingClient) GetHandlers() map[string]app_core.LogHandler {
	return c.handlers
}

// BaseLogHandler provides common functionality for log handlers
type BaseLogHandler struct {
	level string
}

// NewBaseLogHandler creates a new base log handler
func NewBaseLogHandler(level string) *BaseLogHandler {
	return &BaseLogHandler{
		level: level,
	}
}

// ShouldHandle checks if the handler should handle the given level
func (h *BaseLogHandler) ShouldHandle(level string) bool {
	levels := map[string]int{
		"debug":   0,
		"info":    1,
		"warning": 2,
		"error":   3,
		"fatal":   4,
	}

	handlerLevel := levels[h.level]
	messageLevel := levels[level]

	return messageLevel >= handlerLevel
}

// GetLevel returns the handler's level
func (h *BaseLogHandler) GetLevel() string {
	return h.level
}

// Handle implements LogHandler interface (to be overridden)
func (h *BaseLogHandler) Handle(level string, message string, context map[string]interface{}) error {
	// Base implementation - override in specific handlers
	return nil
}
