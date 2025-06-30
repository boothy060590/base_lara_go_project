package logging_core

import (
	"fmt"
	"time"

	client_core "base_lara_go_project/app/core/laravel_core/clients"
	exceptions_core "base_lara_go_project/app/core/laravel_core/exceptions"
)

// LoggingClient provides logging functionality
type LoggingClient struct {
	*client_core.BaseClient
	level    LogLevel
	handlers map[string]LogHandler
}

// NewLoggingClient creates a new logging client
func NewLoggingClient(config *client_core.ClientConfig) *LoggingClient {
	return &LoggingClient{
		BaseClient: client_core.NewBaseClient(config, "logging"),
		level:      LogLevelInfo,
		handlers:   make(map[string]LogHandler),
	}
}

// convertToException converts a generic error to our framework exception
func convertToException(err error) *exceptions_core.SimpleException {
	return exceptions_core.ConvertToException(err)
}

// Log implements LoggingClientInterface
func (c *LoggingClient) Log(level LogLevel, message string, context map[string]interface{}) error {
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
	level := LogLevelError
	switch ex.GetCode() {
	case 404:
		level = LogLevelWarning
	case 401, 403:
		level = LogLevelInfo
	case 422:
		level = LogLevelWarning
	case 500:
		level = LogLevelError
	default:
		level = LogLevelError
	}

	return c.Log(level, ex.Error(), context)
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
func (c *LoggingClient) SetLevel(level LogLevel) error {
	c.level = level
	return nil
}

// GetLevel implements LoggingClientInterface
func (c *LoggingClient) GetLevel() LogLevel {
	return c.level
}

// AddHandler adds a log handler
func (c *LoggingClient) AddHandler(name string, handler LogHandler) {
	c.handlers[name] = handler
}

// RemoveHandler removes a log handler
func (c *LoggingClient) RemoveHandler(name string) {
	delete(c.handlers, name)
}

// GetHandler returns a log handler by name
func (c *LoggingClient) GetHandler(name string) (LogHandler, bool) {
	handler, exists := c.handlers[name]
	return handler, exists
}

// GetHandlers returns all log handlers
func (c *LoggingClient) GetHandlers() map[string]LogHandler {
	return c.handlers
}
